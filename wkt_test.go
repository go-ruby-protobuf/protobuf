// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "testing"

func TestWellKnownTypeLookup(t *testing.T) {
	for _, name := range []string{
		"Timestamp", "Duration", "Any", "Struct", "Value",
		"ListValue", "FieldMask", "Empty", "StringValue",
	} {
		if WellKnownType(name) == nil {
			t.Errorf("well-known type %q not resolvable", name)
		}
	}
	if WellKnownType("NotARealType") != nil {
		t.Fatal("unknown WKT should be nil")
	}
}

func TestAnyPackUnpackIs(t *testing.T) {
	p := GeneratedPool()
	if err := p.Build(func(b *Builder) {
		b.AddMessage("AnyPayload", func(m *MessageBuilder) {
			m.Optional("note", String, 1)
		})
	}); err != nil {
		t.Fatal(err)
	}
	payloadCls := p.LookupMsgclass("AnyPayload")
	otherCls := WellKnownType("StringValue")

	msg := mustNew(t, payloadCls, map[string]any{"note": "hi"})
	any, err := AnyPack(msg)
	if err != nil {
		t.Fatal(err)
	}
	if url := mustGet(t, any, "type_url"); url != "type.googleapis.com/AnyPayload" {
		t.Fatalf("type_url = %v", url)
	}

	if !AnyIs(any, payloadCls) {
		t.Fatal("AnyIs payload should be true")
	}
	if AnyIs(any, otherCls) {
		t.Fatal("AnyIs other should be false")
	}
	// AnyIs on a message that is not an Any -> false.
	if AnyIs(msg, payloadCls) {
		t.Fatal("AnyIs on non-Any should be false")
	}

	back, err := AnyUnpack(any, payloadCls)
	if err != nil {
		t.Fatal(err)
	}
	if mustGet(t, back, "note") != "hi" {
		t.Fatal("AnyUnpack lost data")
	}
	// Unpacking as the wrong type returns nil.
	if other, err := AnyUnpack(any, otherCls); err != nil || other != nil {
		t.Fatalf("AnyUnpack wrong type = %v,%v", other, err)
	}
}

func TestAnyPackEncodeError(t *testing.T) {
	p := GeneratedPool()
	if err := p.Build(func(b *Builder) {
		b.AddMessage("AnyBad", func(m *MessageBuilder) {
			m.Optional("s", String, 1)
		})
	}); err != nil {
		t.Fatal(err)
	}
	msg := mustNew(t, p.LookupMsgclass("AnyBad"))
	mustSet(t, msg, "s", "\xff") // invalid UTF-8 -> Encode fails inside AnyPack
	if _, err := AnyPack(msg); err == nil {
		t.Fatal("AnyPack should surface encode error")
	}
}

func TestTypeFromURL(t *testing.T) {
	if typeFromURL("type.googleapis.com/foo.Bar") != "foo.Bar" {
		t.Fatal("with slash")
	}
	if typeFromURL("noslash") != "noslash" {
		t.Fatal("without slash")
	}
}
