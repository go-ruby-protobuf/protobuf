// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "testing"

func TestBuilderMapKeyErrors(t *testing.T) {
	// Unknown map key type -> the MessageBuilder.Map guard fires.
	if _, err := NewMap(Symbol("bogus"), Int32); err == nil {
		t.Fatal("unknown map key type should error")
	}
	// A :message key type -> the synthesised key field fails to compile.
	if _, err := NewMap(MessageType, Int32); err == nil {
		t.Fatal("message map key should error")
	}
}

func TestBuilderMapNameCamelCase(t *testing.T) {
	// A snake_case map field name exercises the camel() underscore handling when
	// naming the synthesised map-entry message (My_attrs -> MyAttrsEntry).
	p := NewDescriptorPool()
	if err := p.Build(func(b *Builder) {
		b.AddMessage("Holder", func(m *MessageBuilder) {
			m.Map("my_attrs", String, Int32, 1)
		})
	}); err != nil {
		t.Fatalf("build: %v", err)
	}
	msg := mustNew(t, p.LookupMsgclass("Holder"))
	mp := mustGet(t, msg, "my_attrs").(*Map)
	if err := mp.Set("k", int64(1)); err != nil {
		t.Fatal(err)
	}
	if v, _ := mp.Get("k"); v != int64(1) {
		t.Fatal("snake_case map field round trip")
	}
}

func TestMessageClassDescriptor(t *testing.T) {
	p := newTestPool(t)
	cls := p.LookupMsgclass("All")
	if cls.Descriptor().Name() != "All" {
		t.Fatal("MessageClass.Descriptor().Name()")
	}
}

func TestNewRepeatedFieldPushError(t *testing.T) {
	if _, err := NewRepeatedField(Int32, "not-an-int"); err == nil {
		t.Fatal("NewRepeatedField with a bad initial value should error")
	}
}
