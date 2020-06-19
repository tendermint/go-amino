package amino

// NOTE: We must not depend on protubuf libraries for serialization.

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	timeType       = reflect.TypeOf(time.Time{})
	durationType   = reflect.TypeOf(time.Duration(0))
	gAnyType       = reflect.TypeOf(anypb.Any{})
	gDurationType  = reflect.TypeOf(durationpb.Duration{})
	gEmptyType     = reflect.TypeOf(emptypb.Empty{})
	gStructType    = reflect.TypeOf(structpb.Struct{})
	gValueType     = reflect.TypeOf(structpb.Value{})
	gListType      = reflect.TypeOf(structpb.ListValue{})
	gTimestampType = reflect.TypeOf(timestamppb.Timestamp{})
	gDoubleType    = reflect.TypeOf(wrapperspb.DoubleValue{})
	gFloatType     = reflect.TypeOf(wrapperspb.FloatValue{})
	gInt64Type     = reflect.TypeOf(wrapperspb.Int64Value{})
	gUInt64Type    = reflect.TypeOf(wrapperspb.UInt64Value{})
	gInt32Type     = reflect.TypeOf(wrapperspb.Int32Value{})
	gUInt32Type    = reflect.TypeOf(wrapperspb.UInt32Value{})
	gBoolType      = reflect.TypeOf(wrapperspb.BoolValue{})
	gStringType    = reflect.TypeOf(wrapperspb.StringValue{})
	gBytesType     = reflect.TypeOf(wrapperspb.BytesValue{})
)

// These require special functions for encoding/decoding.
func isBinaryWellKnownType(rt reflect.Type) (wellKnown bool) {
	switch rt {
	// Native types.
	case timeType, durationType:
		return true
	}
	return false
}

// These require special functions for encoding/decoding.
func isJSONWellKnownType(rt reflect.Type) (wellKnown bool) {
	// Special cases based on type.
	switch rt {
	// Native types.
	case timeType, durationType:
		return true
	// Google "well known" types.
	case
		gAnyType, gDurationType, gEmptyType, gStructType, gValueType,
		gListType, gTimestampType, gDoubleType, gFloatType, gInt64Type,
		gUInt64Type, gInt32Type, gUInt32Type, gBoolType, gStringType,
		gBytesType:
		return true
	}
	// General cases based on kind.
	switch rt.Kind() {
	case
		reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.Array, reflect.Slice, reflect.String:
		return true
	default:
		return false
	}
	return false
}

