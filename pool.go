// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// DescriptorPool is a registry of protobuf descriptors, mirroring the gem's
// Google::Protobuf::DescriptorPool. Build compiles a batch of message/enum
// definitions into it (the pool.build DSL) and Lookup resolves a fully-qualified
// name to its *Descriptor / *EnumDescriptor.
type DescriptorPool struct {
	mu    sync.Mutex
	files *protoregistry.Files
	seq   int // monotonic counter for synthesised file names
}

// NewDescriptorPool returns an empty pool pre-loaded with the well-known-type
// files (google.protobuf.Timestamp, Duration, Any, Struct, Value, ListValue,
// FieldMask, the scalar wrappers and Empty), matching the gem: they are always
// resolvable in a fresh pool.
func NewDescriptorPool() *DescriptorPool {
	p := &DescriptorPool{files: new(protoregistry.Files)}
	registerWellKnownTypes(p.files)
	return p
}

var (
	genOnce sync.Once
	genPool *DescriptorPool
)

// GeneratedPool returns the process-wide generated pool, mirroring
// Google::Protobuf::DescriptorPool.generated_pool: the single pool every
// generated message class registers into.
func GeneratedPool() *DescriptorPool {
	genOnce.Do(func() { genPool = NewDescriptorPool() })
	return genPool
}

// Build compiles one batch of definitions into the pool, mirroring
//
//	pool.build do
//	  add_message "Foo" do … end
//	  add_enum "Bar" do … end
//	end
//
// Every message and enum added by fn is placed in a single synthesised proto3
// file so they may reference one another freely. It returns an error (never
// nil-swallowed) if the batch is malformed or fails to compile.
func (p *DescriptorPool) Build(fn func(*Builder)) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.seq++
	b := &Builder{
		pool: p,
		file: &descriptorpb.FileDescriptorProto{
			Name:   proto3Name(p.seq),
			Syntax: strPtr("proto3"),
		},
	}
	fn(b)
	if b.err != nil {
		return b.err
	}

	// Import every file already in the pool so references to prior builds and to
	// the well-known types resolve. Unused imports are harmless.
	p.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		b.file.Dependency = append(b.file.Dependency, fd.Path())
		return true
	})

	fd, err := protodesc.NewFile(b.file, p.files)
	if err != nil {
		return &ArgumentError{Message: err.Error()}
	}
	if err := p.files.RegisterFile(fd); err != nil {
		return &ArgumentError{Message: err.Error()}
	}
	return nil
}

// Lookup resolves a fully-qualified name to its descriptor, mirroring
// pool.lookup(name). It returns a *Descriptor for a message, a *EnumDescriptor
// for an enum, or nil when the name is unknown.
func (p *DescriptorPool) Lookup(name string) any {
	p.mu.Lock()
	defer p.mu.Unlock()
	d, err := p.files.FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		return nil
	}
	switch desc := d.(type) {
	case protoreflect.MessageDescriptor:
		return &Descriptor{md: desc, pool: p}
	case protoreflect.EnumDescriptor:
		return &EnumDescriptor{ed: desc}
	default:
		return nil
	}
}

// LookupMsgclass is a convenience for the common case: it resolves name to a
// message and returns its class, or nil if name is not a message. It is
// equivalent to pool.lookup(name).msgclass in Ruby.
func (p *DescriptorPool) LookupMsgclass(name string) *MessageClass {
	if d, ok := p.Lookup(name).(*Descriptor); ok {
		return d.Msgclass()
	}
	return nil
}

// proto3Name returns the synthesised path for the seq-th built file.
func proto3Name(seq int) *string {
	s := fmt.Sprintf("go-ruby-protobuf/build-%d.proto", seq)
	return &s
}

func strPtr(s string) *string { return &s }
