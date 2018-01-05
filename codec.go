package wire

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Codec struct {
	mtx               sync.RWMutex
	typeInfos         map[reflect.Type]*TypeInfo
	interfaceInfos    []*TypeInfo
	prefixToTypeInfos map[PrefixBytes][]*TypeInfo
	disfixToTypeInfo  map[DisfixBytes]*TypeInfo
}

func NewCodec() *Codec {
	cdc := &Codec{
		typeInfos:         make(map[reflect.Type]*TypeInfo),
		prefixToTypeInfos: make(map[PrefixBytes][]*TypeInfo),
		disfixToTypeInfo:  make(map[DisfixBytes]*TypeInfo),
	}
	return cdc
}

func (cdc *Codec) setTypeInfo_wlock(info *TypeInfo) {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	cdc.setTypeInfo(info)
}

func (cdc *Codec) setTypeInfo(info *TypeInfo) {

	if info.Type.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("unexpected pointer type"))
	}
	if _, ok := cdc.typeInfos[info.Type]; ok {
		panic(fmt.Sprintf("TypeInfo already exists for %v", info.Type))
	}

	cdc.typeInfos[info.Type] = info
	if info.Type.Kind() == reflect.Interface {
		cdc.interfaceInfos = append(cdc.interfaceInfos, info)
	} else if info.Registered {
		prefix := info.Prefix
		disamb := info.Disamb
		disfix := toDisfix(prefix, disamb)
		// XXX panic if conflict with disfix
		cdc.prefixToTypeInfos[prefix] = []*TypeInfo{info}
		cdc.disfixToTypeInfo[disfix] = info
	}
}

func (cdc *Codec) getTypeInfo_wlock(rt reflect.Type) (info *TypeInfo, err error) {
	cdc.mtx.Lock() // write-lock because we might set.
	defer cdc.mtx.Unlock()

	// Transparently "dereference" pointer type.
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	info, ok := cdc.typeInfos[rt]
	if !ok {
		info, err = cdc.newAutoTypeInfo(rt)
		if err != nil {
			return
		}
		cdc.setTypeInfo(info)
	}
	return info, nil
}

// Constructs a *TypeInfo automatically, not from registration.
func (cdc *Codec) newAutoTypeInfo(rt reflect.Type) (info *TypeInfo, err error) {
	if rt.Kind() == reflect.Ptr {
		panic("unexpected pointer type") // should not happen.
	}
	if rt.Kind() == reflect.Interface {
		err = fmt.Errorf("Unregistered interface %v", rt)
		return
	}

	info = new(TypeInfo)
	info.Type = rt
	info.PointerPreferred = false
	info.Registered = false
	// info.Name =
	// info.Prefix, info.Disamb =
	info.Fields = cdc.parseFieldInfos(rt)
	// info.ConcreteOptions =
	return
}

func (cdc *Codec) getTypeInfoFromPrefix_rlock(pb PrefixBytes) (info *TypeInfo, err error) {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	infos, ok := cdc.prefixToTypeInfos[pb]
	if !ok {
		err = fmt.Errorf("unrecognized prefix bytes %X", pb)
		return
	}
	// XXX handle ambiguous case
	info = infos[0]
	return
}

func (cdc *Codec) getTypeInfoFromDisfix_rlock(df DisfixBytes) (info *TypeInfo, err error) {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	info, ok := cdc.disfixToTypeInfo[df]
	if !ok {
		err = fmt.Errorf("unrecognized disambiguation+prefix bytes %X", df)
		return
	}
	return
}

func (cdc *Codec) parseFieldInfos(rt reflect.Type) (infos []FieldInfo) {
	if rt.Kind() != reflect.Struct {
		return nil
	}

	infos = make([]FieldInfo, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue // field is private
		}
		skip, opts := cdc.parseFieldOptions(field)
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

func (cdc *Codec) parseFieldOptions(field reflect.StructField) (skip bool, opts FieldOptions) {
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
