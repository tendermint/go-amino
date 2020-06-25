package main

import (
	pbpkg "github.com/tendermint/go-amino/genproto/example/pb"
	proto "google.golang.org/protobuf/proto"
	amino "github.com/tendermint/go-amino"
)

func (o StructA) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	pb, err := new(pbpkg.StructA), error(nil)
	pb.FieldC = o.FieldC
	pb.FieldD = o.FieldD
	msg = pb
	return
}
func (o *StructA) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	err, pb := error(nil), *pbpkg.StructA(nil)
	pb = msg.(*pbpkg.StructA)
	pb.FieldC = o.FieldC
	pb.FieldD = o.FieldD
	msg = pb
	return
}
func (_ StructA) GetTypeURL() (typeURL string) {
	return "/main.StructA"
}
func (o StructB) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	pb, err := new(pbpkg.StructB), error(nil)
	pb.FieldC = o.FieldC
	pb.FieldD = o.FieldD
	pb.FieldE = o.FieldE
	pb.FieldF = o.FieldF
	typeUrl := o.GetTypeUrl()
	bz, err := cdc.MarshalBinaryBare(o.FieldG)
	if err!=nil {
		return
	}
	pb.FieldG = anypb.Any{TypeUrl: typeUrl, Value: bz}
	msg = pb
	return
}
func (o *StructB) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	err, pb := error(nil), *pbpkg.StructB(nil)
	pb = msg.(*pbpkg.StructB)
	pb.FieldC = o.FieldC
	pb.FieldD = o.FieldD
	pb.FieldE = o.FieldE
	pb.FieldF = o.FieldF
	any := pb.FieldG
	typeUrl, any.TypeUrl, bz := any.Value
	err := cdc.UnmarshalBinaryAny(typeUrl, bz,  &o.FieldG)
	if err!=nil {
		return
	}
	msg = pb
	return
}
func (_ StructB) GetTypeURL() (typeURL string) {
	return "/main.StructB"
}
