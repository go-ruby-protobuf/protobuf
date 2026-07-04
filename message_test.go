// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"math"
	"strings"
	"testing"
)

func TestScalarRoundTripAllKinds(t *testing.T) {
	p := newTestPool(t)
	cls := p.LookupMsgclass("All")
	m := mustNew(t, cls)

	mustSet(t, m, "i32", int64(-7))
	mustSet(t, m, "i64", int64(-1<<40))
	mustSet(t, m, "u32", uint64(4000000000))
	mustSet(t, m, "u64", uint64(1<<63))
	mustSet(t, m, "s32", 12)       // plain int
	mustSet(t, m, "s64", int32(9)) // int32 accepted
	mustSet(t, m, "f32", uint32(11))
	mustSet(t, m, "f64", uint64(12))
	mustSet(t, m, "sf32", int64(-13))
	mustSet(t, m, "sf64", int64(-14))
	mustSet(t, m, "fl", 1.5)
	mustSet(t, m, "db", 2.25)
	mustSet(t, m, "bl", true)
	mustSet(t, m, "st", "hi")
	mustSet(t, m, "by", []byte{1, 2, 3})
	mustSet(t, m, "en", Symbol("BLUE"))

	checks := map[string]any{
		"i32": int64(-7), "i64": int64(-1 << 40),
		"u32": uint64(4000000000), "u64": uint64(1 << 63),
		"s32": int64(12), "s64": int64(9),
		"f32": uint64(11), "f64": uint64(12),
		"sf32": int64(-13), "sf64": int64(-14),
		"fl": float64(1.5), "db": float64(2.25),
		"bl": true, "st": "hi", "en": Symbol("BLUE"),
	}
	for k, want := range checks {
		if got := mustGet(t, m, k); got != want {
			t.Errorf("%s = %v (%T), want %v (%T)", k, got, got, want, want)
		}
	}
	if by := mustGet(t, m, "by").([]byte); string(by) != "\x01\x02\x03" {
		t.Errorf("by = %v", by)
	}
}

func TestBytesFromString(t *testing.T) {
	m := mustNew(t, newTestPool(t).LookupMsgclass("All"))
	mustSet(t, m, "by", "raw") // string accepted for a bytes field
	if string(mustGet(t, m, "by").([]byte)) != "raw" {
		t.Fatal("bytes-from-string")
	}
}

func TestEnumByStringAndNumberAndUnknown(t *testing.T) {
	m := mustNew(t, newTestPool(t).LookupMsgclass("All"))
	mustSet(t, m, "en", "GREEN") // string name
	if mustGet(t, m, "en") != Symbol("GREEN") {
		t.Fatal("enum by string")
	}
	mustSet(t, m, "en", int64(2)) // number
	if mustGet(t, m, "en") != Symbol("BLUE") {
		t.Fatal("enum by number")
	}
	mustSet(t, m, "en", int64(42)) // unknown number -> read back as int64
	if mustGet(t, m, "en") != int64(42) {
		t.Fatalf("unknown enum = %v", mustGet(t, m, "en"))
	}
}

func TestMessageFieldGetSetClear(t *testing.T) {
	p := newTestPool(t)
	all := p.LookupMsgclass("All")
	inner := p.LookupMsgclass("Inner")
	m := mustNew(t, all)

	if mustGet(t, m, "inner") != nil {
		t.Fatal("unset message field must be nil")
	}
	in := mustNew(t, inner, map[string]any{"s": "deep"})
	mustSet(t, m, "inner", in)
	got := mustGet(t, m, "inner").(*Message)
	if mustGet(t, got, "s") != "deep" {
		t.Fatal("nested value")
	}
	mustSet(t, m, "inner", nil) // clear
	if mustGet(t, m, "inner") != nil {
		t.Fatal("message field not cleared")
	}
}

func TestOneof(t *testing.T) {
	m := mustNew(t, newTestPool(t).LookupMsgclass("All"))
	mustSet(t, m, "oa", int64(5))
	if mustGet(t, m, "oa") != int64(5) {
		t.Fatal("oneof oa")
	}
	mustSet(t, m, "ob", "x") // setting the other member clears oa
	if mustGet(t, m, "ob") != "x" {
		t.Fatal("oneof ob")
	}
}

