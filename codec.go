package wire

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	PrefixBytesLen = 4
	DisambBytesLen = 3
	DisfixBytesLen = PrefixBytesLen + DisambBytesLen
)

type PrefixBytes [PrefixBytesLen]byte
type DisambBytes [DisambBytesLen]byte
type DisfixBytes [DisfixBytesLen]byte // Disamb+Prefix

type TypeInfo struct {
	Type      reflect.Type // Interface type.
	PtrToType reflect.Type
	ZeroValue reflect.Value
	ZeroProto interface{}
	InterfaceInfo
	ConcreteInfo
}

type InterfaceInfo struct {
	Priority     []DisfixBytes               // Disfix priority.
	Implementers map[PrefixBytes][]*TypeInfo // Mutated over time.
	InterfaceOptions
}

type InterfaceOptions struct {
	Priority           []string // Disamb priority.
	AlwaysDisambiguate bool     // If true, include disamb for all types.
}

type ConcreteInfo struct {
	PointerPreferred bool        // Deserialize to pointer type if possible.
	Registered       bool        // Manually regsitered.
	Name             string      // Ignored if !Registered.
	Prefix           PrefixBytes // Ignored if !Registered.
	Disamb           DisambBytes // Ignored if !Registered.
	Fields           []FieldInfo // If a struct.
	ConcreteOptions
}

type ConcreteOptions struct {
}

type FieldInfo struct {
	Type         reflect.Type // Struct field type
	Index        int          // Struct field index
	FieldOptions              // Encoding options
}

type FieldOptions struct {
	JSONName      string // (JSON) field name
	JSONOmitEmpty bool   // (JSON) omitempty
	BinVarint     bool   // (Binary) Use length-prefixed encoding for (u)int64.
	Unsafe        bool   // e.g. if this field is a float.
}

//----------------------------------------
// Codec

type Codec struct {
	mtx              sync.RWMutex
	typeInfos        map[reflect.Type]*TypeInfo
	interfaceInfos   []*TypeInfo
	concreteInfos    []*TypeInfo
	disfixToTypeInfo map[DisfixBytes]*TypeInfo
}

func NewCodec() *Codec {
	cdc := &Codec{
		typeInfos:        make(map[reflect.Type]*TypeInfo),
		disfixToTypeInfo: make(map[DisfixBytes]*TypeInfo),
	}
	return cdc
}

// This function should be used to register all interfaces that will be
// encoded/decoded by go-wire.
// Usage:
// `wire.RegisterInterface((*MyInterface1)(nil), nil)`
func (cdc *Codec) RegisterInterface(ptr interface{}, opts *InterfaceOptions) {

	// Get reflect.Type from ptr.
	rt := getTypeFromPointer(ptr)
	if rt.Kind() != reflect.Interface {
		panic(fmt.Sprintf("RegisterInterface expects an interface, got %v", rt))
	}

	// Construct InterfaceInfo
	var info = cdc.newTypeInfoFromInterfaceType(rt, opts)

	// Finally, check conflicts and register.
	func() {
		cdc.mtx.Lock()
		defer cdc.mtx.Unlock()

		cdc.collectImplementers_nolock(info)
		err := cdc.checkConflictsInPrio_nolock(info)
		if err != nil {
			panic(err)
		}
		cdc.setTypeInfo_nolock(info)
	}()
}

// This function should be used to register concrete types that will appear in
// interface fields/elements to be encoded/decoded by go-wire.
// Usage:
// `wire.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)`
func (cdc *Codec) RegisterConcrete(o interface{}, name string, opts *ConcreteOptions) {

	var pointerPreferred bool

	// Get reflect.Type.
	rt := reflect.TypeOf(o)
	if rt.Kind() == reflect.Interface {
		panic(fmt.Sprintf("expected a non-interface: %v", rt))
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		if rt.Kind() == reflect.Ptr {
			// We can encode/decode pointer-pointers, but not register them.
			panic(fmt.Sprintf("registering pointer-pointers not yet supported: *%v", rt))
		}
		if rt.Kind() == reflect.Interface {
			panic(fmt.Sprintf("registering interface-pointers not yet supported: *%v", rt))
		}
		pointerPreferred = true
	}

	// Construct ConcreteInfo.
	var info = cdc.newTypeInfoFromConcreteType(rt, pointerPreferred, name, opts)

	// Finally, check conflicts and register.
	func() {
		cdc.mtx.Lock()
		defer cdc.mtx.Unlock()

		cdc.addCheckConflictsWithConcrete_nolock(info)
		cdc.setTypeInfo_nolock(info)
	}()
}

//----------------------------------------

