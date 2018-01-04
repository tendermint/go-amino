package tmarrayencoder

type TMArrayEncoderUnlength interface {
	TMArrayEncoder
	PrefixStatus(TMArrayEncoderUnlength)
}
