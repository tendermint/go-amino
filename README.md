# Wire encoding for Golang

This software implements Go bindings for the Wire encoding protocol.  The goal
of the Wire encoding protocol is to be a simple language-agnostic encoding
protocol for rapid prototyping of blockchain applications.

This package also includes a compatible (and slower) JSON codec.

## Interfaces and concrete types

Wire is an encoding library that can handle interfaces (like Protobuf "oneof")
well.  This is achieved by prefixing bytes before each "concrete type".

A concrete type is some non-interface value (generally a struct) which
implements the interface to be (de)serialized. Not all structures need to be
registered as concrete types -- only when they will be stored in interface type
fields (or interface type slices) do they need to be registered.

### Registering types

All interfaces and the concrete types that implement them must be registered.

```golang
wire.RegisterInterface((*MyInterface1)(nil), nil)
wire.RegisterInterface((*MyInterface2)(nil), nil)
wire.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)
wire.RegisterConcrete(MyStruct2{}, "com.tendermint/MyStruct2", nil)
wire.RegisterConcrete(&MyStruct3{}, "anythingcangoinhereifitsunique", nil)
```

Notice that an interface is represented by a nil pointer.

Structures that must be deserialized as pointer values must be registered with
a pointer value as well.  It's OK to (de)serialize such structures in
non-pointer (value) form, but when deserializing such structures into an
interface field, they will always be deserialized as pointers.

### How it works

