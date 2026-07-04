// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"sort"
	"testing"
)

func TestDescriptorAccessors(t *testing.T) {
	p := newTestPool(t)
	d := p.Lookup("All").(*Descriptor)

	if d.Msgclass().Name() != "All" {
		t.Fatal("Msgclass name")
	}
	if d.Lookup("i32") == nil {
		t.Fatal("Lookup i32")
	}
	if d.Lookup("missing") != nil {
		t.Fatal("Lookup missing should be nil")
	}

	var names []string
	types := map[string]Symbol{}
	labels := map[string]Symbol{}
	d.Each(func(f *FieldDescriptor) {
		names = append(names, f.Name())
		types[f.Name()] = f.Type()
		labels[f.Name()] = f.Label()
	})
	sort.Strings(names)
	if len(names) == 0 {
		t.Fatal("Each yielded no fields")
	}
	if types["i32"] != Int32 || types["en"] != Enum || types["inner"] != MessageType {
		t.Fatalf("field types wrong: %v", types)
	}
	if labels["ri"] != "repeated" || labels["i32"] != "optional" {
		t.Fatalf("labels wrong: %v", labels)
	}

	fd := d.Lookup("i32")
	if fd.Number() != 1 {
		t.Fatalf("i32 number = %d", fd.Number())
	}
}

func TestEnumDescriptor(t *testing.T) {
	p := newTestPool(t)
	e := p.Lookup("Color").(*EnumDescriptor)

	if n, ok := e.LookupName("GREEN"); !ok || n != 1 {
		t.Fatalf("LookupName GREEN = %d,%v", n, ok)
	}
	if _, ok := e.LookupName("PURPLE"); ok {
		t.Fatal("LookupName PURPLE should be absent")
	}
	if s, ok := e.LookupValue(2); !ok || s != Symbol("BLUE") {
		t.Fatalf("LookupValue 2 = %v,%v", s, ok)
	}
	if _, ok := e.LookupValue(99); ok {
		t.Fatal("LookupValue 99 should be absent")
	}
}
