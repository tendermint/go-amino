# Amino encoding for Go

This software implements Go bindings for the Amino encoding protocol.

Amino is an object encoding specification. Think of it as an object-oriented
Protobuf3 with native JSON support.

The goal of the Amino encoding protocol is to bring parity between application
logic objects and persistence objects.

**CAVEAT:** we're still building out the ecosystem, which is currently most
developed in Go.  But Amino is not just for Go — if you'd like to
contribute by creating supporting libraries in various languages from scratch,
or by adapting existing Protobuf3 libraries, please [open an issue on GitHub](https://github.com/tendermint/go-amino/issues)!

# Why Amino?

## Amino vs JSON

JavaScript Object Notation (JSON) is human readable, well structured and great for interoperability with Javascript, but it is inefficient.
Protobuf3, BER, RLP all exist because we need a more compact and efficient binary encoding standard.
Amino provides efficient binary encoding for complex objects (e.g. embedded objects) that integrate naturally
with your favorite modern programming language. Additionally, Amino is fully compatible with JSON encoding.

## Amino vs Protobuf3

Amino wants to be Protobuf4. The bulk of this spec will
explain how Amino differs from Protobuf3. Here, we will illustrate two key
selling points for Amino.

* In Protobuf3, [embedded message are `varint` byte-length prefixed]((https://github.com/tendermint/go-amino/wiki/aminoscan));
  However, this makes the binary encoding naturally more inefficient, as bytes cannot
  simply be written to a memory array (buffer) in sequence without allocating a
  new buffer for each embedded message. Amino is encoded in such a way that the
  complete structure of the message (not just the top-level structure) can be determined by
  scanning the byte encoding without any type information other than what is
  available in the binary bytes. This makes encoding faster with no penalty when
  decoding.

* Protobuf3 has `oneof`, but it's clunky.  For example, Go Protobuf's
  implementation is [excessively verbose](https://github.com/gogo/protobuf/issues/168).
  But this isn't just an implementation issue. The real problem is that `oneof`
  doesn't match how modern languages already work to provide `oneof`-like
  features. Protobuf3's `oneof` support feels more like a (bad) encoding for C union
  types.  For example, you can't declare a union type and re-use it in Protobuf.
  
  What we want is a way to encode *objects*, and a new type to represent a set of object
  types (often called interfaces).
  
  * In C++, classes.  Unions are still useful (e.g. for performance) but not as widely used as classes.
  * In Java, Java-interfaces and classes.
  * In Go, the replacement is Golang-interfaces and all (even primitive) types.
  * Javascript naturally lends itself well to `oneof` support, as it only has a few native types including the ubiquitous Object type.

Protobuf would be better if it were object-oriented. Since it isn't, the generated code
is often not the logical objects that you really want to use in your
application, so you end up duplicating the structure in the Protobuf schema file and
writing translators to and from your logic objects.  Amino can eliminate this extra duplication.

## Amino in the Wild

* Amino:binary spec in [Tendermint](
https://github.com/tendermint/tendermint/blob/develop/docs/specification/new-spec/encoding.md)

## Amino Spec

### Interfaces and concrete types

Amino is an encoding library that can handle interfaces (like Protobuf `oneof`)
exceptionally well.  This is achieved by prefixing bytes before each "concrete
type".

A concrete type is some non-interface value (generally a struct) which
implements the interface to be (de)serialized. Not all structures need to be
registered as concrete types — only when they will be stored in interface type
fields (or interface type slices) do they need to be registered.

### Registering types

To encode and decode an interface, it has to be registered with `codec.RegisterInterface`
and its respective concrete type implementers should be registered with `codec.RegisterConcrete`

```go
amino.RegisterInterface((*MyInterface1)(nil), nil)
amino.RegisterInterface((*MyInterface2)(nil), nil)
amino.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)
amino.RegisterConcrete(MyStruct2{}, "com.tendermint/MyStruct2", nil)
amino.RegisterConcrete(&MyStruct3{}, "anythingcangoinhereifitsunique", nil)
```

Notice that an interface is represented by a nil pointer of that interface.

Amino tries to transparently deal with pointers (and pointer-pointers) when it can.
When it comes to decoding a concrete type into an interface value, Go gives
the user the option to register the concrete type as a pointer or non-pointer.
If and only if the value is registered as a pointer the decoded value will be a pointer as well.
...

### Prefix bytes to identify the concrete type
All registered concrete types are encoded with leading 4 bytes (called "prefix
bytes"), even when it's not held in an interface field/element.  In this way,
Amino ensures that concrete types (almost) always have the same canonical
representation.  The first byte of the prefix bytes must not be a zero byte, and
the last 3 bits are reserved for the [`typ3` bits](#the-typ3-byte), so there
are `2^(8x4-3)-2^(8x3-3) = 534,773,760` possible values.

When there are 1024 concrete types registered that implement the same interface,
the probability of there being a conflict is ~ 0.1%.

This is assuming that all registered concrete types have unique natural names
(e.g.  prefixed by a unique entity name such as "com.tendermint/", and not
"mined/grinded" to produce a particular sequence of "prefix bytes"). Do not
mine/grind to produce a particular sequence of prefix bytes, and avoid using
dependencies that do so.

```
The Birthday Paradox: 1024 random registered types, Amino prefix bytes
https://instacalc.com/51339

possible = 534773760                                = 534,773,760
registered = 1024                                   = 1,024
pairs = ((registered)*(registered-1)) / 2           = 523,776
no_collisions = ((possible-1) / possible)^pairs     = 0.99902104475
any_collisions = 1 - no_collisions                  = 0.00097895525
percent_any_collisions = any_collisions * 100       = 0.09789552533
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

### Computing the prefix and disambiguation bytes
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
> <0xA8 0xFC 0x54> [0xBB 0x9C 9x83 9xDD] // Before stripping typ3 bits
```

We reserve the last 3 bits for the typ3 of the concrete type, so in
this case the final prefix bytes become `(0xDD & 0xF8) | <typ3-byte>`.
The type byte for a struct is 0x03, so if the concrete type were a struct,
the final prefix byte would be `0xDB`.

```
> <0xA8 0xFC 0x54> [0xBB 0x9C 9x83 9xDB] // Final <Disamb Bytes> and [Prefix Bytes]
```

### Supported types

#### Primary Types
`uvarint`, `varint`, `byte`, `uint[8,16,32,64]`, `int[8,16,32,64]`, `string`, and `time`.

#### Arrays
Arrays can hold items of any arbitrary type.  For example, byte-arrays and
byte-array-arrays are supported.

#### Structs
Struct fields are encoded by value (without the key name) in the order that they
are declared in the struct.  In this way it is similar to Apache Avro.

#### Interfaces
Interfaces are like union types where the value can be any
non-interface type. The actual value is preceded by a single "type byte" that
shows which concrete is encoded.

#### Pointers
Pointers are like optional fields.  The first byte is 0x00 to
denote a null pointer (i.e. no value), otherwise it is 0x01.

### Supported (but discouraged) types

#### Maps
In most languages, iteration order is nondeterministic, so in those
cases Amino (should) provide a standard Map library for compatibility.

#### Floating points
Floating point number types are discouraged as [they are generally
non-deterministic](http://gafferongames.com/networking-for-game-programmers/floating-point-determinism/).
If you need to use them, use the field tag `amino:"unsafe"`.

### Unsupported types

#### Enums
Enum types are not supported in all languages, and they're simple enough to
model as integers anyways.

### Amino vs Protobuf3 in detail

From the [Protocol Buffers encoding
guide](https://developers.google.com/protocol-buffers/docs/encoding):

> As you know, a protocol buffer message is a series of key-value pairs. The
> binary version of a message just uses the field's number as the key – the
> name and declared type for each field can only be determined on the decoding
> end by referencing the message type's definition (i.e. the .proto file).
>
> When a message is encoded, the keys and values are concatenated into a byte
> stream. When the message is being decoded, the parser needs to be able to
> skip fields that it doesn't recognize. This way, new fields can be added to a
> message without breaking old programs that do not know about them. To this
> end, the "key" for each pair in a wire-format message is actually two values
> – the field number from your .proto file, plus a wire type that provides just
> enough information to find the length of the following value.
>
> The available wire types are as follows:
>
> Type | Meaning | Used For
> ---- | ------- | --------
> 0    | Varint  | int32, int64, uint32, uint64, sint32, sint64, bool, enum
> 1    | 64-bit  | fixed64, sfixed64, double
> 2    | Length-prefixed | string, bytes, embedded messages, packed repeated fields
> 3    | Start group | groups (deprecated)
> 4    | End group | groups (deprecated)
> 5    | 32-bit  | fixed32, sfixed32, float
>
> Each key in the streamed message is a varint with the value (field_number <<
> 3) | wire_type – in other words, the last three bits of the number store the
> wire type.

#### The Typ3 Byte

In Amino, the "type" is similarly enocded by 3 bits, called the "typ3". When it
appears alone in a byte, it is called a "typ3 byte".

In Amino, `varint` is the Protobuf equivalent of "signed varint" aka `sint32`,
and `uvarint` is the equivalent of "varint" aka `int32`.

Typ3 | Meaning          | Used For
---- | ---------------- | --------
0    | Varint           | bool, byte, [u]int16, and varint-[u]int[64/32]
1    | 8-Byte           | int64, uint64, float64(unsafe)
2    | Byte-Length      | string, bytes, raw?
3    | Struct           | struct (e.g. Protobuf message)
4    | Struct Term      | end of struct
5    | 4-Byte           | int32, uint32, float32(unsafe)
6    | List             | array, slice; followed by element `<typ4-byte>`, then `<uvarint(num-items)>`
7    | Interface        | registered concrete types; followed by `<prefix-bytes>` or `<disfix-bytes>`, the last byte which ends with the concrete type's `<typ3-byte>`.

#### Structs

Struct fields are encoded in order, and a null/empty/zero field is represented
by the absence of a field in the encoding, similar to Protobuf. Unlike Protobuf,
in Amino, the total byte-size of a Amino encoded struct cannot in general be
determined in a stream until each field's size has been determined by scanning
all fields and elements recursively.

As in Protobuf, each struct field is keyed by a uvarint with the value
`(field_number << 3) | type`, where `type` is 3 bits long.

When the typ3 bits are represented as a single byte (using the least
significant bits of the byte), we call it the "typ3 byte".  For example, the
typ3 byte for a "list" is `0x06`.

In Amino, when encoding elements of a "list" (Go slice or array), the typ3
byte isn't enough.  Specifically, when the element type of the list is a
pointer type, the element value may be nil.  We encode the element type of this
kind of list with a typ4 byte, which is like a typ3 byte, but uses the 4th
least significant bit to encode the "pointer bit".  In other words, the element
typ4 byte follows the List typ3 bits--where the List typ3 bits appear as (1)
the last 3 bits of a struct's field key, (2) the last 3 bits of an interface's
prefix bytes, or (3) the typ4 byte of a parent list's element type declaration.

Inner structs that are embedded in outer structs are encoded by the field typ3
"Struct" (e.g. `0x03`).  (In Protobuf3, embedded messages are encoded as
"Byte-Length (prefixed)".  In Amino, the "Byte-Length" typ3 is only used for
byteslices and bytearrays.)

#### Lists

Unlike Protobuf, Amino deprecates "repeated fields" in favor of "lists". A list
is encoded by first writing the typ4 byte of the element type, followed by the
uvarint encoding of the length of the list, followed by the encoding of each
element.

A list of structs with `n` elements is encoded by the byte `0x03` followed by
the uvarint encoding of `n`, followed by the binary encoding of each element.
Each struct element is encoded starting with the first field key, and is
terminated with the `StructTerm` typ3 byte (`0x04`, which could be interpreted
as a special struct key with field number 0).

```go
type Item struct {
	Number int
}

type List struct {
	MyList []Item
}

list := List{
	MyList: []Item{
		Item{1},		// Item #0
		Item{3},		// Item #1
	}
}

bz, err := amino.MarshalBinary(list)
if err != nil { ... }

// dump bz:
// BINARY      HEX   NOTE
// b0000 1110  0x0E  Field number (1) and type (List) for `MyList`
// b0000 0011  0x03  Type of element (Struct) of `MyList`
// b0000 0010  0x02  Length of list (uvarint(2))
//                   `Item` #0
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyList[0].Number`
// b0000 0010  0x02  Field value (varint(1))
// b0000 0100  0x04  StructTerm for `Item #0`
//                   `Item` #1
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyList[1].Number`
// b0000 0110  0x06  Field value (varint(3))
// b0000 0100  0x04  StructTerm for `Item #1`
// b0000 0100  0x04  StructTerm for `List`
```

A list of `n` elements [where the elements are list-of-structs] is encoded by
the byte `0x06` followed by the uvarint encoding of `n`, followed by the binary
encoding of each element each which start with the byte `0x03` followed by the
uvarint encoding of `m` (the size of the first child list item).  Each struct
element is encoded starting with the first field key, as in the previous
example.

```go
type Item struct {
	Number int
}

type List []Item

type ListOfLists struct {
	MyLists []List
}

llist := ListOfLists{
	MyLists: []List{
		[]Item{			// List #0
			Item{1},	// Item #0
			Item{3},	// Item #1
		},
	}
}

bz, err := amino.MarshalBinary(llist)
if err != nil { ... }

// dump bz:
// BINARY      HEX   NOTE
// b0000 1110  0x0E  Field number (1) and type ([]List) for `MyLists`
// b0000 0110  0x06  Type of element (List) of `MyLists`
// b0000 0001  0x01  Length of list (uvarint(1))
//                   `List` #0
// b0000 0011  0x03  Type of element (Struct) of `MyLists[0]`
// b0000 0010  0x02  Length of list (uvarint(2))
//                   `Item` #0
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyLists[0][0].Number`
// b0000 0010  0x02  Field value (varint(1))
// b0000 0100  0x04  StructTerm for `Item #0`
//                   `Item` #1
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyLists[0][1].Number`
// b0000 0110  0x06  Field value (varint(3))
// b0000 0100  0x04  StructTerm for `Item #1`
// b0000 0100  0x04  StructTerm for `ListOfLists`
```

Unlike fields of a struct where nil (or in the future, perhaps zero/empty)
pointers are denoted by the absence of its encoding (both field key and value),
elements of a list are encoded without an index or key.  They are just encoded
one after the other, with no need to prefix each element with a key, index
number nor typ3 byte.

To declare that the List may contain nil elements (e.g. the list "nillable"),
the Lists's element typ4 byte should set the 4th least-significant bit (the
"pointer bit") to 1.  If (and only if) the pointer bit is 1, each element is
prefixed by a "nil byte" — a 0x00 byte to declare that a non-nil item follows,
or a 0x01 byte to declare that the next item is nil.  Note that the byte values
are flipped (typically 0 is used to denote nil).  This is to open the
possibility of supporting sparse encoding of nil lists in the future by
encoding the number of nil items to skip as a uvarint.

Nil slices, interfaces, and pointers are all encoded as nil in a nillable list.

NOTE: A nil interface in a nillable list is encoded with a single byte 0x01,
while a nil interface in a non-nillable list is encoded with two bytes 0x0000.

```go
type Item struct {
	Number int
}

type List struct {
	MyList []*Item
}

list := List{
	MyList: []*Item{
		Item{1},		// Item #0
		nil,			// Item #1
	}
}

bz, err := amino.MarshalBinary(list)
if err != nil { ... }

// dump bz:
// BINARY      HEX   NOTE
// b0000 1110  0x0E  Field number (1) and type (List) for `MyList`
// b0000 1011  0x03  Type of element (nillable Struct) of `MyList`
// b0000 0010  0x02  Length of list (uvarint(2))
// b0000 0000  0x00  Byte to denote non-nil element
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyList[0].Number`
// b0000 0010  0x02  Field value (varint(1))
// b0000 0100  0x04  StructTerm for `Item #0`
// b0000 0001  0x01  Byte to denote nil element
// b0000 0100  0x04  StructTerm for `List`
```

In theory, List encoding could be similar to struct encoding, e.g. by prefixing
each element with a key that includes the index number. Instead, the Amino
encoding specified here is more compact for dense lists because the index
number is implied.

In the future, for sparse lists we could support encoding of more than one nil
items at a time, which could be even more compact.

NOTE: The current spec makes the byte-length of the input more-or-less
representative of the amount of memory it takes to decode the input. A 200-byte
go-amino binary blob shouldn't decode into a 1GB object in memory, but it might
with sparse encoding, so we should be aware of that.

#### Interfaces

An interface value is typically a `struct`, but it doesn't need to be.  The last
3 bits of the written prefix bytes are the concrete type's typ3 bits, so a
scanner can recursively traverse the fields and elements of the value.  A nil
interface value is encoded by 2 zero bytes (0x0000) in place of the 4 prefix
bytes.  As in Protobuf, a nil struct field value is not encoded at all.

# Amino in other langauges

[Open an Issue on GitHub](https://github.com/tendermint/go-amino/issues), as we
will pay out bounties for implementations in other languages.  In Golang, we are
are primarily interested in codec generators.
