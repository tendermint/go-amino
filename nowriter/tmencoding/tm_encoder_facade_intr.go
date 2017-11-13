package tmencoding

type TMEncoderFacadeIntr interface {
	TMEncoderColumnarBuilder
	Bytes() []byte
}
