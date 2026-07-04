// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"math"
	"testing"
)

func TestAsInt64(t *testing.T) {
	cases := []struct {
		in   any
		want int64
		ok   bool
	}{
		{int(1), 1, true},
		{int8(2), 2, true},
		{int16(3), 3, true},
		{int32(4), 4, true},
		{int64(5), 5, true},
		{uint32(6), 6, true},
		{uint64(7), 7, true},
		{uint64(math.MaxUint64), 0, false}, // overflows int64
		{"str", 0, false},                  // wrong type
	}
	for _, c := range cases {
		got, ok := asInt64(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("asInt64(%v[%T]) = %d,%v want %d,%v", c.in, c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestAsUint64(t *testing.T) {
	cases := []struct {
		in   any
		want uint64
		ok   bool
	}{
		{int(1), 1, true},
		{int(-1), 0, false},
		{int32(2), 2, true},
		{int32(-2), 0, false},
		{int64(3), 3, true},
		{int64(-3), 0, false},
		{uint32(4), 4, true},
		{uint64(5), 5, true},
		{"str", 0, false},
	}
	for _, c := range cases {
		got, ok := asUint64(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("asUint64(%v[%T]) = %d,%v want %d,%v", c.in, c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestAsFloat(t *testing.T) {
	cases := []struct {
		in   any
		want float64
		ok   bool
	}{
		{float64(1.5), 1.5, true},
		{float32(2.5), 2.5, true},
		{int(3), 3, true},
		{int64(4), 4, true},
		{"str", 0, false},
	}
	for _, c := range cases {
		got, ok := asFloat(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("asFloat(%v[%T]) = %g,%v want %g,%v", c.in, c.in, got, ok, c.want, c.ok)
		}
	}
}
