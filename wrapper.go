package data

/*
Mapper is the main entry point in the package.

On init, you should call NewMapper() for each interface type you want
to support flexible de-serialization, and then
RegisterInterface() in the init() function for each implementation of these
interfaces.

Note that unlike go-wire, you can call RegisterInterface separately from
different locations with each implementation, not all in one place.
Just be careful not to use the same key or byte, of init will *panic*
*/
type Mapper struct {
	*JSONMapper
	*binaryMapper
}

// NewMapper creates a Mapper.
//
// If you have:
//   type Foo interface {....}
//   type FooS struct { Foo }
// then you should pass in FooS{} in NewMapper, and implementations of Foo
// in RegisterInterface
func NewMapper(base interface{}) Mapper {
	return Mapper{
		JSONMapper:   newJSONMapper(base),
		binaryMapper: newBinaryMapper(base),
	}
}

// RegisterInterface should be called once for each implementation of the
// interface that we wish to support.
//
// kind is the type string used in the json representation, while b is the
// type byte used in the go-wire representation. data is one instance of this
// concrete type, like Bar{}
func (m Mapper) RegisterInterface(kind string, b byte, data interface{}) Mapper {
	m.JSONMapper.registerInterface(kind, b, data)
	m.binaryMapper.registerInterface(kind, b, data)
	return m
}
