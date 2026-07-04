// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// RepeatedField is a typed, ordered list of protobuf values, mirroring
// Google::Protobuf::RepeatedField. It offers the gem's core Enumerable surface
// (Push/<<, [], []=, each, to_a, +, ==, clear). Values are type-checked against
// the field's element type on insertion (a wrong type raises TypeError).
type RepeatedField struct {
	list protoreflect.List
	fd   protoreflect.FieldDescriptor
	pool *DescriptorPool
}

// NewRepeatedField builds a standalone repeated field of scalar element type t,
// mirroring Google::Protobuf::RepeatedField.new(:int32, [1, 2, 3]). Message and
// enum element types are not supported standalone (obtain such a list from a
// message field).
func NewRepeatedField(t Symbol, initial ...any) (*RepeatedField, error) {
	fd, err := scratchList(t)
	if err != nil {
		return nil, err
	}
	rf := &RepeatedField{list: backingList(fd), fd: fd, pool: scratchPool}
	if err := rf.Push(initial...); err != nil {
		return nil, err
	}
	return rf, nil
}

// Length returns the number of elements, mirroring #length / #size.
func (r *RepeatedField) Length() int { return r.list.Len() }

// Push appends values, mirroring #push / #<<. It type-checks each value.
func (r *RepeatedField) Push(vals ...any) error {
	for _, v := range vals {
		pv, err := toProtoScalar(r.fd, v)
		if err != nil {
			return err
		}
		r.list.Append(pv)
	}
	return nil
}

// At returns the element at index i, mirroring rf[i]. Negative indices count
// from the end; an out-of-range index returns nil (as Ruby's Array#[] does).
func (r *RepeatedField) At(i int) any {
	n := r.list.Len()
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		return nil
	}
	return elemToRuby(r.fd, r.list.Get(i), r.pool)
}

// SetAt writes the element at index i, mirroring rf[i] = value. An out-of-range
// index is a RangeError (the gem raises IndexError). Negative indices count from
// the end.
func (r *RepeatedField) SetAt(i int, v any) error {
	n := r.list.Len()
	orig := i
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		return &RangeError{Message: "index " + itoa(orig) + " out of range"}
	}
	pv, err := toProtoScalar(r.fd, v)
	if err != nil {
		return err
	}
	r.list.Set(i, pv)
	return nil
}

// Each iterates the elements in order, mirroring #each.
func (r *RepeatedField) Each(fn func(any)) {
	for i := 0; i < r.list.Len(); i++ {
		fn(elemToRuby(r.fd, r.list.Get(i), r.pool))
	}
}

// ToArray returns the elements as a []any, mirroring #to_a.
func (r *RepeatedField) ToArray() []any {
	out := make([]any, r.list.Len())
	for i := range out {
		out[i] = elemToRuby(r.fd, r.list.Get(i), r.pool)
	}
	return out
}

// Concat appends every element of other (a *RepeatedField or []any), mirroring
// #concat / #+.
func (r *RepeatedField) Concat(other any) error {
	switch o := other.(type) {
	case *RepeatedField:
		return r.Push(o.ToArray()...)
	case []any:
		return r.Push(o...)
	default:
		return newTypeError("concat expects an Array or RepeatedField")
	}
}

// Clear removes all elements, mirroring #clear.
func (r *RepeatedField) Clear() { r.list.Truncate(0) }

// Equal reports whether r and other hold equal elements in the same order,
// mirroring #==.
func (r *RepeatedField) Equal(other *RepeatedField) bool {
	if other == nil || r.list.Len() != other.list.Len() {
		return false
	}
	for i := 0; i < r.list.Len(); i++ {
		if !valueEqual(r.fd, r.list.Get(i), other.list.Get(i)) {
			return false
		}
	}
	return true
}

// Dup returns a shallow copy with the same element type, mirroring #dup.
func (r *RepeatedField) Dup() *RepeatedField {
	dup := &RepeatedField{list: backingList(r.fd), fd: r.fd, pool: r.pool}
	for i := 0; i < r.list.Len(); i++ {
		dup.list.Append(r.list.Get(i))
	}
	return dup
}

// Inspect renders the list, mirroring #inspect: [a, b, c].
func (r *RepeatedField) Inspect() string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < r.list.Len(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(inspectValue(elemToRuby(r.fd, r.list.Get(i), r.pool)))
	}
	b.WriteByte(']')
	return b.String()
}

// valueEqual reports whether two protoreflect values of field fd are equal.
func valueEqual(fd protoreflect.FieldDescriptor, a, b protoreflect.Value) bool {
	if isMessageField(fd) {
		return protoMessageEqual(a.Message(), b.Message())
	}
	return a.Equal(b)
}