func (cdc *Codec) setTypeInfo_wlock(info *TypeInfo) {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	cdc.setTypeInfo_nolock(info)
}

func (cdc *Codec) setTypeInfo_nolock(info *TypeInfo) {

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
		cdc.concreteInfos = append(cdc.concreteInfos, info)
		prefix := info.Prefix
		disamb := info.Disamb
		disfix := toDisfix(disamb, prefix)
		if existing, ok := cdc.disfixToTypeInfo[disfix]; ok {
			panic(fmt.Sprintf("disfix <%X> already registered for %v", disfix, existing.Type))
		}
		cdc.disfixToTypeInfo[disfix] = info
		//cdc.prefixToTypeInfos[prefix] =
		//	append(cdc.prefixToTypeInfos[prefix], info)
	}
}

func (cdc *Codec) getTypeInfo_wlock(rt reflect.Type) (info *TypeInfo, err error) {
	cdc.mtx.Lock() // requires wlock because we might set.
	defer cdc.mtx.Unlock()

	// Transparently "dereference" pointer type.
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	info, ok := cdc.typeInfos[rt]
	if !ok {
		if rt.Kind() == reflect.Interface {
			err = fmt.Errorf("Unregistered interface %v", rt)
			return
		}

		info = cdc.newTypeInfoUnregistered(rt)
		cdc.setTypeInfo_nolock(info)
	}
	return info, nil
}

// iinfo: TypeInfo for the interface for which we must decode a
// concrete type with prefix bytes pb.
func (cdc *Codec) getTypeInfoFromPrefix_rlock(iinfo *TypeInfo, pb PrefixBytes) (info *TypeInfo, err error) {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	infos, ok := iinfo.Implementers[pb]
	if !ok {
		err = fmt.Errorf("unrecognized prefix bytes %X", pb)
		return
	}
	if len(infos) > 1 {
		err = fmt.Errorf("Conflicting concrete types registered for %X: e.g. %v and %v.", pb)
		return
	}
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

// Constructs a *TypeInfo automatically, not from registration.
func (cdc *Codec) newTypeInfoUnregistered(rt reflect.Type) *TypeInfo {
	if rt.Kind() == reflect.Ptr {
		panic("unexpected pointer type") // should not happen.
	}
	if rt.Kind() == reflect.Interface {
		panic("unexpected interface type") // should not happen.
	}

	var info = new(TypeInfo)
	info.Type = rt
	info.PtrToType = reflect.PtrTo(rt)
	info.ZeroValue = reflect.Zero(rt)
	info.ZeroProto = reflect.Zero(rt).Interface()
	// info.InterfaceInfo =
	info.ConcreteInfo.PointerPreferred = false
	info.ConcreteInfo.Registered = false
	// info.ConcreteInfo.Name =
	// info.ConcreteInfo.Prefix =
	// info.ConcreteInfo.Disamb =
	info.ConcreteInfo.Fields = cdc.parseFieldInfos(rt)
	return info
}

func (cdc *Codec) newTypeInfoFromInterfaceType(rt reflect.Type, opts *InterfaceOptions) *TypeInfo {
	if rt.Kind() != reflect.Interface {
		panic(fmt.Sprintf("expected interface type, got %v", rt))
	}

	var info = new(TypeInfo)
	info.Type = rt
	info.PtrToType = reflect.PtrTo(rt)
	info.ZeroValue = reflect.Zero(rt)
	info.ZeroProto = reflect.Zero(rt).Interface()
	info.InterfaceInfo.Implementers = make(map[PrefixBytes][]*TypeInfo)
	if opts != nil {
		info.InterfaceInfo.InterfaceOptions = *opts
		info.InterfaceInfo.Priority = make([]DisfixBytes, len(opts.Priority))
		// Construct Priority []DisfixBytes
		for i, name := range opts.Priority {
			disamb, prefix := nameToDisfix(name)
			disfix := toDisfix(disamb, prefix)
			info.InterfaceInfo.Priority[i] = disfix
		}
	}
	// info.ConcreteInfo.PointerPreferred =
	// info.ConcreteInfo.Registered =
	// info.ConcreteInfo.Name =
	// info.ConcreteInfo.Prefix
	// info.ConcreteInfo.Disamb =
	// info.ConcreteInfo.Fields =
	return info
}

func (cdc *Codec) newTypeInfoFromConcreteType(rt reflect.Type, pointerPreferred bool, name string, opts *ConcreteOptions) *TypeInfo {
	if rt.Kind() == reflect.Interface ||
		rt.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("expected non-interface non-pointer concrete type, got %v", rt))
	}

	var info = new(TypeInfo)
	info.Type = rt
	info.PtrToType = reflect.PtrTo(rt)
	info.ZeroValue = reflect.Zero(rt)
	info.ZeroProto = reflect.Zero(rt).Interface()
	// info.InterfaceOptions =
	info.ConcreteInfo.PointerPreferred = pointerPreferred
	info.ConcreteInfo.Registered = true
	info.ConcreteInfo.Name = name
	info.ConcreteInfo.Disamb = nameToDisamb(name)
	info.ConcreteInfo.Prefix = nameToPrefix(name)
	info.ConcreteInfo.Fields = cdc.parseFieldInfos(rt)
	if opts != nil {
		info.ConcreteOptions = *opts
	}
	return info
}

