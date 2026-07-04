// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"reflect"
	"testing"
)

func TestRepeatedFieldStandalone(t *testing.T) {
	r, err := NewRepeatedField(Int32, int64(1), int64(2), int64(3))
	if err != nil {
		t.Fatal(err)
	}
	if r.Length() != 3 {
		t.Fatalf("len = %d", r.Length())
	}
	if r.At(0) != int64(1) || r.At(-1) != int64(3) {
		t.Fatalf("At: %v %v", r.At(0), r.At(-1))
	}
	if r.At(99) != nil || r.At(-99) != nil {
		t.Fatal("out-of-range At must be nil")
	}
	if err := r.SetAt(1, int64(20)); err != nil {
		t.Fatal(err)
	}
	if r.At(1) != int64(20) {
		t.Fatal("SetAt")
	}
	if err := r.SetAt(-1, int64(30)); err != nil {
		t.Fatal(err)
	}
	if r.At(2) != int64(30) {
		t.Fatal("SetAt negative")
	}
	if err := r.SetAt(99, int64(0)); err == nil {
		t.Fatal("SetAt out of range should error")
	}
	if err := r.SetAt(0, "bad"); err == nil {
		t.Fatal("SetAt wrong type should error")
	}

	var sum int64
	r.Each(func(v any) { sum += v.(int64) })
	if sum != 1+20+30 {
		t.Fatalf("Each sum = %d", sum)
	}
	if !reflect.DeepEqual(r.ToArray(), []any{int64(1), int64(20), int64(30)}) {
		t.Fatalf("ToArray = %v", r.ToArray())
	}
	if r.Inspect() != "[1, 20, 30]" {
		t.Fatalf("Inspect = %s", r.Inspect())
	}

	// Push error.
	if err := r.Push("bad"); err == nil {
		t.Fatal("Push wrong type should error")
	}

	// Concat.
	r2, _ := NewRepeatedField(Int32)
	if err := r2.Concat(r); err != nil {
		t.Fatal(err)
	}
	if err := r2.Concat([]any{int64(99)}); err != nil {
		t.Fatal(err)
	}
	if r2.Length() != 4 {
		t.Fatalf("after concat len = %d", r2.Length())
	}
	if err := r2.Concat("bad"); err == nil {
		t.Fatal("Concat wrong type should error")
	}

	// Dup + Clear + Equal.
	d := r.Dup()
	if !r.Equal(d) {
		t.Fatal("Dup not equal")
	}
	d.Clear()
	if d.Length() != 0 {
		t.Fatal("Clear")
	}
	if r.Equal(d) {
		t.Fatal("cleared should differ in length")
	}
	if r.Equal(nil) {
		t.Fatal("Equal(nil)")
	}
	// Same length, different element.
	e1, _ := NewRepeatedField(Int32, int64(1))
	e2, _ := NewRepeatedField(Int32, int64(2))
	if e1.Equal(e2) {
		t.Fatal("different elements equal")
	}

	// Unsupported standalone element type.
	if _, err := NewRepeatedField(MessageType); err == nil {
		t.Fatal("standalone message list should error")
	}
	// Caching: second call for the same type reuses the descriptor.
	if _, err := NewRepeatedField(Int32); err != nil {
		t.Fatal(err)
	}
}

