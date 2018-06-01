# Amino Spec (and impl for Go)

This software implements Go bindings for the Amino encoding protocol.

Amino is an object encoding specification. Think of it as an object-oriented
Protobuf3 with native JSON support.

The goal of the Amino encoding protocol is to bring parity into logic objects
and persistence objects.

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

Protobuf would be better if it supported interfaces/implementations as in most
modern object-oriented languages. Since it is not, the generated code is often
not the logical objects that you really want to use in your application, so you
end up duplicating the structure in the Protobuf schema file and writing
translators to and from your logic objects.  Amino can eliminate this extra
duplication and help streamline development from inception to maturity.

* In Protobuf3, [embedded message are `uvarint` byte-length prefixed]((https://github.com/tendermint/go-amino/wiki/aminoscan));
  However, this makes the binary encoding naturally more inefficient, as bytes
cannot simply be written to a memory array (buffer) in sequence without
allocating a new buffer for each embedded message. Amino is encoded in such a
way that the complete structure of the message (not just the top-level
structure) can be determined by scanning the byte encoding without any type
information other than what is available in the binary bytes. This makes
encoding faster with no penalty when decoding.

* In Protobuf3, lists are encoded as repeated field values.  This makes
  decoding more expensive than necessary, since it is generally faster to know
the length of a list/array to allocate prior to decoding to it.

## Amino in the Wild

* Amino:binary spec in [Tendermint](
https://github.com/tendermint/tendermint/blob/develop/docs/specification/new-spec/encoding.md)


# Amino Spec

## Supported types

Amino supports 7 type families: These are Varint, 4-Byte, 8-Byte, Byte-Length
(delimited), Struct, List, and Interface types.

### Varint

Varints in Amino are like Protobuf Varints.  They can be signed (uvarint) or
unsigned (varint).  This type is used to encode `bool`, `byte`, `[u]int8`,
`[u]int16` and optionally varint-encoded `[u]int32` and `[u]int64` numeric
langauge types.

The binary encoding for signed and unsigned Varints is such that the length is
self-defined, but the encoding doesn't say whether the type is signed or
unsigned.  As in Protobuf, signed varints are zig-zag encoded whereas unsigned
varints are not. The schema is thus required in order to determine the correct
numeric value.

See https://developers.google.com/protocol-buffers/docs/encoding#varints for
more information.

### 4-Byte and 8-Byte

These represent fixed-length 4-byte and 8-byte encodings for `[u]int32` and
[u]int64` numeric langauge types.  They are big-endian encoded.  Like Varint
and Uvarints, these types may be signed or unsigned, so the schema is required
to determine the correct numeric value.

### Byte-Length

Byte-Length delimited fields are used to encode `[]byte` and `string` language
types.

The encoding schema for a Byte-Length types is as follows:

```
<uvarint(len(bytes))><bytes>
e.g. <03><66 6f 6f> encodes the string "foo"
```

### Lists

Lists can hold zero or many items of any single Amino type.

The encoding schema for a List is as follows:

```
<typ3 of list items><uvarint(len(list))><encoding(first item)>...<encoding(last item)>
e.g. <00><02><01><03> encodes a List of two Varints, 1 (or -1 if signed) and 3 (or -2).
e.g. <06><01><00 02 01 03> encodes a List of Lists, with 1 item identical to the above List.
```

See example 3 below for more information on nillable Lists. The encoding schema
for a nillable List is as follows:

```
<typ4 of list items w/ pointer-bit set><uvarint(len(list))>
	<00, or 01 if first item is nil><encoding(first item)>...
	<00, or 01 if last item is nil><encoding(last item)>
```

#### List example 1

A List of Structs with `n` elements is encoded by the byte `0x03` followed by
the uvarint encoding of `n`, followed by the binary encoding of each element.
Each Struct element is encoded starting with the first field key, and is
terminated with the Struct-terminator byte 0x04.

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
// b0000 0010  0x02  Length of List (uvarint(2))
//                   `Item` #0
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyList[0].Number`
// b0000 0010  0x02  Field value (varint(1))
// b0000 0100  0x04  StructTerm for `Item #0`
//                   `Item` #1
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyList[1].Number`
// b0000 0110  0x06  Field value (varint(3))
// b0000 0100  0x04  StructTerm for `Item #1`
```

#### List example 2

A List of `n` elements [where the elements are List-of-Structs] is encoded by
the byte `0x06` followed by the uvarint encoding of `n`, followed by the binary
encoding of each element each which start with the byte `0x03` followed by the
uvarint encoding of `m` (the size of the first child List item).  Each Struct
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
// b0000 0001  0x01  Length of List (uvarint(1))
//                   `List` #0
// b0000 0011  0x03  Type of element (Struct) of `MyLists[0]`
// b0000 0010  0x02  Length of List (uvarint(2))
//                   `Item` #0
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyLists[0][0].Number`
// b0000 0010  0x02  Field value (varint(1))
// b0000 0100  0x04  StructTerm for `Item #0`
//                   `Item` #1
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyLists[0][1].Number`
// b0000 0110  0x06  Field value (varint(3))
// b0000 0100  0x04  StructTerm for `Item #1`
```

#### List example 3 (nilliable List)

Unlike fields of a Struct where nil (or in the future, perhaps zero/empty)
pointers are denoted by the absence of its encoding (both field key and value),
elements of a List are encoded without an index or key.  They are just encoded
one after the other, with no need to prefix each element with a key, index
number nor typ3 byte.

To declare that the List may contain nil elements (e.g. the List is
"nillable"), the Lists's element typ4 byte should set the 4th least-significant
bit (the "pointer bit") to 1.  If (and only if) the pointer bit is 1, each
element is prefixed by a "nil byte" — a 0x00 byte to declare that a non-nil
item follows, or a 0x01 byte to declare that the next item is nil.  Note that
the byte values are flipped (typically 0 is used to denote nil).  This is to
open the possibility of supporting sparse encoding of nil Lists in the future
by encoding the number of nil items to skip as a uvarint.

Nil Lists, Interfaces, (and in Go-Amino, nil pointers) are all encoded as nil
in a nillable List.

NOTE: A nil Interface in a nillable List is encoded with a single byte 0x01,
while a nil Interface in a non-nillable List is encoded with two bytes 0x0000.

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
// b0000 0010  0x02  Length of List (uvarint(2))
// b0000 0000  0x00  Byte to denote non-nil element
// b0000 1000  0x08  Field number (1) and type (Varint) for `MyList[0].Number`
// b0000 0010  0x02  Field value (varint(1))
// b0000 0100  0x04  StructTerm for `Item #0`
// b0000 0001  0x01  Byte to denote nil element
```

### Struct

As in Protobuf, each Struct field is keyed by an unsigned Varint with the value
`(field_number << 3) | field_typ3`, where `field_typ3` is the 3 bit typ3 of the
field value.  The fields of a Struct are ordered by field number.

All inner Structs end with a Struct-terminator byte 0x04.  Top-level Structs
(e.g. structs as encoded by cdc.MarshalBinaryBare() in Go-Amino) does not end
with a Struct-terminator, because top-level Structs can be implicitly
terminated by the length of the input bytes.

```
<field_key><field_value> <field_key><field_value>... <Struct_term?>
e.g. <08><14> <10><28> <20><3C> encodes a Struct with three Varint fields, of
    values 20 (or 10 if signed) for field #1, 40 (or 20) for #2, and 60 (or 30) for
    #3 respectively.
e.g. <0E><06><01><<00><02><01><03>> encodes a Struct with a single field #1
    which is a List type with one element which is itself a List with two Varints, 
    1 (or -1 if signed) and 3 (or -2).
e.g. <0A><<03><66 6F 6F>> <13><<08><01><04>> encodes a Struct with two fields,
    a string or byteslice ("foo") for field #1, and an inner Struct with a single
    Varint field with value 1 (or -1 if signed).  Notice the Struct-terminator
    byte 0x04 which terminates the inner Struct, but not the root Struct.
```

Time is encoded as a Struct with two fields:
1. Unix seconds since 1970 as an int64 (may be negative)
2. Nanoseconds as an int32 (must be non-negative)

### Interface

Amino is an encoding library that can handle Interfaces. This is achieved by
prefixing bytes before each "concrete type".

A concrete type is a non-Interface type which implements a registered
Interface. Not all types need to be registered as concrete types — only when
they will be stored in Interface type fields (or in a List with Interface
elements) do they need to be registered.  Registration of Interfaces and the
implementing concrete types should happen upon initialization of the program to
detect any problems (such as conflicting prefix bytes -- more on that later).

```
<field_key><field_value> <field_key><field_value>... <Struct_term?>
```

A concrete type is typically a Struct type, but it doesn't need to be.  The
last 3 bits of the written prefix bytes are the concrete type's typ3 bits, so a
scanner can recursively traverse the fields and elements of the value.  A nil
Interface value is encoded by 2 zero bytes (0x0000) in place of the 4 prefix
bytes.  As in Protobuf, a nil Struct field value is not encoded at all.

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
representation.  The first byte of the prefix bytes must not be a zero byte, and
the last 3 bits are reserved for the [`typ3` bits](#the-typ3-byte), so there
are `2^(8x4-3)-2^(8x3-3) = 534,773,760` possible values.

When there are 1024 concrete types registered that implement the same Interface,
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
> <0xA8 0xFC 0x54> [0xBB 0x9C 9x83 9xDD] // Before stripping typ3 bits
```

We reserve the last 3 bits for the typ3 of the concrete type, so in
this case the final prefix bytes become `(0xDD & 0xF8) | <typ3-byte>`.
The type byte for a Struct is 0x03, so if the concrete type were a Struct,
the final prefix byte would be `0xDB`.

```
> <0xA8 0xFC 0x54> [0xBB 0x9C 9x83 9xDB] // Final <Disamb Bytes> and [Prefix Bytes]
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

## Amino vs Protobuf3 in detail

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

### The Typ3 Byte

In Amino, the "type" is similarly encoded by 3 bits, called the "typ3". When it
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

### Structs

Struct fields are encoded in order, and a null/empty/zero field is represented
by the absence of a field in the encoding, similar to Protobuf. Unlike Protobuf,
in Amino, the total byte-size of a Amino encoded struct cannot in general be
determined in a stream until each field's size has been determined by scanning
all fields and elements recursively.

As in Protobuf, each struct field is keyed by a uvarint with the value
`(field_number << 3) | type`, where `type` is 3 bits long.

When the typ3 bits are represented as a single byte (using the least
significant bits of the byte), we call it the "typ3 byte".  For example, the
typ3 byte for a List is `0x06`.

Inner structs that are embedded in outer structs are encoded by the field typ3
"Struct" (e.g. `0x03`).  (In Protobuf3, embedded messages are encoded as
"Byte-Length (prefixed)".  In Amino, the "Byte-Length" typ3 is only used for
byteslices and bytearrays.)

### Lists

Unlike Protobuf, Amino deprecates "repeated fields" in favor of Lists. A List
is encoded by first writing the typ4 byte of the element type, followed by the
uvarint encoding of the length of the List, followed by the encoding of each
element.

In Amino, when encoding elements of a List (Go slice or array), the typ3
byte isn't enough.  Specifically, when the element type of the List is a
pointer type, the element value may be nil.  We encode the element type of this
kind of List with a typ4 byte, which is like a typ3 byte, but uses the 4th
least significant bit to encode the "pointer bit".  In other words, the element
typ4 byte follows the List typ3 bits--where the List typ3 bits appear as (1)
the last 3 bits of a struct's field key, (2) the last 3 bits of an Interface's
prefix bytes, or (3) the typ4 byte of a parent List's element type declaration.

### Interfaces and concrete types

Amino drops support of `oneof` and instead supports Interfaces. See the Amino
Interface spec in this README for more information on how interfaces work.


# Amino in other langauges

[Open an Issue on GitHub](https://github.com/tendermint/go-amino/issues), as we
will pay out bounties for implementations in other languages.  In Golang, we are
are primarily interested in codec generators.
