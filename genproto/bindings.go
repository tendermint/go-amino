package genproto

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tendermint/go-amino"
)

// Given genproto generated schema files for Go objects, generate
// mappers to and from pb messages.  The purpose of this is to let Amino
// use already-optimized probuf logic for serialization.
func GenerateProtoBindingsForTypes(pkg *amino.Package, rtz ...reflect.Type) (file *ast.File, err error) {

	// for TypeInfos.
	cdc := amino.NewCodec()
	cdc.RegisterPackage(pkg)

	file = &ast.File{
		Name:    _i(pkg.GoPkgName),
		Decls:   nil,
		Imports: nil,
	}

	// Generate Imports
	var scope = ast.NewScope(nil)
	var imports = _imports(
		"proto", "google.golang.org/protobuf/proto",
		"amino", "github.com/tendermint/go-amino")
	addImportAuto(imports, scope, pkg.Name+"pb", pkg.P3GoPkgPath)
	file.Decls = append(file.Decls, imports)

	// Generate Decls
	for _, type_ := range rtz {
		info, err := cdc.GetTypeInfo(type_)
		if err != nil {
			return file, err
		}
		if info.Type.Kind() != reflect.Struct {
			continue // Maybe consider supporting more.
		}

		// Generate methods for each type.
		methods, err := generateMethodsForType(imports, scope, pkg, info)
		if err != nil {
			return file, err
		}
		file.Decls = append(file.Decls, methods...)
	}
	return file, nil
}

// Writes in the same directory as the origin package.
// Assumes pb imports in origGoPkgPath+"/pb".
func WriteProtoBindings(pkg *amino.Package) {
	filename := path.Join(pkg.DirName, "pbbindings.go")
	fmt.Printf("writing proto3 bindings to %v for package %v\n", filename, pkg)
	err := WriteProtoBindingsForTypes(filename, pkg, pkg.Types...)
	if err != nil {
		panic(err)
	}
}

func WriteProtoBindingsForTypes(filename string, pkg *amino.Package, rtz ...reflect.Type) (err error) {
	var buf bytes.Buffer
	var fset = token.NewFileSet()
	var file *ast.File
	file, err = GenerateProtoBindingsForTypes(pkg, rtz...)
	if err != nil {
		return
	}
	err = printer.Fprint(&buf, fset, file)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		return
	}
	return
}

// modified imports if necessary.
func generateMethodsForType(imports *ast.GenDecl, scope *ast.Scope, pkg *amino.Package, info *amino.TypeInfo) (methods []ast.Decl, err error) {
	if info.Type.Kind() != reflect.Struct {
		panic("not yet supported")
	}

	p3pkgName, ok := importNameForPath(pkg.P3GoPkgPath, imports)
	if !ok {
		panic("should not happen")
	}

	//////////////////
	// ToPBMessage()
	{
		scope2 := ast.NewScope(scope)
		addVars(scope2, "cdc", "goo", "pbo", "msg", "err")
		// Set toProto function.
		methods = append(methods, _func("ToPBMessage",
			"goo", info.Type.Name(),
			_fields("cdc", "*amino.Codec"),
			_fields("msg", "proto.Message", "err", "error"),
			_block(
				// Body: declaration for pb message.
				_var("pbo", _x("*%v.%v", p3pkgName, info.Type.Name()), nil),
				// Body: copying over fields.
				_block(go2pbStmts(true, imports, scope2, _i("pbo"), _i("goo"), false, info, 0)...),
				// Body: return value.
				_a("msg", "=", "pbo"),
				_return(),
			),
		))
	}

	//////////////////
	// FromPBMessage()
	{
		scope2 := ast.NewScope(scope)
		addVars(scope2, "cdc", "goo", "pbo", "msg", "err")
		methods = append(methods, _func("FromPBMessage",
			"goo", "*"+info.Type.Name(),
			_fields("cdc", "*amino.Codec", "msg", "proto.Message"),
			_fields("err", "error"),
			_block(
				// Body: declaration for pb message.
				_var("pbo", _x("*%v.%v", p3pkgName, info.Type.Name()),
					_x("%v.~(~*%v.%v~)", "msg", p3pkgName, info.Type.Name())),
				// Body: copying over fields.
				_block(pb2goStmts(pkg, true, imports, scope2, _i("goo"), true, info, _i("pbo"))...),
				// Body: return.
				_return(),
			),
		))
	}

	//////////////////
	// TypeUrl()
	{
		methods = append(methods, _func("GetTypeURL",
			"", info.Type.Name(),
			_fields(),
			_fields("typeURL", "string"),
			_block(
				_return(_s(info.TypeURL)),
			),
		))
	}

	//////////////////
	// IsEmpty()
	{
		scope2 := ast.NewScope(scope)
		addVars(scope2, "goo", "empty")
		methods = append(methods, _func("IsEmpty",
			"goo", info.Type.Name(),
			_fields(),
			_fields("empty", "bool"),
			_block(
				// Body: check fields.
				_block(append(
					[]ast.Stmt{_a("empty", "=", "true")},
					isEmptyStmts(true, imports, scope2, _i("goo"), false, info)...,
				)...),
				// Body: return.
				_return(),
			),
		))
	}
	return
}

// These don't have ToPBMessage functions.
// TODO make this a property of the package?
var noBindings = struct{}{}
var noBindingsPkgs = map[string]struct{}{
	"":     noBindings,
	"time": noBindings,
}

func hasPBBindings(info *amino.TypeInfo) bool {
	if info.Type.Kind() == reflect.Ptr {
		return false
	}
	pkg := info.Package.GoPkgPath
	_, ok := noBindingsPkgs[pkg]
	return !ok
}

// END

