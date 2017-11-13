package tmencoding

// This scaffold may go away when `go-wire` refactoring completes.
type TMEncoderSmoothIntr struct {
	TMEncoderFacadeIntr
	Legacy TMEncoderFastIOWriterIntr
}
