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
* The serialization must be reasonably compact.
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

Amino wants to be Protobuf4. The bulk of this spec will explain how Amino
differs from Protobuf3. Here, we will illustrate two key selling points for
Amino.

* Protobuf3 doesn't support interfaces.  It supports `oneof`, which works as a
  kind of union type, but it doesn't translate well to "interfaces" and
"implementations" in modern langauges such as C++ classes, Java
interfaces/classes, Go interfaces/implementations, and Rust traits.  

If Protobuf supported interfaces, users of externally defined schema files
would be able to support caller-defined concrete types of an interface.
Instead, the `oneof` feature of Protobuf3 requires the concrete types to be
pre-declared in the definition of the `oneof` field.

Proto3 now supports Any, but its usage is still being defined.  Amino is a
gg
proposal for how this could work, and go-amino is a proposal in the context of
Go.

## Amino in the Wild

* Amino:binary spec in [Tendermint](
https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/encoding.md)


# Amino Spec

### Interface

Amino is an encoding library that can handle Interfaces. This is achieved by
prefixing bytes before each "concrete type".

A concrete type is a non-Interface type which implements a registered
Interface. Not all types need to be registered as concrete types — only when
they will be stored in Interface type fields (or in a List with Interface
elements) do they need to be registered.  Registration of Interfaces and the
implementing concrete types should happen upon initialization of the program to
detect any problems (such as conflicting prefix bytes -- more on that later).

#### Registering types

To encode and decode an Interface, it has to be registered with `codec.RegisterInterface`
and its respective concrete type implementers should be registered with `codec.RegisterConcrete`

```go
amino.RegisterInterface((*MyInterface1)(nil), nil)
amino.RegisterInterface((*MyInterface2)(nil), nil)
amino.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)
amino.RegisterConcrete(MyStruct2{}, "com.tendermint/MyStruct2", nil)
amino.RegisterConcrete(&MyStruct3{}, "anythingcangoinhereifitsunique", nil)
```

Notice that an Interface is represented by a nil pointer of that Interface.

NOTE: Go-Amino tries to transparently deal with pointers (and pointer-pointers)
when it can.  When it comes to decoding a concrete type into an Interface
value, Go gives the user the option to register the concrete type as a pointer
or non-pointer.  If and only if the value is registered as a pointer is the
decoded value will be a pointer as well.

#### Prefix bytes to identify the concrete type

All registered concrete types are encoded with leading 4 bytes (called "prefix
bytes"), even when it's not held in an Interface field/element.  In this way,
Amino ensures that concrete types (almost) always have the same canonical
representation.  The first byte of the prefix bytes must not be a zero byte,
so there are `2^(8x4)-2^(8x3) = 4,278,190,080` possible values.

When there are 1024 concrete types registered that implement the same Interface,
the probability of there being a conflict is ~ 0.01%.

This is assuming that all registered concrete types have unique natural names
(e.g.  prefixed by a unique entity name such as "com.tendermint/", and not
"mined/grinded" to produce a particular sequence of "prefix bytes"). Do not
mine/grind to produce a particular sequence of prefix bytes, and avoid using
dependencies that do so.

```
The Birthday Paradox: 1024 random registered types, Wire prefix bytes
https://instacalc.com/51554

possible = 4278190080                               = 4,278,190,080 
registered = 1024                                   = 1,024 
pairs = ((registered)*(registered-1)) / 2           = 523,776 
no_collisions = ((possible-1) / possible)^pairs     = 0.99987757816 
any_collisions = 1 - no_collisions                  = 0.00012242184 
percent_any_collisions = any_collisions * 100       = 0.01224218414 
```

Since 4 bytes are not sufficient to ensure no conflicts, sometimes it is
necessary to prepend more than the 4 prefix bytes for disambiguation.  Like the
prefix bytes, the disambiguation bytes are also computed from the registered
name of the concrete type.  There are 3 disambiguation bytes, and in binary
form they always precede the prefix bytes.  The first byte of the
disambiguation bytes must not be a zero byte, so there are 2^(8x3)-2^(8x2)
possible values.

```
// Sample Amino encoded binary bytes with 4 prefix bytes.
> [0xBB 0x9C 0x83 0xDD] [...]

// Sample Amino encoded binary bytes with 3 disambiguation bytes and 4
// prefix bytes.
> 0x00 <0xA8 0xFC 0x54> [0xBB 0x9C 0x83 0xDD] [...]
```

The prefix bytes never start with a zero byte, so the disambiguation bytes are
escaped with 0x00.

The 4 prefix bytes always immediately precede the binary encoding of the
concrete type.

#### Computing the prefix and disambiguation bytes

To compute the disambiguation bytes, we take `hash := sha256(concreteTypeName)`,
and drop the leading 0x00 bytes.

```
> hash := sha256("com.tendermint.consensus/MyConcreteName")
> hex.EncodeBytes(hash) // 0x{00 00 A8 FC 54 00 00 00 BB 9C 83 DD ...} (example)
```

In the example above, hash has two leading 0x00 bytes, so we drop them.

```
> rest = dropLeadingZeroBytes(hash) // 0x{A8 FC 54 00 00 00 BB 9C 83 DD ...}
> disamb = rest[0:3]
> rest = dropLeadingZeroBytes(rest[3:])
> prefix = rest[0:4]
```

The first 3 bytes are called the "disambiguation bytes" (in angle brackets).
The next 4 bytes are called the "prefix bytes" (in square brackets).

```
> <0xA8 0xFC 0x54> [0xBB 0x9C 9x83 9xDD] // <Disamb Bytes> and [Prefix Bytes]
```

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