// isRoot: true if goo is the rootPkg, false if nested fields, even if gooType is rootPkg.
// imports: global imports -- may be modified.
// pbo: protobuf variable or field.
// goo: native go variable or field.
// gooIsPtr: whether goo is ptr.
// gooType: type info for goo's type (elem type if pointer).
// CONTRACT: pbo is assignable.
//  * The general case is `_a(pbo, "=", goo)`
//  * The struct case is like `_a(_sel(pbo, field.Name), "=", goo)`
// CONTRACT: for arrays and lists, memory must be allocated beforehand, but new
// instances are created within this function.
func go2pbStmts(isRoot bool, imports *ast.GenDecl, scope *ast.Scope, pbo ast.Expr, goo ast.Expr, gooIsPtr bool, gooType *amino.TypeInfo, options uint64) (b []ast.Stmt) {

	const (
		option_bytes = 0x01 // if uint8 is an element of bytes.
	)

	// Special case if nil-pointer.
	if gooIsPtr || gooType.Type.Kind() == reflect.Interface {
		defer func(goo ast.Expr) {
			// Wrap penultimate b with if statement.
			b = []ast.Stmt{_if(_b(goo, "!=", _i("nil")),
				b...,
			)}
		}(goo)
	}
	// Below, we can assume that goo isn't nil.

	// Declare dgoo before it's used if needed.
	// dgoo() returns goo or _deref(goo) depending.
	dgoo_ := ""
	dgoo := func() ast.Expr {
		if gooIsPtr {
			if dgoo_ == "" {
				dgoo_ = addVarUniq(scope, "dgoo")
				b = append(b,
					_a(dgoo_, ":=", _deref(goo)))
			}
			return _i(dgoo_)
		} else {
			return goo
		}
	}

	// External case.
	// If gooType is registered, just call ToPBMessage.
	// TODO If not registered?
	if !isRoot && gooType.Registered && hasPBBindings(gooType) {
		// Call ToPBMessage().
		pbote_ := p3goTypeExprString(imports, scope, gooType)
		pbom_ := addVarUniq(scope, "pbom")
		b = append(b,
			_a(pbom_, ":=", _x("proto.Message~(~nil~)")),
			_a(pbom_, _i("err"), "=", _call(_sel(goo, "ToPBMessage"), _i("cdc"))),
			_if(_x("err__!=__nil"),
				_return(),
			),
			_a(pbo, "=", _x("%v.~(~%v~)", pbom_, pbote_)),
		)
		if gooIsPtr {
			if pbote_[0] != '*' {
				panic("expected pointer kind for p3goTypeExprString (of registered type)")
			}
			dpbote_ := pbote_[1:]
			b = append(b,
				_if(_b(pbo, "==", "nil"),
					_a(pbo, "=", _x("new~(~%v~)", dpbote_))))
		}
		return
	}

	// Special case if IsAminoMarshaler.
	if gooType.IsAminoMarshaler {
		// First, derive repr instance.
		goor_ := addVarUniq(scope, "goor")
		b = append(b,
			_a(goor_, "err", ":=", _call(_sel(goo, "MarshalAmino"))),
			_if(_x("err__!=__nil"),
				_return(_x("nil"), _i("err")),
			),
		)
		goo = _i(goor_) // switcharoo
		gooType = gooType.ReprType
	}
	// Below, we can assume that gooType isn't amino.Marshaler

	// Special case, time & duration.
	switch gooType.Type {
	case timeType:
		pkgName := addImportAuto(
			imports, scope, "timestamppb", "google.golang.org/protobuf/types/known/timestamppb")
		if gooIsPtr { // (non-nil)
			b = append(b,
				_a(pbo, "=", _call(_sel(_x(pkgName), "New"), dgoo())))
		} else {
			b = append(b,
				_if(_not(_call(_x("amino.IsEmptyTime"), dgoo())),
					_a(pbo, "=", _call(_sel(_x(pkgName), "New"), dgoo()))))
		}
		return
	case durationType:
		pkgName := addImportAuto(
			imports, scope, "durationpb", "google.golang.org/protobuf/types/known/durationpb")
		if gooIsPtr { // (non-nil)
			b = append(b,
				_a(pbo, "=", _call(_sel(_x(pkgName), "New"), dgoo())))
		} else {
			b = append(b,
				_if(_b(_call(_sel(goo, "Nanoseconds")), "!=", "0"),
					_a(pbo, "=", _call(_sel(_x(pkgName), "New"), dgoo()))))
		}
		return
	}

	// Special case, external empty types.
	if gooType.Registered && hasPBBindings(gooType) {
		if isRoot {
			pbote_ := p3goTypeExprString(imports, scope, gooType)
			pbov_ := addVarUniq(scope, "pbov")
			b = append(b,
				_if(_call(_sel(goo, "IsEmpty")),
					_var(pbov_, _x(pbote_), nil),
					_a("msg", "=", pbov_),
					_return()))
		} else if !gooIsPtr {
			oldb := b
			b = []ast.Stmt(nil) // switcharoo
			defer func(goo ast.Expr) {
				newb := b
				b = append(oldb,
					_if(_not(_call(_sel(goo, "IsEmpty"))),
						newb...))
			}(goo)
		}
	}

	// General case
	switch gooType.Type.Kind() {

	case reflect.Interface:
		typeUrl_ := addVarUniq(scope, "typeUrl")
		bz_ := addVarUniq(scope, "bz")
		danyte_ := p3goTypeExprString(imports, scope, gooType)[1:]
		b = append(b,
			_a(typeUrl_, ":=", _call(_sel(_ta(goo, _x("amino.Object")), "GetTypeURL"))),
			_a(bz_, ":=", "[]byte~(~nil~)"),
			_a(bz_, "err", "=", _call(_sel(_i("cdc"), "MarshalBinaryBare"), goo)),
			_if(_x("err__!=__nil"),
				_return(),
			),
			_a(pbo, "=", _x("&%v~{~TypeUrl:typeUrl,Value:bz~}", danyte_)),
		)

	case reflect.Int:
		b = append(b,
			_a(pbo, "=", _call(_i("int64"), dgoo())))
	case reflect.Int16, reflect.Int8:
		b = append(b,
			_a(pbo, "=", _call(_i("int32"), dgoo())))
	case reflect.Uint:
		b = append(b,
			_a(pbo, "=", _call(_i("uint64"), dgoo())))
	case reflect.Uint16:
		b = append(b,
			_a(pbo, "=", _call(_i("uint32"), dgoo())))
	case reflect.Uint8:
		if options&option_bytes == 0 {
			b = append(b,
				_a(pbo, "=", _call(_i("uint32"), dgoo())))
		} else {
			b = append(b,
				_a(pbo, "=", _call(_i("byte"), dgoo())))
		}

	case reflect.Array, reflect.Slice:
		var options uint64
		var gooeIsPtr = gooType.ElemIsPtr
		var gooeType = gooType.Elem
		var pboete_ string
		switch gooeType.ReprType.Type.Kind() {
		case reflect.Interface:
			pboete_ = "*anypb.Any"
		case reflect.Array, reflect.Slice:
			// nested list
			// nested lists should be declared at the rootPkg package.
			// this is a workaround due to Proto deficiencies.
			pboete_ = p3goListTypeExprString(gooeType)
		case reflect.Struct:
			pboete_ = p3goTypeExprString(imports, scope, gooeType)
		case reflect.Int:
			pboete_ = "int64"
		case reflect.Uint:
			pboete_ = "uint64"
		case reflect.Int8:
			pboete_ = "int32"
		case reflect.Uint8:
			pboete_ = "uint8" // bytes
			options |= option_bytes
		case reflect.Int16:
			pboete_ = "int32"
		case reflect.Uint16:
			pboete_ = "uint32"
		default:
			pboete_ = gooeType.Type.String()
			if pboete_ == "" {
				panic("unexpected empty type expr string")
			}
		}

		// Construct, translate, assign.
		gool_ := addVarUniq(scope, "gool")
		pbos_ := addVarUniq(scope, "pbos")
		scope2 := ast.NewScope(scope)
		addVars(scope2, "i", "gooe")
		b = append(b,
			_a(gool_, ":=", _len(dgoo())),
			_ife(_x("%v__==__0", gool_),
				_block( // then
					// Prefer nil for empty slices for less gc overhead.
					_a(pbo, "=", _i("nil")),
				),
				_block( // else
					_var(pbos_, nil, _x("make~(~[]%v,%v~)", pboete_, gool_)),
					_for(
						_a("i", ":=", "0"),
						_x("i__<__%v", gool_),
						_a("i", "+=", "1"),
						_block(
							// Translate in place.
							_a("gooe", ":=", _ix(dgoo(), _i("i"))),
							_block(go2pbStmts(false, imports, scope2, _x("%v~[~i~]", pbos_), _i("gooe"), gooeIsPtr, gooeType, options)...),
						),
					),
					_a(pbo, "=", pbos_),
				)))

	case reflect.Struct:
		pbote_ := p3goTypeExprString(imports, scope, gooType)
		if pbote_[0] != '*' {
			panic("expected pointer kind for p3goTypeExprString of struct type")
		}
		dpbote_ := pbote_[1:]

		b = append(b,
			_a(pbo, "=", _x("new~(~%v~)", dpbote_)))

		for _, field := range gooType.Fields {
			var goofIsPtr = field.IsPtr()
			var goofType = field.TypeInfo.ReprType
			var goof = _sel(dgoo(), field.Name) // next goo
			var pbof = _sel(pbo, field.Name)    // next pbo

			// Translate in place.
			scope2 := ast.NewScope(scope)
			b = append(b,
				_block(go2pbStmts(false, imports, scope2, pbof, goof, goofIsPtr, goofType, 0)...),
			)
		}

	default:
		// General translation.
		b = append(b, _a(pbo, "=", dgoo()))

	}
	return b
}

