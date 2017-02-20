package data

type Mapper struct {
	*JSONMapper
	*BinaryMapper
}

func NewMapper(base interface{}) Mapper {
	return Mapper{
		JSONMapper:   NewJSONMapper(base),
		BinaryMapper: NewBinaryMapper(base),
	}
}

func (m Mapper) RegisterInterface(kind string, b byte, data interface{}) Mapper {
	m.JSONMapper.RegisterInterface(kind, b, data)
	m.BinaryMapper.RegisterInterface(kind, b, data)
	return m
}
