package tmencoding

type TMEncoderFacadeIntr interface {
	TMEncoderEasyIntr
	Bytes() []byte
}