// package: the package for the concrete type for which we are generating go2pbStmts.
// isRoot: true if goo is the root, false for fields and elems which are not inlined.
// imports: global imports -- used to look up package names.
// goo: native go variable or field.
// gooIsPtr: is goo a pointer?
// gooType: type info for goo's ultimate type (elem if pointer)..
// pbo: protobuf variable or field.
// CONTRACT: goo is addressable.
// CONTRACT: for arrays and lists, memory must be allocated beforehand, but new
// instances are created within this function.
func pb2goStmts(rootPkg *amino.Package, isRoot bool, imports *ast.GenDecl, scope *ast.Scope, goo ast.Expr, gooIsPtr bool, gooType *amino.TypeInfo, pbo ast.Expr) (b []ast.Stmt) {

	// Special case if pbo is a nil struct pointer (that isn't timestamp)
	//
	// We especially want this behavior (and optimization) for for
	// amino.Marshalers, because of the construction cost.
	switch gooType.ReprType.Type.Kind() {
	case reflect.Struct:
		if gooType.ReprType.Type != timeType {
			defer func(pbo ast.Expr) {
				// Wrap penultimate b with if statement.
				b = []ast.Stmt{_if(_b(pbo, "!=", "nil"),
					b...,
				)}
			}(pbo)
		}
	}
	// Below, we can assume that pbo isn't a nil struct (that isn't timestamp).

	// First we need to construct the goo.
	// NOTE Unlike go2pb, due to the asymmetry of FromPBMessage/ToPBMessage,
	// and MarshalAmino/UnmarshalAmino, both pairs which require that goo not
	// be nil (so we must instantiate new() here).  On the other hand, go2pb's
	// instantiation of corresponding pb objects depends on the kind, so it
	// cannot be done before the switch cases like here.
	if gooIsPtr && !isRoot {
		dgoote_ := goTypeExprString(rootPkg, imports, scope, false, gooType)
		b = append(b,
			_a(goo, "=", _x("new~(~%v~)", dgoote_)))
		goo = _deref(goo)
	}
	// Below, we can assume that goo is a valid non-pointer.

	// External case.
	// If gooType is registered, just call FromPBMessage.
	// TODO If not registered?
	if !isRoot && gooType.Registered && hasPBBindings(gooType) {
		b = append(b,
			_a(_i("err"), "=", _call(_sel(goo, "FromPBMessage"), _i("cdc"), pbo)),
			_if(_x("err__!=__nil"),
				_return(),
			),
		)
		return
	}

	// Special case if IsAminoMarshaler.
	// NOTE: doesn't matter whether goo is ptr or not.
	if gooType.IsAminoMarshaler {
		// First, construct new repr instance.
		goorte_ := goTypeExprString(rootPkg, imports, scope, false, gooType.ReprType)
		goor_ := addVarUniq(scope, "goor")
		scope2 := ast.NewScope(scope)
		b = append(b,
			_var(goor_, _x(goorte_), nil))
		// Then, transcribe to repr var.
		b = append(b, _block(
			pb2goStmts(rootPkg, isRoot, imports, scope2, _i(goor_), false, gooType.ReprType, pbo)...))
		// Finally, unmarshal to goo.
		b = append(b,
			_a("err", "=", _call(_sel(goo, "UnmarshalAmino"), _i(goor_))),
			_if(_x("err__!=__nil"),
				_return(),
			),
		)
		return
	}
	// Below, we can assume that gooType isn't amino.Marshaler.

	// Special case for time/duration.
	switch gooType.Type {
	case timeType:
		b = append(b,
			_a(goo, "=", _call(_sel(pbo, "AsTime"))))
		return
	case durationType:
		b = append(b,
			_a(goo, "=", _call(_sel(pbo, "AsDuration"))))
		return
	}

	// General case
	switch gooType.Type.Kind() {

	case reflect.Interface:
		typeUrl_ := addVarUniq(scope, "typeUrl")
		bz_ := addVarUniq(scope, "bz")
		goop_ := addVarUniq(scope, "goop")
		b = append(b,
			_a(typeUrl_, ":=", _sel(pbo, "TypeUrl")),
			_a(bz_, ":=", _sel(pbo, "Value")),
			_a(goop_, ":=", _ref(goo)), // goo is addressable. NOTE &*a == a if a != nil.
			_a("err", "=", _x("cdc.UnmarshalBinaryAny~(~%v,%v,%v~)",
				typeUrl_, bz_, goop_)),
			_if(_x("err__!=__nil"),
				_return(),
			),
		)

	case reflect.Int:
		b = append(b,
			_a(goo, "=", _call(_i("int"), pbo)))
	case reflect.Int16:
		b = append(b,
			_a(goo, "=", _call(_i("int16"), pbo)))
	case reflect.Int8:
		b = append(b,
			_a(goo, "=", _call(_i("int8"), pbo)))
	case reflect.Uint:
		b = append(b,
			_a(goo, "=", _call(_i("uint"), pbo)))
	case reflect.Uint16:
		b = append(b,
			_a(goo, "=", _call(_i("uint16"), pbo)))
	case reflect.Uint8:
		b = append(b,
			_a(goo, "=", _call(_i("uint8"), pbo)))

	case reflect.Array:
		var gooLen = gooType.Type.Len()
		var gooeType = gooType.Elem
		var gooeIsPtr = gooType.ElemIsPtr
		var gooete_ = goTypeExprString(rootPkg, imports, scope, gooeIsPtr, gooeType)
		goos_ := addVarUniq(scope, "goos")
		scope2 := ast.NewScope(scope)
		addVars(scope2, "i", "pboe")

		// Construct, translate, assign.
		b = append(b,
			_var(goos_, nil, _x("[%v]%v~{~~}", gooLen, gooete_)),
			_for(
				_a("i", ":=", "0"),
				_x("i__<__%v", gooLen),
				_a("i", "+=", "1"),
				_block(
					// Translate in place.
					_a("pboe", ":=", _ix(pbo, _i("i"))),
					_block(pb2goStmts(rootPkg, false, imports, scope2, _x("%v~[~i~]", goos_), gooeIsPtr, gooeType, _i("pboe"))...),
				),
			),
			_a(goo, "=", goos_),
		)

	case reflect.Slice:
		var gooeType = gooType.Elem
		var gooeIsPtr = gooType.ElemIsPtr
		var gooete_ = goTypeExprString(rootPkg, imports, scope, gooeIsPtr, gooeType)
		pbol_ := addVarUniq(scope, "pbol")
		goos_ := addVarUniq(scope, "goos")
		scope2 := ast.NewScope(scope)
		addVars(scope2, "i", "pboe")

		// Construct, translate, assign.
		b = append(b,
			_a(pbol_, ":=", _len(pbo)),
			_ife(_x("%v__==__0", pbol_),
				_block( // then
					// Prefer nil for empty slices for less gc overhead.
					_a(goo, "=", _i("nil")),
				),
				_block( // else
					_var(goos_, nil, _x("make~(~[]%v,%v~)", gooete_, pbol_)),
					_for(
						_a("i", ":=", "0"),
						_x("i__<__%v", pbol_),
						_a("i", "+=", "1"),
						_block(
							// Translate in place.
							_a("pboe", ":=", _ix(pbo, _i("i"))),
							_block(pb2goStmts(rootPkg, false, imports, scope2, _x("%v~[~i~]", goos_), gooeIsPtr, gooeType, _i("pboe"))...),
						),
					),
					_a(goo, "=", goos_),
				)),
		)

	case reflect.Struct:
		for _, field := range gooType.Fields {
			var pbof = _sel(pbo, field.Name) // next pbo.
			var goofIsPtr = field.IsPtr()
			var goofType = field.TypeInfo
			var goof = _sel(goo, field.Name) // next goo.

			// Translate in place.
			scope2 := ast.NewScope(scope)
			b = append(b,
				_block(pb2goStmts(rootPkg, false, imports, scope2, goof, goofIsPtr, goofType, pbof)...),
			)
		}

	default:
		// General translation.
		b = append(b, _a(goo, "=", pbo))
	}
	return b
}

