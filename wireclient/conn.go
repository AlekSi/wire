// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wireclient

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/xdg-go/scram"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
)

// nextRequestID stores the last generated request ID.
var nextRequestID atomic.Int32

// Conn represents a single client connection.
//
// It is not safe for concurrent use.
type Conn struct {
	c net.Conn
	r *bufio.Reader
	w *bufio.Writer
	l *slog.Logger // debug-level only
}

// New wraps the given connection.
//
// The passed logger will be used only for debug-level messages.
func New(c net.Conn, l *slog.Logger) *Conn {
	return &Conn{
		c: c,
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
		l: l,
	}
}

// Connect creates a new connection for the given MongoDB URI. See [Credentials].
//
// Context can be used to cancel the connection attempt.
// Canceling the context after the connection is established has no effect.
//
// The passed logger will be used only for debug-level messages.
func Connect(ctx context.Context, uri string, l *slog.Logger) (*Conn, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("wireclient.Connect: %w", err)
	}

	if u.Scheme == "mongodb+srv" {
		if err = lookupSrvURI(ctx, u); err != nil {
			return nil, fmt.Errorf("wireclient.Connect: %w", err)
		}
	}

	if u.Scheme != "mongodb" {
		return nil, fmt.Errorf("wireclient.Connect: invalid scheme %q", u.Scheme)
	}

	if u.Opaque != "" {
		return nil, fmt.Errorf("wireclient.Connect: invalid URI %q", uri)
	}

	if u.User != nil {
		return nil, fmt.Errorf("wireclient.Connect: credentials must be absent")
	}

	if u.Path != "/" {
		return nil, fmt.Errorf("wireclient.Connect: unsupported path %q", u.Path)
	}

	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		return nil, fmt.Errorf("wireclient.Connect: %w", err)
	}

	var tlsParam bool
	var tlsCaFileParam string

	for k, vs := range u.Query() {
		switch k {
		case "replicaSet":
			// safe to ignore

		case "tls":
			if len(vs) != 1 {
				return nil, fmt.Errorf("wireclient.Connect: query parameter %q must have exactly one value", k)
			}

			if tlsParam, err = strconv.ParseBool(vs[0]); err != nil {
				return nil, fmt.Errorf("wireclient.Connect: query parameter %q has invalid value %q", k, vs[0])
			}

		case "tlsCaFile":
			if len(vs) != 1 {
				return nil, fmt.Errorf("wireclient.Connect: query parameter %q must have exactly one value", k)
			}

			tlsCaFileParam = vs[0]
			if _, err = os.Stat(tlsCaFileParam); err != nil {
				return nil, fmt.Errorf("wireclient.Connect: query parameter %q error %w", k, err)
			}

		default:
			return nil, fmt.Errorf("wireclient.Connect: query parameter %q is not supported", k)
		}
	}

	l.DebugContext(ctx, "Connecting", slog.String("uri", uri))

	dial := (&net.Dialer{}).DialContext

	if tlsParam {
		var config tls.Config

		if tlsCaFileParam != "" {
			var b []byte
			if b, err = os.ReadFile(tlsCaFileParam); err != nil {
				return nil, fmt.Errorf("wireclient.Connect: %w", err)
			}

			ca := x509.NewCertPool()
			if ok := ca.AppendCertsFromPEM(b); !ok {
				return nil, fmt.Errorf("wireclient.Connect: failed to parse tlsCaFile")
			}

			config.RootCAs = ca
		}

		dial = (&tls.Dialer{Config: &config}).DialContext
	}

	c, err := dial(ctx, "tcp", u.Host)
	if err != nil {
		return nil, fmt.Errorf("wireclient.Connect: %w", err)
	}

	return New(c, l), nil
}

