package tmarrayencoder

type TMArrayEncoderLength interface {
	TMArrayEncoder
	PrefixStatus(TMArrayEncoderLength)
}
