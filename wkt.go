// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// wellKnownFiles are the well-known-type descriptors pre-registered in every
// pool, so the builder DSL can reference them (e.g. optional :ts, :message, 1,
// "google.protobuf.Timestamp") and the gem's Google::Protobuf::Timestamp /
// Duration / Any / Struct / Value / ListValue / FieldMask / wrappers / Empty are
// resolvable, matching the gem's generated_pool.
func wellKnownFiles() []protoreflect.FileDescriptor {
	return []protoreflect.FileDescriptor{
		timestamppb.File_google_protobuf_timestamp_proto,
		durationpb.File_google_protobuf_duration_proto,
		anypb.File_google_protobuf_any_proto,
		structpb.File_google_protobuf_struct_proto,
		wrapperspb.File_google_protobuf_wrappers_proto,
		fieldmaskpb.File_google_protobuf_field_mask_proto,
		emptypb.File_google_protobuf_empty_proto,
	}
}

// registerWellKnownTypes registers the WKT files into files. Each file is
// self-contained (no cross imports), so registration cannot fail here.
func registerWellKnownTypes(files *protoregistry.Files) {
	for _, fd := range wellKnownFiles() {
		_ = files.RegisterFile(fd)
	}
}

// WellKnownType returns the message class for a well-known type by its short
// name (e.g. "Timestamp", "Duration", "Any", "Struct", "Value", "ListValue",
// "FieldMask", "Empty", "StringValue"), from the generated pool. It returns nil
// for an unknown name.
func WellKnownType(shortName string) *MessageClass {
	return GeneratedPool().LookupMsgclass("google.protobuf." + shortName)
}

// anyTypePrefix is the conventional type-URL prefix the gem uses when packing an
// Any (Any#pack).
const anyTypePrefix = "type.googleapis.com/"

// AnyPack wraps msg in a new google.protobuf.Any, mirroring
// Google::Protobuf::Any#pack: it sets type_url to the conventional
// type.googleapis.com/<full name> and value to the binary encoding of msg.
func AnyPack(msg *Message) (*Message, error) {
	anyCls := WellKnownType("Any")
	any, err := anyCls.New()
	if err != nil {
		return nil, err
	}
	data, err := Encode(msg)
	if err != nil {
		return nil, err
	}
	if err := any.Set("type_url", anyTypePrefix+string(msg.m.Descriptor().FullName())); err != nil {
		return nil, err
	}
	if err := any.Set("value", data); err != nil {
		return nil, err
	}
	return any, nil
}

// AnyIs reports whether the Any holds a message of class, mirroring
// Google::Protobuf::Any#is?(Klass).
func AnyIs(any *Message, class *MessageClass) bool {
	url, err := any.Get("type_url")
	if err != nil {
		return false
	}
	s, ok := url.(string)
	if !ok {
		return false
	}
	return typeFromURL(s) == class.Name()
}

// AnyUnpack decodes the message packed inside the Any as an instance of class,
// mirroring Google::Protobuf::Any#unpack(Klass). It returns nil when the Any
// does not hold a message of that type.
func AnyUnpack(any *Message, class *MessageClass) (*Message, error) {
	if !AnyIs(any, class) {
		return nil, nil
	}
	val, err := any.Get("value")
	if err != nil {
		return nil, err
	}
	data, _ := val.([]byte)
	return Decode(class, data)
}

// typeFromURL returns the trailing type name of an Any type_url.
func typeFromURL(url string) string {
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			return url[i+1:]
		}
	}
	return url
}
