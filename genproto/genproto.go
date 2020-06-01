package genproto

import (
	"reflect"

	"github.com/tendermint/go-amino"
)

// Given a codec and some reflection type, generate the Proto3 message
// (partial) schema.
//
func GenerateProto3MessageSchema(cdc *amino.Codec, rt reflect.Type) (P3Message, error) {

	// XXX
	return P3Message{}, nil
}
