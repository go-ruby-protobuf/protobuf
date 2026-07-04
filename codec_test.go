// Copyright (c) the go-ruby-protobuf/protobuf authors
//
// SPDX-License-Identifier: BSD-3-Clause

package protobuf

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestBinaryRoundTrip(t *testing.T) {
	p := newTestPool(t)
	cls := p.LookupMsgclass("All")
	m := mustNew(t, cls, map[string]any{
		"i32": int64(-5), "u64": uint64(1 << 40), "st": "hello", "bl": true,
	})
	mustSet(t, m, "en", Symbol("GREEN"))
	_ = mustGet(t, m, "ri").(*RepeatedField).Push(int64(9), int64(8))
	_ = mustGet(t, m, "mi").(*Map).Set("k", int64(3))

	data, err := Encode(m)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Decode(cls, data)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Equal(back) {
		t.Fatalf("round trip mismatch:\n%s\n%s", m.Inspect(), back.Inspect())
	}
}

// TestWireCompatDynamicOracle proves our encode/decode is wire-compatible with
// the canonical google.golang.org/protobuf runtime, in both directions, using a
// dynamic message of the same descriptor as the oracle. (Byte-identical output
// is deliberately NOT asserted: the runtime does not guarantee a stable field
// order between independent Marshal calls, so the invariant is semantic.)
func TestWireCompatDynamicOracle(t *testing.T) {
	p := newTestPool(t)
	cls := p.LookupMsgclass("All")

	// Our encode -> canonical decode: the oracle reads back the exact values.
	m := mustNew(t, cls, map[string]any{"i32": int64(42), "st": "wire"})
	ours, err := Encode(m)
	if err != nil {
		t.Fatal(err)
	}
	oracle := dynamicpb.NewMessage(cls.md)
	if err := proto.Unmarshal(ours, oracle); err != nil {
		t.Fatalf("canonical runtime cannot decode our bytes: %v", err)
	}
	if oracle.Get(cls.md.Fields().ByName("i32")).Int() != 42 ||
		oracle.Get(cls.md.Fields().ByName("st")).String() != "wire" {
		t.Fatalf("canonical decode of our bytes wrong: %v", oracle)
	}

	// Canonical encode -> our decode: we read back the exact values.
	canon := dynamicpb.NewMessage(cls.md)
	canon.Set(cls.md.Fields().ByName("i32"), protoreflect.ValueOfInt32(42))
	canon.Set(cls.md.Fields().ByName("st"), protoreflect.ValueOfString("wire"))
	canonBytes, err := proto.Marshal(canon)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Decode(cls, canonBytes)
	if err != nil {
		t.Fatal(err)
	}
	if mustGet(t, back, "i32") != int64(42) || mustGet(t, back, "st") != "wire" {
		t.Fatalf("our decode of canonical bytes wrong: %s", back.Inspect())
	}
}

// TestWireCompatGeneratedWKT proves wire-compatibility against REAL generated Go
// types: our Timestamp/Duration/StringValue encodings decode into the canonical
// generated structs and vice-versa.
func TestWireCompatGeneratedWKT(t *testing.T) {
	ts := mustNew(t, WellKnownType("Timestamp"), map[string]any{
		"seconds": int64(1600000000), "nanos": int64(123),
	})
	data, err := Encode(ts)
	if err != nil {
		t.Fatal(err)
	}
	var gen timestamppb.Timestamp
	if err := proto.Unmarshal(data, &gen); err != nil {
		t.Fatal(err)
	}
	if gen.Seconds != 1600000000 || gen.Nanos != 123 {
		t.Fatalf("Timestamp decoded wrong: %+v", &gen)
	}
	// Reverse direction: canonical bytes -> our message.
	genBytes, _ := proto.Marshal(&timestamppb.Timestamp{Seconds: 7, Nanos: 8})
	back, err := Decode(WellKnownType("Timestamp"), genBytes)
	if err != nil {
		t.Fatal(err)
	}
	if mustGet(t, back, "seconds") != int64(7) || mustGet(t, back, "nanos") != int64(8) {
		t.Fatal("reverse Timestamp mismatch")
	}

	// Duration.
	dur := mustNew(t, WellKnownType("Duration"), map[string]any{"seconds": int64(90)})
	db, _ := Encode(dur)
	var gd durationpb.Duration
	if err := proto.Unmarshal(db, &gd); err != nil || gd.Seconds != 90 {
		t.Fatalf("Duration mismatch: %+v %v", &gd, err)
	}

	// StringValue wrapper.
	sv := mustNew(t, WellKnownType("StringValue"), map[string]any{"value": "wrapped"})
	sb, _ := Encode(sv)
	var gsv wrapperspb.StringValue
	if err := proto.Unmarshal(sb, &gsv); err != nil || gsv.Value != "wrapped" {
		t.Fatalf("StringValue mismatch: %+v %v", &gsv, err)
	}
}