// Find all conflicting prefixes for concrete types
// that "implement" the interface.  "Implement" in quotes because
// we only consider the pointer, for extra safety.
func (cdc *Codec) collectImplementers_nolock(info *TypeInfo) {
	for _, cinfo := range cdc.concreteInfos {
		if cinfo.PtrToType.Implements(info.Type) {
			info.Implementers[cinfo.Prefix] = append(
				info.Implementers[cinfo.Prefix], cinfo)
		}
	}
}

// Ensure that prefix-conflicting implementing concrete types
// are all registered in the priority list.
// Returns an error if a disamb conflict is found.
func (cdc *Codec) checkConflictsInPrio_nolock(iinfo *TypeInfo) error {

	for _, cinfos := range iinfo.Implementers {
		if len(cinfos) < 2 {
			continue
		}
		for _, cinfo := range cinfos {
			var inPrio = false
			for _, disfix := range iinfo.InterfaceInfo.Priority {
				if toDisfix(cinfo.Disamb, cinfo.Prefix) == disfix {
					inPrio = true
				}
			}
			if !inPrio {
				return fmt.Errorf("%v conflicts with %v other(s). Add it to the priority list for %v.", cinfo.Type, iinfo.Type)
			}
		}
	}
	return nil
}

func (cdc *Codec) addCheckConflictsWithConcrete_nolock(cinfo *TypeInfo) {

	// Iterate over registered interfaces that this "implements".
	// "Implement" in quotes because we only consider the pointer, for extra
	// safety.
	for _, iinfo := range cdc.interfaceInfos {
		if !cinfo.PtrToType.Implements(iinfo.Type) {
			continue
		}

		// Add cinfo to iinfo.Implementers.
		var origImpls = iinfo.Implementers[cinfo.Prefix]
		iinfo.Implementers[cinfo.Prefix] = append(origImpls, cinfo)

		// Finally, check that all conflicts are in `.Priority`.
		// NOTE: This could be optimized, but it's non-trivial.
		err := cdc.checkConflictsInPrio_nolock(iinfo)
		if err != nil {
			// Return to previous state.
			iinfo.Implementers[cinfo.Prefix] = origImpls
			panic(err)
		}
	}
}

//----------------------------------------
// .String()

func (ti TypeInfo) String() string {
	buf := new(bytes.Buffer)
	buf.Write([]byte("TypeInfo{"))
	buf.Write([]byte(fmt.Sprintf("Type:%v,", ti.Type)))
	if ti.Type.Kind() == reflect.Interface {
		buf.Write([]byte(fmt.Sprintf("Priority:%v,", ti.Priority)))
		buf.Write([]byte("Implementers:{"))
		for pb, cinfos := range ti.Implementers {
			buf.Write([]byte(fmt.Sprintf("\"%X\":", pb)))
			buf.Write([]byte(fmt.Sprintf("%v,", cinfos)))
		}
		buf.Write([]byte("}"))
		buf.Write([]byte(fmt.Sprintf("Priority:%v,", ti.InterfaceOptions.Priority)))
		buf.Write([]byte(fmt.Sprintf("AlwaysDisambiguate:%v,", ti.InterfaceOptions.AlwaysDisambiguate)))
	}
	if ti.Type.Kind() != reflect.Interface {
		if ti.ConcreteInfo.Registered {
			buf.Write([]byte("Registered:true,"))
			buf.Write([]byte(fmt.Sprintf("PointerPreferred:%v,", ti.PointerPreferred)))
			buf.Write([]byte(fmt.Sprintf("Name:\"%v\",", ti.Name)))
			buf.Write([]byte(fmt.Sprintf("Prefix:\"%X\",", ti.Prefix)))
			buf.Write([]byte(fmt.Sprintf("Disamb:\"%X\",", ti.Disamb)))
		} else {
			buf.Write([]byte("Registered:false,"))
		}
		if ti.Type.Kind() == reflect.Struct {
			buf.Write([]byte(fmt.Sprintf("Fields:%v,", ti.Fields)))
		}
	}
	buf.Write([]byte("}"))
	return buf.String()
}
