package wire

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// cdc.encodeReflectJSON

func (cdc *Codec) encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	if printLog {
		spew.Printf("(e) encodeReflectJSON(info: %v, rv: %#v (%v), opts: %v)\n",
			info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Write the disfix wrapper if it is a registered concrete type.
	if info.Registered {
		// Part 1:
		disfix := toDisfix(info.Disamb, info.Prefix)
		err = writeStr(w, _fmt(`{"_df":"%X","_v":`, disfix))
		if err != nil {
			return
		}
		// Part 2:
		defer func() {
			err = writeStr(w, `}`)
		}()
	}

	// Dereference pointers all the way if any.
	// This works for pointer-pointers.
	var foundPointer = false
	for rv.Kind() == reflect.Ptr {
		foundPointer = true
		rv = rv.Elem()
	}

	// Write null if necessary.
	if foundPointer {
		if !rv.IsValid() {
			err = writeStr(w, `null`)
			return
		}
	}

	// If a pointer to the dereferenced type implements json.Marshaller...
	if rv.Addr().Type().Implements(marshalerType) {
		err = invokeMarshalJSON(w, rv)
		if err != nil {
			return
		}
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		return cdc.encodeReflectJSONInterface(w, info, rv, opts)

	case reflect.Array, reflect.Slice:
		return cdc.encodeReflectJSONArrayOrSlice(w, info, rv, opts)

	case reflect.Struct:
		return cdc.encodeReflectJSONStruct(w, info, rv, opts)

	//----------------------------------------
	// Signed, Unsigned

	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int,
		reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		return invokeStdlibJSONMarshal(w, rv.Interface())

	//----------------------------------------
	// Misc

	case reflect.Bool, reflect.Float64, reflect.Float32, reflect.String:
		if !opts.Unsafe {
			return errors.New("Wire.JSON float* support requires `wire:\"unsafe\"`.")
		}
		return invokeStdlibJSONMarshal(w, rv.Interface())

	//----------------------------------------
	// Default

	default:
		panic(fmt.Sprintf("unsupported type %v", info.Type.Kind()))
	}
}

func (cdc *Codec) encodeReflectJSONInterface(w io.Writer, iinfo *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	// Special case when rv is nil, just write "null".
	if rv.IsNil() {
		err = writeStr(w, `null`)
		return
	}

	// Get concrete non-pointer reflect value & type.
	var crv = rv.Elem()
	crv, err = derefForInterface(crv, iinfo)
	if err != nil {
		return
	}
	var crt = crv.Type()

	// Get *TypeInfo for concrete type.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfo_wlock(crt)
	if err != nil {
		return
	}
	if !cinfo.Registered {
		err = fmt.Errorf("Cannot encode unregistered concrete type %v.", crt)
		return
	}

	// NOTE: In the future, we may write disambiguation bytes
	// here, if it is only to be written for interface values.
	// Currently, go-wire JSON *always* writes disfix bytes for
	// all registered concrete types.

	err = cdc.encodeReflectJSON(w, cinfo, crv, opts)
	return
}

func (cdc *Codec) encodeReflectJSONArrayOrSlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	length := rv.Len()

	// Special case when length is 0, just write "null".
	if length == 0 {
		err = writeStr(w, `null`)
		return
	}

	// Open square bracket.
	err = writeStr(w, `[`)
	if err != nil {
		return
	}
	// Close square bracket.
	defer func() {
		err = writeStr(w, `]`)
	}()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		bz := []byte(nil)
		if rv.CanAddr() {
			origBytes := rv.Slice(0, length).Bytes()
			jsonBytes := []byte(nil)
			jsonBytes, err = json.Marshal(origBytes) // base64 encode
			if err != nil {
				return
			}
			bz = jsonBytes
		} else {
			bz = make([]byte, length)
			reflect.Copy(reflect.ValueOf(bz), rv) // XXX: looks expensive!
		}
		_, err = w.Write(bz)
		return

	default:
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			err = cdc.encodeReflectJSON(w, einfo, erv, opts)
			if err != nil {
				return
			}
			// Add a comma if it isn't the last item.
			if i != length-1 {
				err = writeStr(w, `,`)
				if err != nil {
					return
				}
			}
		}
		return
	}
}

func (cdc *Codec) encodeReflectJSONStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	err = writeStr(w, `{`)
	if err != nil {
		return
	}

	for i, field := range info.Fields {
		// Get field value and info.
		var frv = rv.Field(field.Index)
		var finfo *TypeInfo
		finfo, err = cdc.getTypeInfo_wlock(field.Type)
		if err != nil {
			return
		}
		// Write field JSON name.
		err = invokeStdlibJSONMarshal(w, field.JSONName)
		if err != nil {
			return
		}
		// Write colon.
		err = writeStr(w, `:`)
		if err != nil {
			return
		}
		// Write field value.
		err = cdc.encodeReflectJSON(w, finfo, frv, field.FieldOptions)
		if err != nil {
			return
		}
		// Add a comma if it isn't the last item.
		if i != len(info.Fields)-1 {
			err = writeStr(w, `,`)
			if err != nil {
				return
			}
		}
	}
	return

}

//----------------------------------------
// Misc.

// CONTRACT: rv implements json.Marshaler.
func invokeMarshalJSON(w io.Writer, rv reflect.Value) error {
	blob, err := rv.Interface().(json.Marshaler).MarshalJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

func invokeStdlibJSONMarshal(w io.Writer, v interface{}) error {
	// Note: Please don't stream out the output because that adds a newline
	// using json.NewEncoder(w).Encode(data)
	// as per https://golang.org/pkg/encoding/json/#Encoder.Encode
	blob, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

func writeStr(w io.Writer, s string) (err error) {
	_, err = w.Write([]byte(s))
	return
}

func _fmt(s string, args ...interface{}) string {
	return fmt.Sprintf(s, args...)
}
