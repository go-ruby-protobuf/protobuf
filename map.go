// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Map is a typed protobuf map, mirroring Google::Protobuf::Map. Keys and values
// are type-checked against the map's key/value types on insertion. Iteration
// order is deterministic (keys sorted) so Inspect / Keys / Each are reproducible;
// protobuf maps are unordered on the wire, so this only affects presentation.
type Map struct {
	m    protoreflect.Map
	fd   protoreflect.FieldDescriptor
	pool *DescriptorPool
}

// NewMap builds a standalone map with scalar key type k and scalar value type v,
// mirroring Google::Protobuf::Map.new(:string, :int32). Message/enum value types
// are not supported standalone (obtain such a map from a message field).
func NewMap(k, v Symbol) (*Map, error) {
	fd, err := scratchMap(k, v)
	if err != nil {
		return nil, err
	}
	return &Map{m: backingMap(fd), fd: fd, pool: scratchPool}, nil
}

// keyValue converts a Ruby key to its protoreflect.MapKey.
func (m *Map) keyValue(key any) (protoreflect.MapKey, error) {
	kv, err := toProtoScalar(m.fd.MapKey(), key)
	if err != nil {
		return protoreflect.MapKey{}, err
	}
	return kv.MapKey(), nil
}

// Length returns the number of entries, mirroring #length / #size.
func (m *Map) Length() int { return m.m.Len() }

// Get returns the value for key and whether it is present, mirroring map[key].
func (m *Map) Get(key any) (any, bool) {
	mk, err := m.keyValue(key)
	if err != nil || !m.m.Has(mk) {
		return nil, false
	}
	return elemToRuby(m.fd.MapValue(), m.m.Get(mk), m.pool), true
}

// Set inserts or replaces the entry for key, mirroring map[key] = value.
func (m *Map) Set(key, val any) error {
	mk, err := m.keyValue(key)
	if err != nil {
		return err
	}
	vv, err := toProtoScalar(m.fd.MapValue(), val)
	if err != nil {
		return err
	}
	m.m.Set(mk, vv)
	return nil
}

// Delete removes key, returning whether it was present, mirroring #delete.
func (m *Map) Delete(key any) bool {
	mk, err := m.keyValue(key)
	if err != nil || !m.m.Has(mk) {
		return false
	}
	m.m.Clear(mk)
	return true
}

// Has reports whether key is present, mirroring #key? / #has_key?.
func (m *Map) Has(key any) bool {
	mk, err := m.keyValue(key)
	if err != nil {
		return false
	}
	return m.m.Has(mk)
}

// Keys returns the keys in deterministic (sorted) order, mirroring #keys.
func (m *Map) Keys() []any {
	type kv struct {
		mk protoreflect.MapKey
		rb any
	}
	all := make([]kv, 0, m.m.Len())
	m.m.Range(func(k protoreflect.MapKey, _ protoreflect.Value) bool {
		all = append(all, kv{k, fromProtoScalar(m.fd.MapKey(), k.Value(), m.pool)})
		return true
	})
	sort.Slice(all, func(i, j int) bool { return lessKey(all[i].rb, all[j].rb) })
	out := make([]any, len(all))
	for i := range all {
		out[i] = all[i].rb
	}
	return out
}

// Values returns the values ordered by their (sorted) keys, mirroring #values.
func (m *Map) Values() []any {
	keys := m.Keys()
	out := make([]any, len(keys))
	for i, k := range keys {
		out[i], _ = m.Get(k)
	}
	return out
}

// Each iterates entries in sorted-key order, mirroring #each.
func (m *Map) Each(fn func(key, val any)) {
	for _, k := range m.Keys() {
		v, _ := m.Get(k)
		fn(k, v)
	}
}

// ToHash returns the map as a map[any]any, mirroring #to_h.
func (m *Map) ToHash() map[any]any {
	out := map[any]any{}
	m.Each(func(k, v any) { out[k] = v })
	return out
}

// Clear removes all entries, mirroring #clear.
func (m *Map) Clear() {
	for _, k := range m.Keys() {
		m.Delete(k)
	}
}

// Equal reports whether m and other hold the same entries, mirroring #==.
func (m *Map) Equal(other *Map) bool {
	if other == nil || m.m.Len() != other.m.Len() {
		return false
	}
	equal := true
	m.m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		if !other.m.Has(k) || !valueEqual(m.fd.MapValue(), v, other.m.Get(k)) {
			equal = false
			return false
		}
		return true
	})
	return equal
}

// Dup returns a shallow copy with the same key/value types, mirroring #dup.
func (m *Map) Dup() *Map {
	dup := &Map{m: backingMap(m.fd), fd: m.fd, pool: m.pool}
	m.m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		dup.m.Set(k, v)
		return true
	})
	return dup
}

// Inspect renders the map, mirroring #inspect: {k=>v, …}.
func (m *Map) Inspect() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	m.Each(func(k, v any) {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(inspectValue(k))
		b.WriteString("=>")
		b.WriteString(inspectValue(v))
	})
	b.WriteByte('}')
	return b.String()
}

// lessKey orders map keys of the same (scalar) type for deterministic iteration.
// Protobuf map keys are always one of string, int64 (signed 32/64-bit kinds),
// uint64 (unsigned 32/64-bit kinds) or bool, so those four cases are exhaustive.
func lessKey(a, b any) bool {
	switch x := a.(type) {
	case string:
		return x < b.(string)
	case int64:
		return x < b.(int64)
	case uint64:
		return x < b.(uint64)
	default: // bool
		return !x.(bool) && b.(bool)
	}
}
