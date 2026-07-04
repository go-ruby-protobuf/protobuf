// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import "testing"

func TestErrorTaxonomy(t *testing.T) {
	cases := []struct {
		err   Error
		msg   string
		class string
	}{
		{&TypeError{Message: "t"}, "t", "Google::Protobuf::TypeError"},
		{&RangeError{Message: "r"}, "r", "RangeError"},
		{&ArgumentError{Message: "a"}, "a", "ArgumentError"},
		{&ParseError{Message: "p"}, "p", "Google::Protobuf::ParseError"},
	}
	for _, c := range cases {
		if c.err.Error() != c.msg {
			t.Errorf("Error() = %q, want %q", c.err.Error(), c.msg)
		}
		if c.err.RubyClass() != c.class {
			t.Errorf("RubyClass() = %q, want %q", c.err.RubyClass(), c.class)
		}
	}
}
