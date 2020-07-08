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
	"strconv"
	"strings"

	"github.com/tendermint/go-amino"
)

// Given genproto generated schema files for Go objects, generate
// mappers to and from pb messages.  The purpose of this is to let Amino
// use already-optimized probuf logic for serialization.
//
// pbpkg is the import path to the pb compiled message structs.
// Regardless of the import path, the local pkg identifier is
// always "pbpkg"
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
	var imports = _imports(
		"pbpkg", pkg.GoP3PkgPath,
		"proto", "google.golang.org/protobuf/proto",
		"amino", "github.com/tendermint/go-amino")
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

		// Generate translation functions.
		bindings, err := generateTranslationMethodsForType(imports, pkg, info)
		if err != nil {
			return file, err
		}
		file.Decls = append(file.Decls, bindings.toProto)
		file.Decls = append(file.Decls, bindings.fromProto)

		// Generate common methods.
		decls, err := generateCommonMethodsForType(imports, pkg, info)
		if err != nil {
			return file, err
		}
		file.Decls = append(file.Decls, decls...)
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

type translationBindings struct {
	// func (Obj) ToPBMessage() (proto.Message, error)
	toProto *ast.FuncDecl
	// func (*Obj) FromPBMessage(proto.Message) (error)
	fromProto *ast.FuncDecl
}

// modified imports if necessary.
func generateTranslationMethodsForType(imports *ast.GenDecl, pkg *amino.Package, info *amino.TypeInfo) (bindings translationBindings, err error) {
	if info.Type.Kind() != reflect.Struct {
		panic("not yet supported")
	}

	//////////////////
	// ToPBMessage()
	{
		var b = []ast.Stmt{}
		// Body: constructor for pb message.
		b = append(b, _a("pb", ":=", _x("new~(~pbpkg.%v~)", info.Type.Name())))
		b = append(b, _a("err", ":=", "error(nil)"))

		// Body: copying over fields.
		b = append(b, _block(go2pbStmts(imports, _i("pb"), _i("o"), true, info)...))

		// Body: return value.
		b = append(b, _a("msg", "=", "pb"))
		b = append(b, _return())

		// Set toProto function.
		bindings.toProto = _func("ToPBMessage",
			"o", info.Type.Name(),
			_fields("cdc", "*amino.Codec"),
			_fields("msg", "proto.Message", "err", "error"),
			_block(b...),
		)
	}

	//////////////////
	// FromPBMessage()
	{
		var b = []ast.Stmt{}

		// Body: constructor for pb message.
		b = append(b, _a("err", ":=", "pb"))
		b = append(b, _a("pb", ":=", _x("*pbpkg.%v~(~nil~)", info.Type.Name())))

		// Body: type assert to pb.
		b = append(b, _a("pb", "=", _x("msg.~(~*pbpkg.%v~)", info.Type.Name())))

		// Body: copying over fields.
		b = append(b, _block(pb2goStmts(imports, _i("o"), true, info, _i("pb"))...))

		// Body: return value.
		b = append(b, _a("msg", "=", "pb"))
		b = append(b, _return())

		bindings.fromProto = _func("FromPBMessage",
			"o", "*"+info.Type.Name(),
			_fields("cdc", "*amino.Codec", "msg", "proto.Message"),
			_fields("err", "error"),
			_block(b...),
		)
	}
	return
}

