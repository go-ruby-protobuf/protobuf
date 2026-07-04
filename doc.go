// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package protobuf is a pure-Go (CGO-free) reimplementation of the runtime and
// builder surface of Ruby's google-protobuf gem — the object model a Ruby
// program sees as the Google::Protobuf namespace — without any Ruby runtime and
// without a C extension (upstream google-protobuf ships as a C extension around
// libprotobuf/upb).
//
// It does not reimplement the protobuf wire format: it is a Ruby-faithful API
// layer built on top of google.golang.org/protobuf, the official pure-Go
// protobuf runtime. Descriptors are compiled with protodesc, messages are
// google.golang.org/protobuf/types/dynamicpb dynamic messages, and binary/JSON
// encoding delegate to proto.Marshal / protojson. Every byte on the wire is
// therefore produced by the canonical Go runtime, so encode/decode is
// wire-compatible with real protobuf by construction.
//
// # Mapping to the gem
//
//	Ruby (google-protobuf)                Go (this package)
//	----------------------                -----------------
//	Google::Protobuf::DescriptorPool      *DescriptorPool
//	  .generated_pool                       GeneratedPool()
//	  #build { … }                          (*DescriptorPool).Build
//	  #lookup(name)                         (*DescriptorPool).Lookup
//	Google::Protobuf::Descriptor          *Descriptor  (.Msgclass, .Lookup, .Each)
//	Google::Protobuf::EnumDescriptor      *EnumDescriptor
//	Google::Protobuf::FieldDescriptor     *FieldDescriptor
//	a generated message class             *MessageClass  (.New)
//	a message instance                    *Message       (.Get/.Set/.ToH/.Equal/.Dup/.Inspect)
//	Google::Protobuf::RepeatedField       *RepeatedField
//	Google::Protobuf::Map                 *Map
//	Google::Protobuf.encode / .decode     Encode / Decode
//	Google::Protobuf.encode_json/.decode_json  EncodeJSON / DecodeJSON
//	Google::Protobuf::TypeError           *TypeError
//	Google::Protobuf::ParseError          *ParseError
//
// # Ruby value model
//
// Message field values cross the boundary as a small, fixed set of Go types, so
// a host (such as go-embedded-ruby) can map its own object graph to and from
// this package:
//
//	protobuf type                 Go value
//	-------------                 --------
//	bool                          bool
//	int32/int64/sint*/sfixed*     int64
//	uint32/uint64/fixed*          uint64
//	float                         float64
//	double                        float64
//	string                        string
//	bytes                         []byte
//	enum                          Symbol (known value) or int64 (unknown number)
//	message                       *Message (nil when unset)
//	repeated                      *RepeatedField
//	map                           *Map
//
// # Builder DSL
//
// (*DescriptorPool).Build mirrors the gem's pool.build block:
//
//	pool.Build(func(b *protobuf.Builder) {
//		b.AddMessage("Person", func(m *protobuf.MessageBuilder) {
//			m.Optional("name", protobuf.String, 1)
//			m.Optional("id", protobuf.Int32, 2)
//			m.Repeated("emails", protobuf.String, 3)
//			m.Map("attrs", protobuf.String, protobuf.String, 4)
//		})
//	})
//	cls := pool.Lookup("Person").(*protobuf.Descriptor).Msgclass()
//	p := cls.New(map[string]any{"name": "Ada"})
//
// # Scope
//
// The runtime + builder are covered faithfully. What is deliberately out of
// scope (a comment, per the task): the gem's full protoc-generated codegen DSL
// (the giant serialized-FileDescriptorProto string a .proto compiles to) — this
// package offers the equivalent builder DSL instead — and proto2 group wire
// syntax. Well-known types (Any, Timestamp, Duration, Struct, Value, ListValue,
// FieldMask, the wrappers and Empty) are pre-registered in the generated pool
// and round-trip through the canonical runtime.
package protobuf