func isEmptyStmts(isRoot bool, imports *ast.GenDecl, scope *ast.Scope, goo ast.Expr, gooIsPtr bool, gooType *amino.TypeInfo) (b []ast.Stmt) {

	// Special case if non-nil struct-pointer.
	// TODO: this could be precompiled and optimized (when !isRoot).
	if gooIsPtr && gooType.ReprType.Type.Kind() == reflect.Struct {
		b = []ast.Stmt{_if(_b(goo, "!=", _i("nil")),
			_return(_i("false")),
		)}
		return
	}

	// Special case if nil-pointer.
	if gooIsPtr || gooType.Type.Kind() == reflect.Interface {
		defer func(goo ast.Expr) {
			// Wrap penultimate b with if statement.
			b = []ast.Stmt{_if(_b(goo, "!=", _i("nil")),
				b...,
			)}
		}(goo)

	}
	// Below, we can assume that goo isn't nil.
	// NOTE: just because it's not nil doesn't mean it's empty, specifically
	// for time. Amino marshallers are empty iff nil.

	// Declare dgoo before it's used if needed.
	// dgoo() returns goo or _deref(goo) depending.
	dgoo_ := ""
	dgoo := func() ast.Expr {
		if gooIsPtr {
			if dgoo_ == "" {
				dgoo_ = addVarUniq(scope, "dgoo")
				b = append(b,
					_a(dgoo_, ":=", _deref(goo)))
			}
			return _i(dgoo_)
		} else {
			return goo
		}
	}

	// External case.
	// If gooType is registered, just call ToPBMessage.
	// TODO If not registered?
	if !isRoot && gooType.Registered && hasPBBindings(gooType) {
		e_ := addVarUniq(scope, "e")
		b = append(b,
			_a(e_, ":=", _call(_sel(dgoo(), "IsEmpty"))),
			_if(_x("%v__==__false", e_),
				_return(_i("false")),
			),
		)
		return
	}

	// Special case if IsAminoMarshaler.
	if gooType.IsAminoMarshaler {
		// First, derive repr instance.
		goor_ := addVarUniq(scope, "goor")
		b = append(b,
			_a(goor_, "err", ":=", _call(_sel(goo, "MarshalAmino"))),
			_if(_x("err__!=__nil"),
				_return(_x("nil"), _i("err")),
			),
		)
		goo = _i(goor_) // switcharoo
		gooType = gooType.ReprType
	}
	// Below, we can assume that gooType isn't amino.Marshaler

	// General case
	switch gooType.Type.Kind() {

	case reflect.Interface:
		b = append(b,
			_return(_i("false")))

	case reflect.Array, reflect.Slice:
		b = append(b,
			_if(_b(_len(dgoo()), "!=", "0"),
				_return(_i("false"))))

	case reflect.Struct:
		// Special case for time.  The default behavior is fine for time.Duration.
		switch gooType.Type {
		case timeType:
			b = append(b,
				_if(_not(_call(_x("amino.IsEmptyTime"), dgoo())),
					_return(_x("false"))))
			return
		default:
			for _, field := range gooType.Fields {
				var goof = _sel(dgoo(), field.Name) // next goo
				var goofIsPtr = field.IsPtr()
				var goofType = field.TypeInfo.ReprType

				// Translate in place.
				scope2 := ast.NewScope(scope)
				b = append(b,
					_block(isEmptyStmts(false, imports, scope2, goof, goofIsPtr, goofType)...),
				)
			}
		}

	default:
		// General translation.
		b = append(b,
			_if(_b(dgoo(), "!=", defaultExpr(gooType.Type.Kind())),
				_return(_i("false"))))
	}
	return b
}

//----------------------------------------
// other....