// imports: global imports -- may be modified.
// pbo: protobuf variable or field.
// goo: native go variable or field.
// gooIsPtr: whether goo is ptr.
// gooType: type info for goo's type (elem type if pointer).
// CONTRACT: pbo is assignable.
//  * The general case is `_a(pbo, "=", goo)`
//  * The struct case is like `_a(_sel(pbo, field.Name), "=", goo)`
func go2pbStmts(imports *ast.GenDecl, pbo ast.Expr, goo ast.Expr, gooIsPtr bool, gooType *amino.TypeInfo) (b []ast.Stmt) {

	// Special case if nil-pointer.
	if gooIsPtr || gooType.Type.Kind() == reflect.Interface {
		defer func() {
			// Wrap penultimate b with if statement.
			b = []ast.Stmt{_if(_b(goo, "!=", _i("nil")),
				b...,
			)}
		}()
	}
	// Below, we can assume that goo isn't nil.

	// Special case if IsAminoMarshaler.
	if gooType.IsAminoMarshaler {
		// First, derive repr instance.
		b = append(b,
			_a("goor", "err", ":=", _call(_sel(goo, "MarshalAmino"))),
			_if(_x("err__!=__nil"),
				_return(_x("nil"), _i("err")),
			),
		)
		goo = _i("goor") // switcharoo
		gooType = gooType.ReprType
	} else {
		// Special case if goo is pointer.
		// NOTE: IsAminoMarshaler never returns pointers.
		if gooIsPtr {
			b = append(b,
				_a("dgoo", ":=", _deref(goo)))
			goo = _i("dgoo") // switcharoo
		}
	}
	// Below, we can assume that goo isn't a pointer.
	// Below, we can assume that gooType isn't amino.Marshaler

	// General case
	switch gooType.Type.Kind() {

	case reflect.Interface:
		b = append(b,
			// see generateCommonMethodForType().
			_a("typeUrl", ":=", _call(_sel(goo, "GetTypeUrl"), nil)),
			_a("bz", "err", ":=", _call(_sel(_i("cdc"), "MarshalBinaryBare"), goo)),
			_if(_x("err__!=__nil"),
				_return(),
			),
			_a(pbo, "=", "&anypb.Any{TypeUrl:typeUrl,Value:bz}"),
		)

	case reflect.Int:
		b = append(b,
			_a(pbo, "=", _call(_i("int64"), goo)))
	case reflect.Int16, reflect.Int8:
		b = append(b,
			_a(pbo, "=", _call(_i("int32"), goo)))
	case reflect.Uint:
		b = append(b,
			_a(pbo, "=", _call(_i("uint64"), goo)))
	case reflect.Uint16, reflect.Uint8:
		b = append(b,
			_a(pbo, "=", _call(_i("uint32"), goo)))

	case reflect.Array, reflect.Slice:
		var gooeIsPtr = gooType.ElemIsPtr
		var gooeType = gooType.Elem
		var pboeTyp string
		switch gooeType.Type.Kind() {
		case reflect.Interface:
			pboeTyp = ("*anypb.Any")
		case reflect.Struct:
			pkg := gooeType.Package
			pkgName := addImportAuto(pkg.Name, pkg.GoP3PkgPath, imports)
			pboeTyp = fmt.Sprintf("*%v.%v", pkgName, gooeType.Type.Name())
		default:
			pboeTyp = gooeType.Type.Name()
		}

		// Construct, translate, assign.
		b = append(b,
			_a("gool", ":=", _len(goo)),
			_var("pbos", nil, _x("make~(~[]%v,gool~)", pboeTyp)),
			_for(
				_a("i", ":=", "0"),
				_x("i__<__gool"),
				_a("i", "+=", "1"),
				_block(
					// Translate in place.
					_a("gooe", ":=", _ix(goo, _i("i"))),
					_block(go2pbStmts(imports, _x("pbos~[~i~]"), _i("gooe"), gooeIsPtr, gooeType)...),
				),
			),
			_a(pbo, "=", "pbos"),
		)

	case reflect.Struct:
		pkg := gooType.Package
		pkgName := addImportAuto(pkg.Name, pkg.GoP3PkgPath, imports)
		dpboTyp := fmt.Sprintf("%v.%v", pkgName, gooType.Type.Name())

		b = append(b,
			_a(pbo, "=", _x("new~(~%v~)", dpboTyp)))

		for _, field := range gooType.Fields {
			var goofIsPtr = field.IsPtr()
			var goofType = field.TypeInfo.ReprType
			var goof = _sel(goo, field.Name) // next goo
			var pbof = _sel(pbo, field.Name) // next pbo

			// Translate in place.
			b = append(b,
				_block(go2pbStmts(imports, pbof, goof, goofIsPtr, goofType)...),
			)
		}

	default:
		// General translation.
		b = append(b, _a(pbo, "=", goo))

	}
	return b
}