// ConnectPing uses a combination of [Connect] and [Conn.Ping] to establish a working connection.
//
// nil is returned on context expiration.
func ConnectPing(ctx context.Context, uri string, l *slog.Logger) *Conn {
	var err error

	for ctx.Err() == nil {
		var conn *Conn

		conn, err = Connect(ctx, uri, l)
		if err != nil {
			sleep(ctx, time.Second)
			continue
		}

		if err = conn.Ping(ctx); err != nil {
			_ = conn.Close()
			sleep(ctx, time.Second)
			continue
		}

		return conn
	}

	if err != nil {
		l.DebugContext(ctx, "Connection unsuccessful", slog.String("error", err.Error()))
	}

	return nil
}

// Close closes the connection.
func (c *Conn) Close() error {
	c.l.Debug("Closing")

	if err := c.c.Close(); err != nil {
		return fmt.Errorf("wireclient.Conn.Close: %w", err)
	}

	return nil
}

// Read reads the next message from the connection.
//
// Passed context's deadline is honored if set.
func (c *Conn) Read(ctx context.Context) (*wire.MsgHeader, wire.MsgBody, error) {
	d, _ := ctx.Deadline()
	c.c.SetReadDeadline(d)

	header, body, err := wire.ReadMessage(c.r)
	if err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Read: %w", err)
	}

	c.l.DebugContext(
		ctx,
		fmt.Sprintf("<<<\n%s", body.StringIndent()),
		slog.Int("length", int(header.MessageLength)),
		slog.Int("id", int(header.RequestID)),
		slog.Int("response_to", int(header.ResponseTo)),
		slog.String("opcode", header.OpCode.String()),
	)

	return header, body, nil
}

// Write writes the given message to the connection.
//
// Passed context's deadline is honored if set.
func (c *Conn) Write(ctx context.Context, header *wire.MsgHeader, body wire.MsgBody) error {
	c.l.DebugContext(
		ctx,
		fmt.Sprintf(">>>\n%s", body.StringIndent()),
		slog.Int("length", int(header.MessageLength)),
		slog.Int("id", int(header.RequestID)),
		slog.Int("response_to", int(header.ResponseTo)),
		slog.String("opcode", header.OpCode.String()),
	)

	if d, ok := ctx.Deadline(); ok {
		c.c.SetWriteDeadline(d)
	}

	if err := wire.WriteMessage(c.w, header, body); err != nil {
		return fmt.Errorf("wireclient.Conn.Write: %w", err)
	}

	if err := c.w.Flush(); err != nil {
		return fmt.Errorf("wireclient.Conn.Write: %w", err)
	}

	return nil
}

// WriteRaw writes the given raw bytes to the connection.
//
// Passed context's deadline is honored if set.
func (c *Conn) WriteRaw(ctx context.Context, b []byte) error {
	c.l.DebugContext(ctx, ">>> raw bytes", slog.Int("length", len(b)))

	d, _ := ctx.Deadline()
	c.c.SetWriteDeadline(d)

	if _, err := c.w.Write(b); err != nil {
		return fmt.Errorf("wireclient.Conn.WriteRaw: %w", err)
	}

	if err := c.w.Flush(); err != nil {
		return fmt.Errorf("wireclient.Conn.WriteRaw: %w", err)
	}

	return nil
}

// Request sends the given request to the connection and returns the response.
// The header is generated automatically.
//
// Passed context's deadline is honored if set.
//
// It returns errors only for request/response parsing or connection issues.
// All protocol-level errors are stored inside response.
func (c *Conn) Request(ctx context.Context, body wire.MsgBody) (*wire.MsgHeader, wire.MsgBody, error) {
	b, err := body.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Request: %w", err)
	}

	header := &wire.MsgHeader{
		MessageLength: int32(len(b) + wire.MsgHeaderLen),
		RequestID:     nextRequestID.Add(1),
	}

	switch body.(type) {
	case *wire.OpMsg:
		header.OpCode = wire.OpCodeMsg
	case *wire.OpQuery:
		header.OpCode = wire.OpCodeQuery
	default:
		return nil, nil, fmt.Errorf("wireclient.Conn.Request:unsupported body type %T", body)
	}

	if err = c.Write(ctx, header, body); err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Request: %w", err)
	}

	resHeader, resBody, err := c.Read(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Request: %w", err)
	}

	if resHeader.ResponseTo != header.RequestID {
		err = fmt.Errorf(
			"wireclient.Conn.Request: response's response_to=%d is not equal to request's request_id=%d",
			resHeader.ResponseTo,
			header.RequestID,
		)
	}

	return resHeader, resBody, err
}