All registered concrete types are encoded with leading 4 bytes (called "prefix
bytes"), even when it's not held in an interface field/element.  In this way,
Wire ensures that concrete types (almost) always have the same canonical
representation.  The first byte of the prefix bytes must not be a zero byte, so
there are 2^(8x4)-2^(8x3) possible values.

When there are 4096 types registered at once, the probability of there being a
conflict is ~ 0.2%. See https://instacalc.com/51189 for estimation.  This is
assuming that all registered concrete types have unique natural names (e.g.
prefixed by a unique entity name such as "com.tendermint/", and not
"mined/grinded" to produce a particular sequence of "prefix bytes").

TODO Update instacalc.com link with 255/256 since 0x00 is an escape.

Do not mine/grind to produce a particular sequence of prefix bytes, and avoid
using dependencies that do so.

Since 4 bytes are not sufficient to ensure no conflicts, sometimes it is
necessary to prepend more than the 4 prefix bytes for disambiguation.  Like the
prefix bytes, the disambiguation bytes are also computed from the registered
name of the concrete type.  There are 3 disambiguation bytes, and in binary
form they always precede the prefix bytes.  The first byte of the
disambiguation bytes must not be a zero byte, so there are 2^(8x3)-2^(8x2)
possible values.

```
// Sample Wire encoded binary bytes with 4 prefix bytes.
> [0xBB 0x9C 0x83 0xDD] [...]

// Sample Wire encoded binary bytes with 3 disambiguation bytes and 4
// prefix bytes.
> 0x00 <0xA8 0xFC 0x54> [0xBB 0x9C 0x83 0xDD] [...]
```

The prefix bytes never start with a zero byte, so the disambiguation bytes are
escaped with 0x00.

Notice that the 4 prefix bytes always immediately precede the binary encoding
of the concrete type.

### Computing prefix bytes

To compute the disambiguation bytes, we take `hash := sha256(concreteTypeName)`,
and drop the leading 0x00 bytes.

```
> hash := sha256("com.tendermint.consensus/MyConcreteName")
> hex.EncodeBytes(hash) // 0x{00 00 A8 FC 54 00 00 00 BB 9C 83 DD ...} (example)
```

In the example above, hash has two leading 0x00 bytes, so we drop them.

```
> rest = dropLeadingZeroBytes(hash) // 0x{A8 FC 54 00 00 BB 9C 83 DD ...}
> disamb = rest[0:3]
> rest = dropLeadingZeroBytes(rest[3:])
> prefix = rest[0:4]
```

The first 3 bytes are called the "disambiguation bytes" (in angle brackets).
The next 4 bytes are called the "prefix bytes" (in square brackets).

```
> <0xA8 0xFC 0x54> [0xBB 0x9C 9x83 9xDD]
```

### Supported types

**Primary types**: `uvarint`, `varint`, `byte`, `uint[8,16,32,64]`, `int[8,16,32,64]`, `string`, and `time` types are supported

**Arrays**: Arrays can hold items of any arbitrary type.  For example, byte-arrays and byte-array-arrays are supported.

**Structs**: Struct fields are encoded by value (without the key name) in the order that they are declared in the struct.  In this way it is similar to Apache Avro.

**Interfaces**: Interfaces are like union types where the value can be any non-interface type. The actual value is preceded by a single "type byte" that shows which concrete is encoded.

**Pointers**: Pointers are like optional fields.  The first byte is 0x00 to denote a null pointer (i.e. no value), otherwise it is 0x01.

### Unsupported types

**Maps**: Maps are not supported because for most languages, key orders are nondeterministic.
If you need to encode/decode maps of arbitrary key-value pairs, encode an array of {key,value} structs instead.

**Floating points**: Floating point number types are discouraged because [of reasons](http://gafferongames.com/networking-for-game-programmers/floating-point-determinism/).  If you need to use them, use the field tag `wire:"unsafe"`.

**Enums**: Enum types are not supported in all languages, and they're simple enough to model as integers anyways.

## Forward and Backward compatibility

TODO

## Wire vs JSON

TODO

## Wire vs Protobuf

XXX Why Protobuf3 isn't good enough

From the [Protocol Buffers encoding guide](https://developers.google.com/protocol-buffers/docs/encoding):

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

In Wire, the "type" is similarly enocded by 3 bits, called the `typ3`. When it
appears alone in a byte, it is called a `typ3 byte`.

In Wire, "varint" means Protobuf's "signed varint", and "uvarint" means "varint".

Typ3 | Meaning          | Used For
---- | ---------------- | --------
0    | Varint           | bool, byte, [u]int16, and varint-[u]int[64/32]
1    | 64-bit           | int64, uint64, float64(unsafe)
2    | Length-prefixed  | string, bytes, raw?
3    | Start struct     | conceptually, '{'
4    | End struct       | conceptually, '}'; always a single byte, 0x04
5    | 32-bit           | int32, uint32, float32(unsafe)
6    | List             | array, slice; followed by element `<typ4-byte>`, then `<uvarint(num-items)>`
7    | Interface        | registered concrete types; followed by `<prefix-bytes>` or `<disfix-bytes>`, then `<typ3-byte>`.

### Structs 

Struct fields are encoded in order, and a null/empty/zero field is represented
by the absence of a field in the encoding, similar to Protobuf. In general, the
total byte-size of a Wire:binary encoded struct cannot be determined in a
stream until each field's size has been determined by scanning all fields and
elements recursively.

As in Protobuf, each struct field is keyed by a uvarint with the value
`(field_number << 3) | type`, where `type` is 3 bits long.  We call these 3
bits in byte form (using the least significant bits) a `typ3 byte`.  For
instance, the typ3 byte for List is 0x06.

When encoding elements of a list (Golang slice or array) in Wire, the typ3 byte
isn't enough.  Specifically, when the element type of the list is a pointer
type, the element value may be nil.  We encode the element type of this kind of
list with a typ4 byte, which is like a typ3 byte, but uses 4 least significant
bits.  This 4th bit is called the "pointer bit".

Inner structs that are embedded in outer structs are encoded by the field typ3
"Start struct".  (In Protobuf3, embedded messages are encoded as a
Length-prefixed).

TODO: How does this work in Protobuf3?

### Lists

Unlike Protobuf, Wire deprecates "repeated fields" in favor of Lists. A List is
encoded by first writing the typ4 byte of the element type, followed by the
uvarint encoding of the length of the list, followed by the encoding of each
element.

A list of structs with `n` elements is encoded by the byte `0x03` followed by
the uvarint encoding of `n`, followed by the binary encoding of each
element.  Each struct element is encoded as if followed by a `Start struct`
field key (i.e. it is implied), and terminates with an `End struct` field key
(specifically, the byte 0x04).

```golang
type Item struct {
	Number int
}
type List struct {
	MyList []Item
}

list := List{[]Item{Item{1}, Item{3}}}

bz, err := wire.MarshalBinary(list)
if err != nil { ... }

// dump bz:
// BINARY      HEX   NOTE
// b0000 0011  0x03  "Start struct" for `List`
// b0000 1110  0x0E  field number and type for `MyList` (1, "List")
// b0000 0011  0x03  type of element of `MyList`, a struct
// b0000 0010  0x02  length of list, uvarint for 2
//                   "Start struct" for `Item` #0 (implied)
// b0000 1000  0x08  field number and type for Number (1, "Varint)
// b0000 0010  0x02  field value varint for 1
// b0000 0100  0x04  "End struct" for `Item` #0
//                   "Start struct" for `Item` #1 (implied)
// b0000 1000  0x08  field number and type for Number (1, "Varint)
// b0000 0110  0x06  field value varint for 3
// b0000 0100  0x04  "End struct" for `Item` #1
// b0000 0100  0x04  "End struct" for `List`
```

A list of `n` elements where the elements are list of structs is encoded by the
byte `0x06` followed by the uvarint encoding of `n`, followed by the binary
encoding of each element, which starts with the byte `0x03` followed by the
uvarint encoding of `m` (the size of the first child list).  Each struct
element is encoded as if followed by a `Start struct` field key, as in the
previous example.

```golang
type Item struct {
	Number int
}
type List []Item
type ListOfLists struct {
	MyLists []List
}

llist := ListOfLists{List{[]Item{Item{1}, Item{3}}}}

bz, err := wire.MarshalBinary(llist)
if err != nil { ... }

// dump bz:
// BINARY      HEX   NOTE
// b0000 0011  0x03  "Start struct" for `ListOfLists`
// b0000 1110  0x0E  field number and type for `MyLists` (1, "[]List")
// b0000 0110  0x06  type of element of `MyLists`, a list
// b0000 0001  0x01  length of list, uvarint for 1
//                   "List" for `List` #0 (implied)
// b0000 0011  0x03  type of list element, struct
// b0000 0010  0x02  length of list, uvarint for 2
//                   "Start struct" for `Item` #0 (implied)
// b0000 1000  0x08  field number and type for Number (1, "Varint)
// b0000 0010  0x02  field value varint for 1
// b0000 0100  0x04  "End struct" for `Item` #0
//                   "Start struct" for `Item` #1 (implied)
// b0000 1000  0x08  field number and type for Number (1, "Varint)
// b0000 0110  0x06  field value varint for 3
// b0000 0100  0x04  "End struct" for `Item` #1
//                   "End list" for `List` #0 implied
// b0000 0100  0x04  "End struct" for `ListOfLists`
```

Unlike fields of a struct where nil pointers are denoted by the absence of
encoding, elements of a list are encoded without an index or key.  They are
just encoded one after the other, with the first typ3 byte implied by the
list's declared element typ4 byte (technically the first byte of the List).

To declare that the List may contain nil elements, the Lists's element typ4
byte should set the 4th least-significant bit (the "pointer bit") to 1.  If
(and only if) the pointer bit is 1, elements are prefixed by a 0x00 byte to
denote a non-nil item, or a 0x01 byte to denote a nil item.  Note that the byte
values are flipped (typically 0 is used to denote nil).  This is to open the
possibility of supporting sparse encoding of nil lists in the future by
encoding the number of nil items to skip as a uvarint.

```golang
type Item struct {
	Number int
}
type List struct {
	MyList []*Item
}

list := List{[]*Item{Item{1}, nil}}

bz, err := wire.MarshalBinary(list)
if err != nil { ... }

// dump bz:
// BINARY      HEX   NOTE
// b0000 0011  0x03  "Start struct" for `List`
// b0000 1110  0x0E  field number and type for `MyList` (1, "List")
// b0000 1011  0x03  type of element of `MyList`, a struct; and "nullable" elements
// b0000 0010  0x02  length of list, uvarint for 2
// b0000 0000  0x00  byte to denote non-nil element
//                   "Start struct" for `Item` #0 (implied)
// b0000 1000  0x08  field number and type for Number (1, "Varint)
// b0000 0010  0x02  field value varint for 1
// b0000 0100  0x04  "End struct" for `Item` #0
// b0000 0001  0x01  byte to denote nil element
// b0000 0100  0x04  "End struct" for `List`
```

### Interfaces

Finally, Protobuf's "oneof" gets a facelift.  Instead of "oneof", Wire has
Interfaces.

An interface value is typically a struct, but it doesn't need to be.  If the value
of the interface isn't nil, the prefix or disfix bytes are followed by a typ3
byte of the value type, so a scanner can recursively traverse the fields and
elements of the value.


## Wire in other langauges

Coming soon...
