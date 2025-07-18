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

package wirebson

import (
	"log/slog"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/internal/util/must"
)

// RawDocument represents a single BSON document a.k.a object in the binary encoded form.
//
// It generally references a part of a larger slice, not a copy.
type RawDocument []byte

func (raw RawDocument) document() {}

// Encode returns itself to implement the [AnyDocument] interface.
//
// Receiver must not be nil.
func (raw RawDocument) Encode() (RawDocument, error) {
	if raw == nil {
		panic("raw is nil")
	}

	return raw, nil
}

// Decode decodes a single non-nil BSON document that takes the whole non-nil byte slice.
//
// Only top-level fields are decoded;
// nested documents and arrays are converted to RawDocument and RawArray respectively,
// using raw's subslices without copying.
func (raw RawDocument) Decode() (*Document, error) {
	if raw == nil {
		panic("raw is nil")
	}

	res, err := raw.decode(decodeShallow)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// DecodeDeep decodes a single non-nil BSON document that takes the whole non-nil byte slice.
//
// All nested documents and arrays are decoded recursively.
func (raw RawDocument) DecodeDeep() (*Document, error) {
	if raw == nil {
		panic("raw is nil")
	}

	res, err := raw.decode(decodeDeep)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// decode decodes a single BSON document that takes the whole byte slice.
func (raw RawDocument) decode(mode decodeMode) (*Document, error) {
	l, err := FindRaw(raw)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if rl := len(raw); rl != l {
		return nil, lazyerrors.Errorf("len(raw) = %d, l = %d: %w", rl, l, ErrDecodeInvalidInput)
	}

	res := MakeDocument(0)

	offset := 4

	for {
		if err = decodeCheckOffset(raw, offset, 1); err != nil {
			return nil, lazyerrors.Error(err)
		}

		t := tag(raw[offset])
		if t == 0 {
			if rl := len(raw); rl != offset+1 {
				return nil, lazyerrors.Errorf("len(raw) = %d, offset = %d, got %s: %w", rl, offset, t, ErrDecodeInvalidInput)
			}

			return res, nil
		}

		offset++

		if err = decodeCheckOffset(raw, offset, 1); err != nil {
			return nil, lazyerrors.Error(err)
		}

		var name string
		if name, err = DecodeCString(raw[offset:]); err != nil {
			return nil, lazyerrors.Error(err)
		}

		offset += SizeCString(name)

		var v any

		// to check if we can even `raw[offset:]` below
		if err = decodeCheckOffset(raw, offset, 0); err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch t { //nolint:exhaustive // other tags are handled by decodeScalarField
		case tagDocument:
			if l, err = FindRaw(raw[offset:]); err != nil {
				return nil, lazyerrors.Errorf("no document at offset = %d: %w", offset, err)
			}

			rawDoc := RawDocument(raw[offset : offset+l])
			offset += l

			switch mode {
			case decodeShallow:
				v = rawDoc
			case decodeDeep:
				v, err = rawDoc.decode(decodeDeep)
			}

		case tagArray:
			if l, err = FindRaw(raw[offset:]); err != nil {
				return nil, lazyerrors.Errorf("no array at offset = %d: %w", offset, err)
			}

			rawArr := RawArray(raw[offset : offset+l])
			offset += l

			switch mode {
			case decodeShallow:
				v = rawArr
			case decodeDeep:
				v, err = rawArr.decode(decodeDeep)
			}

		default:
			v, l, err = decodeScalarField(raw[offset:], t)
			offset += l
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		must.NoError(res.Add(name, v))
	}
}

// LogValue implements [slog.LogValuer].
func (raw RawDocument) LogValue() slog.Value {
	return slogValue(raw, 1)
}

// LogMessage implements [AnyDocument].
func (raw RawDocument) LogMessage() string {
	return LogMessage(raw)
}

// LogMessageIndent implements [AnyDocument].
func (raw RawDocument) LogMessageIndent() string {
	return LogMessageIndent(raw)
}

// check interfaces
var (
	_ AnyDocument = RawDocument(nil)
)