func TestRepeatedFieldOnMessage(t *testing.T) {
	p := newTestPool(t)
	m := mustNew(t, p.LookupMsgclass("All"))

	// Assign a []any.
	mustSet(t, m, "ri", []any{int64(1), int64(2)})
	if mustGet(t, m, "ri").(*RepeatedField).Length() != 2 {
		t.Fatal("setList []any")
	}
	// Assign a *RepeatedField.
	src, _ := NewRepeatedField(Int32, int64(7))
	mustSet(t, m, "ri", src)
	if mustGet(t, m, "ri").(*RepeatedField).At(0) != int64(7) {
		t.Fatal("setList *RepeatedField")
	}
	// Wrong container type.
	if err := m.Set("ri", "notarray"); err == nil {
		t.Fatal("setList wrong type should error")
	}
	// Element type error.
	if err := m.Set("ri", []any{"bad"}); err == nil {
		t.Fatal("setList element error")
	}

	// Repeated message field with valueEqual over messages.
	inner := p.LookupMsgclass("Inner")
	rm := mustGet(t, m, "rm").(*RepeatedField)
	_ = rm.Push(mustNew(t, inner, map[string]any{"s": "a"}))
	got := rm.At(0).(*Message)
	if mustGet(t, got, "s") != "a" {
		t.Fatal("repeated message element")
	}
	dup := rm.Dup()
	if !rm.Equal(dup) {
		t.Fatal("repeated message Dup equal")
	}
	_ = rm.Push(mustNew(t, inner, map[string]any{"s": "b"}))
	if rm.Equal(dup) {
		t.Fatal("repeated message differing")
	}
}

func TestMapStandalone(t *testing.T) {
	m, err := NewMap(String, Int32)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Set("a", int64(1)); err != nil {
		t.Fatal(err)
	}
	_ = m.Set("b", int64(2))
	if m.Length() != 2 {
		t.Fatalf("len = %d", m.Length())
	}
	if v, ok := m.Get("a"); !ok || v != int64(1) {
		t.Fatalf("Get a = %v,%v", v, ok)
	}
	if _, ok := m.Get("zzz"); ok {
		t.Fatal("absent key present")
	}
	if !m.Has("a") || m.Has("zzz") {
		t.Fatal("Has")
	}
	if !reflect.DeepEqual(m.Keys(), []any{"a", "b"}) {
		t.Fatalf("Keys = %v", m.Keys())
	}
	if !reflect.DeepEqual(m.Values(), []any{int64(1), int64(2)}) {
		t.Fatalf("Values = %v", m.Values())
	}
	seen := map[any]any{}
	m.Each(func(k, v any) { seen[k] = v })
	if !reflect.DeepEqual(seen, map[any]any{"a": int64(1), "b": int64(2)}) {
		t.Fatalf("Each = %v", seen)
	}
	if !reflect.DeepEqual(m.ToHash(), map[any]any{"a": int64(1), "b": int64(2)}) {
		t.Fatal("ToHash")
	}
	if m.Inspect() != `{"a"=>1, "b"=>2}` {
		t.Fatalf("Inspect = %s", m.Inspect())
	}

	// Delete.
	if !m.Delete("a") || m.Delete("a") {
		t.Fatal("Delete")
	}

	// Errors.
	if err := m.Set(123, int64(1)); err == nil {
		t.Fatal("Set wrong key type")
	}
	if err := m.Set("k", "notint"); err == nil {
		t.Fatal("Set wrong value type")
	}
	if _, ok := m.Get(123); ok {
		t.Fatal("Get wrong key type ok")
	}
	if m.Delete(123) {
		t.Fatal("Delete wrong key type")
	}
	if m.Has(123) {
		t.Fatal("Has wrong key type")
	}

	// Dup + Clear + Equal.
	m.Set("a", int64(1))
	d := m.Dup()
	if !m.Equal(d) {
		t.Fatal("Dup equal")
	}
	d.Clear()
	if d.Length() != 0 {
		t.Fatal("Clear")
	}
	if m.Equal(d) {
		t.Fatal("length differs")
	}
	if m.Equal(nil) {
		t.Fatal("Equal(nil)")
	}
	// Same length, differing value and differing key.
	x, _ := NewMap(String, Int32)
	x.Set("a", int64(1))
	x.Set("b", int64(2))
	y, _ := NewMap(String, Int32)
	y.Set("a", int64(1))
	y.Set("b", int64(999))
	if x.Equal(y) {
		t.Fatal("differing value equal")
	}
	z, _ := NewMap(String, Int32)
	z.Set("a", int64(1))
	z.Set("c", int64(2))
	if x.Equal(z) {
		t.Fatal("differing key equal")
	}

	// Unsupported standalone value type.
	if _, err := NewMap(String, MessageType); err == nil {
		t.Fatal("standalone message-valued map should error")
	}
	// Cache hit.
	if _, err := NewMap(String, Int32); err != nil {
		t.Fatal(err)
	}
}

