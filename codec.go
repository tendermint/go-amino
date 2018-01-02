package wire

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Codec struct {
	mtx               sync.RWMutex
	typeInfos         map[string]*TypeInfo
	interfaceInfos    []*TypeInfo
	prefixToTypeInfos map[PrefixBytes]*TypeInfo
	disfixToTypeInfos map[DisfixBytes]*TypeInfo
}

func NewCodec() *Codec {
	cdc := &Codec{
		typeInfos:         make(map[reflect.Type]*TypeInfo),
		prefixToTypeInfos: make(map[PrefixBytes]*TypeInfo),
		disfixToTypeInfos: make(map[DisfixBytes]*TypeInfo),
	}
	return cdc
}

func (cdc *Codec) setTypeInfo(info *TypeInfo) {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	if _, ok := cdc.typeInfos[info.TypeKey()]; ok {
		//if !info.Registered {
		panic(fmt.Sprintf("TypeInfo already exists for %v", info.TypeKey()))
		//}
	}

	fmt.Println("SET TYPE INFO", info.Type, info.TypeKey())
	cdc.typeInfos[info.TypeKey()] = info
	if info.Type.Kind() == reflect.Interface {
		cdc.interfaceInfos = append(cdc.interfaceInfos, info)
	} else if info.Registered {
		prefix := info.Prefix
		disamb := info.Disamb
		disfix := toDisfix(prefix, disamb)
		cdc.prefixToTypeInfos[prefix] = info
		cdc.disfixToTypeInfos[disfix] = info
	}
}

func (cdc *Codec) getTypeInfo(rt reflect.Type) (*TypeInfo, error) {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	info, ok := cdc.typeInfos[rt]
	if !ok {
		fmt.Println("constructing type", rt)
		// Construct info
		var info = new(TypeInfo)
		info.Type = rt
		info.PointerPreferred = false // TODO
		info.Registered = false
		info.Fields = parseFieldInfos(rt)

		// set the info
		fmt.Println("SET TYPE INFO CONSTRUCTION", rt)
		cdc.typeInfos[info.TypeKey()] = info
		return info, nil
	}
	fmt.Println("returning type", info)
	return info, nil
}

func (cdc *Codec) getTypeInfoFromPrefix(pb PrefixBytes) (info *TypeInfo, err error) {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	info, ok := cdc.prefixToTypeInfos[pb]
	if !ok {
		err = fmt.Errorf("unrecognized prefix bytes %X", pb)
	}
	return
}

func (cdc *Codec) getTypeInfoFromDisfix(df DisfixBytes) (info *TypeInfo, err error) {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	info, ok := cdc.disfixToTypeInfos[df]
	if !ok {
		err = fmt.Errorf("unrecognized disambiguation+prefix bytes %X", df)
	}
	return
}

func parseFieldInfos(rt reflect.Type) (infos []FieldInfo) {
	if rt.Kind() != reflect.Struct {
		return nil
	}

	infos = make([]FieldInfo, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue // field is private
		}
		skip, opts := parseFieldOptions(field)
		if skip {
			continue // e.g. json:"-"
		}
		fieldInfo := FieldInfo{
			Index:        i,
			Type:         field.Type,
			ZeroProto:    reflect.Zero(field.Type).Interface(),
			FieldOptions: opts,
		}
		checkUnsafe(fieldInfo)
		infos = append(infos, fieldInfo)
	}
	return infos
}

func parseFieldOptions(field reflect.StructField) (skip bool, opts FieldOptions) {
	binTag := field.Tag.Get("binary")
	wireTag := field.Tag.Get("wire")
	jsonTag := field.Tag.Get("json")

	// If `json:"-"`, don't encode.
	// NOTE: This skips binary as well.
	if jsonTag == "-" {
		skip = true
		return
	}

	// Get JSON field name.
	jsonTagParts := strings.Split(jsonTag, ",")
	if jsonTagParts[0] == "" {
		opts.JSONName = field.Name
	} else {
		opts.JSONName = jsonTagParts[0]
	}

	// Get JSON omitempty.
	if len(jsonTagParts) > 1 {
		if jsonTagParts[1] == "omitempty" {
			opts.JSONOmitEmpty = true
		}
	}

	// Parse binary tags.
	if binTag == "varint" { // TODO: extend
		opts.BinVarint = true
	}

	// Parse wire tags.
	if wireTag == "unsafe" {
		opts.Unsafe = true
	}

	return
}