func TestErrorPaths(t *testing.T) {
	p := newTestPool(t)
	all := p.LookupMsgclass("All")
	inner := p.LookupMsgclass("Inner")
	m := mustNew(t, all)

	assertErr := func(err error, class string) {
		t.Helper()
		if err == nil {
			t.Fatal("expected error")
		}
		if e, ok := err.(Error); !ok || e.RubyClass() != class {
			t.Fatalf("want RubyClass %s, got %v (%T)", class, err, err)
		}
	}

	// Unknown field.
	_, err := m.Get("nope")
	assertErr(err, "ArgumentError")
	assertErr(m.Set("nope", 1), "ArgumentError")

	// Wrong scalar types.
	assertErr(m.Set("i32", "notint"), "Google::Protobuf::TypeError")
	assertErr(m.Set("i64", "x"), "Google::Protobuf::TypeError")
	assertErr(m.Set("u32", "x"), "Google::Protobuf::TypeError")
	assertErr(m.Set("u64", "x"), "Google::Protobuf::TypeError")
	assertErr(m.Set("fl", "x"), "Google::Protobuf::TypeError")
	assertErr(m.Set("db", "x"), "Google::Protobuf::TypeError")
	assertErr(m.Set("bl", 1), "Google::Protobuf::TypeError")
	assertErr(m.Set("st", 1), "Google::Protobuf::TypeError")
	assertErr(m.Set("by", 1), "Google::Protobuf::TypeError")

	// Integer overflow.
	assertErr(m.Set("i32", int64(1<<40)), "RangeError")
	assertErr(m.Set("u32", uint64(1<<40)), "RangeError")

	// Negative into unsigned.
	assertErr(m.Set("u32", -1), "Google::Protobuf::TypeError")

	// uint64 too big to fit int64 target.
	assertErr(m.Set("i64", uint64(math.MaxUint64)), "Google::Protobuf::TypeError")

	// Enum errors.
	assertErr(m.Set("en", Symbol("NOPE")), "RangeError")
	assertErr(m.Set("en", "NOPE"), "RangeError")
	assertErr(m.Set("en", 1.5), "Google::Protobuf::TypeError")
	assertErr(m.Set("en", int64(1<<40)), "RangeError")

	// Wrong message type.
	wrongType := mustNew(t, all) // an All where an Inner is expected
	assertErr(m.Set("inner", wrongType), "Google::Protobuf::TypeError")
	// Non-*Message into a message field.
	assertErr(m.Set("inner", "x"), "Google::Protobuf::TypeError")
	_ = inner

	// New with a bad init value.
	if _, e := all.New(map[string]any{"i32": "bad"}); e == nil {
		t.Fatal("New should surface init error")
	}
}

func TestToHEqualDupCloneInspect(t *testing.T) {
	p := newTestPool(t)
	all := p.LookupMsgclass("All")
	inner := p.LookupMsgclass("Inner")

	m := mustNew(t, all, map[string]any{"i32": int64(1), "st": "s"})
	mustSet(t, m, "inner", mustNew(t, inner, map[string]any{"s": "y"}))
	ri := mustGet(t, m, "ri").(*RepeatedField)
	_ = ri.Push(int64(1), int64(2))
	mi := mustGet(t, m, "mi").(*Map)
	_ = mi.Set("k", int64(9))

	h := m.ToH()
	if h["i32"] != int64(1) || h["st"] != "s" {
		t.Fatalf("ToH scalars: %v", h)
	}
	if h["inner"].(map[string]any)["s"] != "y" {
		t.Fatal("ToH nested message")
	}
	if len(h["ri"].([]any)) != 2 {
		t.Fatal("ToH repeated")
	}
	if h["mi"].(map[any]any)["k"] != int64(9) {
		t.Fatal("ToH map")
	}
	// Unset message field is nil in ToH.
	m2 := mustNew(t, all)
	if m2.ToH()["inner"] != nil {
		t.Fatal("unset message in ToH should be nil")
	}

	// Equal / Dup / Clone.
	d := m.Dup()
	if !m.Equal(d) {
		t.Fatal("Dup not equal")
	}
	c := m.Clone()
	if !m.Equal(c) {
		t.Fatal("Clone not equal")
	}
	// Mutating the dup does not affect the original.
	mustSet(t, d, "i32", int64(999))
	if m.Equal(d) {
		t.Fatal("dup should be independent")
	}
	if m.Equal(nil) {
		t.Fatal("Equal(nil) must be false")
	}
	if m.Equal(m2) {
		t.Fatal("different contents equal")
	}

	// Inspect.
	s := m.Inspect()
	if !strings.HasPrefix(s, "<All: ") || !strings.Contains(s, "i32: 1") {
		t.Fatalf("Inspect = %s", s)
	}
	if m.Class().Name() != "All" {
		t.Fatal("Class name")
	}
}

func TestInspectValueVariants(t *testing.T) {
	p := newTestPool(t)
	m := mustNew(t, p.LookupMsgclass("All"))
	// Exercise nil (unset message), string, bytes, symbol, repeated, map, numeric.
	mustSet(t, m, "st", "hello")
	mustSet(t, m, "by", []byte("bb"))
	mustSet(t, m, "en", Symbol("RED"))
	s := m.Inspect()
	for _, want := range []string{`st: "hello"`, `by: "bb"`, `en: :RED`, `inner: nil`, `ri: []`, `mi: {}`} {
		if !strings.Contains(s, want) {
			t.Errorf("Inspect missing %q in %s", want, s)
		}
	}
}
