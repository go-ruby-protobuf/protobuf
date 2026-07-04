// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "testing"

// newTestPool builds a pool with a rich schema exercising every field kind,
// enums, repeated (scalar + message), maps (scalar + message value), a oneof and
// a nested message referencing a well-known type.
func newTestPool(t *testing.T) *DescriptorPool {
	t.Helper()
	p := NewDescriptorPool()
	err := p.Build(func(b *Builder) {
		b.AddEnum("Color", func(e *EnumBuilder) {
			e.Value("RED", 0)
			e.Value("GREEN", 1)
			e.Value("BLUE", 2)
		})
		b.AddMessage("Inner", func(m *MessageBuilder) {
			m.Optional("s", String, 1)
		})
		b.AddMessage("All", func(m *MessageBuilder) {
			m.Optional("i32", Int32, 1)
			m.Optional("i64", Int64, 2)
			m.Optional("u32", Uint32, 3)
			m.Optional("u64", Uint64, 4)
			m.Optional("s32", Sint32, 5)
			m.Optional("s64", Sint64, 6)
			m.Optional("f32", Fixed32, 7)
			m.Optional("f64", Fixed64, 8)
			m.Optional("sf32", Sfixed32, 9)
			m.Optional("sf64", Sfixed64, 10)
			m.Optional("fl", Float, 11)
			m.Optional("db", Double, 12)
			m.Optional("bl", Bool, 13)
			m.Optional("st", String, 14)
			m.Optional("by", Bytes, 15)
			m.Optional("en", Enum, 16, "Color")
			m.Optional("inner", MessageType, 17, "Inner")
			m.Repeated("ri", Int32, 18)
			m.Repeated("rm", MessageType, 19, "Inner")
			m.Map("mi", String, Int32, 20)
			m.Map("mm", Int32, MessageType, 21, "Inner")
			m.Oneof("choice", func(o *OneofBuilder) {
				o.Optional("oa", Int32, 22)
				o.Optional("ob", String, 23)
			})
		})
	})
	if err != nil {
		t.Fatalf("build test pool: %v", err)
	}
	return p
}

func mustNew(t *testing.T, c *MessageClass, init ...map[string]any) *Message {
	t.Helper()
	m, err := c.New(init...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}

func mustGet(t *testing.T, m *Message, name string) any {
	t.Helper()
	v, err := m.Get(name)
	if err != nil {
		t.Fatalf("Get(%q): %v", name, err)
	}
	return v
}

func mustSet(t *testing.T, m *Message, name string, v any) {
	t.Helper()
	if err := m.Set(name, v); err != nil {
		t.Fatalf("Set(%q, %v): %v", name, v, err)
	}
}