// imports: global imports -- used to look up package names.
// goo: native go variable or field.
// gooIsPtr: is goo a pointer?
// gooType: type info for goo's ultimate type (elem if pointer)..
// pbo: protobuf variable or field.
// CONTRACT: goo is addressable.
func pb2goStmts(imports *ast.GenDecl, goo ast.Expr, gooIsPtr bool, gooType *amino.TypeInfo, pbo ast.Expr) (b []ast.Stmt) {

	// Special case if pbo is zero.
	//
	// We especially want this behavior (and optimization) for for
	// amino.Marshalers, because of the construction cost.
	//
	// Ignoring the optimization, we could duplicate these checks for every
	// switch case in the main body of this function, but that would be
	// duplicating a lot of code.
	var pboZero ast.Expr
	// Determine pbo type from gooType.ReprType.
	switch gooType.ReprType.Type.Kind() {
	case reflect.Struct:
		pboZero = _x("nil") // In protobuf is pointer.
	case reflect.Array:
		pboZero = nil // Do not wrap b.
	default:
		pboZero = defaultExpr(gooType.ReprType.Type.Kind())
	}
	if pboZero != nil {
		defer func() {
			// Wrap penultimate b with if statement.
			b = []ast.Stmt{_if(_b(pbo, "!=", pboZero),
				b...,
			)}
		}()
	}
	// Below, we can assume that pbo isn't nil or zero.

	// Special case if IsAminoMarshaler.
	// NOTE: doesn't matter whether goo is ptr or not.
	if gooType.IsAminoMarshaler {
		// First, construct new repr instance.
		b = append(b,
			_var(
				"goor",
				_x(gooType.ReprType.Type.String()), // NOTE: never pointer.
				nil,
			),
			_var( // for checking whether goor is zero
				"goor2",
				_x(gooType.ReprType.Type.String()),
				nil,
			),
		)
		// Then, transcribe to repr var.
		b = append(b, _block(
			pb2goStmts(imports, _i("goor"), false, gooType.ReprType, pbo)...,
		))
		// Finally, maybe assign.
		b = append(b,
			_if(_x("goor__!=__goor2"),
				_a("err", ":=", _call(_sel(goo, "UnmarshalAmino"), _i("goor"))),
				_if(_x("err__!=__nil"),
					_return(),
				),
			),
		)
		return
	} else {
		// Special case if goo is pointer.
		if gooIsPtr {
			pkg := gooType.Package
			pkgName := addImportAuto(pkg.Name, pkg.GoPkgPath, imports)
			dgooTyp := fmt.Sprintf("%v.%v", pkgName, gooType.Type.Name())
			b = append(b,
				_a(goo, ":=", _x("new~(~%v~)", dgooTyp)),
				_a("dgoo", ":=", _deref(goo)),
			)
			goo = _i("dgoo") // switcheroo
		}
	}
	// Below, we can assume that goo isn't a pointer.

	// General case
	switch gooType.Type.Kind() {

	case reflect.Interface:
		b = append(b,
			// see generateCommonMethodForType().
			_a("typeUrl", ":=", _sel(pbo, "TypeUrl")),
			_a("bz", ":=", _sel(pbo, "Value")),
			_a("goop", ":=", _ref(goo)),
			_a("err", ":=", "cdc.UnmarshalBinaryAny(typeUrl,bz,goop)"),
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

	case reflect.Array, reflect.Slice:
		var gooeType = gooType.Elem
		var gooeIsPtr = gooType.ElemIsPtr
		var gooeTyp string
		switch gooeType.Type.Kind() {
		case reflect.Interface:
			pkg := gooeType.Package
			pkgName := addImportAuto(pkg.Name, pkg.GoPkgPath, imports)
			gooeTyp = fmt.Sprintf("%v.%v", pkgName, gooeType.Type.Name())
		case reflect.Struct:
			pkg := gooeType.Package
			pkgName := addImportAuto(pkg.Name, pkg.GoPkgPath, imports)
			if gooeIsPtr {
				gooeTyp = fmt.Sprintf("*%v.%v", pkgName, gooeType.Type.Name())
			} else {
				gooeTyp = fmt.Sprintf("%v.%v", pkgName, gooeType.Type.Name())
			}
		default:
			gooeTyp = gooeType.Type.Name()
		}

		// Construct, translate, assign.
		b = append(b,
			_a("pbol", ":=", _len(pbo)),
			_var("goos", nil, _x("make~(~[]%v,pbol~)", gooeTyp)),
			_for(
				_a("i", ":=", "0"),
				_x("i__<__pbol"),
				_a("i", "+=", "1"),
				_block(
					// Translate in place.
					_a("pboe", ":=", _ix(pbo, _i("i"))),
					_block(pb2goStmts(imports, _x("goos~[~i~]"), gooeIsPtr, gooeType, _i("pboe"))...),
				),
			),
			_a(goo, "=", "goos"),
		)

	case reflect.Struct:
		for _, field := range gooType.Fields {
			var pbof = _sel(pbo, field.Name) // next pbo.
			var goofIsPtr = field.IsPtr()
			var goofType = field.TypeInfo.ReprType
			var goof = _sel(goo, field.Name) // next goo.

			// Translate in place.
			b = append(b,
				_block(pb2goStmts(imports, goof, goofIsPtr, goofType, pbof)...),
			)
		}

	default:
		// General translation.
		b = append(b, _a(goo, "=", pbo))
	}
	return b
}

func generateCommonMethodsForType(imports *ast.GenDecl, pkg *amino.Package, info *amino.TypeInfo) (decls []ast.Decl, err error) {
	return []ast.Decl{
		_func("GetTypeURL",
			"", info.Type.Name(),
			_fields(),
			_fields("typeURL", "string"),
			_block(
				_return(_s(info.TypeURL)),
			),
		),
	}, nil
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
	lastType := ""
	maybePop := func() {
		if len(names) > 0 {
			list = append(list, &ast.Field{
				Names: names,
				Type:  _x(lastType),
			})
			names = []*ast.Ident{}
		}
	}
	for i := 0; i < len(args); i++ {
		name := args[i]
		typ_ := args[i+1]
		i += 1
		if typ_ == "" {
			panic("empty types not allowed")
		}
		if lastType == typ_ {
			names = append(names, _i(name))
			continue
		} else {
			maybePop()
			names = append(names, _i(name))
			lastType = typ_
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
	last := expr[len(expr)-1]

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
			Op: token.Lookup(expr[:1]),
			X:  _x(expr[1:]),
		}
	case '<':
		second := expr[1] // is required.
		if second != '-' {
			panic("unparseable expression " + expr)
		}
		return &ast.UnaryExpr{
			Op: token.Lookup("<-"),
			X:  _x(expr[2:]),
		}
	}

	// 3: Unary operators or literals that don't depend on the first letter.
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
		parts := strings.Split(right, ",")
		var ty = _x(left)
		var elts = []ast.Expr{}
		for _, part := range parts {
			elts = append(elts, _kv(part))
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

	for _, arg := range args {
		if s, ok := arg.(string); ok {
			if tok == token.ILLEGAL {
				switch s {
				case "=":
					tok = token.ASSIGN
					continue
				case ":=":
					tok = token.DEFINE
					continue
				default:
					arg = _x(s)
				}
			} else {
				panic("too many assignment operators")
			}
		}
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
		Op: token.Lookup(op),
		Y:  yx,
	}
}

func _call(fn ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fn,
		Args: args,
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

func addImport(name, path string, imports *ast.GenDecl) {
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
	}
}

func addImportAuto(name, path string, imports *ast.GenDecl) string {
	if _, exists := importPathForName(name, imports); exists {
		for i := 1; ; i++ {
			n := fmt.Sprintf("%v%v", name, i)
			if _, exists := importPathForName(n, imports); !exists {
				addImport(n, path, imports)
				return n
			}
		}
	} else {
		addImport(name, path, imports)
		return name
	}
}
