// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "google.golang.org/protobuf/reflect/protoreflect"

// Descriptor wraps a message descriptor, mirroring
// Google::Protobuf::Descriptor. It yields the message class (Msgclass) and its
// fields (Each / Lookup).
type Descriptor struct {
	md   protoreflect.MessageDescriptor
	pool *DescriptorPool
}

// Name returns the message's fully-qualified name, mirroring descriptor.name.
func (d *Descriptor) Name() string { return string(d.md.FullName()) }

// Msgclass returns the message class for this descriptor, mirroring
// descriptor.msgclass.
func (d *Descriptor) Msgclass() *MessageClass {
	return &MessageClass{md: d.md, pool: d.pool}
}

// Lookup returns the field named name, or nil, mirroring descriptor.lookup(name).
func (d *Descriptor) Lookup(name string) *FieldDescriptor {
	fd := d.md.Fields().ByName(protoreflect.Name(name))
	if fd == nil {
		return nil
	}
	return &FieldDescriptor{fd: fd}
}

// Each iterates the message's fields in declaration order, mirroring
// descriptor.each { |field| … }.
func (d *Descriptor) Each(fn func(*FieldDescriptor)) {
	fs := d.md.Fields()
	for i := 0; i < fs.Len(); i++ {
		fn(&FieldDescriptor{fd: fs.Get(i)})
	}
}

// FieldDescriptor wraps a field descriptor, mirroring
// Google::Protobuf::FieldDescriptor.
type FieldDescriptor struct {
	fd protoreflect.FieldDescriptor
}

// Name returns the field name, mirroring field.name.
func (f *FieldDescriptor) Name() string { return string(f.fd.Name()) }

// Number returns the field's tag number, mirroring field.number.
func (f *FieldDescriptor) Number() int { return int(f.fd.Number()) }

// Type returns the field's type as a Symbol (:int32, :string, :message, …),
// mirroring field.type.
func (f *FieldDescriptor) Type() Symbol {
	switch f.fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return MessageType
	case protoreflect.EnumKind:
		return Enum
	default:
		return Symbol(f.fd.Kind().String())
	}
}

// Label returns the field's label as a Symbol (:optional or :repeated),
// mirroring field.label.
func (f *FieldDescriptor) Label() Symbol {
	if f.fd.Cardinality() == protoreflect.Repeated {
		return "repeated"
	}
	return "optional"
}

// EnumDescriptor wraps an enum descriptor, mirroring
// Google::Protobuf::EnumDescriptor.
type EnumDescriptor struct {
	ed protoreflect.EnumDescriptor
}

// Name returns the enum's fully-qualified name, mirroring enum.name.
func (e *EnumDescriptor) Name() string { return string(e.ed.FullName()) }

// LookupName returns the number for the value named name, mirroring
// enum.lookup_name(name). The bool reports whether name is a defined value.
func (e *EnumDescriptor) LookupName(name string) (int, bool) {
	v := e.ed.Values().ByName(protoreflect.Name(name))
	if v == nil {
		return 0, false
	}
	return int(v.Number()), true
}

// LookupValue returns the Symbol name for number, mirroring
// enum.lookup_value(number). The bool reports whether number is defined.
func (e *EnumDescriptor) LookupValue(number int) (Symbol, bool) {
	v := e.ed.Values().ByNumber(protoreflect.EnumNumber(number))
	if v == nil {
		return "", false
	}
	return Symbol(v.Name()), true
}
