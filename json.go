// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

// JSONOptions mirror the keyword options of the gem's encode_json / decode_json.
type JSONOptions struct {
	// EmitDefaults includes fields set to their default value, mirroring
	// emit_defaults: true (protojson's EmitUnpopulated).
	EmitDefaults bool
	// PreserveProtoNames emits/accepts the proto field names rather than the
	// lowerCamelCase JSON names, mirroring preserve_proto_fieldnames: true
	// (protojson's UseProtoNames / decode's field-name handling).
	PreserveProtoNames bool
	// IgnoreUnknownFields skips unknown fields on decode, mirroring
	// ignore_unknown_fields: true (protojson's DiscardUnknown).
	IgnoreUnknownFields bool
}

func firstJSONOpts(opts []JSONOptions) JSONOptions {
	if len(opts) > 0 {
		return opts[0]
	}
	return JSONOptions{}
}

// EncodeJSON renders a message as proto3-JSON, mirroring
// Google::Protobuf.encode_json(msg). The mapping follows the proto3 JSON spec
// (via protojson). Note protojson deliberately varies insignificant whitespace
// between runs, so compare decoded values, not raw bytes, for equality.
func EncodeJSON(m *Message, opts ...JSONOptions) ([]byte, error) {
	o := firstJSONOpts(opts)
	mo := protojson.MarshalOptions{
		EmitUnpopulated: o.EmitDefaults,
		UseProtoNames:   o.PreserveProtoNames,
	}
	b, err := mo.Marshal(m.m.Interface())
	if err != nil {
		return nil, &ArgumentError{Message: err.Error()}
	}
	return b, nil
}

// DecodeJSON parses proto3-JSON into a new instance of class, mirroring
// Google::Protobuf.decode_json(Klass, json). Malformed input is a ParseError.
func DecodeJSON(class *MessageClass, data []byte, opts ...JSONOptions) (*Message, error) {
	o := firstJSONOpts(opts)
	msg := dynamicpb.NewMessage(class.md)
	uo := protojson.UnmarshalOptions{DiscardUnknown: o.IgnoreUnknownFields}
	if err := uo.Unmarshal(data, msg); err != nil {
		return nil, &ParseError{Message: err.Error()}
	}
	return &Message{m: msg, pool: class.pool}, nil
}