// Returns ok=false if nothing was done because the default behavior is fine (or if err).
// TODO: remove proto dependency.
func encodeReflectJSONWellKnown(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (ok bool, err error) {
	switch info.Type {
	// Native types.
	case timeType:
		// See https://github.com/golang/protobuf/blob/d04d7b157bb510b1e0c10132224b616ac0e26b17/jsonpb/encode.go#L308,
		// "RFC 3339, where generated output will always be Z-normalized
		//  and uses 0, 3, 6 or 9 fractional digits."
		t := rv.Interface().(time.Time)
		err = EncodeJSONTime(w, t)
		if err != nil {
			return false, err
		}
		return true, nil
	case durationType:
		// "Generated output always contains 0, 3, 6, or 9 fractional digits,
		//  depending on required precision."
		d := rv.Interface().(time.Duration)
		err = EncodeJSONDuration(w, d)
		if err != nil {
			return false, err
		}
		return true, nil
	// Google "well known" types.
	case gTimestampType:
		t := rv.Interface().(timestamppb.Timestamp)
		err = EncodeJSONPBTimestamp(w, t)
		if err != nil {
			return false, err
		}
		return true, nil
	case gDurationType:
		d := rv.Interface().(durationpb.Duration)
		err = EncodeJSONPBDuration(w, d)
		if err != nil {
			return false, err
		}
		return true, nil
	// TODO: port each below to above without proto dependency
	// for marshaling code, to minimize dependencies.
	case
		gAnyType, gEmptyType, gStructType, gValueType,
		gListType, gDoubleType, gFloatType, gInt64Type,
		gUInt64Type, gInt32Type, gUInt32Type, gBoolType, gStringType,
		gBytesType:
		bz, err := proto.Marshal(rv.Interface().(proto.Message))
		if err != nil {
			return false, err
		}
		_, err = w.Write(bz)
		return true, err
	}
	return false, nil
}

// Returns ok=false if nothing was done because the default behavior is fine.
// CONTRACT: rv is a concrete type.
func decodeReflectJSONWellKnown(bz []byte, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (ok bool, err error) {
	if rv.Kind() == reflect.Interface {
		panic("expected a concrete type to decode to")
	}
	switch info.Type {
	// Native types.
	case timeType:
		var t time.Time
		t, err = DecodeJSONTime(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.Set(reflect.ValueOf(t))
		return true, nil
	case durationType:
		var d time.Duration
		d, err = DecodeJSONDuration(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.Set(reflect.ValueOf(d))
		return true, nil
	// Google "well known" types.
	case gTimestampType:
		var t timestamppb.Timestamp
		t, err = DecodeJSONPBTimestamp(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.Set(reflect.ValueOf(t))
		return true, nil
	case gDurationType:
		var d durationpb.Duration
		d, err = DecodeJSONPBDuration(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.Set(reflect.ValueOf(d))
		return true, nil
	// TODO: port each below to above without proto dependency
	// for unmarshaling code, to minimize dependencies.
	case
		gAnyType, gEmptyType, gStructType, gValueType,
		gListType, gDoubleType, gFloatType, gInt64Type,
		gUInt64Type, gInt32Type, gUInt32Type, gBoolType, gStringType,
		gBytesType:
		err := proto.Unmarshal(bz, rv.Addr().Interface().(proto.Message))
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Returns ok=false if nothing was done because the default behavior is fine.
func encodeReflectBinaryWellKnown(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions, bare bool) (ok bool, err error) {
	// Validations.
	if rv.Kind() == reflect.Interface {
		panic("expected a concrete type to decode to")
	}
	// Maybe recurse with length-prefixing.
	if !bare {
		buf := bytes.NewBuffer(nil)
		ok, err = encodeReflectBinaryWellKnown(buf, info, rv, fopts, true)
		if err != nil {
			return false, err
		}
		err = EncodeByteSlice(w, buf.Bytes())
		if err != nil {
			return false, err
		}
		return true, nil
	}
	switch info.Type {
	// Native types.
	case timeType:
		var t time.Time
		t = rv.Interface().(time.Time)
		err = EncodeTime(w, t)
		if err != nil {
			return false, err
		}
		return true, nil
	case durationType:
		var d time.Duration
		d = rv.Interface().(time.Duration)
		err = EncodeDuration(w, d)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Returns ok=false if nothing was done because the default behavior is fine.
func decodeReflectBinaryWellKnown(bz []byte, info *TypeInfo, rv reflect.Value, fopts FieldOptions, bare bool) (ok bool, n int, err error) {
	// Validations.
	if rv.Kind() == reflect.Interface {
		panic("expected a concrete type to decode to")
	}
	// Strip if needed.
	bz, err = decodeMaybeBare(bz, &n, bare)
	if err != nil {
		return false, n, err
	}
	switch info.Type {
	// Native types.
	case timeType:
		var t time.Time
		var n_ int
		t, n_, err = DecodeTime(bz)
		if slide(&bz, &n, n_) && err != nil {
			return false, n, err
		}
		rv.Set(reflect.ValueOf(t))
		return true, n, nil
	case durationType:
		var d time.Duration
		var n_ int
		d, n_, err = DecodeDuration(bz)
		if slide(&bz, &n, n_) && err != nil {
			return false, n, err
		}
		rv.Set(reflect.ValueOf(d))
		return true, n, nil
	}
	return false, 0, nil
}

//----------------------------------------
// Well known JSON encoders and decoders

func EncodeJSONTimeValue(w io.Writer, s int64, ns int32) (err error) {
	err = validateTimeValue(s, ns)
	if err != nil {
		return err
	}
	// time.RFC3339Nano isn't exactly right (we need to get 3/6/9 fractional digits).
	t := time.Unix(s, int64(ns)).Round(0).UTC()
	x := t.Format("2006-01-02T15:04:05.000000000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	_, err = w.Write([]byte(fmt.Sprintf(`"%vZ"`, x)))
	return err
}

func EncodeJSONTime(w io.Writer, t time.Time) (err error) {
	t = t.Round(0).UTC()
	return EncodeJSONTimeValue(w, t.Unix(), int32(t.Nanosecond()))
}

func EncodeJSONPBTimestamp(w io.Writer, t timestamppb.Timestamp) (err error) {
	return EncodeJSONTimeValue(w, t.GetSeconds(), t.GetNanos())
}

func EncodeJSONDurationValue(w io.Writer, s int64, ns int32) (err error) {
	err = validateDurationValue(s, ns)
	if err != nil {
		return err
	}
	if s < 0 {
		ns = -ns
	}
	x := fmt.Sprintf("%d.%09d", s, ns)
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	_, err = w.Write([]byte(fmt.Sprintf(`"%vs"`, x)))
	return err
}

func EncodeJSONDuration(w io.Writer, d time.Duration) (err error) {
	return EncodeJSONDurationValue(w, int64(d)/1e9, int32(int64(d)%1e9))
}

func EncodeJSONPBDuration(w io.Writer, d durationpb.Duration) (err error) {
	return EncodeJSONDurationValue(w, d.GetSeconds(), d.GetNanos())
}

func DecodeJSONTime(bz []byte, fopts FieldOptions) (t time.Time, err error) {
	t = zeroTime // defensive
	v, err := unquoteString(string(bz))
	if err != nil {
		return
	}
	t, err = time.Parse(time.RFC3339Nano, v)
	if err != nil {
		err = fmt.Errorf("bad time: %v", err)
		return
	}
	return
}

// NOTE: probably not needed after protobuf v1.25 and after, replace with New().
func newPBTimestamp(t time.Time) timestamppb.Timestamp {
	return timestamppb.Timestamp{Seconds: int64(t.Unix()), Nanos: int32(t.Nanosecond())}
}

func DecodeJSONPBTimestamp(bz []byte, fopts FieldOptions) (t timestamppb.Timestamp, err error) {
	var t_ time.Time
	t_, err = DecodeJSONTime(bz, fopts)
	if err != nil {
		return
	}
	return newPBTimestamp(t_), nil
}

func DecodeJSONDuration(bz []byte, fopts FieldOptions) (d time.Duration, err error) {
	v, err := unquoteString(string(bz))
	if err != nil {
		return
	}
	d, err = time.ParseDuration(v)
	if err != nil {
		err = fmt.Errorf("bad time: %v", err)
		return
	}
	return
}

// NOTE: probably not needed after protobuf v1.25 and after, replace with New().
func newPBDuration(d time.Duration) durationpb.Duration {
	nanos := d.Nanoseconds()
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return durationpb.Duration{Seconds: int64(secs), Nanos: int32(nanos)}
}

func DecodeJSONPBDuration(bz []byte, fopts FieldOptions) (d durationpb.Duration, err error) {
	var d_ time.Duration
	d_, err = DecodeJSONDuration(bz, fopts)
	if err != nil {
		return
	}
	return newPBDuration(d_), nil
}