// Ping sends a ping command.
// It returns error for unexpected server response or any other error.
func (c *Conn) Ping(ctx context.Context) error {
	cmd := wire.MustOpMsg("ping", int32(1), "$db", "test")

	_, resBody, err := c.Request(ctx, cmd)
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Ping: %w", err)
	}

	res, err := resBody.(*wire.OpMsg).DocumentDeep()
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Ping: %w", err)
	}

	if ok := res.Get("ok"); ok != 1.0 {
		return fmt.Errorf("wireclient.Conn.Ping: failed (ok was %v)", ok)
	}

	return nil
}

// Login authenticates the connection with the given credentials
// using some unspecified sequences of commands.
//
// It should not be used to test various authentication scenarios.
//
// The authMechanism is used to enforce a specific authentication mechanism.
// If empty, the negotiation is performed instead using `hello` command.
// Note that in practice it often fails due to incorrect handling of the `loadBalanced` parameter.
func (c *Conn) Login(ctx context.Context, userinfo *url.Userinfo, authSource, authMechanism string) error {
	username := userinfo.Username()
	password, _ := userinfo.Password()

	mechs := []string{authMechanism}
	if authMechanism == "" {
		var err error
		if mechs, err = c.hello(ctx, username, authSource); err != nil {
			return fmt.Errorf("wireclient.Conn.Login: %w", err)
		}
	}

	switch {
	case slices.Contains(mechs, "SCRAM-SHA-256"):
		return c.loginScramSHA256(ctx, username, password, authSource)
	case slices.Contains(mechs, "PLAIN"):
		return c.loginPlain(ctx, username, password, authSource)
	default:
		return fmt.Errorf("wireclient.Conn.Login: unsupported authentication mechanisms: %v", mechs)
	}
}

// hello sends a hello command to the server
// and returns the list of supported authentication mechanisms.
//
// Note that in practice it often fails due to incorrect handling of the `loadBalanced` parameter.
func (c *Conn) hello(ctx context.Context, username, authDB string) ([]string, error) {
	body := wire.MustOpMsg(
		"hello", int32(1),
		"saslSupportedMechs", authDB+"."+username,
		"$db", authDB,
	)

	_, resBody, err := c.Request(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("wireclient.Conn.hello: %w", err)
	}

	res, err := resBody.(*wire.OpMsg).DocumentDeep()
	if err != nil {
		return nil, fmt.Errorf("wireclient.Conn.hello: %w", err)
	}

	saslSupportedMech, ok := res.Get("saslSupportedMechs").(*wirebson.Array)
	if !ok {
		return nil, fmt.Errorf("wireclient.Conn.hello: invalid saslSupportedMechs in hello response")
	}

	var mechs []string

	for _, mech := range saslSupportedMech.All() {
		var m string
		if m, ok = mech.(string); !ok {
			return nil, fmt.Errorf("wireclient.Conn.hello: invalid saslSupportedMechs value %v", mech)
		}

		mechs = append(mechs, m)
	}

	return mechs, nil
}

// loginPlain authenticates the connection using the PLAIN mechanism.
func (c *Conn) loginPlain(ctx context.Context, username, password, authDB string) error {
	plainCredentials := "\000" + username + "\000" + password

	body := wire.MustOpMsg(
		"saslStart", int32(1),
		"mechanism", "PLAIN",
		"payload", wirebson.Binary{B: []byte(plainCredentials)},
		"$db", authDB,
	)

	_, resBody, err := c.Request(ctx, body)
	if err != nil {
		return fmt.Errorf("wireclient.Conn.loginPlain: %w", err)
	}

	res, err := resBody.(*wire.OpMsg).DocumentDeep()
	if err != nil {
		return fmt.Errorf("wireclient.Conn.loginPlain: %w", err)
	}

	if ok := res.Get("ok"); ok != 1.0 {
		return fmt.Errorf("wireclient.Conn.loginPlain: authentication failed (ok was %v)", ok)
	}

	return c.checkAuth(ctx)
}