// Splits a Go expression into left and right parts.
// Returns ok=false if not a binary operator.
// Never panics.
// If ok=true, left+op+right == expr.
//
// Examples:
//  - "5 * 2":       left="5 ", op="*", right=" 2", ok=true
//  - " 5*2 ":       left=" 5", op="*", right="2 ", ok=true
//  - "1*2+ 3":      left="1*2", op="+", right=" 3", ok=true
//  - "1*2+(3+ 4)":  left="1*2", op="+", right="(3+ 4)", ok=true
//  - "'foo'+'bar'": left="'foo'", op="+", right="'bar'", ok=true
//  - "'x'":         ok=false
func chopBinary(expr string) (left, op, right string, ok bool) {
	// XXX implementation redacted for CHALLENGE1.
	// TODO restore implementation and replace '__'
	parts := strings.Split(expr, "__")
	if len(parts) != 3 {
		return
	}
	left = parts[0]
	op = parts[1]
	right = parts[2]
	ok = true
	return
}

// Given that 'in' ends with ')', '}', or ']',
// find the matching opener, while processing escape
// sequences of strings and rune literals.
// `tok` is the corresponding opening token.
// `right` excludes the last character (closer).
func chopRight(expr string) (left string, tok rune, right string) {
	// XXX implementation redacted for CHALLENGE1.
	// TODO restore implementation and replace '~'
	parts := strings.Split(expr, "~")
	if len(parts) != 4 {
		return
	}
	left = parts[0]
	tok = rune(parts[1][0])
	right = parts[2]
	// close = parts[3]
	return
}

//----------------------------------------
// AST Construction (Expr)

func _i(name string) *ast.Ident {
	if name == "" {
		panic("unexpected empty identifier")
	}
	return &ast.Ident{Name: name}
}

func _iOrNil(name string) *ast.Ident {
	if name == "" {
		return nil
	} else {
		return _i(name)
	}
}

// recvTypeName is empty if there are no receivers.
// recvTypeName cannot contain any dots.
func _func(name string, recvRef string, recvTypeName string, params *ast.FieldList, results *ast.FieldList, b *ast.BlockStmt) *ast.FuncDecl {
	fn := &ast.FuncDecl{
		Name: _i(name),
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: b,
	}
	if recvRef == "" {
		recvRef = "_"
	}
	if recvTypeName != "" {
		fn.Recv = &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{_i(recvRef)},
					Type:  _i(recvTypeName),
				},
			},
		}
	}
	return fn
}

// Usage: _fields("a", "int", "b", "int32", ...) and so on.
// The types get parsed by _x().
// Identical types are compressed into Names automatically.
// args must always be even in length.
func _fields(args ...string) *ast.FieldList {
	list := []*ast.Field{}
	names := []*ast.Ident{}
	lastte_ := ""
	maybePop := func() {
		if len(names) > 0 {
			list = append(list, &ast.Field{
				Names: names,
				Type:  _x(lastte_),
			})
			names = []*ast.Ident{}
		}
	}
	for i := 0; i < len(args); i++ {
		name := args[i]
		te_ := args[i+1]
		i += 1
		if te_ == "" {
			panic("empty types not allowed")
		}
		if lastte_ == te_ {
			names = append(names, _i(name))
			continue
		} else {
			maybePop()
			names = append(names, _i(name))
			lastte_ = te_
		}
	}
	maybePop()
	return &ast.FieldList{
		List: list,
	}
}

// Parses simple expressions (but not all).
// Useful for parsing strings to ast nodes, like foo.bar["qwe"](),
// new(bytes.Buffer), *bytes.Buffer, package.MyStruct{FieldA:1}, numeric
//
//  * num/char (e.g. e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f')
//  * strings (e.g. "foo" or `\m\n\o`), nil, function calls
//  * square bracket indexing
//  * dot notation
//  * star expression for pointers
//  * struct construction
//  * nil
//  * type assertions, for EXPR.(EXPR) and also EXPR.(type)
//  * []type slice types
//  * [n]type array types
//  * &something referencing
//  * unary operations, namely
//    "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
//  * binary operations, namely
//    "||", "&&",
//    "==" | "!=" | "<" | "<=" | ">" | ">="
//    "+" | "-" | "|" | "^"
//    "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .
//
// NOTE: This isn't trying to implement everything -- just what is
// intuitively elegant to implement.  Why don't we use a parser generator?
// Cuz I'm testing the hypothesis that for the scope of what we're trying
// to do here, given this language, that this is easier to understand and
// maintain than using advanced tooling.
func _x(expr string, args ...interface{}) ast.Expr {
	if expr == "" {
		panic("_x requires argument")
	}
	expr = fmt.Sprintf(expr, args...)
	expr = strings.TrimSpace(expr)
	first := expr[0]

	// 1: Binary operators have a lower predecence than unary operators (or
	// monoids).
	left, op, right, ok := chopBinary(expr)
	if ok {
		return _b(_x(left), op, _x(right))
	}

	// 2: Unary operators that depend on the first letter.
	switch first {
	case '*':
		return &ast.StarExpr{
			X: _x(expr[1:]),
		}
	case '+', '-', '!', '^', '&':
		return &ast.UnaryExpr{
			Op: _op(expr[:1]),
			X:  _x(expr[1:]),
		}
	case '<':
		second := expr[1] // is required.
		if second != '-' {
			panic("unparseable expression " + expr)
		}
		return &ast.UnaryExpr{
			Op: _op("<-"),
			X:  _x(expr[2:]),
		}
	}

	// 3: Unary operators or literals that don't depend on the first letter,
	// and have some distinct suffix.
	if len(expr) > 1 {
		last := expr[len(expr)-1]
		switch last {
		case 'l':
			if expr == "nil" {
				return _i("nil")
			}
		case 'i':
			num := _x(expr[:len(expr)-1]).(*ast.BasicLit)
			if num.Kind != token.INT && num.Kind != token.FLOAT {
				panic("expected int or float before 'i'")
			}
			num.Kind = token.IMAG
			return num
		case '\'':
			if first != last {
				panic("unmatched quote")
			}
			return &ast.BasicLit{
				Kind:  token.CHAR,
				Value: string(expr[1 : len(expr)-1]),
			}
		case '"', '`':
			if first != last {
				panic("unmatched quote")
			}
			return &ast.BasicLit{
				Kind:  token.STRING,
				Value: string(expr),
			}
		case ')':
			left, _, right := chopRight(expr)
			if left == "" {
				// Special case, not a function call.
				return _x(right)
			} else if left[len(left)-1] == '.' {
				// Special case, a type assert.
				var x, t ast.Expr = _x(left[:len(left)-1]), nil
				if right == "type" {
					t = nil
				} else {
					t = _x(right)
				}
				return &ast.TypeAssertExpr{
					X:    x,
					Type: t,
				}
			}

			var fn = _x(left)
			var args = []ast.Expr{}
			parts := strings.Split(right, ",")
			for _, part := range parts {
				// NOTE: repeated commas have no effect,
				// nor do trailing commas.
				if len(part) > 0 {
					args = append(args, _x(part))
				}
			}
			return &ast.CallExpr{
				Fun:  fn,
				Args: args,
			}
		case '}':
			left, _, right := chopRight(expr)
			var ty = _x(left)
			var elts = []ast.Expr{}
			parts := strings.Split(right, ",")
			for _, part := range parts {
				if strings.TrimSpace(part) != "" {
					elts = append(elts, _kv(part))
				}
			}
			return &ast.CompositeLit{
				Type:       ty,
				Elts:       elts,
				Incomplete: false,
			}
		case ']':
			left, _, right := chopRight(expr)
			return &ast.IndexExpr{
				X:     _x(left),
				Index: _x(right),
			}
		}
	}
	// 4.  Monoids of array or slice type.
	// NOTE: []foo.bar requires this to have lower predence than dots.
	switch first {
	case '[':
		if expr[1] == ']' {
			return &ast.ArrayType{
				Len: nil,
				Elt: _x(expr[2:]),
			}
		} else {
			idx := strings.Index(expr, "]")
			if idx == -1 {
				panic(fmt.Sprintf(
					"mismatched '[' in slice expr %v",
					expr))
			}
			return &ast.ArrayType{
				Len: _x(expr[1:idx]),
				Elt: _x(expr[idx+1:]),
			}
		}
	}
	// Numeric int?  We do these before dots, because dots are legal in numbers.
	const (
		DGTS = `(?:[0-9]+)`
		HExX = `(?:0[xX][0-9a-fA-F]+)`
		PSCI = `(?:[eE]+?[0-9]+)`
		NSCI = `(?:[eE]-[1-9][0-9]+)`
		ASCI = `(?:[eE][-+]?[0-9]+)`
	)
	isInt, err := regexp.Match(
		`^-?(?:`+
			DGTS+`|`+
			HExX+`)`+PSCI+`?$`,
		[]byte(expr),
	)
	if err != nil {
		panic("should not happen")
	}
	if isInt {
		return &ast.BasicLit{
			Kind:  token.INT,
			Value: string(expr),
		}
	}
	// Numeric float?  We do these before dots, because dots are legal in floats.
	isFloat, err := regexp.Match(
		`^-?(?:`+
			DGTS+`\.`+DGTS+ASCI+`?|`+
			DGTS+NSCI+`)$`,
		[]byte(expr),
	)
	if err != nil {
		panic("should not happen")
	}
	if isFloat {
		return &ast.BasicLit{
			Kind:  token.FLOAT,
			Value: string(expr),
		}
	}
	// Last case, handle dots.
	// It's last, meaning it's got the highest precedence.
	if idx := strings.LastIndex(expr, "."); idx != -1 {
		return &ast.SelectorExpr{
			X:   _x(expr[:idx]),
			Sel: _i(expr[idx+1:]),
		}
	}
	return _i(expr)
}

