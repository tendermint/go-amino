package submodule2

import (
	pbpkg "github.com/tendermint/go-amino/genproto/example/submodule2/pb"
	proto "google.golang.org/protobuf/proto"
	amino "github.com/tendermint/go-amino"
)

func (o StructSM2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	pb, err := new(pbpkg.StructSM2), error(nil)
	pb.FieldA = o.FieldA
	pb.FieldB = o.FieldB
	msg = pb
	return
}
func (o *StructSM2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	err, pb := error(nil), *pbpkg.StructSM2(nil)
	pb = msg.(*pbpkg.StructSM2)
	pb.FieldA = o.FieldA
	pb.FieldB = o.FieldB
	msg = pb
	return
}
func (_ StructSM2) GetTypeURL() (typeURL string) {
	return "/submodule2.StructSM2"
}
