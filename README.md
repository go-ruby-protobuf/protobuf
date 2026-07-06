<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-protobuf/brand/main/social/go-ruby-protobuf-protobuf.png" alt="go-ruby-protobuf/protobuf" width="720"></p>

# protobuf — go-ruby-protobuf

[![CI](https://github.com/go-ruby-protobuf/protobuf/actions/workflows/ci.yml/badge.svg)](https://github.com/go-ruby-protobuf/protobuf/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-ruby-protobuf/protobuf.svg)](https://pkg.go.dev/github.com/go-ruby-protobuf/protobuf)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)

**A pure-Go (no cgo) reimplementation of the runtime and builder surface of
Ruby's [`google-protobuf`](https://rubygems.org/gems/google-protobuf) gem** — the
object model a Ruby program sees as the `Google::Protobuf` namespace — with **no
Ruby runtime and no C extension**. Upstream `google-protobuf` ships as a C
extension around libprotobuf/upb; this package offers the same API in ordinary
Go.

It does **not** reimplement the protobuf wire format. It is a Ruby-faithful API
layer built on top of [`google.golang.org/protobuf`](https://pkg.go.dev/google.golang.org/protobuf),
the official pure-Go protobuf runtime: descriptors are compiled with `protodesc`,
messages are `dynamicpb` dynamic messages, and binary/JSON encoding delegate to
`proto.Marshal` / `protojson`. **Every byte on the wire is produced by the
canonical Go runtime, so encode/decode is wire-compatible with real protobuf by
construction** — verified in CI by round-tripping through real generated types
(e.g. `timestamppb.Timestamp`).

It is a foundational sibling of the other `go-ruby-*` libraries and the intended
protobuf backend for [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
and for `go-ruby-grpc`.

## Features

- **`DescriptorPool` + builder DSL** — `NewDescriptorPool()`, the process-wide
  `GeneratedPool()`, and a `Build` block mirroring the gem's `pool.build`
  (`add_message` / `add_enum` / `optional` / `repeated` / `map` / `oneof` /
  `value`), plus `Lookup`.
- **Dynamic message objects** — typed field `Get`/`Set`, `ToH`, `Equal` (`==`),
  `Dup`, `Clone`, `Inspect`.
- **Binary + JSON** — `Encode` / `Decode` (binary wire) and `EncodeJSON` /
  `DecodeJSON` (proto3 JSON mapping, with `emit_defaults` /
  `preserve_proto_fieldnames` / `ignore_unknown_fields` options).
- **Repeated fields & maps** — `RepeatedField` and `Map` with Ruby Enumerable
  semantics (`push`/`<<`, `[]`, `[]=`, `each`, `to_a`/`to_h`, `==`, `clear`, …).
- **Well-known types** — `Any` (with `pack`/`unpack`/`is?`), `Timestamp`,
  `Duration`, `Struct`, `Value`, `ListValue`, `FieldMask`, the scalar wrappers
  and `Empty` are pre-registered in every pool and round-trip through the
  canonical runtime.
- **Error taxonomy** — `TypeError`, `RangeError`, `ArgumentError`, `ParseError`,
  each reporting the Ruby exception class a host should raise.

## Example

```go
pool := protobuf.NewDescriptorPool()
_ = pool.Build(func(b *protobuf.Builder) {
	b.AddEnum("Color", func(e *protobuf.EnumBuilder) {
		e.Value("RED", 0)
		e.Value("GREEN", 1)
	})
	b.AddMessage("Person", func(m *protobuf.MessageBuilder) {
		m.Optional("name", protobuf.String, 1)
		m.Optional("id", protobuf.Int32, 2)
		m.Repeated("emails", protobuf.String, 3)
		m.Map("attrs", protobuf.String, protobuf.String, 4)
		m.Optional("fav", protobuf.Enum, 5, "Color")
	})
})

cls := pool.LookupMsgclass("Person")
p, _ := cls.New(map[string]any{"name": "Ada", "id": int64(42)})

emails, _ := p.Get("emails")
_ = emails.(*protobuf.RepeatedField).Push("ada@example.com")
_ = p.Set("fav", protobuf.Symbol("GREEN"))

bytes, _ := protobuf.Encode(p)          // canonical protobuf wire bytes
back, _ := protobuf.Decode(cls, bytes)  // == p
json, _ := protobuf.EncodeJSON(p)       // proto3 JSON
```

## Ruby value model

Message field values cross the boundary as a small, fixed set of Go types, so a
host (such as go-embedded-ruby) can map its own object graph to and from this
package:

| protobuf type                       | Go value                              |
| ----------------------------------- | ------------------------------------- |
| `bool`                              | `bool`                                |
| `int32` / `int64` / `sint*` / `sfixed*` | `int64`                           |
| `uint32` / `uint64` / `fixed*`      | `uint64`                              |
| `float` / `double`                  | `float64`                             |
| `string`                            | `string`                              |
| `bytes`                             | `[]byte`                              |
| `enum`                              | `Symbol` (known) or `int64` (unknown) |
| `message`                           | `*Message` (`nil` when unset)         |
| `repeated`                          | `*RepeatedField`                      |
| `map`                               | `*Map`                                |

## Scope

The runtime + builder are covered faithfully. Deliberately out of scope: the
gem's full protoc-generated codegen DSL (the serialized-`FileDescriptorProto`
string a `.proto` compiles to) — this package offers the equivalent builder DSL
instead — and proto2 group wire syntax. See the package doc comment for details.

## Tests & coverage

```
go test -race -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -1
```

The suite holds coverage at **100%** and validates wire-compatibility against
`google.golang.org/protobuf` (including real generated well-known types) on
Linux/macOS/Windows and on all six 64-bit architectures — amd64, arm64, riscv64,
loong64, ppc64le and **s390x** (big-endian) — in CI.

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright (c) 2026, the
go-ruby-protobuf/protobuf authors.

## WebAssembly

Being pure Go (CGO=0), this library also compiles to **WebAssembly** — both
`GOOS=js GOARCH=wasm` (browser / Node.js) and `GOOS=wasip1 GOARCH=wasm` (WASI).
CI builds both targets on every push, alongside the six 64-bit native/qemu arches.

```sh
GOOS=js     GOARCH=wasm go build ./...   # browser / Node
GOOS=wasip1 GOARCH=wasm go build ./...   # WASI (wasmtime, wasmer, wasmedge, …)
```