// Returns idx=-1 if not a binary operator.
// Precedence    Operator
//     5             *  /  %  <<  >>  &  &^
//     4             +  -  |  ^
//     3             ==  !=  <  <=  >  >=
//     2             &&
//     1             ||
var sp = " "
var prec5 = strings.Split("*  /  %  <<  >>  &  &^", sp)
var prec4 = strings.Split("+ - | ^", sp)
var prec3 = strings.Split("== != < <= > >=", sp)
var prec2 = strings.Split("&&", sp)
var prec1 = strings.Split("||", sp)
var precs = [][]string{prec1, prec2, prec3, prec4, prec5}

// 0 for prec1... -1 if no match.
func lowestMatch(op string) int {
	for i, prec := range precs {
		for _, op2 := range prec {
			if op == op2 {
				return i
			}
		}
	}
	return -1
}

func _kv(kv string) *ast.KeyValueExpr {
	parts := strings.Split(kv, ":")
	if len(parts) != 2 {
		panic("_kv requires 1 colon")
	}
	return &ast.KeyValueExpr{
		Key:   _x(parts[0]),
		Value: _x(parts[1]),
	}
}

func _block(b ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: b,
	}
}

func _xs(exprs ...ast.Expr) []ast.Expr {
	return exprs
}

// Usage: _a(lhs1, lhs2, ..., ":=", rhs1, rhs2, ...)
// Token can be ":=", "=", "+=", etc.
// Other strings are automatically parsed as _x(arg).
func _a(args ...interface{}) *ast.AssignStmt {
	lhs := []ast.Expr(nil)
	tok := token.ILLEGAL
	rhs := []ast.Expr(nil)

	setTok := func(t token.Token) {
		if tok != token.ILLEGAL {
			panic("too many assignment operators")
		}
		tok = t
	}

	for _, arg := range args {
		if s, ok := arg.(string); ok {
			switch s {
			case "=", ":=", "+=", "-=", "*=", "/=", "%=",
				"&=", "|=", "^=", "<<=", ">>=", "&^=":
				setTok(_aop(s))
				continue
			default:
				arg = _x(s)
			}
		}
		// append to lhs or rhs depending on tok.
		if tok == token.ILLEGAL {
			lhs = append(lhs, arg.(ast.Expr))
		} else {
			rhs = append(rhs, arg.(ast.Expr))
		}
	}

	return &ast.AssignStmt{
		Lhs: lhs,
		Tok: tok,
		Rhs: rhs,
	}
}

func _not(x ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{
		Op: _op("!"),
		X:  x,
	}
}

// Binary expression.  x, y can be ast.Expr or string.
func _b(x interface{}, op string, y interface{}) ast.Expr {
	var xx, yx ast.Expr
	if xstr, ok := x.(string); ok {
		xx = _x(xstr)
	} else {
		xx = x.(ast.Expr)
	}
	if ystr, ok := y.(string); ok {
		yx = _x(ystr)
	} else {
		yx = y.(ast.Expr)
	}
	return &ast.BinaryExpr{
		X:  xx,
		Op: _op(op),
		Y:  yx,
	}
}

func _call(fn ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fn,
		Args: args,
	}
}

func _ta(x ast.Expr, t ast.Expr) *ast.TypeAssertExpr {
	return &ast.TypeAssertExpr{
		X:    x,
		Type: t,
	}
}

func _sel(x ast.Expr, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   x,
		Sel: _i(sel),
	}
}

func _ix(x ast.Expr, idx ast.Expr) *ast.IndexExpr {
	return &ast.IndexExpr{
		X:     x,
		Index: idx,
	}
}

func _s(s string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(s),
	}
}

func _ref(x ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{
		Op: token.AND,
		X:  x,
	}
}

func _deref(x ast.Expr) *ast.StarExpr {
	return &ast.StarExpr{
		X: x,
	}
}

// NOTE: Same as _deref, but different contexts.
func _ptr(x ast.Expr) *ast.StarExpr {
	return &ast.StarExpr{
		X: x,
	}
}

//----------------------------------------
// AST Construction (Stmt)

