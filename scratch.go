// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// The scratch pool backs standalone RepeatedField / Map values (the gem's
// Google::Protobuf::RepeatedField.new(:int32) and Map.new(:string, :int32)),
// which have no owning message. Each requested element/key-value type shape is
// compiled once into a throwaway message with a single field, and instances draw
// a fresh backing List/Map from a new dynamic message of that shape.
var (
	scratchPool    = NewDescriptorPool()
	scratchMu      sync.Mutex
	scratchListFDs = map[Symbol]protoreflect.FieldDescriptor{}
	scratchMapFDs  = map[[2]Symbol]protoreflect.FieldDescriptor{}
	scratchSeq     int
)

// scratchList returns a repeated-field descriptor of scalar element type t,
// building it on first use. Message/enum element types are not supported
// standalone (they need a named type); obtain such a field from a message.
func scratchList(t Symbol) (protoreflect.FieldDescriptor, error) {
	scratchMu.Lock()
	defer scratchMu.Unlock()
	if fd, ok := scratchListFDs[t]; ok {
		return fd, nil
	}
	scratchSeq++
	name := fmt.Sprintf("ScratchList%d", scratchSeq)
	// Build validates the element type: an unknown, :message or :enum element
	// type fails here (a standalone list of those has no named type to point at).
	if err := scratchPool.Build(func(b *Builder) {
		b.AddMessage(name, func(mb *MessageBuilder) { mb.Repeated("v", t, 1) })
	}); err != nil {
		return nil, err
	}
	fd := scratchPool.Lookup(name).(*Descriptor).md.Fields().ByName("v")
	scratchListFDs[t] = fd
	return fd, nil
}

// scratchMap returns a map-field descriptor with scalar key type k and scalar
// value type v, building it on first use.
func scratchMap(k, v Symbol) (protoreflect.FieldDescriptor, error) {
	scratchMu.Lock()
	defer scratchMu.Unlock()
	key := [2]Symbol{k, v}
	if fd, ok := scratchMapFDs[key]; ok {
		return fd, nil
	}
	scratchSeq++
	name := fmt.Sprintf("ScratchMap%d", scratchSeq)
	// Build validates key/value types: an unknown key/value type, or a :message /
	// :enum value type (unsupported standalone), fails here.
	if err := scratchPool.Build(func(b *Builder) {
		b.AddMessage(name, func(mb *MessageBuilder) { mb.Map("v", k, v, 1) })
	}); err != nil {
		return nil, err
	}
	fd := scratchPool.Lookup(name).(*Descriptor).md.Fields().ByName("v")
	scratchMapFDs[key] = fd
	return fd, nil
}

// backingList returns a fresh, empty protoreflect.List for field fd, drawn from a
// new dynamic message of fd's containing type.
func backingList(fd protoreflect.FieldDescriptor) protoreflect.List {
	msg := dynamicpb.NewMessage(fd.ContainingMessage())
	return msg.Mutable(fd).List()
}

// backingMap returns a fresh, empty protoreflect.Map for field fd.
func backingMap(fd protoreflect.FieldDescriptor) protoreflect.Map {
	msg := dynamicpb.NewMessage(fd.ContainingMessage())
	return msg.Mutable(fd).Map()
}
