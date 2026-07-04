// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

// Error is the base of this package's error taxonomy. Every error this package
// returns satisfies Error and reports, via RubyClass, the Ruby exception class a
// host (go-embedded-ruby) should raise. It mirrors the exceptions the
// google-protobuf gem raises.
type Error interface {
	error
	// RubyClass is the fully-qualified Ruby exception class this error maps to
	// (e.g. "Google::Protobuf::TypeError").
	RubyClass() string
}

// TypeError mirrors the gem's Google::Protobuf::TypeError: a value handed to a
// field setter (or to a repeated/map container) whose Ruby type does not match
// the field's protobuf type. In MRI Google::Protobuf::TypeError subclasses the
// core ::TypeError.
type TypeError struct{ Message string }

func (e *TypeError) Error() string     { return e.Message }
func (e *TypeError) RubyClass() string { return "Google::Protobuf::TypeError" }

// RangeError mirrors MRI raising ::RangeError when an integer value does not fit
// the target field's integer type (e.g. 1<<40 into an int32 field), or when an
// enum number is out of the valid range.
type RangeError struct{ Message string }

func (e *RangeError) Error() string     { return e.Message }
func (e *RangeError) RubyClass() string { return "RangeError" }

// ArgumentError mirrors MRI raising ::ArgumentError — used here for an unknown
// field name, an unknown enum symbol, or a malformed builder specification.
type ArgumentError struct{ Message string }

func (e *ArgumentError) Error() string     { return e.Message }
func (e *ArgumentError) RubyClass() string { return "ArgumentError" }

// ParseError mirrors the gem's Google::Protobuf::ParseError, raised by
// Google::Protobuf.decode / decode_json on malformed input.
type ParseError struct{ Message string }

func (e *ParseError) Error() string     { return e.Message }
func (e *ParseError) RubyClass() string { return "Google::Protobuf::ParseError" }

// newTypeError builds a *TypeError with the given message.
func newTypeError(msg string) *TypeError { return &TypeError{Message: msg} }