func _if(cond ast.Expr, b ...ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: cond,
		Body: _block(b...),
	}
}

func _ife(cond ast.Expr, bdy, els ast.Stmt) *ast.IfStmt {
	if _, ok := bdy.(*ast.BlockStmt); !ok {
		bdy = _block(bdy)
	}
	if _, ok := els.(*ast.BlockStmt); !ok {
		if _, ok := els.(*ast.IfStmt); !ok {
			els = _block(els)
		}
	}
	return &ast.IfStmt{
		Cond: cond,
		Body: bdy.(*ast.BlockStmt),
		Else: els,
	}
}

func _return(results ...ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: results,
	}
}

func _continue(label string) *ast.BranchStmt {
	return &ast.BranchStmt{
		Tok:   token.CONTINUE,
		Label: _i(label),
	}
}

func _break(label string) *ast.BranchStmt {
	return &ast.BranchStmt{
		Tok:   token.BREAK,
		Label: _i(label),
	}
}

func _goto(label string) *ast.BranchStmt {
	return &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: _i(label),
	}
}

func _fallthrough(label string) *ast.BranchStmt {
	return &ast.BranchStmt{
		Tok:   token.FALLTHROUGH,
		Label: _i(label),
	}
}

// even/odd args are paired,
// name1, path1, name2, path2, etc.
func _imports(nameAndPaths ...string) *ast.GenDecl {
	decl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{},
	}
	for i := 0; i < len(nameAndPaths); i += 2 {
		spec := &ast.ImportSpec{
			Name: _i(nameAndPaths[i]),
			Path: _s(nameAndPaths[i+1]),
		}
		decl.Specs = append(decl.Specs, spec)
	}
	return decl
}

func _for(init ast.Stmt, cond ast.Expr, post ast.Stmt, b ...ast.Stmt) *ast.ForStmt {
	return &ast.ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: _block(b...),
	}
}

func _loop(b ...ast.Stmt) *ast.ForStmt {
	return _for(nil, nil, nil, b...)
}

func _once(b ...ast.Stmt) *ast.ForStmt {
	b = append(b, _break(""))
	return _for(nil, nil, nil, b...)
}

func _len(x ast.Expr) *ast.CallExpr {
	return _call(_i("len"), x)
}

func _var(name string, type_ ast.Expr, value ast.Expr) *ast.DeclStmt {
	if value == nil {
		return &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names:  []*ast.Ident{_i(name)},
						Type:   type_,
						Values: nil,
					},
				},
			},
		}
	} else {
		return &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names:  []*ast.Ident{_i(name)},
						Type:   type_,
						Values: []ast.Expr{value},
					},
				},
			},
		}
	}
}

func defaultExpr(k reflect.Kind) ast.Expr {
	switch k {
	case reflect.Interface, reflect.Ptr, reflect.Slice:
		return _x("nil")
	case reflect.String:
		return _x("\"\"")
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16,
		reflect.Int8, reflect.Uint, reflect.Uint64, reflect.Uint32,
		reflect.Uint16, reflect.Uint8:
		return _x("0")
	case reflect.Bool:
		return _x("false")
	default:
		panic(fmt.Sprintf("no fixed default expression for kind %v", k))
	}
}

// binary and unary operators, excluding assignment operators.
func _op(op string) token.Token {
	switch op {
	case "+":
		return token.ADD
	case "-":
		return token.SUB
	case "*":
		return token.MUL
	case "/":
		return token.QUO
	case "%":
		return token.REM
	case "&":
		return token.AND
	case "|":
		return token.OR
	case "^":
		return token.XOR
	case "<<":
		return token.SHL
	case ">>":
		return token.SHR
	case "&^":
		return token.AND_NOT
	case "&&":
		return token.LAND
	case "||":
		return token.LOR
	case "<-":
		return token.ARROW
	case "++":
		return token.INC
	case "--":
		return token.DEC
	case "==":
		return token.EQL
	case "<":
		return token.LSS
	case ">":
		return token.GTR
	case "!":
		return token.NOT
	case "!=":
		return token.NEQ
	case "<=":
		return token.LEQ
	case ">=":
		return token.GEQ
	default:
		panic("unrecognized binary/unary operator " + op)
	}
}

// assignment operators.
func _aop(op string) token.Token {
	switch op {
	case "=":
		return token.ASSIGN
	case ":=":
		return token.DEFINE
	case "+=":
		return token.ADD_ASSIGN
	case "-=":
		return token.SUB_ASSIGN
	case "*=":
		return token.MUL_ASSIGN
	case "/=":
		return token.QUO_ASSIGN
	case "%=":
		return token.REM_ASSIGN
	case "&=":
		return token.AND_ASSIGN
	case "|=":
		return token.OR_ASSIGN
	case "^=":
		return token.XOR_ASSIGN
	case "<<=":
		return token.SHL_ASSIGN
	case ">>=":
		return token.SHR_ASSIGN
	case "&^=":
		return token.AND_NOT_ASSIGN
	default:
		panic("unrecognized assignment operator " + op)
	}
}

//----------------------------------------
// AST Compile-Time

func _ctif(cond bool, then_, else_ ast.Stmt) ast.Stmt {
	if cond {
		return then_
	} else if else_ != nil {
		return else_
	} else {
		return &ast.EmptyStmt{Implicit: true} // TODO
	}
}

//----------------------------------------
// AST query and manipulation.

func importPathForName(name string, imports *ast.GenDecl) (path string, exists bool) {
	if imports.Tok != token.IMPORT {
		panic("unexpected ast.GenDecl token " + imports.Tok.String())
	}
	for _, spec := range imports.Specs {
		if ispec, ok := spec.(*ast.ImportSpec); ok {
			if ispec.Name.Name == name {
				path, err := strconv.Unquote(ispec.Path.Value)
				if err != nil {
					panic("malformed path " + ispec.Path.Value)
				}
				return path, true
			}
		}
	}
	return "", false
}

func importNameForPath(path string, imports *ast.GenDecl) (name string, exists bool) {
	if imports.Tok != token.IMPORT {
		panic("unexpected ast.GenDecl token " + imports.Tok.String())
	}
	for _, spec := range imports.Specs {
		if ispec, ok := spec.(*ast.ImportSpec); ok {
			specPath, err := strconv.Unquote(ispec.Path.Value)
			if err != nil {
				panic("malformed path " + ispec.Path.Value)
			}
			if specPath == path {
				return ispec.Name.Name, true
			}
		}
	}
	return "", false
}

func rootScope(scope *ast.Scope) *ast.Scope {
	for scope.Outer != nil {
		scope = scope.Outer
	}
	return scope
}

func addImport(imports *ast.GenDecl, scope *ast.Scope, name, path string) {
	epath, exists := importPathForName(name, imports)
	if path == epath {
		return
	} else if exists {
		panic(fmt.Sprintf("import already exists for name %v", name))
	} else {
		imports.Specs = append(imports.Specs, &ast.ImportSpec{
			Name: _i(name),
			Path: _s(path),
		})
		addPkgNameToRootScope(name, rootScope(scope))
	}
}

