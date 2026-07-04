// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"fmt"
	"math"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// toProtoScalar converts a Ruby value v to the protoreflect.Value a scalar (or
// message) field fd expects, applying the gem's type checks (TypeError on a
// wrong type, RangeError on integer overflow / unknown enum). It is not used for
// list/map fields, which the container types handle element-wise via this same
// function on the element descriptor.
func toProtoScalar(fd protoreflect.FieldDescriptor, v any) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		b, ok := v.(bool)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "true/false")
		}
		return protoreflect.ValueOfBool(b), nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		n, ok := asInt64(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Integer")
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return protoreflect.Value{}, rangeErr(n, fd)
		}
		return protoreflect.ValueOfInt32(int32(n)), nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		n, ok := asInt64(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Integer")
		}
		return protoreflect.ValueOfInt64(n), nil

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		n, ok := asUint64(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Integer")
		}
		if n > math.MaxUint32 {
			return protoreflect.Value{}, rangeErr(n, fd)
		}
		return protoreflect.ValueOfUint32(uint32(n)), nil

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		n, ok := asUint64(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Integer")
		}
		return protoreflect.ValueOfUint64(n), nil

	case protoreflect.FloatKind:
		f, ok := asFloat(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Float")
		}
		return protoreflect.ValueOfFloat32(float32(f)), nil

	case protoreflect.DoubleKind:
		f, ok := asFloat(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Float")
		}
		return protoreflect.ValueOfFloat64(f), nil

	case protoreflect.StringKind:
		s, ok := v.(string)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "String")
		}
		return protoreflect.ValueOfString(s), nil

	case protoreflect.BytesKind:
		switch b := v.(type) {
		case []byte:
			return protoreflect.ValueOfBytes(b), nil
		case string:
			return protoreflect.ValueOfBytes([]byte(b)), nil
		default:
			return protoreflect.Value{}, typeMismatch(fd, v, "String")
		}

	case protoreflect.EnumKind:
		return enumValue(fd, v)

	default:
		// MessageKind / GroupKind — every scalar Kind is handled above, so a
		// field reaching the default is a message field.
		m, ok := v.(*Message)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, string(fd.Message().FullName()))
		}
		if m.m.Descriptor().FullName() != fd.Message().FullName() {
			return protoreflect.Value{}, newTypeError(fmt.Sprintf(
				"expected message %s, got %s", fd.Message().FullName(), m.m.Descriptor().FullName()))
		}
		return protoreflect.ValueOfMessage(m.m), nil
	}
}

// enumValue converts a Ruby enum assignment (Symbol name, String name or
// Integer number) to its protoreflect enum Value.
func enumValue(fd protoreflect.FieldDescriptor, v any) (protoreflect.Value, error) {
	ed := fd.Enum()
	switch x := v.(type) {
	case Symbol:
		ev := ed.Values().ByName(protoreflect.Name(string(x)))
		if ev == nil {
			return protoreflect.Value{}, &RangeError{Message: fmt.Sprintf("unknown enum value :%s", x)}
		}
		return protoreflect.ValueOfEnum(ev.Number()), nil
	case string:
		ev := ed.Values().ByName(protoreflect.Name(x))
		if ev == nil {
			return protoreflect.Value{}, &RangeError{Message: fmt.Sprintf("unknown enum value :%s", x)}
		}
		return protoreflect.ValueOfEnum(ev.Number()), nil
	default:
		n, ok := asInt64(v)
		if !ok {
			return protoreflect.Value{}, typeMismatch(fd, v, "Symbol or Integer")
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return protoreflect.Value{}, rangeErr(n, fd)
		}
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(n)), nil
	}
}

// fromProtoScalar converts a protoreflect scalar/message Value read from field
// fd back to its Ruby representation.
func fromProtoScalar(fd protoreflect.FieldDescriptor, v protoreflect.Value, pool *DescriptorPool) any {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return v.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return v.Uint()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return v.Float()
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.BytesKind:
		return v.Bytes()
	case protoreflect.EnumKind:
		num := v.Enum()
		if ev := fd.Enum().Values().ByNumber(num); ev != nil {
			return Symbol(ev.Name())
		}
		return int64(num)
	default:
		// MessageKind / GroupKind.
		return &Message{m: v.Message(), pool: pool}
	}
}

// asInt64 accepts the Go integer types a host uses for a Ruby Integer.
func asInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint32:
		return int64(n), true
	case uint64:
		if n > math.MaxInt64 {
			return 0, false
		}
		return int64(n), true
	default:
		return 0, false
	}
}

// asUint64 accepts a non-negative Ruby Integer for an unsigned field.
func asUint64(v any) (uint64, bool) {
	switch n := v.(type) {
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int32:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	default:
		return 0, false
	}
}

// asFloat accepts a Ruby Float, or an Integer the gem coerces to Float.
func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// typeMismatch builds the gem-style TypeError for a wrong value type.
func typeMismatch(fd protoreflect.FieldDescriptor, v any, want string) *TypeError {
	return newTypeError(fmt.Sprintf("field %q expects %s, got %T", fd.Name(), want, v))
}

// rangeErr builds the gem-style RangeError for an out-of-range integer/enum.
func rangeErr(n any, fd protoreflect.FieldDescriptor) *RangeError {
	return &RangeError{Message: fmt.Sprintf("%v is out of range for field %q", n, fd.Name())}
}
