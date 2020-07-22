package tests

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/tendermint/go-amino"
	testspb "github.com/tendermint/go-amino/tests/pb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	time "time"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func (goo EmptyStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmptyStruct
	{
		if isEmptyStructEmptyRepr(goo) {
			var pbov *testspb.EmptyStruct
			msg = pbov
			return
		}
		pbo = new(testspb.EmptyStruct)
	}
	msg = pbo
	return
}
func (goo *EmptyStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmptyStruct = msg.(*testspb.EmptyStruct)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ EmptyStruct) GetTypeURL() (typeURL string) {
	return "/tests.EmptyStruct"
}
func isEmptyStructEmptyRepr(goor EmptyStruct) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo PrimitivesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStruct
	{
		if isPrimitivesStructEmptyRepr(goo) {
			var pbov *testspb.PrimitivesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PrimitivesStruct)
		{
			pbo.Int8 = int32(goo.Int8)
		}
		{
			pbo.Int16 = int32(goo.Int16)
		}
		{
			pbo.Int32 = goo.Int32
		}
		{
			pbo.Int32Fixed = goo.Int32Fixed
		}
		{
			pbo.Int64 = goo.Int64
		}
		{
			pbo.Int64Fixed = goo.Int64Fixed
		}
		{
			pbo.Int = int64(goo.Int)
		}
		{
			pbo.Byte = uint32(goo.Byte)
		}
		{
			pbo.Uint8 = uint32(goo.Uint8)
		}
		{
			pbo.Uint16 = uint32(goo.Uint16)
		}
		{
			pbo.Uint32 = goo.Uint32
		}
		{
			pbo.Uint32Fixed = goo.Uint32Fixed
		}
		{
			pbo.Uint64 = goo.Uint64
		}
		{
			pbo.Uint64Fixed = goo.Uint64Fixed
		}
		{
			pbo.Uint = uint64(goo.Uint)
		}
		{
			pbo.Str = goo.Str
		}
		{
			goorl := len(goo.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			if !amino.IsEmptyTime(goo.Time) {
				pbo.Time = timestamppb.New(goo.Time)
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Empty.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Empty = pbom.(*testspb.EmptyStruct)
		}
	}
	msg = pbo
	return
}
func (goo *PrimitivesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStruct = msg.(*testspb.PrimitivesStruct)
	{
		if pbo != nil {
			{
				goo.Int8 = int8(pbo.Int8)
			}
			{
				goo.Int16 = int16(pbo.Int16)
			}
			{
				goo.Int32 = pbo.Int32
			}
			{
				goo.Int32Fixed = pbo.Int32Fixed
			}
			{
				goo.Int64 = pbo.Int64
			}
			{
				goo.Int64Fixed = pbo.Int64Fixed
			}
			{
				goo.Int = int(pbo.Int)
			}
			{
				goo.Byte = uint8(pbo.Byte)
			}
			{
				goo.Uint8 = uint8(pbo.Uint8)
			}
			{
				goo.Uint16 = uint16(pbo.Uint16)
			}
			{
				goo.Uint32 = pbo.Uint32
			}
			{
				goo.Uint32Fixed = pbo.Uint32Fixed
			}
			{
				goo.Uint64 = pbo.Uint64
			}
			{
				goo.Uint64Fixed = pbo.Uint64Fixed
			}
			{
				goo.Uint = uint(pbo.Uint)
			}
			{
				goo.Str = pbo.Str
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					goo.Bytes = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.Bytes = goos
				}
			}
			{
				goo.Time = pbo.Time.AsTime()
			}
			{
				if pbo.Empty != nil {
					err = goo.Empty.FromPBMessage(cdc, pbo.Empty)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PrimitivesStruct) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStruct"
}
func isPrimitivesStructEmptyRepr(goor PrimitivesStruct) (empty bool) {
	{
		empty = true
		{
			if goor.Int8 != 0 {
				return false
			}
		}
		{
			if goor.Int16 != 0 {
				return false
			}
		}
		{
			if goor.Int32 != 0 {
				return false
			}
		}
		{
			if goor.Int32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int64 != 0 {
				return false
			}
		}
		{
			if goor.Int64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int != 0 {
				return false
			}
		}
		{
			if goor.Byte != 0 {
				return false
			}
		}
		{
			if goor.Uint8 != 0 {
				return false
			}
		}
		{
			if goor.Uint16 != 0 {
				return false
			}
		}
		{
			if goor.Uint32 != 0 {
				return false
			}
		}
		{
			if goor.Uint32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint64 != 0 {
				return false
			}
		}
		{
			if goor.Uint64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint != 0 {
				return false
			}
		}
		{
			if goor.Str != "" {
				return false
			}
		}
		{
			if len(goor.Bytes) != 0 {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.Time) {
				return false
			}
		}
		{
			e := isEmptyStructEmptyRepr(goor.Empty)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo ShortArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ShortArraysStruct
	{
		if isShortArraysStructEmptyRepr(goo) {
			var pbov *testspb.ShortArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ShortArraysStruct)
		{
			goorl := len(goo.TimeAr)
			if goorl == 0 {
				pbo.TimeAr = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeAr[i]
						{
							if !amino.IsEmptyTime(goore) {
								pbos[i] = timestamppb.New(goore)
							}
						}
					}
				}
				pbo.TimeAr = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *ShortArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ShortArraysStruct = msg.(*testspb.ShortArraysStruct)
	{
		if pbo != nil {
			{
				var goos = [0]time.Time{}
				for i := 0; i < 0; i += 1 {
					{
						pboe := pbo.TimeAr[i]
						{
							goos[i] = pboe.AsTime()
						}
					}
				}
				goo.TimeAr = goos
			}
		}
	}
	return
}
func (_ ShortArraysStruct) GetTypeURL() (typeURL string) {
	return "/tests.ShortArraysStruct"
}
func isShortArraysStructEmptyRepr(goor ShortArraysStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.TimeAr) != 0 {
				return false
			}
		}
	}
	return
}
func (goo ArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ArraysStruct
	{
		if isArraysStructEmptyRepr(goo) {
			var pbov *testspb.ArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ArraysStruct)
		{
			goorl := len(goo.Int8Ar)
			if goorl == 0 {
				pbo.Int8Ar = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int8Ar = pbos
			}
		}
		{
			goorl := len(goo.Int16Ar)
			if goorl == 0 {
				pbo.Int16Ar = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int16Ar = pbos
			}
		}
		{
			goorl := len(goo.Int32Ar)
			if goorl == 0 {
				pbo.Int32Ar = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32Ar[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int32Ar = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedAr)
			if goorl == 0 {
				pbo.Int32FixedAr = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedAr[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int32FixedAr = pbos
			}
		}
		{
			goorl := len(goo.Int64Ar)
			if goorl == 0 {
				pbo.Int64Ar = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64Ar[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int64Ar = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedAr)
			if goorl == 0 {
				pbo.Int64FixedAr = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedAr[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int64FixedAr = pbos
			}
		}
		{
			goorl := len(goo.IntAr)
			if goorl == 0 {
				pbo.IntAr = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntAr[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.IntAr = pbos
			}
		}
		{
			goorl := len(goo.ByteAr)
			if goorl == 0 {
				pbo.ByteAr = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteAr[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ByteAr = pbos
			}
		}
		{
			goorl := len(goo.Uint8Ar)
			if goorl == 0 {
				pbo.Uint8Ar = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8Ar[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Uint8Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint16Ar)
			if goorl == 0 {
				pbo.Uint16Ar = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16Ar[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint16Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint32Ar)
			if goorl == 0 {
				pbo.Uint32Ar = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32Ar[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint32Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedAr)
			if goorl == 0 {
				pbo.Uint32FixedAr = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedAr[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint32FixedAr = pbos
			}
		}
		{
			goorl := len(goo.Uint64Ar)
			if goorl == 0 {
				pbo.Uint64Ar = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64Ar[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint64Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedAr)
			if goorl == 0 {
				pbo.Uint64FixedAr = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedAr[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint64FixedAr = pbos
			}
		}
		{
			goorl := len(goo.UintAr)
			if goorl == 0 {
				pbo.UintAr = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintAr[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.UintAr = pbos
			}
		}
		{
			goorl := len(goo.StrAr)
			if goorl == 0 {
				pbo.StrAr = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrAr[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.StrAr = pbos
			}
		}
		{
			goorl := len(goo.BytesAr)
			if goorl == 0 {
				pbo.BytesAr = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.BytesAr = pbos
			}
		}
		{
			goorl := len(goo.TimeAr)
			if goorl == 0 {
				pbo.TimeAr = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeAr[i]
						{
							if !amino.IsEmptyTime(goore) {
								pbos[i] = timestamppb.New(goore)
							}
						}
					}
				}
				pbo.TimeAr = pbos
			}
		}
		{
			goorl := len(goo.EmptyAr)
			if goorl == 0 {
				pbo.EmptyAr = nil
			} else {
				var pbos = make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptyAr[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*testspb.EmptyStruct)
						}
					}
				}
				pbo.EmptyAr = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *ArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ArraysStruct = msg.(*testspb.ArraysStruct)
	{
		if pbo != nil {
			{
				var goos = [4]int8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int8Ar[i]
						{
							goos[i] = int8(pboe)
						}
					}
				}
				goo.Int8Ar = goos
			}
			{
				var goos = [4]int16{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int16Ar[i]
						{
							goos[i] = int16(pboe)
						}
					}
				}
				goo.Int16Ar = goos
			}
			{
				var goos = [4]int32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int32Ar[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Int32Ar = goos
			}
			{
				var goos = [4]int32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int32FixedAr[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Int32FixedAr = goos
			}
			{
				var goos = [4]int64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int64Ar[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Int64Ar = goos
			}
			{
				var goos = [4]int64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int64FixedAr[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Int64FixedAr = goos
			}
			{
				var goos = [4]int{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.IntAr[i]
						{
							goos[i] = int(pboe)
						}
					}
				}
				goo.IntAr = goos
			}
			{
				var goos = [4]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.ByteAr[i]
						{
							goos[i] = uint8(pboe)
						}
					}
				}
				goo.ByteAr = goos
			}
			{
				var goos = [4]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint8Ar[i]
						{
							goos[i] = uint8(pboe)
						}
					}
				}
				goo.Uint8Ar = goos
			}
			{
				var goos = [4]uint16{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint16Ar[i]
						{
							goos[i] = uint16(pboe)
						}
					}
				}
				goo.Uint16Ar = goos
			}
			{
				var goos = [4]uint32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint32Ar[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Uint32Ar = goos
			}
			{
				var goos = [4]uint32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint32FixedAr[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Uint32FixedAr = goos
			}
			{
				var goos = [4]uint64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint64Ar[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Uint64Ar = goos
			}
			{
				var goos = [4]uint64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint64FixedAr[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.Uint64FixedAr = goos
			}
			{
				var goos = [4]uint{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.UintAr[i]
						{
							goos[i] = uint(pboe)
						}
					}
				}
				goo.UintAr = goos
			}
			{
				var goos = [4]string{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.StrAr[i]
						{
							goos[i] = pboe
						}
					}
				}
				goo.StrAr = goos
			}
			{
				var goos = [4][]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.BytesAr[i]
						{
							var pbol int = 0
							if pboe != nil {
								pbol = len(pboe)
							}
							if pbol == 0 {
								goos[i] = nil
							} else {
								var goos1 = make([]uint8, pbol)
								for i := 0; i < pbol; i += 1 {
									{
										pboe := pboe[i]
										{
											goos1[i] = uint8(pboe)
										}
									}
								}
								goos[i] = goos1
							}
						}
					}
				}
				goo.BytesAr = goos
			}
			{
				var goos = [4]time.Time{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.TimeAr[i]
						{
							goos[i] = pboe.AsTime()
						}
					}
				}
				goo.TimeAr = goos
			}
			{
				var goos = [4]EmptyStruct{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.EmptyAr[i]
						{
							if pboe != nil {
								err = goos[i].FromPBMessage(cdc, pboe)
								if err != nil {
									return
								}
							}
						}
					}
				}
				goo.EmptyAr = goos
			}
		}
	}
	return
}
func (_ ArraysStruct) GetTypeURL() (typeURL string) {
	return "/tests.ArraysStruct"
}
func isArraysStructEmptyRepr(goor ArraysStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.IntAr) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.UintAr) != 0 {
				return false
			}
		}
		{
			if len(goor.StrAr) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesAr) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeAr) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptyAr) != 0 {
				return false
			}
		}
	}
	return
}
func (goo ArraysArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ArraysArraysStruct
	{
		if isArraysArraysStructEmptyRepr(goo) {
			var pbov *testspb.ArraysArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ArraysArraysStruct)
		{
			goorl := len(goo.Int8ArAr)
			if goorl == 0 {
				pbo.Int8ArAr = nil
			} else {
				var pbos = make([]*testspb.Int8List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int8List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int8ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int16ArAr)
			if goorl == 0 {
				pbo.Int16ArAr = nil
			} else {
				var pbos = make([]*testspb.Int16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int16ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int32ArAr)
			if goorl == 0 {
				pbo.Int32ArAr = nil
			} else {
				var pbos = make([]*testspb.Int32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Int32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedArAr)
			if goorl == 0 {
				pbo.Int32FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed32Int32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed32Int32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.Int64ArAr)
			if goorl == 0 {
				pbo.Int64ArAr = nil
			} else {
				var pbos = make([]*testspb.Int64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Int64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedArAr)
			if goorl == 0 {
				pbo.Int64FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed64Int64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed64Int64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.IntArAr)
			if goorl == 0 {
				pbo.IntArAr = nil
			} else {
				var pbos = make([]*testspb.IntList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.IntList{Value: pbos1}
							}
						}
					}
				}
				pbo.IntArAr = pbos
			}
		}
		{
			goorl := len(goo.ByteArAr)
			if goorl == 0 {
				pbo.ByteArAr = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.ByteArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint8ArAr)
			if goorl == 0 {
				pbo.Uint8ArAr = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.Uint8ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint16ArAr)
			if goorl == 0 {
				pbo.Uint16ArAr = nil
			} else {
				var pbos = make([]*testspb.Uint16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.Uint16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint16ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint32ArAr)
			if goorl == 0 {
				pbo.Uint32ArAr = nil
			} else {
				var pbos = make([]*testspb.Uint32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Uint32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedArAr)
			if goorl == 0 {
				pbo.Uint32FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed32Uint32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed32Uint32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint64ArAr)
			if goorl == 0 {
				pbo.Uint64ArAr = nil
			} else {
				var pbos = make([]*testspb.Uint64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Uint64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedArAr)
			if goorl == 0 {
				pbo.Uint64FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed64Uint64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed64Uint64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.UintArAr)
			if goorl == 0 {
				pbo.UintArAr = nil
			} else {
				var pbos = make([]*testspb.UintList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.UintList{Value: pbos1}
							}
						}
					}
				}
				pbo.UintArAr = pbos
			}
		}
		{
			goorl := len(goo.StrArAr)
			if goorl == 0 {
				pbo.StrArAr = nil
			} else {
				var pbos = make([]*testspb.StringList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]string, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.StringList{Value: pbos1}
							}
						}
					}
				}
				pbo.StrArAr = pbos
			}
		}
		{
			goorl := len(goo.BytesArAr)
			if goorl == 0 {
				pbo.BytesArAr = nil
			} else {
				var pbos = make([]*testspb.BytesList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([][]byte, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											goorl2 := len(goore)
											if goorl2 == 0 {
												pbos1[i] = nil
											} else {
												var pbos2 = make([]uint8, goorl2)
												for i := 0; i < goorl2; i += 1 {
													{
														goore := goore[i]
														{
															pbos2[i] = byte(goore)
														}
													}
												}
												pbos1[i] = pbos2
											}
										}
									}
								}
								pbos[i] = &testspb.BytesList{Value: pbos1}
							}
						}
					}
				}
				pbo.BytesArAr = pbos
			}
		}
		{
			goorl := len(goo.TimeArAr)
			if goorl == 0 {
				pbo.TimeArAr = nil
			} else {
				var pbos = make([]*testspb.TimeList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*timestamppb.Timestamp, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											if !amino.IsEmptyTime(goore) {
												pbos1[i] = timestamppb.New(goore)
											}
										}
									}
								}
								pbos[i] = &testspb.TimeList{Value: pbos1}
							}
						}
					}
				}
				pbo.TimeArAr = pbos
			}
		}
		{
			goorl := len(goo.EmptyArAr)
			if goorl == 0 {
				pbo.EmptyArAr = nil
			} else {
				var pbos = make([]*testspb.EmptyStructList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptyArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*testspb.EmptyStruct, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbom := proto.Message(nil)
											pbom, err = goore.ToPBMessage(cdc)
											if err != nil {
												return
											}
											pbos1[i] = pbom.(*testspb.EmptyStruct)
										}
									}
								}
								pbos[i] = &testspb.EmptyStructList{Value: pbos1}
							}
						}
					}
				}
				pbo.EmptyArAr = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *ArraysArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ArraysArraysStruct = msg.(*testspb.ArraysArraysStruct)
	{
		if pbo != nil {
			{
				var goos = [2][2]int8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int8ArAr[i]
						{
							var goos1 = [2]int8{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = int8(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Int8ArAr = goos
			}
			{
				var goos = [2][2]int16{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int16ArAr[i]
						{
							var goos1 = [2]int16{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = int16(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Int16ArAr = goos
			}
			{
				var goos = [2][2]int32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int32ArAr[i]
						{
							var goos1 = [2]int32{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Int32ArAr = goos
			}
			{
				var goos = [2][2]int32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int32FixedArAr[i]
						{
							var goos1 = [2]int32{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Int32FixedArAr = goos
			}
			{
				var goos = [2][2]int64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int64ArAr[i]
						{
							var goos1 = [2]int64{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Int64ArAr = goos
			}
			{
				var goos = [2][2]int64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int64FixedArAr[i]
						{
							var goos1 = [2]int64{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Int64FixedArAr = goos
			}
			{
				var goos = [2][2]int{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.IntArAr[i]
						{
							var goos1 = [2]int{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = int(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.IntArAr = goos
			}
			{
				var goos = [2][2]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.ByteArAr[i]
						{
							var goos1 = [2]uint8{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe[i]
									{
										goos1[i] = uint8(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.ByteArAr = goos
			}
			{
				var goos = [2][2]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint8ArAr[i]
						{
							var goos1 = [2]uint8{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe[i]
									{
										goos1[i] = uint8(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Uint8ArAr = goos
			}
			{
				var goos = [2][2]uint16{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint16ArAr[i]
						{
							var goos1 = [2]uint16{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = uint16(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Uint16ArAr = goos
			}
			{
				var goos = [2][2]uint32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint32ArAr[i]
						{
							var goos1 = [2]uint32{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Uint32ArAr = goos
			}
			{
				var goos = [2][2]uint32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint32FixedArAr[i]
						{
							var goos1 = [2]uint32{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Uint32FixedArAr = goos
			}
			{
				var goos = [2][2]uint64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint64ArAr[i]
						{
							var goos1 = [2]uint64{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Uint64ArAr = goos
			}
			{
				var goos = [2][2]uint64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint64FixedArAr[i]
						{
							var goos1 = [2]uint64{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.Uint64FixedArAr = goos
			}
			{
				var goos = [2][2]uint{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.UintArAr[i]
						{
							var goos1 = [2]uint{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = uint(pboe)
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.UintArAr = goos
			}
			{
				var goos = [2][2]string{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.StrArAr[i]
						{
							var goos1 = [2]string{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.StrArAr = goos
			}
			{
				var goos = [2][2][]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.BytesArAr[i]
						{
							var goos1 = [2][]uint8{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										var pbol int = 0
										if pboe != nil {
											pbol = len(pboe)
										}
										if pbol == 0 {
											goos1[i] = nil
										} else {
											var goos2 = make([]uint8, pbol)
											for i := 0; i < pbol; i += 1 {
												{
													pboe := pboe[i]
													{
														goos2[i] = uint8(pboe)
													}
												}
											}
											goos1[i] = goos2
										}
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.BytesArAr = goos
			}
			{
				var goos = [2][2]time.Time{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.TimeArAr[i]
						{
							var goos1 = [2]time.Time{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										goos1[i] = pboe.AsTime()
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.TimeArAr = goos
			}
			{
				var goos = [2][2]EmptyStruct{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.EmptyArAr[i]
						{
							var goos1 = [2]EmptyStruct{}
							for i := 0; i < 2; i += 1 {
								{
									pboe := pboe.Value[i]
									{
										if pboe != nil {
											err = goos1[i].FromPBMessage(cdc, pboe)
											if err != nil {
												return
											}
										}
									}
								}
							}
							goos[i] = goos1
						}
					}
				}
				goo.EmptyArAr = goos
			}
		}
	}
	return
}
func (_ ArraysArraysStruct) GetTypeURL() (typeURL string) {
	return "/tests.ArraysArraysStruct"
}
func isArraysArraysStructEmptyRepr(goor ArraysArraysStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.IntArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.UintArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.StrArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptyArAr) != 0 {
				return false
			}
		}
	}
	return
}
func (goo SlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.SlicesStruct
	{
		if isSlicesStructEmptyRepr(goo) {
			var pbov *testspb.SlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.SlicesStruct)
		{
			goorl := len(goo.Int8Sl)
			if goorl == 0 {
				pbo.Int8Sl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int8Sl = pbos
			}
		}
		{
			goorl := len(goo.Int16Sl)
			if goorl == 0 {
				pbo.Int16Sl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int16Sl = pbos
			}
		}
		{
			goorl := len(goo.Int32Sl)
			if goorl == 0 {
				pbo.Int32Sl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32Sl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int32Sl = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedSl)
			if goorl == 0 {
				pbo.Int32FixedSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedSl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int32FixedSl = pbos
			}
		}
		{
			goorl := len(goo.Int64Sl)
			if goorl == 0 {
				pbo.Int64Sl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64Sl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int64Sl = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedSl)
			if goorl == 0 {
				pbo.Int64FixedSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedSl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Int64FixedSl = pbos
			}
		}
		{
			goorl := len(goo.IntSl)
			if goorl == 0 {
				pbo.IntSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntSl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.IntSl = pbos
			}
		}
		{
			goorl := len(goo.ByteSl)
			if goorl == 0 {
				pbo.ByteSl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteSl[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ByteSl = pbos
			}
		}
		{
			goorl := len(goo.Uint8Sl)
			if goorl == 0 {
				pbo.Uint8Sl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8Sl[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Uint8Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint16Sl)
			if goorl == 0 {
				pbo.Uint16Sl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16Sl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint16Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint32Sl)
			if goorl == 0 {
				pbo.Uint32Sl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32Sl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint32Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedSl)
			if goorl == 0 {
				pbo.Uint32FixedSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedSl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint32FixedSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64Sl)
			if goorl == 0 {
				pbo.Uint64Sl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64Sl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint64Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedSl)
			if goorl == 0 {
				pbo.Uint64FixedSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedSl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.Uint64FixedSl = pbos
			}
		}
		{
			goorl := len(goo.UintSl)
			if goorl == 0 {
				pbo.UintSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintSl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.UintSl = pbos
			}
		}
		{
			goorl := len(goo.StrSl)
			if goorl == 0 {
				pbo.StrSl = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrSl[i]
						{
							pbos[i] = goore
						}
					}
				}
				pbo.StrSl = pbos
			}
		}
		{
			goorl := len(goo.BytesSl)
			if goorl == 0 {
				pbo.BytesSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.BytesSl = pbos
			}
		}
		{
			goorl := len(goo.TimeSl)
			if goorl == 0 {
				pbo.TimeSl = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeSl[i]
						{
							if !amino.IsEmptyTime(goore) {
								pbos[i] = timestamppb.New(goore)
							}
						}
					}
				}
				pbo.TimeSl = pbos
			}
		}
		{
			goorl := len(goo.EmptySl)
			if goorl == 0 {
				pbo.EmptySl = nil
			} else {
				var pbos = make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptySl[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*testspb.EmptyStruct)
						}
					}
				}
				pbo.EmptySl = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *SlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.SlicesStruct = msg.(*testspb.SlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8Sl != nil {
					pbol = len(pbo.Int8Sl)
				}
				if pbol == 0 {
					goo.Int8Sl = nil
				} else {
					var goos = make([]int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8Sl[i]
							{
								goos[i] = int8(pboe)
							}
						}
					}
					goo.Int8Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int16Sl != nil {
					pbol = len(pbo.Int16Sl)
				}
				if pbol == 0 {
					goo.Int16Sl = nil
				} else {
					var goos = make([]int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16Sl[i]
							{
								goos[i] = int16(pboe)
							}
						}
					}
					goo.Int16Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int32Sl != nil {
					pbol = len(pbo.Int32Sl)
				}
				if pbol == 0 {
					goo.Int32Sl = nil
				} else {
					var goos = make([]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32Sl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Int32Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedSl != nil {
					pbol = len(pbo.Int32FixedSl)
				}
				if pbol == 0 {
					goo.Int32FixedSl = nil
				} else {
					var goos = make([]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedSl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Int32FixedSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int64Sl != nil {
					pbol = len(pbo.Int64Sl)
				}
				if pbol == 0 {
					goo.Int64Sl = nil
				} else {
					var goos = make([]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64Sl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Int64Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedSl != nil {
					pbol = len(pbo.Int64FixedSl)
				}
				if pbol == 0 {
					goo.Int64FixedSl = nil
				} else {
					var goos = make([]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedSl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Int64FixedSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.IntSl != nil {
					pbol = len(pbo.IntSl)
				}
				if pbol == 0 {
					goo.IntSl = nil
				} else {
					var goos = make([]int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntSl[i]
							{
								goos[i] = int(pboe)
							}
						}
					}
					goo.IntSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.ByteSl != nil {
					pbol = len(pbo.ByteSl)
				}
				if pbol == 0 {
					goo.ByteSl = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ByteSl[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.ByteSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8Sl != nil {
					pbol = len(pbo.Uint8Sl)
				}
				if pbol == 0 {
					goo.Uint8Sl = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8Sl[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.Uint8Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16Sl != nil {
					pbol = len(pbo.Uint16Sl)
				}
				if pbol == 0 {
					goo.Uint16Sl = nil
				} else {
					var goos = make([]uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16Sl[i]
							{
								goos[i] = uint16(pboe)
							}
						}
					}
					goo.Uint16Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32Sl != nil {
					pbol = len(pbo.Uint32Sl)
				}
				if pbol == 0 {
					goo.Uint32Sl = nil
				} else {
					var goos = make([]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32Sl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Uint32Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedSl != nil {
					pbol = len(pbo.Uint32FixedSl)
				}
				if pbol == 0 {
					goo.Uint32FixedSl = nil
				} else {
					var goos = make([]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedSl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Uint32FixedSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64Sl != nil {
					pbol = len(pbo.Uint64Sl)
				}
				if pbol == 0 {
					goo.Uint64Sl = nil
				} else {
					var goos = make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64Sl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Uint64Sl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedSl != nil {
					pbol = len(pbo.Uint64FixedSl)
				}
				if pbol == 0 {
					goo.Uint64FixedSl = nil
				} else {
					var goos = make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedSl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.Uint64FixedSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.UintSl != nil {
					pbol = len(pbo.UintSl)
				}
				if pbol == 0 {
					goo.UintSl = nil
				} else {
					var goos = make([]uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintSl[i]
							{
								goos[i] = uint(pboe)
							}
						}
					}
					goo.UintSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.StrSl != nil {
					pbol = len(pbo.StrSl)
				}
				if pbol == 0 {
					goo.StrSl = nil
				} else {
					var goos = make([]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrSl[i]
							{
								goos[i] = pboe
							}
						}
					}
					goo.StrSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.BytesSl != nil {
					pbol = len(pbo.BytesSl)
				}
				if pbol == 0 {
					goo.BytesSl = nil
				} else {
					var goos = make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe[i]
											{
												goos1[i] = uint8(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.BytesSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.TimeSl != nil {
					pbol = len(pbo.TimeSl)
				}
				if pbol == 0 {
					goo.TimeSl = nil
				} else {
					var goos = make([]time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimeSl[i]
							{
								goos[i] = pboe.AsTime()
							}
						}
					}
					goo.TimeSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.EmptySl != nil {
					pbol = len(pbo.EmptySl)
				}
				if pbol == 0 {
					goo.EmptySl = nil
				} else {
					var goos = make([]EmptyStruct, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.EmptySl[i]
							{
								if pboe != nil {
									err = goos[i].FromPBMessage(cdc, pboe)
									if err != nil {
										return
									}
								}
							}
						}
					}
					goo.EmptySl = goos
				}
			}
		}
	}
	return
}
func (_ SlicesStruct) GetTypeURL() (typeURL string) {
	return "/tests.SlicesStruct"
}
func isSlicesStructEmptyRepr(goor SlicesStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.IntSl) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.UintSl) != 0 {
				return false
			}
		}
		{
			if len(goor.StrSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesSl) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeSl) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptySl) != 0 {
				return false
			}
		}
	}
	return
}
func (goo SlicesSlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.SlicesSlicesStruct
	{
		if isSlicesSlicesStructEmptyRepr(goo) {
			var pbov *testspb.SlicesSlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.SlicesSlicesStruct)
		{
			goorl := len(goo.Int8SlSl)
			if goorl == 0 {
				pbo.Int8SlSl = nil
			} else {
				var pbos = make([]*testspb.Int8List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int8List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int8SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int16SlSl)
			if goorl == 0 {
				pbo.Int16SlSl = nil
			} else {
				var pbos = make([]*testspb.Int16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int16SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int32SlSl)
			if goorl == 0 {
				pbo.Int32SlSl = nil
			} else {
				var pbos = make([]*testspb.Int32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Int32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedSlSl)
			if goorl == 0 {
				pbo.Int32FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed32Int32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed32Int32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.Int64SlSl)
			if goorl == 0 {
				pbo.Int64SlSl = nil
			} else {
				var pbos = make([]*testspb.Int64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Int64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedSlSl)
			if goorl == 0 {
				pbo.Int64FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed64Int64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed64Int64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.IntSlSl)
			if goorl == 0 {
				pbo.IntSlSl = nil
			} else {
				var pbos = make([]*testspb.IntList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.IntList{Value: pbos1}
							}
						}
					}
				}
				pbo.IntSlSl = pbos
			}
		}
		{
			goorl := len(goo.ByteSlSl)
			if goorl == 0 {
				pbo.ByteSlSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.ByteSlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint8SlSl)
			if goorl == 0 {
				pbo.Uint8SlSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.Uint8SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint16SlSl)
			if goorl == 0 {
				pbo.Uint16SlSl = nil
			} else {
				var pbos = make([]*testspb.Uint16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.Uint16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint16SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint32SlSl)
			if goorl == 0 {
				pbo.Uint32SlSl = nil
			} else {
				var pbos = make([]*testspb.Uint32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Uint32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedSlSl)
			if goorl == 0 {
				pbo.Uint32FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed32Uint32List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed32Uint32List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64SlSl)
			if goorl == 0 {
				pbo.Uint64SlSl = nil
			} else {
				var pbos = make([]*testspb.Uint64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Uint64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedSlSl)
			if goorl == 0 {
				pbo.Uint64FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed64Uint64List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.Fixed64Uint64List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.UintSlSl)
			if goorl == 0 {
				pbo.UintSlSl = nil
			} else {
				var pbos = make([]*testspb.UintList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.UintList{Value: pbos1}
							}
						}
					}
				}
				pbo.UintSlSl = pbos
			}
		}
		{
			goorl := len(goo.StrSlSl)
			if goorl == 0 {
				pbo.StrSlSl = nil
			} else {
				var pbos = make([]*testspb.StringList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]string, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = goore
										}
									}
								}
								pbos[i] = &testspb.StringList{Value: pbos1}
							}
						}
					}
				}
				pbo.StrSlSl = pbos
			}
		}
		{
			goorl := len(goo.BytesSlSl)
			if goorl == 0 {
				pbo.BytesSlSl = nil
			} else {
				var pbos = make([]*testspb.BytesList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([][]byte, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											goorl2 := len(goore)
											if goorl2 == 0 {
												pbos1[i] = nil
											} else {
												var pbos2 = make([]uint8, goorl2)
												for i := 0; i < goorl2; i += 1 {
													{
														goore := goore[i]
														{
															pbos2[i] = byte(goore)
														}
													}
												}
												pbos1[i] = pbos2
											}
										}
									}
								}
								pbos[i] = &testspb.BytesList{Value: pbos1}
							}
						}
					}
				}
				pbo.BytesSlSl = pbos
			}
		}
		{
			goorl := len(goo.TimeSlSl)
			if goorl == 0 {
				pbo.TimeSlSl = nil
			} else {
				var pbos = make([]*testspb.TimeList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*timestamppb.Timestamp, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											if !amino.IsEmptyTime(goore) {
												pbos1[i] = timestamppb.New(goore)
											}
										}
									}
								}
								pbos[i] = &testspb.TimeList{Value: pbos1}
							}
						}
					}
				}
				pbo.TimeSlSl = pbos
			}
		}
		{
			goorl := len(goo.EmptySlSl)
			if goorl == 0 {
				pbo.EmptySlSl = nil
			} else {
				var pbos = make([]*testspb.EmptyStructList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptySlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*testspb.EmptyStruct, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbom := proto.Message(nil)
											pbom, err = goore.ToPBMessage(cdc)
											if err != nil {
												return
											}
											pbos1[i] = pbom.(*testspb.EmptyStruct)
										}
									}
								}
								pbos[i] = &testspb.EmptyStructList{Value: pbos1}
							}
						}
					}
				}
				pbo.EmptySlSl = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *SlicesSlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.SlicesSlicesStruct = msg.(*testspb.SlicesSlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8SlSl != nil {
					pbol = len(pbo.Int8SlSl)
				}
				if pbol == 0 {
					goo.Int8SlSl = nil
				} else {
					var goos = make([][]int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = int8(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Int8SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int16SlSl != nil {
					pbol = len(pbo.Int16SlSl)
				}
				if pbol == 0 {
					goo.Int16SlSl = nil
				} else {
					var goos = make([][]int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int16, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = int16(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Int16SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int32SlSl != nil {
					pbol = len(pbo.Int32SlSl)
				}
				if pbol == 0 {
					goo.Int32SlSl = nil
				} else {
					var goos = make([][]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int32, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Int32SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedSlSl != nil {
					pbol = len(pbo.Int32FixedSlSl)
				}
				if pbol == 0 {
					goo.Int32FixedSlSl = nil
				} else {
					var goos = make([][]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int32, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Int32FixedSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int64SlSl != nil {
					pbol = len(pbo.Int64SlSl)
				}
				if pbol == 0 {
					goo.Int64SlSl = nil
				} else {
					var goos = make([][]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int64, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Int64SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedSlSl != nil {
					pbol = len(pbo.Int64FixedSlSl)
				}
				if pbol == 0 {
					goo.Int64FixedSlSl = nil
				} else {
					var goos = make([][]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int64, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Int64FixedSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.IntSlSl != nil {
					pbol = len(pbo.IntSlSl)
				}
				if pbol == 0 {
					goo.IntSlSl = nil
				} else {
					var goos = make([][]int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]int, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = int(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.IntSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.ByteSlSl != nil {
					pbol = len(pbo.ByteSlSl)
				}
				if pbol == 0 {
					goo.ByteSlSl = nil
				} else {
					var goos = make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ByteSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe[i]
											{
												goos1[i] = uint8(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.ByteSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8SlSl != nil {
					pbol = len(pbo.Uint8SlSl)
				}
				if pbol == 0 {
					goo.Uint8SlSl = nil
				} else {
					var goos = make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe[i]
											{
												goos1[i] = uint8(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Uint8SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16SlSl != nil {
					pbol = len(pbo.Uint16SlSl)
				}
				if pbol == 0 {
					goo.Uint16SlSl = nil
				} else {
					var goos = make([][]uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint16, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = uint16(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Uint16SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32SlSl != nil {
					pbol = len(pbo.Uint32SlSl)
				}
				if pbol == 0 {
					goo.Uint32SlSl = nil
				} else {
					var goos = make([][]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint32, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Uint32SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedSlSl != nil {
					pbol = len(pbo.Uint32FixedSlSl)
				}
				if pbol == 0 {
					goo.Uint32FixedSlSl = nil
				} else {
					var goos = make([][]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint32, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Uint32FixedSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64SlSl != nil {
					pbol = len(pbo.Uint64SlSl)
				}
				if pbol == 0 {
					goo.Uint64SlSl = nil
				} else {
					var goos = make([][]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64SlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint64, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Uint64SlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedSlSl != nil {
					pbol = len(pbo.Uint64FixedSlSl)
				}
				if pbol == 0 {
					goo.Uint64FixedSlSl = nil
				} else {
					var goos = make([][]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint64, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.Uint64FixedSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.UintSlSl != nil {
					pbol = len(pbo.UintSlSl)
				}
				if pbol == 0 {
					goo.UintSlSl = nil
				} else {
					var goos = make([][]uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]uint, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = uint(pboe)
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.UintSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.StrSlSl != nil {
					pbol = len(pbo.StrSlSl)
				}
				if pbol == 0 {
					goo.StrSlSl = nil
				} else {
					var goos = make([][]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]string, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.StrSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.BytesSlSl != nil {
					pbol = len(pbo.BytesSlSl)
				}
				if pbol == 0 {
					goo.BytesSlSl = nil
				} else {
					var goos = make([][][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([][]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												var pbol2 int = 0
												if pboe != nil {
													pbol2 = len(pboe)
												}
												if pbol2 == 0 {
													goos1[i] = nil
												} else {
													var goos2 = make([]uint8, pbol2)
													for i := 0; i < pbol2; i += 1 {
														{
															pboe := pboe[i]
															{
																goos2[i] = uint8(pboe)
															}
														}
													}
													goos1[i] = goos2
												}
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.BytesSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.TimeSlSl != nil {
					pbol = len(pbo.TimeSlSl)
				}
				if pbol == 0 {
					goo.TimeSlSl = nil
				} else {
					var goos = make([][]time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimeSlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]time.Time, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												goos1[i] = pboe.AsTime()
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.TimeSlSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.EmptySlSl != nil {
					pbol = len(pbo.EmptySlSl)
				}
				if pbol == 0 {
					goo.EmptySlSl = nil
				} else {
					var goos = make([][]EmptyStruct, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.EmptySlSl[i]
							{
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe.Value)
								}
								if pbol1 == 0 {
									goos[i] = nil
								} else {
									var goos1 = make([]EmptyStruct, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe.Value[i]
											{
												if pboe != nil {
													err = goos1[i].FromPBMessage(cdc, pboe)
													if err != nil {
														return
													}
												}
											}
										}
									}
									goos[i] = goos1
								}
							}
						}
					}
					goo.EmptySlSl = goos
				}
			}
		}
	}
	return
}
func (_ SlicesSlicesStruct) GetTypeURL() (typeURL string) {
	return "/tests.SlicesSlicesStruct"
}
func isSlicesSlicesStructEmptyRepr(goor SlicesSlicesStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.IntSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.UintSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.StrSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptySlSl) != 0 {
				return false
			}
		}
	}
	return
}
func (goo PointersStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PointersStruct
	{
		if isPointersStructEmptyRepr(goo) {
			var pbov *testspb.PointersStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PointersStruct)
		{
			if goo.Int8Pt != nil {
				dgoor := *goo.Int8Pt
				pbo.Int8Pt = int32(dgoor)
			}
		}
		{
			if goo.Int16Pt != nil {
				dgoor := *goo.Int16Pt
				pbo.Int16Pt = int32(dgoor)
			}
		}
		{
			if goo.Int32Pt != nil {
				dgoor := *goo.Int32Pt
				pbo.Int32Pt = dgoor
			}
		}
		{
			if goo.Int32FixedPt != nil {
				dgoor := *goo.Int32FixedPt
				pbo.Int32FixedPt = dgoor
			}
		}
		{
			if goo.Int64Pt != nil {
				dgoor := *goo.Int64Pt
				pbo.Int64Pt = dgoor
			}
		}
		{
			if goo.Int64FixedPt != nil {
				dgoor := *goo.Int64FixedPt
				pbo.Int64FixedPt = dgoor
			}
		}
		{
			if goo.IntPt != nil {
				dgoor := *goo.IntPt
				pbo.IntPt = int64(dgoor)
			}
		}
		{
			if goo.BytePt != nil {
				dgoor := *goo.BytePt
				pbo.BytePt = uint32(dgoor)
			}
		}
		{
			if goo.Uint8Pt != nil {
				dgoor := *goo.Uint8Pt
				pbo.Uint8Pt = uint32(dgoor)
			}
		}
		{
			if goo.Uint16Pt != nil {
				dgoor := *goo.Uint16Pt
				pbo.Uint16Pt = uint32(dgoor)
			}
		}
		{
			if goo.Uint32Pt != nil {
				dgoor := *goo.Uint32Pt
				pbo.Uint32Pt = dgoor
			}
		}
		{
			if goo.Uint32FixedPt != nil {
				dgoor := *goo.Uint32FixedPt
				pbo.Uint32FixedPt = dgoor
			}
		}
		{
			if goo.Uint64Pt != nil {
				dgoor := *goo.Uint64Pt
				pbo.Uint64Pt = dgoor
			}
		}
		{
			if goo.Uint64FixedPt != nil {
				dgoor := *goo.Uint64FixedPt
				pbo.Uint64FixedPt = dgoor
			}
		}
		{
			if goo.UintPt != nil {
				dgoor := *goo.UintPt
				pbo.UintPt = uint64(dgoor)
			}
		}
		{
			if goo.StrPt != nil {
				dgoor := *goo.StrPt
				pbo.StrPt = dgoor
			}
		}
		{
			if goo.BytesPt != nil {
				dgoor := *goo.BytesPt
				goorl := len(dgoor)
				if goorl == 0 {
					pbo.BytesPt = nil
				} else {
					var pbos = make([]uint8, goorl)
					for i := 0; i < goorl; i += 1 {
						{
							goore := dgoor[i]
							{
								pbos[i] = byte(goore)
							}
						}
					}
					pbo.BytesPt = pbos
				}
			}
		}
		{
			if goo.TimePt != nil {
				dgoor := *goo.TimePt
				pbo.TimePt = timestamppb.New(dgoor)
			}
		}
		{
			if goo.EmptyPt != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.EmptyPt.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.EmptyPt = pbom.(*testspb.EmptyStruct)
				if pbo.EmptyPt == nil {
					pbo.EmptyPt = new(testspb.EmptyStruct)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo *PointersStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PointersStruct = msg.(*testspb.PointersStruct)
	{
		if pbo != nil {
			{
				goo.Int8Pt = new(int8)
				*goo.Int8Pt = int8(pbo.Int8Pt)
			}
			{
				goo.Int16Pt = new(int16)
				*goo.Int16Pt = int16(pbo.Int16Pt)
			}
			{
				goo.Int32Pt = new(int32)
				*goo.Int32Pt = pbo.Int32Pt
			}
			{
				goo.Int32FixedPt = new(int32)
				*goo.Int32FixedPt = pbo.Int32FixedPt
			}
			{
				goo.Int64Pt = new(int64)
				*goo.Int64Pt = pbo.Int64Pt
			}
			{
				goo.Int64FixedPt = new(int64)
				*goo.Int64FixedPt = pbo.Int64FixedPt
			}
			{
				goo.IntPt = new(int)
				*goo.IntPt = int(pbo.IntPt)
			}
			{
				goo.BytePt = new(uint8)
				*goo.BytePt = uint8(pbo.BytePt)
			}
			{
				goo.Uint8Pt = new(uint8)
				*goo.Uint8Pt = uint8(pbo.Uint8Pt)
			}
			{
				goo.Uint16Pt = new(uint16)
				*goo.Uint16Pt = uint16(pbo.Uint16Pt)
			}
			{
				goo.Uint32Pt = new(uint32)
				*goo.Uint32Pt = pbo.Uint32Pt
			}
			{
				goo.Uint32FixedPt = new(uint32)
				*goo.Uint32FixedPt = pbo.Uint32FixedPt
			}
			{
				goo.Uint64Pt = new(uint64)
				*goo.Uint64Pt = pbo.Uint64Pt
			}
			{
				goo.Uint64FixedPt = new(uint64)
				*goo.Uint64FixedPt = pbo.Uint64FixedPt
			}
			{
				goo.UintPt = new(uint)
				*goo.UintPt = uint(pbo.UintPt)
			}
			{
				goo.StrPt = new(string)
				*goo.StrPt = pbo.StrPt
			}
			{
				goo.BytesPt = new([]uint8)
				var pbol int = 0
				if pbo.BytesPt != nil {
					pbol = len(pbo.BytesPt)
				}
				if pbol == 0 {
					*goo.BytesPt = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesPt[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					*goo.BytesPt = goos
				}
			}
			{
				goo.TimePt = new(time.Time)
				*goo.TimePt = pbo.TimePt.AsTime()
			}
			{
				if pbo.EmptyPt != nil {
					goo.EmptyPt = new(EmptyStruct)
					err = (*goo.EmptyPt).FromPBMessage(cdc, pbo.EmptyPt)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PointersStruct) GetTypeURL() (typeURL string) {
	return "/tests.PointersStruct"
}
func isPointersStructEmptyRepr(goor PointersStruct) (empty bool) {
	{
		empty = true
		{
			if goor.Int8Pt != nil {
				dgoor := *goor.Int8Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int16Pt != nil {
				dgoor := *goor.Int16Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int32Pt != nil {
				dgoor := *goor.Int32Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int32FixedPt != nil {
				dgoor := *goor.Int32FixedPt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int64Pt != nil {
				dgoor := *goor.Int64Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int64FixedPt != nil {
				dgoor := *goor.Int64FixedPt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.IntPt != nil {
				dgoor := *goor.IntPt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.BytePt != nil {
				dgoor := *goor.BytePt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint8Pt != nil {
				dgoor := *goor.Uint8Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint16Pt != nil {
				dgoor := *goor.Uint16Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint32Pt != nil {
				dgoor := *goor.Uint32Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint32FixedPt != nil {
				dgoor := *goor.Uint32FixedPt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint64Pt != nil {
				dgoor := *goor.Uint64Pt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint64FixedPt != nil {
				dgoor := *goor.Uint64FixedPt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.UintPt != nil {
				dgoor := *goor.UintPt
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.StrPt != nil {
				dgoor := *goor.StrPt
				if dgoor != "" {
					return false
				}
			}
		}
		{
			if goor.BytesPt != nil {
				dgoor := *goor.BytesPt
				if len(dgoor) != 0 {
					return false
				}
			}
		}
		{
			if goor.TimePt != nil {
				return false
			}
		}
		{
			if goor.EmptyPt != nil {
				return false
			}
		}
	}
	return
}
func (goo PointerSlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PointerSlicesStruct
	{
		if isPointerSlicesStructEmptyRepr(goo) {
			var pbov *testspb.PointerSlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PointerSlicesStruct)
		{
			goorl := len(goo.Int8PtSl)
			if goorl == 0 {
				pbo.Int8PtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = int32(dgoor)
							}
						}
					}
				}
				pbo.Int8PtSl = pbos
			}
		}
		{
			goorl := len(goo.Int16PtSl)
			if goorl == 0 {
				pbo.Int16PtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = int32(dgoor)
							}
						}
					}
				}
				pbo.Int16PtSl = pbos
			}
		}
		{
			goorl := len(goo.Int32PtSl)
			if goorl == 0 {
				pbo.Int32PtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Int32PtSl = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedPtSl)
			if goorl == 0 {
				pbo.Int32FixedPtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Int32FixedPtSl = pbos
			}
		}
		{
			goorl := len(goo.Int64PtSl)
			if goorl == 0 {
				pbo.Int64PtSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Int64PtSl = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedPtSl)
			if goorl == 0 {
				pbo.Int64FixedPtSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Int64FixedPtSl = pbos
			}
		}
		{
			goorl := len(goo.IntPtSl)
			if goorl == 0 {
				pbo.IntPtSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = int64(dgoor)
							}
						}
					}
				}
				pbo.IntPtSl = pbos
			}
		}
		{
			goorl := len(goo.BytePtSl)
			if goorl == 0 {
				pbo.BytePtSl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytePtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = byte(dgoor)
							}
						}
					}
				}
				pbo.BytePtSl = pbos
			}
		}
		{
			goorl := len(goo.Uint8PtSl)
			if goorl == 0 {
				pbo.Uint8PtSl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = byte(dgoor)
							}
						}
					}
				}
				pbo.Uint8PtSl = pbos
			}
		}
		{
			goorl := len(goo.Uint16PtSl)
			if goorl == 0 {
				pbo.Uint16PtSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = uint32(dgoor)
							}
						}
					}
				}
				pbo.Uint16PtSl = pbos
			}
		}
		{
			goorl := len(goo.Uint32PtSl)
			if goorl == 0 {
				pbo.Uint32PtSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Uint32PtSl = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedPtSl)
			if goorl == 0 {
				pbo.Uint32FixedPtSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Uint32FixedPtSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64PtSl)
			if goorl == 0 {
				pbo.Uint64PtSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Uint64PtSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedPtSl)
			if goorl == 0 {
				pbo.Uint64FixedPtSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.Uint64FixedPtSl = pbos
			}
		}
		{
			goorl := len(goo.UintPtSl)
			if goorl == 0 {
				pbo.UintPtSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = uint64(dgoor)
							}
						}
					}
				}
				pbo.UintPtSl = pbos
			}
		}
		{
			goorl := len(goo.StrPtSl)
			if goorl == 0 {
				pbo.StrPtSl = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = dgoor
							}
						}
					}
				}
				pbo.StrPtSl = pbos
			}
		}
		{
			goorl := len(goo.BytesPtSl)
			if goorl == 0 {
				pbo.BytesPtSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								goorl1 := len(dgoor)
								if goorl1 == 0 {
									pbos[i] = nil
								} else {
									var pbos1 = make([]uint8, goorl1)
									for i := 0; i < goorl1; i += 1 {
										{
											goore := dgoor[i]
											{
												pbos1[i] = byte(goore)
											}
										}
									}
									pbos[i] = pbos1
								}
							}
						}
					}
				}
				pbo.BytesPtSl = pbos
			}
		}
		{
			goorl := len(goo.TimePtSl)
			if goorl == 0 {
				pbo.TimePtSl = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimePtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								pbos[i] = timestamppb.New(dgoor)
							}
						}
					}
				}
				pbo.TimePtSl = pbos
			}
		}
		{
			goorl := len(goo.EmptyPtSl)
			if goorl == 0 {
				pbo.EmptyPtSl = nil
			} else {
				var pbos = make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptyPtSl[i]
						{
							if goore != nil {
								pbom := proto.Message(nil)
								pbom, err = goore.ToPBMessage(cdc)
								if err != nil {
									return
								}
								pbos[i] = pbom.(*testspb.EmptyStruct)
								if pbos[i] == nil {
									pbos[i] = new(testspb.EmptyStruct)
								}
							}
						}
					}
				}
				pbo.EmptyPtSl = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *PointerSlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PointerSlicesStruct = msg.(*testspb.PointerSlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8PtSl != nil {
					pbol = len(pbo.Int8PtSl)
				}
				if pbol == 0 {
					goo.Int8PtSl = nil
				} else {
					var goos = make([]*int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8PtSl[i]
							{
								goos[i] = new(int8)
								*goos[i] = int8(pboe)
							}
						}
					}
					goo.Int8PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int16PtSl != nil {
					pbol = len(pbo.Int16PtSl)
				}
				if pbol == 0 {
					goo.Int16PtSl = nil
				} else {
					var goos = make([]*int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16PtSl[i]
							{
								goos[i] = new(int16)
								*goos[i] = int16(pboe)
							}
						}
					}
					goo.Int16PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int32PtSl != nil {
					pbol = len(pbo.Int32PtSl)
				}
				if pbol == 0 {
					goo.Int32PtSl = nil
				} else {
					var goos = make([]*int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32PtSl[i]
							{
								goos[i] = new(int32)
								*goos[i] = pboe
							}
						}
					}
					goo.Int32PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedPtSl != nil {
					pbol = len(pbo.Int32FixedPtSl)
				}
				if pbol == 0 {
					goo.Int32FixedPtSl = nil
				} else {
					var goos = make([]*int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedPtSl[i]
							{
								goos[i] = new(int32)
								*goos[i] = pboe
							}
						}
					}
					goo.Int32FixedPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int64PtSl != nil {
					pbol = len(pbo.Int64PtSl)
				}
				if pbol == 0 {
					goo.Int64PtSl = nil
				} else {
					var goos = make([]*int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64PtSl[i]
							{
								goos[i] = new(int64)
								*goos[i] = pboe
							}
						}
					}
					goo.Int64PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedPtSl != nil {
					pbol = len(pbo.Int64FixedPtSl)
				}
				if pbol == 0 {
					goo.Int64FixedPtSl = nil
				} else {
					var goos = make([]*int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedPtSl[i]
							{
								goos[i] = new(int64)
								*goos[i] = pboe
							}
						}
					}
					goo.Int64FixedPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.IntPtSl != nil {
					pbol = len(pbo.IntPtSl)
				}
				if pbol == 0 {
					goo.IntPtSl = nil
				} else {
					var goos = make([]*int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntPtSl[i]
							{
								goos[i] = new(int)
								*goos[i] = int(pboe)
							}
						}
					}
					goo.IntPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.BytePtSl != nil {
					pbol = len(pbo.BytePtSl)
				}
				if pbol == 0 {
					goo.BytePtSl = nil
				} else {
					var goos = make([]*uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytePtSl[i]
							{
								goos[i] = new(uint8)
								*goos[i] = uint8(pboe)
							}
						}
					}
					goo.BytePtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8PtSl != nil {
					pbol = len(pbo.Uint8PtSl)
				}
				if pbol == 0 {
					goo.Uint8PtSl = nil
				} else {
					var goos = make([]*uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8PtSl[i]
							{
								goos[i] = new(uint8)
								*goos[i] = uint8(pboe)
							}
						}
					}
					goo.Uint8PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16PtSl != nil {
					pbol = len(pbo.Uint16PtSl)
				}
				if pbol == 0 {
					goo.Uint16PtSl = nil
				} else {
					var goos = make([]*uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16PtSl[i]
							{
								goos[i] = new(uint16)
								*goos[i] = uint16(pboe)
							}
						}
					}
					goo.Uint16PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32PtSl != nil {
					pbol = len(pbo.Uint32PtSl)
				}
				if pbol == 0 {
					goo.Uint32PtSl = nil
				} else {
					var goos = make([]*uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32PtSl[i]
							{
								goos[i] = new(uint32)
								*goos[i] = pboe
							}
						}
					}
					goo.Uint32PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedPtSl != nil {
					pbol = len(pbo.Uint32FixedPtSl)
				}
				if pbol == 0 {
					goo.Uint32FixedPtSl = nil
				} else {
					var goos = make([]*uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedPtSl[i]
							{
								goos[i] = new(uint32)
								*goos[i] = pboe
							}
						}
					}
					goo.Uint32FixedPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64PtSl != nil {
					pbol = len(pbo.Uint64PtSl)
				}
				if pbol == 0 {
					goo.Uint64PtSl = nil
				} else {
					var goos = make([]*uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64PtSl[i]
							{
								goos[i] = new(uint64)
								*goos[i] = pboe
							}
						}
					}
					goo.Uint64PtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedPtSl != nil {
					pbol = len(pbo.Uint64FixedPtSl)
				}
				if pbol == 0 {
					goo.Uint64FixedPtSl = nil
				} else {
					var goos = make([]*uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedPtSl[i]
							{
								goos[i] = new(uint64)
								*goos[i] = pboe
							}
						}
					}
					goo.Uint64FixedPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.UintPtSl != nil {
					pbol = len(pbo.UintPtSl)
				}
				if pbol == 0 {
					goo.UintPtSl = nil
				} else {
					var goos = make([]*uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintPtSl[i]
							{
								goos[i] = new(uint)
								*goos[i] = uint(pboe)
							}
						}
					}
					goo.UintPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.StrPtSl != nil {
					pbol = len(pbo.StrPtSl)
				}
				if pbol == 0 {
					goo.StrPtSl = nil
				} else {
					var goos = make([]*string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrPtSl[i]
							{
								goos[i] = new(string)
								*goos[i] = pboe
							}
						}
					}
					goo.StrPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.BytesPtSl != nil {
					pbol = len(pbo.BytesPtSl)
				}
				if pbol == 0 {
					goo.BytesPtSl = nil
				} else {
					var goos = make([]*[]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesPtSl[i]
							{
								goos[i] = new([]uint8)
								var pbol1 int = 0
								if pboe != nil {
									pbol1 = len(pboe)
								}
								if pbol1 == 0 {
									*goos[i] = nil
								} else {
									var goos1 = make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboe[i]
											{
												goos1[i] = uint8(pboe)
											}
										}
									}
									*goos[i] = goos1
								}
							}
						}
					}
					goo.BytesPtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.TimePtSl != nil {
					pbol = len(pbo.TimePtSl)
				}
				if pbol == 0 {
					goo.TimePtSl = nil
				} else {
					var goos = make([]*time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimePtSl[i]
							{
								goos[i] = new(time.Time)
								*goos[i] = pboe.AsTime()
							}
						}
					}
					goo.TimePtSl = goos
				}
			}
			{
				var pbol int = 0
				if pbo.EmptyPtSl != nil {
					pbol = len(pbo.EmptyPtSl)
				}
				if pbol == 0 {
					goo.EmptyPtSl = nil
				} else {
					var goos = make([]*EmptyStruct, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.EmptyPtSl[i]
							{
								if pboe != nil {
									goos[i] = new(EmptyStruct)
									err = (*goos[i]).FromPBMessage(cdc, pboe)
									if err != nil {
										return
									}
								}
							}
						}
					}
					goo.EmptyPtSl = goos
				}
			}
		}
	}
	return
}
func (_ PointerSlicesStruct) GetTypeURL() (typeURL string) {
	return "/tests.PointerSlicesStruct"
}
func isPointerSlicesStructEmptyRepr(goor PointerSlicesStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.IntPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytePtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.UintPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.StrPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.TimePtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptyPtSl) != 0 {
				return false
			}
		}
	}
	return
}
func (goo ComplexSt) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ComplexSt
	{
		if isComplexStEmptyRepr(goo) {
			var pbov *testspb.ComplexSt
			msg = pbov
			return
		}
		pbo = new(testspb.ComplexSt)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrField = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ArField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArField = pbom.(*testspb.ArraysStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.SlField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlField = pbom.(*testspb.SlicesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PtField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PtField = pbom.(*testspb.PointersStruct)
		}
	}
	msg = pbo
	return
}
func (goo *ComplexSt) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ComplexSt = msg.(*testspb.ComplexSt)
	{
		if pbo != nil {
			{
				if pbo.PrField != nil {
					err = goo.PrField.FromPBMessage(cdc, pbo.PrField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArField != nil {
					err = goo.ArField.FromPBMessage(cdc, pbo.ArField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlField != nil {
					err = goo.SlField.FromPBMessage(cdc, pbo.SlField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PtField != nil {
					err = goo.PtField.FromPBMessage(cdc, pbo.PtField)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ComplexSt) GetTypeURL() (typeURL string) {
	return "/tests.ComplexSt"
}
func isComplexStEmptyRepr(goor ComplexSt) (empty bool) {
	{
		empty = true
		{
			e := isPrimitivesStructEmptyRepr(goor.PrField)
			if e == false {
				return false
			}
		}
		{
			e := isArraysStructEmptyRepr(goor.ArField)
			if e == false {
				return false
			}
		}
		{
			e := isSlicesStructEmptyRepr(goor.SlField)
			if e == false {
				return false
			}
		}
		{
			e := isPointersStructEmptyRepr(goor.PtField)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo EmbeddedSt1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt1
	{
		if isEmbeddedSt1EmptyRepr(goo) {
			var pbov *testspb.EmbeddedSt1
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt1)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
	}
	msg = pbo
	return
}
func (goo *EmbeddedSt1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt1 = msg.(*testspb.EmbeddedSt1)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					err = goo.PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EmbeddedSt1) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt1"
}
func isEmbeddedSt1EmptyRepr(goor EmbeddedSt1) (empty bool) {
	{
		empty = true
		{
			e := isPrimitivesStructEmptyRepr(goor.PrimitivesStruct)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo EmbeddedSt2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt2
	{
		if isEmbeddedSt2EmptyRepr(goo) {
			var pbov *testspb.EmbeddedSt2
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt2)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ArraysStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArraysStruct = pbom.(*testspb.ArraysStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PointersStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PointersStruct = pbom.(*testspb.PointersStruct)
		}
	}
	msg = pbo
	return
}
func (goo *EmbeddedSt2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt2 = msg.(*testspb.EmbeddedSt2)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					err = goo.PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArraysStruct != nil {
					err = goo.ArraysStruct.FromPBMessage(cdc, pbo.ArraysStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlicesStruct != nil {
					err = goo.SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PointersStruct != nil {
					err = goo.PointersStruct.FromPBMessage(cdc, pbo.PointersStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EmbeddedSt2) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt2"
}
func isEmbeddedSt2EmptyRepr(goor EmbeddedSt2) (empty bool) {
	{
		empty = true
		{
			e := isPrimitivesStructEmptyRepr(goor.PrimitivesStruct)
			if e == false {
				return false
			}
		}
		{
			e := isArraysStructEmptyRepr(goor.ArraysStruct)
			if e == false {
				return false
			}
		}
		{
			e := isSlicesStructEmptyRepr(goor.SlicesStruct)
			if e == false {
				return false
			}
		}
		{
			e := isPointersStructEmptyRepr(goor.PointersStruct)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo EmbeddedSt3) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt3
	{
		if isEmbeddedSt3EmptyRepr(goo) {
			var pbov *testspb.EmbeddedSt3
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt3)
		{
			if goo.PrimitivesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
				if pbo.PrimitivesStruct == nil {
					pbo.PrimitivesStruct = new(testspb.PrimitivesStruct)
				}
			}
		}
		{
			if goo.ArraysStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ArraysStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ArraysStruct = pbom.(*testspb.ArraysStruct)
				if pbo.ArraysStruct == nil {
					pbo.ArraysStruct = new(testspb.ArraysStruct)
				}
			}
		}
		{
			if goo.SlicesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
				if pbo.SlicesStruct == nil {
					pbo.SlicesStruct = new(testspb.SlicesStruct)
				}
			}
		}
		{
			if goo.PointersStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PointersStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PointersStruct = pbom.(*testspb.PointersStruct)
				if pbo.PointersStruct == nil {
					pbo.PointersStruct = new(testspb.PointersStruct)
				}
			}
		}
		{
			if goo.EmptyStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.EmptyStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.EmptyStruct = pbom.(*testspb.EmptyStruct)
				if pbo.EmptyStruct == nil {
					pbo.EmptyStruct = new(testspb.EmptyStruct)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo *EmbeddedSt3) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt3 = msg.(*testspb.EmbeddedSt3)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					goo.PrimitivesStruct = new(PrimitivesStruct)
					err = (*goo.PrimitivesStruct).FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArraysStruct != nil {
					goo.ArraysStruct = new(ArraysStruct)
					err = (*goo.ArraysStruct).FromPBMessage(cdc, pbo.ArraysStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlicesStruct != nil {
					goo.SlicesStruct = new(SlicesStruct)
					err = (*goo.SlicesStruct).FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PointersStruct != nil {
					goo.PointersStruct = new(PointersStruct)
					err = (*goo.PointersStruct).FromPBMessage(cdc, pbo.PointersStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.EmptyStruct != nil {
					goo.EmptyStruct = new(EmptyStruct)
					err = (*goo.EmptyStruct).FromPBMessage(cdc, pbo.EmptyStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EmbeddedSt3) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt3"
}
func isEmbeddedSt3EmptyRepr(goor EmbeddedSt3) (empty bool) {
	{
		empty = true
		{
			if goor.PrimitivesStruct != nil {
				return false
			}
		}
		{
			if goor.ArraysStruct != nil {
				return false
			}
		}
		{
			if goor.SlicesStruct != nil {
				return false
			}
		}
		{
			if goor.PointersStruct != nil {
				return false
			}
		}
		{
			if goor.EmptyStruct != nil {
				return false
			}
		}
	}
	return
}
func (goo EmbeddedSt4) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt4
	{
		if isEmbeddedSt4EmptyRepr(goo) {
			var pbov *testspb.EmbeddedSt4
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt4)
		{
			pbo.Foo1 = int64(goo.Foo1)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbo.Foo2 = goo.Foo2
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ArraysStructField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArraysStructField = pbom.(*testspb.ArraysStruct)
		}
		{
			goorl := len(goo.Foo3)
			if goorl == 0 {
				pbo.Foo3 = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Foo3[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Foo3 = pbos
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
		}
		{
			pbo.Foo4 = goo.Foo4
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PointersStructField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PointersStructField = pbom.(*testspb.PointersStruct)
		}
		{
			pbo.Foo5 = uint64(goo.Foo5)
		}
	}
	msg = pbo
	return
}
func (goo *EmbeddedSt4) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt4 = msg.(*testspb.EmbeddedSt4)
	{
		if pbo != nil {
			{
				goo.Foo1 = int(pbo.Foo1)
			}
			{
				if pbo.PrimitivesStruct != nil {
					err = goo.PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				goo.Foo2 = pbo.Foo2
			}
			{
				if pbo.ArraysStructField != nil {
					err = goo.ArraysStructField.FromPBMessage(cdc, pbo.ArraysStructField)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Foo3 != nil {
					pbol = len(pbo.Foo3)
				}
				if pbol == 0 {
					goo.Foo3 = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Foo3[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.Foo3 = goos
				}
			}
			{
				if pbo.SlicesStruct != nil {
					err = goo.SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				goo.Foo4 = pbo.Foo4
			}
			{
				if pbo.PointersStructField != nil {
					err = goo.PointersStructField.FromPBMessage(cdc, pbo.PointersStructField)
					if err != nil {
						return
					}
				}
			}
			{
				goo.Foo5 = uint(pbo.Foo5)
			}
		}
	}
	return
}
func (_ EmbeddedSt4) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt4"
}
func isEmbeddedSt4EmptyRepr(goor EmbeddedSt4) (empty bool) {
	{
		empty = true
		{
			if goor.Foo1 != 0 {
				return false
			}
		}
		{
			e := isPrimitivesStructEmptyRepr(goor.PrimitivesStruct)
			if e == false {
				return false
			}
		}
		{
			if goor.Foo2 != "" {
				return false
			}
		}
		{
			e := isArraysStructEmptyRepr(goor.ArraysStructField)
			if e == false {
				return false
			}
		}
		{
			if len(goor.Foo3) != 0 {
				return false
			}
		}
		{
			e := isSlicesStructEmptyRepr(goor.SlicesStruct)
			if e == false {
				return false
			}
		}
		{
			if goor.Foo4 != false {
				return false
			}
		}
		{
			e := isPointersStructEmptyRepr(goor.PointersStructField)
			if e == false {
				return false
			}
		}
		{
			if goor.Foo5 != 0 {
				return false
			}
		}
	}
	return
}
func (goo EmbeddedSt5) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt5
	{
		if isEmbeddedSt5EmptyRepr(goo) {
			var pbov *testspb.EmbeddedSt5
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt5)
		{
			pbo.Foo1 = int64(goo.Foo1)
		}
		{
			if goo.PrimitivesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
				if pbo.PrimitivesStruct == nil {
					pbo.PrimitivesStruct = new(testspb.PrimitivesStruct)
				}
			}
		}
		{
			pbo.Foo2 = goo.Foo2
		}
		{
			if goo.ArraysStructField != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ArraysStructField.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ArraysStructField = pbom.(*testspb.ArraysStruct)
				if pbo.ArraysStructField == nil {
					pbo.ArraysStructField = new(testspb.ArraysStruct)
				}
			}
		}
		{
			goorl := len(goo.Foo3)
			if goorl == 0 {
				pbo.Foo3 = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Foo3[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Foo3 = pbos
			}
		}
		{
			if goo.SlicesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
				if pbo.SlicesStruct == nil {
					pbo.SlicesStruct = new(testspb.SlicesStruct)
				}
			}
		}
		{
			pbo.Foo4 = goo.Foo4
		}
		{
			if goo.PointersStructField != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PointersStructField.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PointersStructField = pbom.(*testspb.PointersStruct)
				if pbo.PointersStructField == nil {
					pbo.PointersStructField = new(testspb.PointersStruct)
				}
			}
		}
		{
			pbo.Foo5 = uint64(goo.Foo5)
		}
	}
	msg = pbo
	return
}
func (goo *EmbeddedSt5) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt5 = msg.(*testspb.EmbeddedSt5)
	{
		if pbo != nil {
			{
				goo.Foo1 = int(pbo.Foo1)
			}
			{
				if pbo.PrimitivesStruct != nil {
					goo.PrimitivesStruct = new(PrimitivesStruct)
					err = (*goo.PrimitivesStruct).FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				goo.Foo2 = pbo.Foo2
			}
			{
				if pbo.ArraysStructField != nil {
					goo.ArraysStructField = new(ArraysStruct)
					err = (*goo.ArraysStructField).FromPBMessage(cdc, pbo.ArraysStructField)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Foo3 != nil {
					pbol = len(pbo.Foo3)
				}
				if pbol == 0 {
					goo.Foo3 = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Foo3[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.Foo3 = goos
				}
			}
			{
				if pbo.SlicesStruct != nil {
					goo.SlicesStruct = new(SlicesStruct)
					err = (*goo.SlicesStruct).FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				goo.Foo4 = pbo.Foo4
			}
			{
				if pbo.PointersStructField != nil {
					goo.PointersStructField = new(PointersStruct)
					err = (*goo.PointersStructField).FromPBMessage(cdc, pbo.PointersStructField)
					if err != nil {
						return
					}
				}
			}
			{
				goo.Foo5 = uint(pbo.Foo5)
			}
		}
	}
	return
}
func (_ EmbeddedSt5) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt5"
}
func isEmbeddedSt5EmptyRepr(goor EmbeddedSt5) (empty bool) {
	{
		empty = true
		{
			if goor.Foo1 != 0 {
				return false
			}
		}
		{
			if goor.PrimitivesStruct != nil {
				return false
			}
		}
		{
			if goor.Foo2 != "" {
				return false
			}
		}
		{
			if goor.ArraysStructField != nil {
				return false
			}
		}
		{
			if len(goor.Foo3) != 0 {
				return false
			}
		}
		{
			if goor.SlicesStruct != nil {
				return false
			}
		}
		{
			if goor.Foo4 != false {
				return false
			}
		}
		{
			if goor.PointersStructField != nil {
				return false
			}
		}
		{
			if goor.Foo5 != 0 {
				return false
			}
		}
	}
	return
}
func (goo PrimitivesStructDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructDef
	{
		if isPrimitivesStructDefEmptyRepr(goo) {
			var pbov *testspb.PrimitivesStructDef
			msg = pbov
			return
		}
		pbo = new(testspb.PrimitivesStructDef)
		{
			pbo.Int8 = int32(goo.Int8)
		}
		{
			pbo.Int16 = int32(goo.Int16)
		}
		{
			pbo.Int32 = goo.Int32
		}
		{
			pbo.Int32Fixed = goo.Int32Fixed
		}
		{
			pbo.Int64 = goo.Int64
		}
		{
			pbo.Int64Fixed = goo.Int64Fixed
		}
		{
			pbo.Int = int64(goo.Int)
		}
		{
			pbo.Byte = uint32(goo.Byte)
		}
		{
			pbo.Uint8 = uint32(goo.Uint8)
		}
		{
			pbo.Uint16 = uint32(goo.Uint16)
		}
		{
			pbo.Uint32 = goo.Uint32
		}
		{
			pbo.Uint32Fixed = goo.Uint32Fixed
		}
		{
			pbo.Uint64 = goo.Uint64
		}
		{
			pbo.Uint64Fixed = goo.Uint64Fixed
		}
		{
			pbo.Uint = uint64(goo.Uint)
		}
		{
			pbo.Str = goo.Str
		}
		{
			goorl := len(goo.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			if !amino.IsEmptyTime(goo.Time) {
				pbo.Time = timestamppb.New(goo.Time)
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Empty.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Empty = pbom.(*testspb.EmptyStruct)
		}
	}
	msg = pbo
	return
}
func (goo *PrimitivesStructDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructDef = msg.(*testspb.PrimitivesStructDef)
	{
		if pbo != nil {
			{
				goo.Int8 = int8(pbo.Int8)
			}
			{
				goo.Int16 = int16(pbo.Int16)
			}
			{
				goo.Int32 = pbo.Int32
			}
			{
				goo.Int32Fixed = pbo.Int32Fixed
			}
			{
				goo.Int64 = pbo.Int64
			}
			{
				goo.Int64Fixed = pbo.Int64Fixed
			}
			{
				goo.Int = int(pbo.Int)
			}
			{
				goo.Byte = uint8(pbo.Byte)
			}
			{
				goo.Uint8 = uint8(pbo.Uint8)
			}
			{
				goo.Uint16 = uint16(pbo.Uint16)
			}
			{
				goo.Uint32 = pbo.Uint32
			}
			{
				goo.Uint32Fixed = pbo.Uint32Fixed
			}
			{
				goo.Uint64 = pbo.Uint64
			}
			{
				goo.Uint64Fixed = pbo.Uint64Fixed
			}
			{
				goo.Uint = uint(pbo.Uint)
			}
			{
				goo.Str = pbo.Str
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					goo.Bytes = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.Bytes = goos
				}
			}
			{
				goo.Time = pbo.Time.AsTime()
			}
			{
				if pbo.Empty != nil {
					err = goo.Empty.FromPBMessage(cdc, pbo.Empty)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PrimitivesStructDef) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStructDef"
}
func isPrimitivesStructDefEmptyRepr(goor PrimitivesStructDef) (empty bool) {
	{
		empty = true
		{
			if goor.Int8 != 0 {
				return false
			}
		}
		{
			if goor.Int16 != 0 {
				return false
			}
		}
		{
			if goor.Int32 != 0 {
				return false
			}
		}
		{
			if goor.Int32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int64 != 0 {
				return false
			}
		}
		{
			if goor.Int64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int != 0 {
				return false
			}
		}
		{
			if goor.Byte != 0 {
				return false
			}
		}
		{
			if goor.Uint8 != 0 {
				return false
			}
		}
		{
			if goor.Uint16 != 0 {
				return false
			}
		}
		{
			if goor.Uint32 != 0 {
				return false
			}
		}
		{
			if goor.Uint32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint64 != 0 {
				return false
			}
		}
		{
			if goor.Uint64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint != 0 {
				return false
			}
		}
		{
			if goor.Str != "" {
				return false
			}
		}
		{
			if len(goor.Bytes) != 0 {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.Time) {
				return false
			}
		}
		{
			e := isEmptyStructEmptyRepr(goor.Empty)
			if e == false {
				return false
			}
		}
	}
	return
}
func (goo Concrete1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.Concrete1
	{
		if isConcrete1EmptyRepr(goo) {
			var pbov *testspb.Concrete1
			msg = pbov
			return
		}
		pbo = new(testspb.Concrete1)
	}
	msg = pbo
	return
}
func (goo *Concrete1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.Concrete1 = msg.(*testspb.Concrete1)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ Concrete1) GetTypeURL() (typeURL string) {
	return "/tests.Concrete1"
}
func isConcrete1EmptyRepr(goor Concrete1) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo Concrete2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.Concrete2
	{
		if isConcrete2EmptyRepr(goo) {
			var pbov *testspb.Concrete2
			msg = pbov
			return
		}
		pbo = new(testspb.Concrete2)
	}
	msg = pbo
	return
}
func (goo *Concrete2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.Concrete2 = msg.(*testspb.Concrete2)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ Concrete2) GetTypeURL() (typeURL string) {
	return "/tests.Concrete2"
}
func isConcrete2EmptyRepr(goor Concrete2) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo ConcreteWrappedBytes) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ConcreteWrappedBytes
	{
		if isConcreteWrappedBytesEmptyRepr(goo) {
			var pbov *testspb.ConcreteWrappedBytes
			msg = pbov
			return
		}
		pbo = new(testspb.ConcreteWrappedBytes)
		{
			goorl := len(goo.Value)
			if goorl == 0 {
				pbo.Value = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Value[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Value = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo *ConcreteWrappedBytes) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ConcreteWrappedBytes = msg.(*testspb.ConcreteWrappedBytes)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Value != nil {
					pbol = len(pbo.Value)
				}
				if pbol == 0 {
					goo.Value = nil
				} else {
					var goos = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Value[i]
							{
								goos[i] = uint8(pboe)
							}
						}
					}
					goo.Value = goos
				}
			}
		}
	}
	return
}
func (_ ConcreteWrappedBytes) GetTypeURL() (typeURL string) {
	return "/tests.ConcreteWrappedBytes"
}
func isConcreteWrappedBytesEmptyRepr(goor ConcreteWrappedBytes) (empty bool) {
	{
		empty = true
		{
			if len(goor.Value) != 0 {
				return false
			}
		}
	}
	return
}
func (goo InterfaceFieldsStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.InterfaceFieldsStruct
	{
		if isInterfaceFieldsStructEmptyRepr(goo) {
			var pbov *testspb.InterfaceFieldsStruct
			msg = pbov
			return
		}
		pbo = new(testspb.InterfaceFieldsStruct)
		{
			if goo.F1 != nil {
				typeUrl := goo.F1.(amino.Object).GetTypeURL()
				bz := []byte(nil)
				bz, err = cdc.MarshalBinaryBare(goo.F1)
				if err != nil {
					return
				}
				pbo.F1 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.F2 != nil {
				typeUrl := goo.F2.(amino.Object).GetTypeURL()
				bz := []byte(nil)
				bz, err = cdc.MarshalBinaryBare(goo.F2)
				if err != nil {
					return
				}
				pbo.F2 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.F3 != nil {
				typeUrl := goo.F3.(amino.Object).GetTypeURL()
				bz := []byte(nil)
				bz, err = cdc.MarshalBinaryBare(goo.F3)
				if err != nil {
					return
				}
				pbo.F3 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.F4 != nil {
				typeUrl := goo.F4.(amino.Object).GetTypeURL()
				bz := []byte(nil)
				bz, err = cdc.MarshalBinaryBare(goo.F4)
				if err != nil {
					return
				}
				pbo.F4 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
	}
	msg = pbo
	return
}
func (goo *InterfaceFieldsStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.InterfaceFieldsStruct = msg.(*testspb.InterfaceFieldsStruct)
	{
		if pbo != nil {
			{
				typeUrl := pbo.F1.TypeUrl
				bz := pbo.F1.Value
				goop := &goo.F1
				err = cdc.UnmarshalBinaryAny(typeUrl, bz, goop)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F2.TypeUrl
				bz := pbo.F2.Value
				goop := &goo.F2
				err = cdc.UnmarshalBinaryAny(typeUrl, bz, goop)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F3.TypeUrl
				bz := pbo.F3.Value
				goop := &goo.F3
				err = cdc.UnmarshalBinaryAny(typeUrl, bz, goop)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F4.TypeUrl
				bz := pbo.F4.Value
				goop := &goo.F4
				err = cdc.UnmarshalBinaryAny(typeUrl, bz, goop)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ InterfaceFieldsStruct) GetTypeURL() (typeURL string) {
	return "/tests.InterfaceFieldsStruct"
}
func isInterfaceFieldsStructEmptyRepr(goor InterfaceFieldsStruct) (empty bool) {
	{
		empty = true
		{
			if goor.F1 != nil {
				return false
			}
		}
		{
			if goor.F2 != nil {
				return false
			}
		}
		{
			if goor.F3 != nil {
				return false
			}
		}
		{
			if goor.F4 != nil {
				return false
			}
		}
	}
	return
}