func addImportAuto(imports *ast.GenDecl, scope *ast.Scope, name, path string) string {
	if path0, exists := importPathForName(name, imports); exists {
		if path0 == path {
			return name
		}
		for i := 1; ; i++ {
			n := fmt.Sprintf("%v%v", name, i)
			if _, exists := importPathForName(n, imports); !exists {
				addImport(imports, scope, n, path)
				return n
			}
		}
	} else {
		addImport(imports, scope, name, path)
		return name
	}
}

func addPkgNameToRootScope(name string, scope *ast.Scope) {
	if scope.Outer != nil {
		panic("should not happen")
	}
	scope.Insert(ast.NewObj(ast.Pkg, name))
}

func addVars(scope *ast.Scope, names ...string) {
	for _, name := range names {
		if scope.Lookup(name) != nil {
			panic("already declared in scope")
		} else {
			scope.Insert(ast.NewObj(ast.Var, name))
		}
	}
}

func addVarUniq(scope *ast.Scope, name string) string {
OUTER:
	for i := 0; ; i++ {
		tryName := name
		if i > 0 {
			tryName = fmt.Sprintf("%v%v", name, i)
		}
		s := scope
		for {
			if s.Lookup(tryName) != nil {
				continue OUTER
			}
			if s.Outer == nil {
				break
			} else {
				s = s.Outer
			}
		}
		scope.Insert(ast.NewObj(ast.Var, tryName))
		return tryName
	}
}

func goTypeExprString(rootPkg *amino.Package, imports *ast.GenDecl, scope *ast.Scope, isPtr bool, info *amino.TypeInfo) string {
	if isPtr {
		return "*" + goTypeExprString(rootPkg, imports, scope, false, info)
	}
	// Below, assume isPtr is false.
	k := info.Type.Kind()
	if k == reflect.Array || k == reflect.Slice {
		return fmt.Sprintf("[]%v", goTypeExprString(rootPkg, imports, scope, info.ElemIsPtr, info.Elem))
	}
	pkg := info.Package
	if pkg == nil {
		panic(fmt.Sprintf("package not registered for type %v", info))
	}
	if pkg == rootPkg || pkg.GoPkgPath == "" {
		return fmt.Sprintf("%v", info.Type.Name())
	} else {
		pkgName := addImportAuto(imports, scope, pkg.Name, pkg.GoPkgPath)
		return fmt.Sprintf("%v.%v", pkgName, info.Type.Name())
	}
}

func p3goTypeExprString(imports *ast.GenDecl, scope *ast.Scope, info *amino.TypeInfo) string {
	k := info.ReprType.Type.Kind()
	switch k {
	case reflect.Array, reflect.Slice:
		return p3goListTypeExprString(info.Elem)
	case reflect.Interface:
		anypb := addImportAuto(imports, scope, "anypb", "google.golang.org/protobuf/types/known/anypb")
		return fmt.Sprintf("*%v.Any", anypb)
	default:
		// Special cases.
		// TODO: somehow refactor into wellknown.go
		switch info.ReprType.Type {
		case timeType:
			pkgName := addImportAuto(
				imports, scope, "timestamppb", "google.golang.org/protobuf/types/known/timestamppb")
			return fmt.Sprintf("*%v.%v", pkgName, "Timestamp")
		case durationType:
			pkgName := addImportAuto(
				imports, scope, "durationpb", "google.golang.org/protobuf/types/known/durationpb")
			return fmt.Sprintf("*%v.%v", pkgName, "Duration")
		}
		pkg := info.Package
		if pkg == nil {
			panic(fmt.Sprintf("package not registered for type %v", info))
		}
		pkgName := addImportAuto(imports, scope, pkg.Name+"pb", pkg.P3GoPkgPath)
		return fmt.Sprintf("*%v.%v", pkgName, info.ReprType.Type.Name())
	}
}

// NOTE: assumes same pacakge, so the returned expr isn't a selector.
func p3goListTypeExprString(info *amino.TypeInfo) string {
	einfo := info
	counter := 0
	for einfo.ReprType.Type.Kind() == reflect.Array || einfo.ReprType.Type.Kind() == reflect.Slice {
		counter++
		einfo = einfo.Elem.ReprType
	}
	ename := einfo.Type.Name()
	if ename == "uint8" {
		counter--
		if counter == 0 {
			return "[]byte"
		} else {
			listSfx := strings.Repeat("List", counter)
			return fmt.Sprintf("*Bytes%v", listSfx)
		}
	} else {
		listSfx := strings.Repeat("List", counter)
		return fmt.Sprintf("*%v%v", capitalize(einfo.Type.Name()), listSfx)
	}
}

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

// Find struct fields that are nested list types.
// If not a struct, assume an implicit struct with single field.
// If type is amino.Marshaler, find values/fields from the repr.
// Pointers are ignored, even for the terminal type.
// e.g. if TypeInfo.ReprType.Type is
//  * struct{ [][]int, [][]string } -> return [][]int, [][]string
//  * [][]int -> return [][]int
//  * [][][]int -> return [][][]int, [][]int
//  * [][][]byte -> return [][][]byte (but not [][]byte, which is just repeated bytes).
//  * [][][][]int -> return [][][][]int, [][][]int, [][]int.
// The results are uniq'd and sorted somehow.
func findNestedLists(info *amino.TypeInfo, found *map[reflect.Type]struct{}) {
	if found == nil {
		*found = map[reflect.Type]struct{}{}
	}
	switch info.ReprType.Type.Kind() {
	case reflect.Struct:
		for _, field := range info.ReprType.Fields {
			findNestedLists(field.TypeInfo, found)
		}
		return
	case reflect.Array, reflect.Slice:
		ert := info.ReprType.Elem.ReprType.Type
		if ert.Kind() == reflect.Array ||
			ert.Kind() == reflect.Slice {

			eert := info.ReprType.Elem.ReprType.Elem.ReprType.Type
			if eert.Kind() == reflect.Uint8 {
				return
			} else {
				llrt := reprTypeToType(info)
				(*found)[llrt] = struct{}{}
				return
			}
		}
	}
}

func sortFound(found map[reflect.Type]struct{}) (res []reflect.Type) {
	for rt, _ := range found {
		res = append(res, rt)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].String() < res[j].String() })
	return res
}

func reprTypeToType(info *amino.TypeInfo) reflect.Type {
	info = info.ReprType
	switch info.Type.Kind() {
	case reflect.Ptr:
		panic("should not happen")
	case reflect.Array:
		return reflect.ArrayOf(info.Type.Len(), reprTypeToType(info.Elem))
	case reflect.Slice:
		return reflect.SliceOf(reprTypeToType(info.Elem))
	default:
		return info.ReprType.Type
	}
}
