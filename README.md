

# data
`import "github.com/tendermint/go-data"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Data is designed to provide a standard interface and helper functions to
easily allow serialization and deserialization of your data structures
in both binary and json representations.

This is commonly needed for interpreting transactions or stored data in the
abci app, as well as accepting json input in the light-client proxy. If we
can standardize how we pass data around the app, we can also allow more
extensions, like data storage that can interpret the meaning of the []byte
passed in, and use that to index multiple fields for example.

Serialization of data is pretty automatic using standard json and go-wire
encoders.  The main issue is deserialization, especially when using interfaces
where there are many possible concrete types.

go-wire handles this by registering the types and providing a custom
deserializer:


	var _ = wire.RegisterInterface(
	  struct{ PubKey }{},
	  wire.ConcreteType{PubKeyEd25519{}, PubKeyTypeEd25519},
	  wire.ConcreteType{PubKeySecp256k1{}, PubKeyTypeSecp256k1},
	)
	
	func PubKeyFromBytes(pubKeyBytes []byte) (pubKey PubKey, err error) {
	  err = wire.ReadBinaryBytes(pubKeyBytes, &pubKey)
	  return
	}
	
	func (pubKey PubKeyEd25519) Bytes() []byte {
	  return wire.BinaryBytes(struct{ PubKey }{pubKey})
	}

This prepends a type-byte to the binary representation upon serialization and
using that byte to switch between various representations on deserialization.
go-wire also supports something similar in json, but it leads to kind of ugly
mixed-types arrays, and requires using the go-wire json parser, which is
limited relative to the standard library encoding/json library.

In json, the typical idiom is to use a type string and message data:


	{
	  "type": "this part tells you how to interpret the message",
	  "msg": ...the actual message is here, in some kind of json...
	}

I took inspiration from two blog posts, that demonstrate how to use this
to build (de)serialization in a go-wire like way.

* <a href="http://eagain.net/articles/go-dynamic-json/">http://eagain.net/articles/go-dynamic-json/</a>
* <a href="http://eagain.net/articles/go-json-kind/">http://eagain.net/articles/go-json-kind/</a>

This package unifies these two in a single Mapper.

You app needs to do three things to take full advantage of this:

1. For every interface you wish to serialize, define a holder struct


	with some helper methods, like FooerS wraps Fooer in common_test.go

2. In all structs that include this interface, include the wrapping struct


	instead.  Functionally, this also fulfills the interface, so except for
	setting it or casting it to a sub-type it works the same.

3. Register the interface implementations as in the last init of common_test.go


	If you are currently using go-wire, you should be doing this already

The benefits here is you can now run any of the following methods, both for
efficient storage in our go app, and a common format for rpc / humans.


	orig := FooerS{foo}
	
	// read/write binary a la tendermint/go-wire
	bparsed := FooerS{}
	err := wire.ReadBinaryBytes(
	  wire.BinaryBytes(orig), &bparsed)
	
	// read/write json a la encoding/json
	jparsed := FooerS{}
	j, err := json.MarshalIndent(orig, "", "\t")
	err = json.Unmarshal(j, &jparsed)




## <a name="pkg-index">Index</a>
* [type BinaryMapper](#BinaryMapper)
  * [func NewBinaryMapper(base interface{}) *BinaryMapper](#NewBinaryMapper)
  * [func (m *BinaryMapper) RegisterInterface(kind string, b byte, data interface{})](#BinaryMapper.RegisterInterface)
* [type JSONMapper](#JSONMapper)
  * [func NewJSONMapper(base interface{}) *JSONMapper](#NewJSONMapper)
  * [func (m *JSONMapper) FromJSON(data []byte) (interface{}, error)](#JSONMapper.FromJSON)
  * [func (m *JSONMapper) RegisterInterface(kind string, b byte, data interface{})](#JSONMapper.RegisterInterface)
  * [func (m *JSONMapper) ToJSON(data interface{}) ([]byte, error)](#JSONMapper.ToJSON)
* [type Mapper](#Mapper)
  * [func NewMapper(base interface{}) Mapper](#NewMapper)
  * [func (m Mapper) RegisterInterface(kind string, b byte, data interface{}) Mapper](#Mapper.RegisterInterface)


#### <a name="pkg-files">Package files</a>
[binary.go](/src/github.com/tendermint/go-data/binary.go) [docs.go](/src/github.com/tendermint/go-data/docs.go) [json.go](/src/github.com/tendermint/go-data/json.go) [wrapper.go](/src/github.com/tendermint/go-data/wrapper.go) 






## <a name="BinaryMapper">type</a> [BinaryMapper](/src/target/binary.go?s=59:133#L1)
``` go
type BinaryMapper struct {
    // contains filtered or unexported fields
}
```






### <a name="NewBinaryMapper">func</a> [NewBinaryMapper](/src/target/binary.go?s=135:187#L1)
``` go
func NewBinaryMapper(base interface{}) *BinaryMapper
```




### <a name="BinaryMapper.RegisterInterface">func</a> (\*BinaryMapper) [RegisterInterface](/src/target/binary.go?s=424:503#L10)
``` go
func (m *BinaryMapper) RegisterInterface(kind string, b byte, data interface{})
```
RegisterInterface allows you to register multiple concrete types.

We call wire.RegisterInterface with the entire (growing list) each time,
as we do not know when the end is near.




## <a name="JSONMapper">type</a> [JSONMapper](/src/target/json.go?s=80:178#L1)
``` go
type JSONMapper struct {
    // contains filtered or unexported fields
}
```






### <a name="NewJSONMapper">func</a> [NewJSONMapper](/src/target/json.go?s=180:228#L5)
``` go
func NewJSONMapper(base interface{}) *JSONMapper
```




### <a name="JSONMapper.FromJSON">func</a> (\*JSONMapper) [FromJSON](/src/target/json.go?s=1096:1159#L39)
``` go
func (m *JSONMapper) FromJSON(data []byte) (interface{}, error)
```



### <a name="JSONMapper.RegisterInterface">func</a> (\*JSONMapper) [RegisterInterface](/src/target/json.go?s=459:536#L15)
``` go
func (m *JSONMapper) RegisterInterface(kind string, b byte, data interface{})
```
RegisterInterface allows you to register multiple concrete types.

Returns itself to allow calls to be chained




### <a name="JSONMapper.ToJSON">func</a> (\*JSONMapper) [ToJSON](/src/target/json.go?s=1490:1551#L57)
``` go
func (m *JSONMapper) ToJSON(data interface{}) ([]byte, error)
```



## <a name="Mapper">type</a> [Mapper](/src/target/wrapper.go?s=14:64#L1)
``` go
type Mapper struct {
    *JSONMapper
    *BinaryMapper
}
```






### <a name="NewMapper">func</a> [NewMapper](/src/target/wrapper.go?s=66:105#L1)
``` go
func NewMapper(base interface{}) Mapper
```




### <a name="Mapper.RegisterInterface">func</a> (Mapper) [RegisterInterface](/src/target/wrapper.go?s=206:285#L5)
``` go
func (m Mapper) RegisterInterface(kind string, b byte, data interface{}) Mapper
```







- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