func TestJSONRoundTripAndOptions(t *testing.T) {
	p := newTestPool(t)
	cls := p.LookupMsgclass("All")
	m := mustNew(t, cls, map[string]any{"i32": int64(3), "st": "j"})
	mustSet(t, m, "en", Symbol("BLUE"))

	js, err := EncodeJSON(m)
	if err != nil {
		t.Fatal(err)
	}
	// Semantic equality against the protojson oracle (protojson jitters
	// whitespace, so compare via a canonical re-marshal of both).
	oracle := dynamicpb.NewMessage(cls.md)
	if err := protojson.Unmarshal(js, oracle); err != nil {
		t.Fatalf("oracle cannot parse our JSON: %v", err)
	}
	if !proto.Equal(m.m.Interface(), oracle.Interface()) {
		t.Fatal("JSON not semantically equal to oracle")
	}

	back, err := DecodeJSON(cls, js)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Equal(back) {
		t.Fatal("JSON round trip mismatch")
	}

	// emit_defaults includes zero fields; default omits them.
	full, _ := EncodeJSON(m, JSONOptions{EmitDefaults: true})
	if !bytes.Contains(full, []byte("i64")) {
		t.Fatal("EmitDefaults should include zero fields")
	}
	if bytes.Contains(js, []byte("i64")) {
		t.Fatal("default JSON should omit zero fields")
	}

	// preserve_proto_fieldnames.
	proto3names, _ := EncodeJSON(m, JSONOptions{PreserveProtoNames: true})
	if !bytes.Contains(proto3names, []byte(`"i32"`)) {
		t.Fatalf("PreserveProtoNames: %s", proto3names)
	}

	// ignore_unknown_fields on decode.
	if _, err := DecodeJSON(cls, []byte(`{"nope":1,"i32":2}`)); err == nil {
		t.Fatal("unknown field should error by default")
	}
	if _, err := DecodeJSON(cls, []byte(`{"nope":1,"i32":2}`), JSONOptions{IgnoreUnknownFields: true}); err != nil {
		t.Fatalf("ignore_unknown_fields: %v", err)
	}
}

func TestCodecErrors(t *testing.T) {
	p := newTestPool(t)
	cls := p.LookupMsgclass("All")

	// Decode malformed bytes.
	if _, err := Decode(cls, []byte{0xff, 0xff, 0xff}); err == nil {
		t.Fatal("Decode should reject malformed bytes")
	} else if _, ok := err.(*ParseError); !ok {
		t.Fatalf("want *ParseError, got %T", err)
	}

	// DecodeJSON malformed.
	if _, err := DecodeJSON(cls, []byte("{")); err == nil {
		t.Fatal("DecodeJSON should reject malformed json")
	}

	// Encode / EncodeJSON of an invalid-UTF-8 proto3 string errors.
	m := mustNew(t, cls)
	mustSet(t, m, "st", "\xff\xfe")
	if _, err := Encode(m); err == nil {
		t.Fatal("Encode should reject invalid UTF-8")
	}
	if _, err := EncodeJSON(m); err == nil {
		t.Fatal("EncodeJSON should reject invalid UTF-8")
	}
}
