# Amino Spec (and impl for Go)

This software implements Go bindings for the Amino encoding protocol.

Amino is an object encoding specification. It is a subset of Proto3 and a
subset of Go with an extension for interface support, but it is tied to
neither.

The goal of the Amino encoding protocol is to enable logic objects to become
persisted/serialized in a future-compatible way, in a way that streamlines
development from prototyping to production and maintenance.

**DISCLAIMER:** We're still building out the ecosystem, which is currently most
developed in Go.  But Amino is not just for Go — if you'd like to contribute by
creating supporting libraries in various languages from scratch, or by adapting
existing Protobuf3 libraries, please [open an issue on
GitHub](https://github.com/tendermint/go-amino/issues)!

# Why Amino?

## Amino Goals

* Bring parity into logic objects and persistent objects
  by supporting interfaces.
* Have a unique/deterministic encoding of value.
* Binary bytes must be decodeable with a schema.
* Schema must be upgradeable.
* Sufficient structure must be parseable without a schema.
* The encoder and decoder logic must be reasonably simple.
* The serialization must be reasonably compact, assuming
  a string/subsequence compaction string for compacting common strings
  as in Any.TypeURL.
* A sufficiently compatible JSON format must be maintained (but not general
  conversion to/from JSON)

## Amino vs JSON

JavaScript Object Notation (JSON) is human readable, well structured and great
for interoperability with Javascript, but it is inefficient.  Protobuf3, BER,
RLP all exist because we need a more compact and efficient binary encoding
standard.  Amino provides efficient binary encoding for complex objects (e.g.
embedded objects) that integrate naturally with your favorite modern
programming language. Additionally, Amino has a fully compatible JSON encoding.

## Amino vs Protobuf3

Work flow with Amino starts with Go structs, and proto3 schemas are generated
for compatibility with clients in other languages.

Amino uses reflection to produce proto3 compatible binary bytes.
(It should, but it does not currently, produce compatible JSON bytes).

The "genproto" package can be used to automatically generate proto3 schema
files from Go code. So it supports projects that want to define message schemas
in Go.

In the near future, Amino defined structs will use auto-generated code to
use proto3's optimized serializer, by translating amino structs to
protoc generate go code via auto-generated shim go code.  The generated proto3
files will reside in the "proto3" folder of their respective packages.

In the far future, Amino may generate more optimized Go code without relying on
Google's protoc toolset.

go-amino supports Any as Go interfaces.  In other OOP languages, the respective
\*-amino libraries are expected to support similar types for Any in a similarly
native way.

### google.protobuf.Any support and Full Names

Previous versions of amino used to have "disambiguation" and "prefix" bytes
to distinguish among registered concrete types.  This system has been
replaced with the canonical proto3 Any system.

* https://github.com/protocolbuffers/protobuf-go/blob/69839c78c3baff8f1fb37f37b24127ecae185e03/reflect/protoreflect/proto.go#L444

The name of concrete types must conform to proto3's
Any.type\_url spec (generally, "\<domain\and\path\>/\<full.name\>").

For performance, unless we are supporting amino/proto3 with some compression
layer (with registered common strings like Any.type\_url)\*\*, the domain and
path should be short.

\*\* NOTE: If the length of domain names become an issue, it may be desirable to
create a general purpose (de)compressor system that can turn registered common
strings into short representations, LZW with preconfiguration if such doesn't exist already.
Perhaps such a system could be used between two parties for negotiated
shorthand communication via the sharing of Alias declarations, so between p2p
parties, as well as for local persistence.

Proto3 only requires the type URL to include at least one slash,
* https://developers.google.com/protocol-buffers/docs/proto3#any
* https://github.com/protocolbuffers/protobuf/blob/7bff8393cab939bfbb9b5c69b3fe76b4d83c41ee/src/google/protobuf/any_lite.cc#L96
thus the shortest possible domain is empty. However, Google's proto3 generated code
always produces type URLs that start with "type.googleapis.com/".

Amino chooses to the empty domain, keeps the representation short, and assumes
no domain information for security purposes (the user must provide them
explicitly anyways).

At the same time, the full name must also be sufficiently distinguishing of the
message's resource ID/URI, as this is what proto3 tooling expects.  (example:
https://github.com/protocolbuffers/protobuf-go/blob/69839c78c3baff8f1fb37f37b24127ecae185e03/reflect/protoregistry/registry.go#L525).

Without compression, the shortest reasonable type URL for Tendermint's crypto
libraries are of the form "/tm.cryp.PubKeySecp256k1", tm for tendermint, cryp
for crypto libs, and the name of the struct (which could be shorter still).
This is about 20 bytes of extra overhead than the previous 4-byte prefix
system, but is more canonical (which we will need if we ever want binary
signing), and can be optimized in the future.

### google.protobuf.Any JSON Representation for Well Known Types

Google made a decsion a while ago that only Google well known types can
be serialized as {@type:...,value:""} fields.

See
https://github.com/golang/protobuf/blob/d04d7b157bb510b1e0c10132224b616ac0e26b17/jsonpb/encode.go#L336,
Also see the proto3 spec wording: "If the embedded message type is
well-known and has a custom JSON representation, that representation
will be embedded adding a field value which holds the custom JSON in
addition to the @type field."

In addition to these types, the native time and duration types are also
supported as "well known types".  The type\_url will be
"/google.protobuf.Timestamp" and "/google.protobuf.Duration" respectively, and
the encoding format will be identitical to those well known types.

When decoding interface values, by default the native types are constructed
unless field options specify otherwise.

## Amino in the Wild

* Amino:binary spec in [Tendermint](
https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/encoding.md)


# Amino Spec

#### Registering types and packages

Previous versions of Amino used to require a local codec where types must be
registered.  With the change to support Any and type URL strings,
we no longer need to keep track of local codecs, unless we want to override
default behavior from global registrations.

Each package should declare in a package local file (by convention called amino.go)
which should look like the following:

```go
// see github.com/tendermint/go-amino/protogen/example/main.go
package main

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto/example/submodule"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"main", // The Go package path
		"main", // The (shorter) Proto3 package path (no slashes).
		amino.GetCallersDirname(),
	).WithDependencies(
		submodule.Package, // Dependencies must be declared (for now).
	).WithTypes(
		StructA{}, // Declaration of all structs to be managed by Amino.
		StructB{}, // For example, protogen to generate proto3 schema files.
		&StructC{}, // If pointer receivers are preferred when decoding to interfaces.
	),
)
```

You can still override global registrations with local \*amino.Codec state.
This is used by genproto.P3Context, which may help development while writing
migration scripts.  Feedback welcome in the issues section.

## Unsupported types

### Floating points
Floating point number types are discouraged as [they are generally
non-deterministic](http://gafferongames.com/networking-for-game-programmers/floating-point-determinism/).
If you need to use them, use the field tag `amino:"unsafe"`.

### Enums
Enum types are not supported in all languages, and they're simple enough to
model as integers anyways.

### Maps
Maps are not currently supported.  There is an unstable experimental support
for maps for the Amino:JSON codec, but it shouldn't be relied on.  Ideally,
each Amino library should decode maps as a List of key-value structs (in the
case of langauges without generics, the library should maybe provide a custom
Map implementation).  TODO specify the standard for key-value items.

## Amino and Proto3

Amino objects are a subset of Proto3.
* Enums are not supported.
* Nested message declarations are not supported.

Amino extends Proto3's Any system with a particular concrete type
identification format (disfix bytes).

## Amino and Go 

Amino objects are a subset of Go.
* Multi-dimensional slices/arrays are not (yet) supported.
* Floats are nondeterministic, so aren't supported by default.
* Complex types are not (yet) supported.
* Chans, funcs, and maps are not supported.
* Pointers are automatically supported in go-amino but it is an extension of
  the theoretical Amino spec.

Amino, unlike Gob, is beyond the Go language, though the initial implementation
and thus the specification happens to be in Go (for now).
