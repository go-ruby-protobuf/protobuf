// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "testing"

func TestGeneratedPoolSingleton(t *testing.T) {
	if GeneratedPool() != GeneratedPool() {
		t.Fatal("GeneratedPool must be a singleton")
	}
	// Well-known types are pre-registered.
	if GeneratedPool().LookupMsgclass("google.protobuf.Timestamp") == nil {
		t.Fatal("Timestamp not in generated pool")
	}
}

func TestLookupVariants(t *testing.T) {
	p := newTestPool(t)

	if d, ok := p.Lookup("All").(*Descriptor); !ok || d.Name() != "All" {
		t.Fatalf("Lookup All = %v", p.Lookup("All"))
	}
	if e, ok := p.Lookup("Color").(*EnumDescriptor); !ok || e.Name() != "Color" {
		t.Fatalf("Lookup Color = %v", p.Lookup("Color"))
	}
	if p.Lookup("Nope") != nil {
		t.Fatal("unknown lookup should be nil")
	}
	// A non-message/enum descriptor (a field) resolves to nil.
	if p.Lookup("All.i32") != nil {
		t.Fatal("field lookup should be nil (not a message/enum)")
	}
	if p.LookupMsgclass("Color") != nil {
		t.Fatal("LookupMsgclass on an enum should be nil")
	}
	if p.LookupMsgclass("All") == nil {
		t.Fatal("LookupMsgclass on a message should be non-nil")
	}
}

func TestBuildBuilderError(t *testing.T) {
	p := NewDescriptorPool()
	err := p.Build(func(b *Builder) {
		b.AddMessage("Bad", func(m *MessageBuilder) {
			m.Optional("x", Symbol("nonsense"), 1) // unknown type -> builder error
		})
	})
	if err == nil {
		t.Fatal("expected builder error")
	}
	if _, ok := err.(*ArgumentError); !ok {
		t.Fatalf("want *ArgumentError, got %T", err)
	}
}

func TestBuildFirstErrorWins(t *testing.T) {
	p := NewDescriptorPool()
	err := p.Build(func(b *Builder) {
		b.AddMessage("Bad", func(m *MessageBuilder) {
			m.Optional("x", Symbol("nope1"), 1)
			m.Optional("y", Symbol("nope2"), 2) // second failure is swallowed
		})
	})
	if err == nil || err.Error() != "unknown field type :nope1" {
		t.Fatalf("want first error, got %v", err)
	}
}

func TestBuildCompileError(t *testing.T) {
	p := NewDescriptorPool()
	// Two fields sharing a tag number: protodesc.NewFile rejects it.
	err := p.Build(func(b *Builder) {
		b.AddMessage("Dup", func(m *MessageBuilder) {
			m.Optional("a", Int32, 1)
			m.Optional("b", Int32, 1)
		})
	})
	if err == nil {
		t.Fatal("expected compile error for duplicate field number")
	}
}

func TestBuildRegisterConflict(t *testing.T) {
	p := NewDescriptorPool()
	build := func() error {
		return p.Build(func(b *Builder) {
			b.AddMessage("Same", func(m *MessageBuilder) { m.Optional("x", Int32, 1) })
		})
	}
	if err := build(); err != nil {
		t.Fatalf("first build: %v", err)
	}
	if err := build(); err == nil {
		t.Fatal("expected register conflict on duplicate message name")
	}
}

func TestBuildReferencePriorFile(t *testing.T) {
	p := NewDescriptorPool()
	if err := p.Build(func(b *Builder) {
		b.AddMessage("A", func(m *MessageBuilder) { m.Optional("x", Int32, 1) })
	}); err != nil {
		t.Fatal(err)
	}
	// A later build in the same pool references a message from an earlier build.
	if err := p.Build(func(b *Builder) {
		b.AddMessage("B", func(m *MessageBuilder) { m.Optional("a", MessageType, 1, "A") })
	}); err != nil {
		t.Fatalf("cross-file reference failed: %v", err)
	}
	msg := mustNew(t, p.LookupMsgclass("B"))
	a := mustNew(t, p.LookupMsgclass("A"), map[string]any{"x": int64(7)})
	mustSet(t, msg, "a", a)
	if got := mustGet(t, msg, "a").(*Message); mustGet(t, got, "x") != int64(7) {
		t.Fatal("cross-file nested value wrong")
	}
}
