package submodule

import (
	pbpkg "github.com/tendermint/go-amino/genproto/example/submodule/pb"
	proto "google.golang.org/protobuf/proto"
	amino "github.com/tendermint/go-amino"
)

func (o StructSM) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	pb, err := new(pbpkg.StructSM), error(nil)
	pb.FieldA = o.FieldA
	pb.FieldB = o.FieldB
	pb.FieldC = o.FieldC
	msg = pb
	return
}
func (o *StructSM) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	err, pb := error(nil), *pbpkg.StructSM(nil)
	pb = msg.(*pbpkg.StructSM)
	pb.FieldA = o.FieldA
	pb.FieldB = o.FieldB
	pb.FieldC = o.FieldC
	msg = pb
	return
}
func (_ StructSM) GetTypeURL() (typeURL string) {
	return "/submodule.StructSM"
}
