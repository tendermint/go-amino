package tmencoding

// Most recommented interface for common use-cases due to superior safety
// and gentler learning curve.
type TMEncoderFacadeIntr interface {
	TMEncoderColumnarBuilder
	Bytes() []byte
}