// loginScramSHA256 authenticates the connection using the SCRAM-SHA-256 mechanism.
func (c *Conn) loginScramSHA256(ctx context.Context, username, password, authDB string) error {
	s, err := scram.SHA256.NewClient(username, password, "")
	if err != nil {
		return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
	}

	conv := s.NewConversation()

	payload, err := conv.Step("")
	if err != nil {
		return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
	}

	cmd := wirebson.MustDocument(
		"saslStart", int32(1),
		"mechanism", "SCRAM-SHA-256",
		"payload", wirebson.Binary{B: []byte(payload)},
		"options", wirebson.MustDocument(
			"skipEmptyExchange", true,
		),
		"$db", authDB,
	)

	// One `saslStart`, two `saslContinue` for servers that don't support `skipEmptyExchange`;
	// one and one for those who do.
	for step := 1; step <= 3; step++ {
		c.l.DebugContext(
			ctx, "wireclient.Conn.loginScramSHA256: client",
			slog.Int("step", step), slog.String("payload", payload),
			slog.Bool("done", conv.Done()), slog.Bool("valid", conv.Valid()),
		)

		var body *wire.OpMsg
		if body, err = wire.NewOpMsg(cmd); err != nil {
			return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
		}

		var resBody wire.MsgBody
		if _, resBody, err = c.Request(ctx, body); err != nil {
			return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
		}

		var res *wirebson.Document
		if res, err = resBody.(*wire.OpMsg).DocumentDeep(); err != nil {
			return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
		}

		if ok := res.Get("ok"); ok != 1.0 {
			return fmt.Errorf("wireclient.Conn.loginScramSHA256: %s failed (ok was %v)", cmd.Command(), ok)
		}

		payload = string(res.Get("payload").(wirebson.Binary).B)

		c.l.DebugContext(
			ctx, "wireclient.Conn.loginScramSHA256: server",
			slog.Int("step", step), slog.String("payload", payload),
		)

		if done := res.Get("done").(bool); !done {
			payload, err = conv.Step(payload)
			if err != nil {
				return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
			}

			cmd = wirebson.MustDocument(
				"saslContinue", int32(1),
				"conversationId", int32(1),
				"payload", wirebson.Binary{B: []byte(payload)},
				"$db", authDB,
			)

			continue
		}

		if step == 2 {
			c.l.DebugContext(
				ctx, "wireclient.Conn.loginScramSHA256: conversation done at the first saslContinue, "+
					"assuming that server supports skipEmptyExchange",
				slog.Int("step", step), slog.String("payload", payload),
				slog.Bool("done", conv.Done()), slog.Bool("valid", conv.Valid()),
			)

			if _, err = conv.Step(payload); err != nil {
				return fmt.Errorf("wireclient.Conn.loginScramSHA256: %w", err)
			}
		}

		if !conv.Done() {
			return fmt.Errorf("wireclient.Conn.loginScramSHA256: conversation is not done")
		}

		if !conv.Valid() {
			return fmt.Errorf("wireclient.Conn.loginScramSHA256: conversation is done, but not valid")
		}

		return c.checkAuth(ctx)
	}

	return fmt.Errorf("wireclient.Conn.loginScramSHA256: too many steps")
}

// checkAuth checks if the connection is authenticated.
func (c *Conn) checkAuth(ctx context.Context) error {
	_, resBody, err := c.Request(ctx, wire.MustOpMsg("listDatabases", int32(1), "$db", "admin"))
	if err != nil {
		return fmt.Errorf("wireclient.Conn.checkAuth: %w", err)
	}

	res, err := resBody.(*wire.OpMsg).DocumentDeep()
	if err != nil {
		return fmt.Errorf("wireclient.Conn.checkAuth: %w", err)
	}

	if ok := res.Get("ok"); ok != 1.0 {
		return fmt.Errorf("wireclient.Conn.checkAuth: failed (ok was %v)", ok)
	}

	return nil
}
