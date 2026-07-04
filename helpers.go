// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// itoa formats an int (small internal helper for index messages).
func itoa(i int) string { return strconv.Itoa(i) }

// protoMessageEqual compares two protoreflect messages for value equality.
func protoMessageEqual(a, b protoreflect.Message) bool {
	return proto.Equal(a.Interface(), b.Interface())
}
