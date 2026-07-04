// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// MessageClass is a generated message class, mirroring the anonymous Class the
// gem attaches to a Descriptor (descriptor.msgclass). New builds instances.
type MessageClass struct {
	md   protoreflect.MessageDescriptor
	pool *DescriptorPool
}

// Name returns the message's fully-qualified name.
func (c *MessageClass) Name() string { return string(c.md.FullName()) }

// Descriptor returns the class's *Descriptor, mirroring klass.descriptor.
func (c *MessageClass) Descriptor() *Descriptor {
	return &Descriptor{md: c.md, pool: c.pool}
}

// New builds a new message, optionally initialised from a field=>value hash,
// mirroring MyMsg.new(field: value, …). It returns an error (the gem raises) if
// an init key is not a field or a value has the wrong type.
func (c *MessageClass) New(init ...map[string]any) (*Message, error) {
	m := &Message{m: dynamicpb.NewMessage(c.md), pool: c.pool}
	for _, h := range init {
		// Deterministic order so a wrong-value error is reproducible.
		keys := make([]string, 0, len(h))
		for k := range h {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if err := m.Set(k, h[k]); err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

// Message is a message instance, mirroring an instance of a generated message
// class. It wraps a dynamicpb dynamic message so binary/JSON encoding go through
// the canonical runtime.
type Message struct {
	m    protoreflect.Message
	pool *DescriptorPool
}

// Class returns the message's class.
func (m *Message) Class() *MessageClass {
	return &MessageClass{md: m.m.Descriptor(), pool: m.pool}
}

// field resolves name to its descriptor or an ArgumentError (unknown field).
func (m *Message) field(name string) (protoreflect.FieldDescriptor, error) {
	fd := m.m.Descriptor().Fields().ByName(protoreflect.Name(name))
	if fd == nil {
		return nil, &ArgumentError{Message: fmt.Sprintf("unknown field %q for %s", name, m.m.Descriptor().FullName())}
	}
	return fd, nil
}

// Get reads field name, mirroring the generated reader msg.name. Repeated and
// map fields return their live *RepeatedField / *Map; an unset message field
// returns nil; scalars return their Ruby value (proto3 default when unset).
func (m *Message) Get(name string) (any, error) {
	fd, err := m.field(name)
	if err != nil {
		return nil, err
	}
	switch {
	case fd.IsMap():
		return &Map{m: m.m.Mutable(fd).Map(), fd: fd, pool: m.pool}, nil
	case fd.IsList():
		return &RepeatedField{list: m.m.Mutable(fd).List(), fd: fd, pool: m.pool}, nil
	case isMessageField(fd):
		if !m.m.Has(fd) {
			return nil, nil
		}
		return &Message{m: m.m.Get(fd).Message(), pool: m.pool}, nil
	default:
		return fromProtoScalar(fd, m.m.Get(fd), m.pool), nil
	}
}

// Set writes field name, mirroring the generated writer msg.name = value.
// Assigning nil to a message field clears it; assigning an Array/[]any to a
// repeated field or a Hash/map to a map field replaces its contents.
func (m *Message) Set(name string, v any) error {
	fd, err := m.field(name)
	if err != nil {
		return err
	}
	switch {
	case fd.IsMap():
		return m.setMap(fd, v)
	case fd.IsList():
		return m.setList(fd, v)
	case isMessageField(fd):
		if v == nil {
			m.m.Clear(fd)
			return nil
		}
		pv, err := toProtoScalar(fd, v)
		if err != nil {
			return err
		}
		m.m.Set(fd, pv)
		return nil
	default:
		pv, err := toProtoScalar(fd, v)
		if err != nil {
			return err
		}
		m.m.Set(fd, pv)
		return nil
	}
}

// setList replaces a repeated field's contents from a *RepeatedField or []any.
func (m *Message) setList(fd protoreflect.FieldDescriptor, v any) error {
	var elems []any
	switch src := v.(type) {
	case *RepeatedField:
		elems = src.ToArray()
	case []any:
		elems = src
	default:
		return newTypeError(fmt.Sprintf("field %q expects an Array, got %T", fd.Name(), v))
	}
	list := m.m.Mutable(fd).List()
	for list.Len() > 0 {
		list.Truncate(list.Len() - 1)
	}
	for _, e := range elems {
		pv, err := toProtoScalar(fd, e)
		if err != nil {
			return err
		}
		list.Append(pv)
	}
	return nil
}

// setMap replaces a map field's contents from a *Map or map[string]any.
func (m *Message) setMap(fd protoreflect.FieldDescriptor, v any) error {
	pairs := map[any]any{}
	switch src := v.(type) {
	case *Map:
		for _, k := range src.Keys() {
			val, _ := src.Get(k)
			pairs[k] = val
		}
	case map[string]any:
		for k, val := range src {
			pairs[k] = val
		}
	case map[any]any:
		pairs = src
	default:
		return newTypeError(fmt.Sprintf("field %q expects a Hash, got %T", fd.Name(), v))
	}
	pm := m.m.Mutable(fd).Map()
	var clearKeys []protoreflect.MapKey
	pm.Range(func(k protoreflect.MapKey, _ protoreflect.Value) bool {
		clearKeys = append(clearKeys, k)
		return true
	})
	for _, k := range clearKeys {
		pm.Clear(k)
	}
	kfd, vfd := fd.MapKey(), fd.MapValue()
	for k, val := range pairs {
		kv, err := toProtoScalar(kfd, k)
		if err != nil {
			return err
		}
		vv, err := toProtoScalar(vfd, val)
		if err != nil {
			return err
		}
		pm.Set(kv.MapKey(), vv)
	}
	return nil
}

// ToH returns the message as a Ruby Hash (map keyed by field name), mirroring
// msg.to_h. Repeated fields become []any, maps become map[any]any, sub-messages
// become nested hashes (nil when unset).
func (m *Message) ToH() map[string]any {
	h := map[string]any{}
	fs := m.m.Descriptor().Fields()
	for i := 0; i < fs.Len(); i++ {
		fd := fs.Get(i)
		name := string(fd.Name())
		switch {
		case fd.IsMap():
			mp := map[any]any{}
			m.m.Get(fd).Map().Range(func(k protoreflect.MapKey, val protoreflect.Value) bool {
				mp[fromProtoScalar(fd.MapKey(), k.Value(), m.pool)] =
					fromProtoScalar(fd.MapValue(), val, m.pool)
				return true
			})
			h[name] = mp
		case fd.IsList():
			list := m.m.Get(fd).List()
			arr := make([]any, list.Len())
			for j := 0; j < list.Len(); j++ {
				arr[j] = elemToRuby(fd, list.Get(j), m.pool)
			}
			h[name] = arr
		case isMessageField(fd):
			if !m.m.Has(fd) {
				h[name] = nil
			} else {
				h[name] = (&Message{m: m.m.Get(fd).Message(), pool: m.pool}).ToH()
			}
		default:
			h[name] = fromProtoScalar(fd, m.m.Get(fd), m.pool)
		}
	}
	return h
}

// Equal reports whether m and other carry the same message contents, mirroring
// msg == other. Messages of different types are never equal.
func (m *Message) Equal(other *Message) bool {
	if other == nil {
		return false
	}
	return proto.Equal(m.m.Interface(), other.m.Interface())
}

// Dup returns a deep copy of the message, mirroring msg.dup.
func (m *Message) Dup() *Message {
	return &Message{m: proto.Clone(m.m.Interface()).ProtoReflect(), pool: m.pool}
}

// Clone returns a deep copy of the message, mirroring msg.clone. For a protobuf
// message dup and clone are identical (no singleton/frozen state to preserve).
func (m *Message) Clone() *Message { return m.Dup() }

// Inspect returns a debug string, mirroring msg.inspect:
// <Name: field: value, …>.
func (m *Message) Inspect() string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(string(m.m.Descriptor().FullName()))
	b.WriteString(": ")
	fs := m.m.Descriptor().Fields()
	for i := 0; i < fs.Len(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fd := fs.Get(i)
		b.WriteString(string(fd.Name()))
		b.WriteString(": ")
		v, _ := m.Get(string(fd.Name()))
		b.WriteString(inspectValue(v))
	}
	b.WriteByte('>')
	return b.String()
}

// inspectValue renders a single Ruby value for Inspect / RepeatedField / Map.
func inspectValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case string:
		return fmt.Sprintf("%q", x)
	case []byte:
		return fmt.Sprintf("%q", string(x))
	case Symbol:
		return ":" + string(x)
	case *Message:
		return x.Inspect()
	case *RepeatedField:
		return x.Inspect()
	case *Map:
		return x.Inspect()
	default:
		return fmt.Sprintf("%v", x)
	}
}

// elemToRuby converts one repeated-field element to its Ruby value.
func elemToRuby(fd protoreflect.FieldDescriptor, v protoreflect.Value, pool *DescriptorPool) any {
	if isMessageField(fd) {
		return &Message{m: v.Message(), pool: pool}
	}
	return fromProtoScalar(fd, v, pool)
}

// isMessageField reports whether fd is a (non-map, non-list-of-scalar) message
// field — including a repeated message field's element kind.
func isMessageField(fd protoreflect.FieldDescriptor) bool {
	k := fd.Kind()
	return k == protoreflect.MessageKind || k == protoreflect.GroupKind
}