func TestMapKeyOrderingAllTypes(t *testing.T) {
	// int64 keys.
	mi, _ := NewMap(Int32, Int32)
	mi.Set(int64(3), int64(0))
	mi.Set(int64(1), int64(0))
	if !reflect.DeepEqual(mi.Keys(), []any{int64(1), int64(3)}) {
		t.Fatalf("int keys = %v", mi.Keys())
	}
	// uint64 keys.
	mu, _ := NewMap(Uint32, Int32)
	mu.Set(uint64(5), int64(0))
	mu.Set(uint64(2), int64(0))
	if !reflect.DeepEqual(mu.Keys(), []any{uint64(2), uint64(5)}) {
		t.Fatalf("uint keys = %v", mu.Keys())
	}
	// bool keys.
	mb, _ := NewMap(Bool, Int32)
	mb.Set(true, int64(0))
	mb.Set(false, int64(0))
	if !reflect.DeepEqual(mb.Keys(), []any{false, true}) {
		t.Fatalf("bool keys = %v", mb.Keys())
	}
}

func TestMapOnMessage(t *testing.T) {
	p := newTestPool(t)
	m := mustNew(t, p.LookupMsgclass("All"))
	inner := p.LookupMsgclass("Inner")

	// setMap via *Map.
	src, _ := NewMap(String, Int32)
	src.Set("k", int64(1))
	mustSet(t, m, "mi", src)
	if v, _ := mustGet(t, m, "mi").(*Map).Get("k"); v != int64(1) {
		t.Fatal("setMap *Map")
	}
	// setMap via map[string]any.
	mustSet(t, m, "mi", map[string]any{"z": int64(9)})
	if v, _ := mustGet(t, m, "mi").(*Map).Get("z"); v != int64(9) {
		t.Fatal("setMap map[string]any")
	}
	// setMap via map[any]any.
	mustSet(t, m, "mi", map[any]any{"q": int64(4)})
	if v, _ := mustGet(t, m, "mi").(*Map).Get("q"); v != int64(4) {
		t.Fatal("setMap map[any]any")
	}
	// Wrong container.
	if err := m.Set("mi", 123); err == nil {
		t.Fatal("setMap wrong type")
	}
	// Key error and value error.
	if err := m.Set("mi", map[any]any{123: int64(1)}); err == nil {
		t.Fatal("setMap key error")
	}
	if err := m.Set("mi", map[any]any{"k": "notint"}); err == nil {
		t.Fatal("setMap value error")
	}

	// Message-valued map field.
	mm := mustGet(t, m, "mm").(*Map)
	if err := mm.Set(int64(1), mustNew(t, inner, map[string]any{"s": "v"})); err != nil {
		t.Fatal(err)
	}
	got, _ := mm.Get(int64(1))
	if mustGet(t, got.(*Message), "s") != "v" {
		t.Fatal("message-valued map get")
	}
	// Equal over message values.
	d := mm.Dup()
	if !mm.Equal(d) {
		t.Fatal("message map Dup equal")
	}

	// ToH over a message-valued map routes the value through fromProtoScalar's
	// message path and yields a nested *Message.
	h := m.ToH()
	mmH := h["mm"].(map[any]any)
	if inner, ok := mmH[int64(1)].(*Message); !ok || mustGet(t, inner, "s") != "v" {
		t.Fatalf("ToH message-valued map = %v", mmH)
	}
}
