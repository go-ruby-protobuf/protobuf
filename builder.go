// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// Builder is the receiver of a Build block, mirroring the object the gem yields
// to pool.build: add_message / add_enum.
type Builder struct {
	pool *DescriptorPool
	file *descriptorpb.FileDescriptorProto
	err  error
}

// fail records the first error seen during a build; later calls become no-ops.
func (b *Builder) fail(msg string) {
	if b.err == nil {
		b.err = &ArgumentError{Message: msg}
	}
}

// AddMessage defines a message named name, mirroring add_message. fn receives a
// *MessageBuilder describing its fields.
func (b *Builder) AddMessage(name string, fn func(*MessageBuilder)) {
	msg := &descriptorpb.DescriptorProto{Name: strPtr(name)}
	mb := &MessageBuilder{b: b, msg: msg}
	fn(mb)
	b.file.MessageType = append(b.file.MessageType, msg)
}

// AddEnum defines an enum named name, mirroring add_enum. fn receives a
// *EnumBuilder describing its values.
func (b *Builder) AddEnum(name string, fn func(*EnumBuilder)) {
	en := &descriptorpb.EnumDescriptorProto{Name: strPtr(name)}
	eb := &EnumBuilder{b: b, enum: en}
	fn(eb)
	b.file.EnumType = append(b.file.EnumType, en)
}

// MessageBuilder describes a message's fields, mirroring the block add_message
// yields (optional / repeated / map / oneof).
type MessageBuilder struct {
	b   *Builder
	msg *descriptorpb.DescriptorProto
}

const (
	labelOptional = descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	labelRepeated = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
)

// Optional adds a singular field, mirroring `optional :name, :type, number
// [, "TypeName"]`. For a :message or :enum field the referenced fully-qualified
// type name is required as typeName.
func (mb *MessageBuilder) Optional(name string, typ Symbol, number int, typeName ...string) {
	mb.addField(name, typ, number, labelOptional, nil, typeName...)
}

// Repeated adds a repeated field, mirroring `repeated :name, :type, number
// [, "TypeName"]`.
func (mb *MessageBuilder) Repeated(name string, typ Symbol, number int, typeName ...string) {
	mb.addField(name, typ, number, labelRepeated, nil, typeName...)
}

// Map adds a map field, mirroring `map :name, :key_type, :value_type, number
// [, "ValueTypeName"]`. It synthesises the map-entry message the protobuf format
// requires.
func (mb *MessageBuilder) Map(name string, keyType, valType Symbol, number int, valTypeName ...string) {
	if _, ok := fieldTypes[keyType]; !ok {
		mb.b.fail("unknown map key type :" + string(keyType))
		return
	}
	entryName := camel(name) + "Entry"
	entry := &descriptorpb.DescriptorProto{
		Name:    strPtr(entryName),
		Options: &descriptorpb.MessageOptions{MapEntry: boolPtr(true)},
	}
	kf, err := makeField("key", keyType, 1, labelOptional, nil)
	if err != nil {
		mb.b.fail(err.Error())
		return
	}
	vf, err := makeField("value", valType, 2, labelOptional, nil, valTypeName...)
	if err != nil {
		mb.b.fail(err.Error())
		return
	}
	entry.Field = []*descriptorpb.FieldDescriptorProto{kf, vf}
	mb.msg.NestedType = append(mb.msg.NestedType, entry)

	// The map field itself: a repeated field of the synthesised entry message.
	fullEntry := "." + mb.msg.GetName() + "." + entryName
	f, err := makeField(name, MessageType, number, labelRepeated, nil, strings.TrimPrefix(fullEntry, "."))
	if err != nil {
		mb.b.fail(err.Error())
		return
	}
	mb.msg.Field = append(mb.msg.Field, f)
}

// Oneof adds a oneof group named name, mirroring `oneof :name do … end`. The
// fields declared in fn become members of the oneof.
func (mb *MessageBuilder) Oneof(name string, fn func(*OneofBuilder)) {
	idx := int32(len(mb.msg.OneofDecl))
	mb.msg.OneofDecl = append(mb.msg.OneofDecl,
		&descriptorpb.OneofDescriptorProto{Name: strPtr(name)})
	fn(&OneofBuilder{mb: mb, index: idx})
}

// addField builds a field descriptor and appends it, recording any error.
func (mb *MessageBuilder) addField(name string, typ Symbol, number int, label descriptorpb.FieldDescriptorProto_Label, oneof *int32, typeName ...string) {
	f, err := makeField(name, typ, number, label, oneof, typeName...)
	if err != nil {
		mb.b.fail(err.Error())
		return
	}
	mb.msg.Field = append(mb.msg.Field, f)
}

// OneofBuilder describes the fields inside a oneof.
type OneofBuilder struct {
	mb    *MessageBuilder
	index int32
}

// Optional adds a field to the oneof, mirroring `optional :name, :type, number`
// inside a oneof block.
func (ob *OneofBuilder) Optional(name string, typ Symbol, number int, typeName ...string) {
	idx := ob.index
	ob.mb.addField(name, typ, number, labelOptional, &idx, typeName...)
}

// EnumBuilder describes an enum's values, mirroring the add_enum block.
type EnumBuilder struct {
	b    *Builder
	enum *descriptorpb.EnumDescriptorProto
}

// Value adds an enum value, mirroring `value :NAME, number`.
func (eb *EnumBuilder) Value(name Symbol, number int) {
	eb.enum.Value = append(eb.enum.Value, &descriptorpb.EnumValueDescriptorProto{
		Name:   strPtr(string(name)),
		Number: int32Ptr(int32(number)),
	})
}

// makeField builds a single FieldDescriptorProto from DSL arguments.
func makeField(name string, typ Symbol, number int, label descriptorpb.FieldDescriptorProto_Label, oneof *int32, typeName ...string) (*descriptorpb.FieldDescriptorProto, error) {
	t, ok := fieldTypes[typ]
	if !ok {
		return nil, &ArgumentError{Message: "unknown field type :" + string(typ)}
	}
	f := &descriptorpb.FieldDescriptorProto{
		Name:       strPtr(name),
		Number:     int32Ptr(int32(number)),
		Label:      label.Enum(),
		Type:       t.Enum(),
		OneofIndex: oneof,
	}
	if typ == MessageType || typ == Enum {
		if len(typeName) == 0 || typeName[0] == "" {
			return nil, &ArgumentError{Message: "field :" + name + " (:" + string(typ) + ") requires a type name"}
		}
		f.TypeName = strPtr("." + strings.TrimPrefix(typeName[0], "."))
	}
	return f, nil
}

// camel converts a snake_case field name to UpperCamelCase for a map-entry
// message name (protoc's convention: "foo_bar" -> "FooBarEntry").
func camel(s string) string {
	var b strings.Builder
	up := true
	for _, r := range s {
		if r == '_' {
			up = true
			continue
		}
		if up && r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		up = false
		b.WriteRune(r)
	}
	return b.String()
}

func boolPtr(b bool) *bool    { return &b }
func int32Ptr(i int32) *int32 { return &i }
