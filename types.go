// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "google.golang.org/protobuf/types/descriptorpb"

// Symbol is a Ruby Symbol (`:name`). The builder DSL takes field types as
// Symbols (`protobuf.Int32` == Ruby `:int32`); enum field values are read back
// as Symbols (the value's name), matching the gem.
type Symbol string

// Field-type symbols accepted by the builder DSL, mirroring the gem's type
// symbols (:int32, :string, :message, …).
const (
	Int32       Symbol = "int32"
	Int64       Symbol = "int64"
	Uint32      Symbol = "uint32"
	Uint64      Symbol = "uint64"
	Sint32      Symbol = "sint32"
	Sint64      Symbol = "sint64"
	Fixed32     Symbol = "fixed32"
	Fixed64     Symbol = "fixed64"
	Sfixed32    Symbol = "sfixed32"
	Sfixed64    Symbol = "sfixed64"
	Float       Symbol = "float"
	Double      Symbol = "double"
	Bool        Symbol = "bool"
	String      Symbol = "string"
	Bytes       Symbol = "bytes"
	MessageType Symbol = "message"
	Enum        Symbol = "enum"
)

// fieldTypes maps a DSL type symbol to its descriptorpb type. message and enum
// additionally require a referenced type name (handled by the caller).
var fieldTypes = map[Symbol]descriptorpb.FieldDescriptorProto_Type{
	Int32:       descriptorpb.FieldDescriptorProto_TYPE_INT32,
	Int64:       descriptorpb.FieldDescriptorProto_TYPE_INT64,
	Uint32:      descriptorpb.FieldDescriptorProto_TYPE_UINT32,
	Uint64:      descriptorpb.FieldDescriptorProto_TYPE_UINT64,
	Sint32:      descriptorpb.FieldDescriptorProto_TYPE_SINT32,
	Sint64:      descriptorpb.FieldDescriptorProto_TYPE_SINT64,
	Fixed32:     descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
	Fixed64:     descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
	Sfixed32:    descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
	Sfixed64:    descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
	Float:       descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
	Double:      descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
	Bool:        descriptorpb.FieldDescriptorProto_TYPE_BOOL,
	String:      descriptorpb.FieldDescriptorProto_TYPE_STRING,
	Bytes:       descriptorpb.FieldDescriptorProto_TYPE_BYTES,
	MessageType: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
	Enum:        descriptorpb.FieldDescriptorProto_TYPE_ENUM,
}
