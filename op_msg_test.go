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

package wire

import (
	"math"
	"testing"

	"github.com/FerretDB/wire/internal/util/testutil"
	"github.com/FerretDB/wire/wirebson"
)

// msgTestCases represents test cases for OP_MSG decoding/encoding.
//
//nolint:lll // for readability
var msgTestCases = []testCase{
	{
		name:    "handshake5",
		headerB: testutil.MustParseDumpFile("testdata", "handshake5_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake5_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 92,
			RequestID:     3,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			sections: []opMsgSection{{
				documents: []wirebson.RawDocument{makeRawDocument(
					"buildInfo", int32(1),
					"lsid", wirebson.MustDocument(
						"id", wirebson.Binary{
							Subtype: wirebson.BinaryUUID,
							B: []byte{
								0xa3, 0x19, 0xf2, 0xb4, 0xa1, 0x75, 0x40, 0xc7,
								0xb8, 0xe7, 0xa3, 0xa3, 0x2e, 0xc2, 0x56, 0xbe,
							},
						},
					),
					"$db", "admin",
				)},
			}},
		},
		command: "buildInfo",
		si: `
		{
		  "FlagBits": "[]",
		  "Checksum": int64(0),
		  "Sections": [
		    {
		      "Kind": 0,
		      "Document": {
		        "buildInfo": 1,
		        "lsid": {
		          "id": Binary(uuid:oxnytKF1QMe456OjLsJWvg==),
		        },
		        "$db": "admin",
		      },
		    },
		  ],
		}`,
	},
	{
		name:    "handshake6",
		headerB: testutil.MustParseDumpFile("testdata", "handshake6_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake6_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 1931,
			RequestID:     292,
			ResponseTo:    3,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			sections: []opMsgSection{{
				documents: []wirebson.RawDocument{makeRawDocument(
					"version", "5.0.0",
					"gitVersion", "1184f004a99660de6f5e745573419bda8a28c0e9",
					"modules", wirebson.MustArray(),
					"allocator", "tcmalloc",
					"javascriptEngine", "mozjs",
					"sysInfo", "deprecated",
					"versionArray", wirebson.MustArray(int32(5), int32(0), int32(0), int32(0)),
					"openssl", wirebson.MustDocument(
						"running", "OpenSSL 1.1.1f  31 Mar 2020",
						"compiled", "OpenSSL 1.1.1f  31 Mar 2020",
					),
					"buildEnvironment", wirebson.MustDocument(
						"distmod", "ubuntu2004",
						"distarch", "x86_64",
						"cc", "/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0",
						"ccflags", "-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables "+
							"-ggdb -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer "+
							"-fno-strict-aliasing -O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 "+
							"-Wno-unused-local-typedefs -Wno-unused-function -Wno-deprecated-declarations "+
							"-Wno-unused-const-variable -Wno-unused-but-set-variable -Wno-missing-braces "+
							"-fstack-protector-strong -Wa,--nocompress-debug-sections -fno-builtin-memcmp",
						"cxx", "/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0",
						"cxxflags", "-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17",
						"linkflags", "-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong "+
							"-Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack "+
							"-Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags",
						"target_arch", "x86_64",
						"target_os", "linux",
						"cppdefines", "SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 "+
							"_GNU_SOURCE _REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 "+
							"BOOST_THREAD_USES_DATETIME BOOST_SYSTEM_NO_DEPRECATED "+
							"BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS BOOST_ENABLE_ASSERT_DEBUG_HANDLER "+
							"BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG "+
							"BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS",
					),
					"bits", int32(64),
					"debug", false,
					"maxBsonObjectSize", int32(16777216),
					"storageEngines", wirebson.MustArray("devnull", "ephemeralForTest", "wiredTiger"),
					"ok", float64(1),
				)},
			}},
		},
		command: "version",
		si: `
		{
		  "FlagBits": "[]",
		  "Checksum": int64(0),
		  "Sections": [
		    {
		      "Kind": 0,
		      "Document": {
		        "version": "5.0.0",
		        "gitVersion": "1184f004a99660de6f5e745573419bda8a28c0e9",
		        "modules": [],
		        "allocator": "tcmalloc",
		        "javascriptEngine": "mozjs",
		        "sysInfo": "deprecated",
		        "versionArray": [
		          5,
		          0,
		          0,
		          0,
		        ],
		        "openssl": {
		          "running": "OpenSSL 1.1.1f  31 Mar 2020",
		          "compiled": "OpenSSL 1.1.1f  31 Mar 2020",
		        },
		        "buildEnvironment": {
		          "distmod": "ubuntu2004",
		          "distarch": "x86_64",
		          "cc": "/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0",
		          "ccflags": "-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables -ggdb -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer -fno-strict-aliasing -O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 -Wno-unused-local-typedefs -Wno-unused-function -Wno-deprecated-declarations -Wno-unused-const-variable -Wno-unused-but-set-variable -Wno-missing-braces -fstack-protector-strong -Wa,--nocompress-debug-sections -fno-builtin-memcmp",
		          "cxx": "/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0",
		          "cxxflags": "-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17",
		          "linkflags": "-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong -Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack -Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags",
		          "target_arch": "x86_64",
		          "target_os": "linux",
		          "cppdefines": "SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 _GNU_SOURCE _REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 BOOST_THREAD_USES_DATETIME BOOST_SYSTEM_NO_DEPRECATED BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS BOOST_ENABLE_ASSERT_DEBUG_HANDLER BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS",
		        },
		        "bits": 64,
		        "debug": false,
		        "maxBsonObjectSize": 16777216,
		        "storageEngines": [
		          "devnull",
		          "ephemeralForTest",
		          "wiredTiger",
		        ],
		        "ok": 1.0,
		      },
		    },
		  ],
		}`,
	},
	{
		name:      "import",
		expectedB: testutil.MustParseDumpFile("testdata", "import.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 327,
			RequestID:     7,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			sections: []opMsgSection{
				{
					documents: []wirebson.RawDocument{makeRawDocument(
						"insert", "actor",
						"ordered", true,
						"writeConcern", wirebson.MustDocument(
							"w", "majority",
						),
						"$db", "monila",
					)},
				},
				{
					kind:       1,
					identifier: "documents",
					documents: []wirebson.RawDocument{
						makeRawDocument(
							"_id", wirebson.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01},
							"actor_id", int32(1),
							"first_name", "PENELOPE",
							"last_name", "GUINESS",
							"last_update", lastUpdate,
						),
						makeRawDocument(
							"_id", wirebson.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
							"actor_id", int32(2),
							"first_name", "NICK",
							"last_name", "WAHLBERG",
							"last_update", lastUpdate,
						),
					},
				},
			},
		},
		command: "insert",
		si: `
		{
		  "FlagBits": "[]",
		  "Checksum": int64(0),
		  "Sections": [
		    {
		      "Kind": 0,
		      "Document": {
		        "insert": "actor",
		        "ordered": true,
		        "writeConcern": {
		          "w": "majority",
		        },
		        "$db": "monila",
		      },
		    },
		    {
		      "Kind": 1,
		      "Identifier": "documents",
		      "Documents": [
		        {
		          "_id": ObjectID(612ec2800000000100000001),
		          "actor_id": 1,
		          "first_name": "PENELOPE",
		          "last_name": "GUINESS",
		          "last_update": 2020-02-15T09:34:33Z,
		        },
		        {
		          "_id": ObjectID(612ec2800000000200000002),
		          "actor_id": 2,
		          "first_name": "NICK",
		          "last_name": "WAHLBERG",
		          "last_update": 2020-02-15T09:34:33Z,
		        },
		      ],
		    },
		  ],
		}`,
	},
	{
		name:      "msg_fuzz1",
		expectedB: testutil.MustParseDumpFile("testdata", "msg_fuzz1.hex"),
		err:       wirebson.ErrDecodeInvalidInput.Error(),
	},
	{
		name: "NaN",
		expectedB: []byte{
			0x79, 0x00, 0x00, 0x00, // MessageLength
			0x11, 0x00, 0x00, 0x00, // RequestID
			0x00, 0x00, 0x00, 0x00, // ResponseTo
			0xdd, 0x07, 0x00, 0x00, // OpCode
			0x00, 0x00, 0x00, 0x00, // FlagBits
			0x00,                   // section kind
			0x64, 0x00, 0x00, 0x00, // document size
			0x02, 0x69, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x00, // string "insert"
			0x07, 0x00, 0x00, 0x00, // "values" length
			0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x00, // "values"
			0x04, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x00, // array "documents"
			0x29, 0x00, 0x00, 0x00, // document (array) size
			0x03, 0x30, 0x00, // element 0 (document)
			0x21, 0x00, 0x00, 0x00, // element 0 size
			0x01, 0x76, 0x00, // double "v"
			0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f, // NaN
			0x07, 0x5f, 0x69, 0x64, 0x00, // ObjectID "_id"
			0x63, 0x77, 0xf2, 0x13, 0x75, 0x7c, 0x0b, 0xab, 0xde, 0xbc, 0x2f, 0x6a, // ObjectID value
			0x00,                                                       // end of element 0 (document)
			0x00,                                                       // end of document (array)
			0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x00, 0x01, // "ordered" true
			0x02, 0x24, 0x64, 0x62, 0x00, // "$db"
			0x05, 0x00, 0x00, 0x00, // "test" length
			0x74, 0x65, 0x73, 0x74, 0x00, // "test"
			0x00, // end of document
		},
		msgHeader: &MsgHeader{
			MessageLength: 121,
			RequestID:     17,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			sections: []opMsgSection{{
				documents: []wirebson.RawDocument{makeRawDocument(
					"insert", "values",
					"documents", wirebson.MustArray(
						wirebson.MustDocument(
							"v", math.NaN(),
							"_id", wirebson.ObjectID{0x63, 0x77, 0xf2, 0x13, 0x75, 0x7c, 0x0b, 0xab, 0xde, 0xbc, 0x2f, 0x6a},
						),
					),
					"ordered", true,
					"$db", "test",
				)},
			}},
		},
		command: "insert",
		si: `
		{
		  "FlagBits": "[]",
		  "Checksum": int64(0),
		  "Sections": [
		    {
		      "Kind": 0,
		      "Document": {
		        "insert": "values",
		        "documents": [
		          {
		            "v": NaN,
		            "_id": ObjectID(6377f213757c0babdebc2f6a),
		          },
		        ],
		        "ordered": true,
		        "$db": "test",
		      },
		    },
		  ],
		}`,
	},
	{
		name: "negative zero",
		expectedB: []byte{
			0x8b, 0x00, 0x00, 0x00,
			0x0c, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xdd, 0x07, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00,
			0x46, 0x00, 0x00, 0x00,
			0x02, 0x69, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x00,
			0x11, 0x00, 0x00, 0x00,
			0x54, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x00,
			0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65,
			0x64, 0x00, 0x01, 0x02, 0x24, 0x64, 0x62, 0x00, 0x11, 0x00,
			0x00, 0x00, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x73, 0x65,
			0x72, 0x74, 0x73, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x00, 0x00,
			0x01,
			0x2f, 0x00, 0x00, 0x00,
			0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x00,
			0x21, 0x00, 0x00, 0x00,
			0x07, 0x5f, 0x69, 0x64, 0x00,
			0x63, 0x7c, 0xfa, 0xd8, 0x8d, 0xc3, 0xce, 0xcd, 0xe3, 0x8e, 0x1e, 0x6b,
			0x01, 0x76, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80,
			0x00,
		},
		msgHeader: &MsgHeader{
			MessageLength: 139,
			RequestID:     12,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			sections: []opMsgSection{
				{
					documents: []wirebson.RawDocument{makeRawDocument(
						"insert", "TestInsertSimple",
						"ordered", true,
						"$db", "testinsertsimple",
					)},
				},
				{
					kind:       1,
					identifier: "documents",
					documents: []wirebson.RawDocument{makeRawDocument(
						"_id", wirebson.ObjectID{0x63, 0x7c, 0xfa, 0xd8, 0x8d, 0xc3, 0xce, 0xcd, 0xe3, 0x8e, 0x1e, 0x6b},
						"v", math.Copysign(0, -1),
					)},
				},
			},
		},
		command: "insert",
		si: `
		{
		  "FlagBits": "[]",
		  "Checksum": int64(0),
		  "Sections": [
		    {
		      "Kind": 0,
		      "Document": {
		        "insert": "TestInsertSimple",
		        "ordered": true,
		        "$db": "testinsertsimple",
		      },
		    },
		    {
		      "Kind": 1,
		      "Identifier": "documents",
		      "Documents": [
		        {
		          "_id": ObjectID(637cfad88dc3cecde38e1e6b),
		          "v": -0.0,
		        },
		      ],
		    },
		  ],
		}`,
	},
	{
		name: "MultiSectionInsert",
		expectedB: []byte{
			0x76, 0x00, 0x00, 0x00, // MessageLength
			0x0f, 0x00, 0x00, 0x00, // RequestID
			0x00, 0x00, 0x00, 0x00, // ResponseTo
			0xdd, 0x07, 0x00, 0x00, // OpCode
			0x01, 0x00, 0x00, 0x00, // FlagBits

			0x01,                   // section kind
			0x2f, 0x00, 0x00, 0x00, // section size
			0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x00, // section identifier "documents"
			0x21, 0x00, 0x00, 0x00, // document size
			0x07, 0x5f, 0x69, 0x64, 0x00, // ObjectID "_id"
			0x63, 0x8c, 0xec, 0x46, 0xaa, 0x77, 0x8b, 0xf3, 0x70, 0x10, 0x54, 0x29, // ObjectID value
			0x01, 0x61, 0x00, // double "a"
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x40, // 3.0
			0x00, // end of document

			0x00,                   // section kind
			0x2d, 0x00, 0x00, 0x00, // document size
			0x02, 0x69, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x00, // string "insert"
			0x04, 0x00, 0x00, 0x00, // "foo" length
			0x66, 0x6f, 0x6f, 0x00, // "foo"
			0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x00, 0x01, // "ordered" true
			0x02, 0x24, 0x64, 0x62, 0x00, // string "$db"
			0x05, 0x00, 0x00, 0x00, // "test" length
			0x74, 0x65, 0x73, 0x74, 0x00, // "test"
			0x00, // end of document

			0xe2, 0xb7, 0x90, 0x67, // checksum
		},
		msgHeader: &MsgHeader{
			MessageLength: 118,
			RequestID:     15,
			ResponseTo:    0,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			Flags: OpMsgFlags(OpMsgChecksumPresent),
			sections: []opMsgSection{
				{
					kind:       1,
					identifier: "documents",
					documents: []wirebson.RawDocument{makeRawDocument(
						"_id", wirebson.ObjectID{0x63, 0x8c, 0xec, 0x46, 0xaa, 0x77, 0x8b, 0xf3, 0x70, 0x10, 0x54, 0x29},
						"a", float64(3),
					)},
				},
				{
					documents: []wirebson.RawDocument{makeRawDocument(
						"insert", "foo",
						"ordered", true,
						"$db", "test",
					)},
				},
			},
			checksum: 1737537506,
		},
		command: "insert",
		si: `
		{
		  "FlagBits": "[checksumPresent]",
		  "Checksum": int64(1737537506),
		  "Sections": [
		    {
		      "Kind": 1,
		      "Identifier": "documents",
		      "Documents": [
		        {
		          "_id": ObjectID(638cec46aa778bf370105429),
		          "a": 3.0,
		        },
		      ],
		    },
		    {
		      "Kind": 0,
		      "Document": {
		        "insert": "foo",
		        "ordered": true,
		        "$db": "test",
		      },
		    },
		  ],
		}`,
	},
	{
		name: "MultiSectionUpdate",
		expectedB: []byte{
			0x9a, 0x00, 0x00, 0x00, // MessageLength
			0x0b, 0x00, 0x00, 0x00, // RequestID
			0x00, 0x00, 0x00, 0x00, // ResponseTo
			0xdd, 0x07, 0x00, 0x00, // OpCode
			0x01, 0x00, 0x00, 0x00, // FlagBits

			0x01,                   // section kind
			0x53, 0x00, 0x00, 0x00, // section size
			0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x00, // section identifier "updates"
			0x47, 0x00, 0x00, 0x00, // document size

			0x03, 0x71, 0x00, // document "q"
			0x10, 0x00, 0x00, 0x00, // document size
			0x01, 0x61, 0x00, // double "a"
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x34, 0x40, // 20.0
			0x00, // end of document

			0x03, 0x75, 0x00, // document "u"
			0x1b, 0x00, 0x00, 0x00, // document size
			0x03, 0x24, 0x69, 0x6e, 0x63, 0x00, // document "$inc"
			0x10, 0x00, 0x00, 0x00, // document size
			0x01, 0x61, 0x00, // double "a"
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f, // 1.0
			0x00, // end of document
			0x00, // end of document

			0x08, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x00, 0x00, // "multi" false
			0x08, 0x75, 0x70, 0x73, 0x65, 0x72, 0x74, 0x00, 0x00, // "upsert" false

			0x00, // end of document

			0x00,                   // section kind
			0x2d, 0x00, 0x00, 0x00, // document size
			0x02, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x00, // string "update"
			0x04, 0x00, 0x00, 0x00, // "foo" length
			0x66, 0x6f, 0x6f, 0x00, // "foo"
			0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x00, 0x01, // "ordered" true
			0x02, 0x24, 0x64, 0x62, 0x00, // string "$db"
			0x05, 0x00, 0x00, 0x00, // "test" length
			0x74, 0x65, 0x73, 0x74, 0x00, // "test"
			0x00, // end of document

			0xf1, 0xfc, 0xd1, 0xae, // checksum
		},
		msgHeader: &MsgHeader{
			MessageLength: 154,
			RequestID:     11,
			ResponseTo:    0,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			Flags: OpMsgFlags(OpMsgChecksumPresent),
			sections: []opMsgSection{
				{
					kind:       1,
					identifier: "updates",
					documents: []wirebson.RawDocument{makeRawDocument(
						"q", wirebson.MustDocument(
							"a", float64(20),
						),
						"u", wirebson.MustDocument(
							"$inc", wirebson.MustDocument(
								"a", float64(1),
							),
						),
						"multi", false,
						"upsert", false,
					)},
				},
				{
					documents: []wirebson.RawDocument{makeRawDocument(
						"update", "foo",
						"ordered", true,
						"$db", "test",
					)},
				},
			},
			checksum: 2932997361,
		},
		command: "update",
		si: `
		{
		  "FlagBits": "[checksumPresent]",
		  "Checksum": int64(2932997361),
		  "Sections": [
		    {
		      "Kind": 1,
		      "Identifier": "updates",
		      "Documents": [
		        {
		          "q": {
		            "a": 20.0,
		          },
		          "u": {
		            "$inc": {
		              "a": 1.0,
		            },
		          },
		          "multi": false,
		          "upsert": false,
		        },
		      ],
		    },
		    {
		      "Kind": 0,
		      "Document": {
		        "update": "foo",
		        "ordered": true,
		        "$db": "test",
		      },
		    },
		  ],
		}`,
	},
	{
		name: "InvalidChecksum",
		expectedB: []byte{
			0x77, 0x00, 0x00, 0x00, // MessageLength
			0x0f, 0x00, 0x00, 0x00, // RequestID
			0x00, 0x00, 0x00, 0x00, // ResponseTo
			0xdd, 0x07, 0x00, 0x00, // OpCode
			0x01, 0x00, 0x00, 0x00, // FlagBits

			0x01,                   // section kind
			0x2f, 0x00, 0x00, 0x00, // section size
			0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x00, // section identifier "documents"
			0x21, 0x00, 0x00, 0x00, // document size
			0x07, 0x5f, 0x69, 0x64, 0x00, // ObjectID "_id"
			0x63, 0x8c, 0xec, 0x46, 0xaa, 0x77, 0x8b, 0xf3, 0x70, 0x10, 0x54, 0x29, // ObjectID value
			0x01, 0x61, 0x00, // double "a"
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x40, // 3.0
			0x00, // end of document

			0x00,                   // section kind
			0x2d, 0x00, 0x00, 0x00, // document size
			0x02, 0x69, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x00, // string "insert"
			0x04, 0x00, 0x00, 0x00, // "foo" length
			0x66, 0x6f, 0x6f, 0x6f, 0x00, // "fooo"
			0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x00, 0x01, // "ordered" true
			0x02, 0x24, 0x64, 0x62, 0x00, // string "$db"
			0x05, 0x00, 0x00, 0x00, // "test" length
			0x74, 0x65, 0x73, 0x74, 0x00, // "test"
			0x00, // end of document

			0xe2, 0xb7, 0x90, 0x67, // invalid checksum value
		},
		msgHeader: &MsgHeader{
			MessageLength: 119,
			RequestID:     15,
			ResponseTo:    0,
			OpCode:        OpCodeMsg,
		},
		msgBody: &OpMsg{
			Flags: OpMsgFlags(OpMsgChecksumPresent),
			sections: []opMsgSection{
				{
					kind:       1,
					identifier: "documents",
					documents: []wirebson.RawDocument{makeRawDocument(
						"_id", wirebson.ObjectID{0x63, 0x8c, 0xec, 0x46, 0xaa, 0x77, 0x8b, 0xf3, 0x70, 0x10, 0x54, 0x29},
						"a", float64(3),
					)},
				},
				{
					documents: []wirebson.RawDocument{makeRawDocument(
						"insert", "fooo",
						"ordered", true,
						"$db", "test",
					)},
				},
			},
			checksum: 1737537506,
		},
		err: "OP_MSG checksum does not match contents.",
	},
}

func TestMsg(t *testing.T) {
	t.Parallel()
	testMessages(t, msgTestCases)
}

func FuzzMsg(f *testing.F) {
	fuzzMessages(f, msgTestCases)
}
