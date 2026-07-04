// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Encode serialises a message to the protobuf binary wire format, mirroring
// Google::Protobuf.encode(msg). The bytes are produced by the canonical
// google.golang.org/protobuf runtime, so they are wire-compatible with real
// protobuf. An encoding failure (e.g. an invalid-UTF-8 proto3 string) is
// reported as an ArgumentError (the gem raises).
func Encode(m *Message) ([]byte, error) {
	b, err := proto.Marshal(m.m.Interface())
	if err != nil {
		return nil, &ArgumentError{Message: err.Error()}
	}
	return b, nil
}

// Decode parses binary protobuf bytes into a new instance of class, mirroring
// Google::Protobuf.decode(Klass, bytes). Malformed input is a ParseError.
func Decode(class *MessageClass, data []byte) (*Message, error) {
	msg := dynamicpb.NewMessage(class.md)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, &ParseError{Message: err.Error()}
	}
	return &Message{m: msg, pool: class.pool}, nil
}
