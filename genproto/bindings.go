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
		Name:  astId(pkg.GoPkg),
		Decls: nil,
	}

	for _, type_ := range rtz {
		info, err := cdc.GetTypeInfo(type_)
		if err != nil {
			return file, err
		}
		if info.Type.Kind() != reflect.Struct {
			continue // Maybe consider supporting more.
		}

		// Generate translation functions.
		bindings, err := generateTranslationMethodsForType(pkg, info)
		if err != nil {
			return file, err
		}
		file.Decls = append(file.Decls, bindings.toProto)
		file.Decls = append(file.Decls, bindings.fromProto)

		// Generate common methods.
		decls, err := generateCommonMethodsForType(pkg, info)
		if err != nil {
			return file, err
		}
		file.Decls = append(file.Decls, decls...)
	}
	return file, nil
}

// Writes in the same directory as the origin package.
// Assumes pb imports in origGoPkg+"/pb".
func WriteProtoBindings(pkgs ...*amino.Package) {
	for _, pkg := range pkgs {
		filename := path.Join(pkg.Dirname, "pb_bindings.go")
		fmt.Printf("writing proto3 bindings to %v for package %v\n", filename, pkg)
		err := WriteProtoBindingsForTypes(filename, pkg, pkg.Types...)
		if err != nil {
			panic(err)
		}
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

func generateTranslationMethodsForType(pkg *amino.Package, info *amino.TypeInfo) (bindings translationBindings, err error) {
	if info.Type.Kind() != reflect.Struct {
		panic("not yet supported")
	}

	//////////////////
	// ToPBMessage()
	{
		var body = []ast.Stmt{}
		// Body: constructor for pb message.
		body = append(body, astDefine1(
			astExpr("pb"), astExpr("err"),
			astExpr("new(pbpkg."+info.Type.Name()+")"),
		))

		// Body: copying over fields.
		for _, field := range info.Fields {
			// If interface/any, special translation.
			if field.Type.Kind() == reflect.Interface {
				body = append(body, astDefine1(
					astExpr("typeUrl"),
					astExpr("o.GetTypeUrl()"),
					// see generateCommonMethodForType().
				))
				body = append(body, astDefine1(
					astExpr("bz"), astExpr("err"),
					astExpr("cdc.MarshalBinaryBare(o."+field.Name+")"),
				))
				body = append(body, astIf(
					astExpr("err!=nil"),
					astReturn(),
				))
				body = append(body, astAssign1(
					astExpr("pb."+field.Name),
					astExpr("anypb.Any{TypeUrl:typeUrl,Value:bz}"),
				))
			} else {
				// General translation.
				// TODO: ensure correctness of casting.
				body = append(body, astAssign1(
					astExpr("pb."+field.Name),
					astExpr("o."+field.Name),
				))
			}
		}
		// Body: return value.
		body = append(body, astAssign2(
			astExpr("msg"),
			astExpr("pb"),
		))
		body = append(body, astReturn())

		// Set toProto function.
		bindings.toProto = astFunc("ToPBMessage",
			"o", info.Type.Name(),
			astFields("cdc", "*amino.Codec"),
			astFields("msg", "proto.Message", "err", "error"),
			astBlock(body...),
		)
	}

	//////////////////
	// FromPBMessage()
	{
		var body = []ast.Stmt{}

		// Body: constructor for pb message.
		body = append(body, astDefine2(
			astExpr("err"),
			astExpr("error(nil)"),
			astExpr("pb"),
			astExpr("*pbpkg."+info.Type.Name()+"(nil)"),
		))

		// Body: type assert to pb.
		body = append(body, astAssign1(
			astExpr("pb"),
			astExpr("msg.(*pbpkg."+info.Type.Name()+")"),
		))

		// Body: copying over fields.
		for _, field := range info.Fields {
			// If interface/any, special translation.
			if field.Type.Kind() == reflect.Interface {
				body = append(body,
					astDefine1(
						astExpr("any"),
						astExpr("pb."+field.Name),
					),
					astDefine1(
						astExpr("typeUrl"),
						astExpr("any.TypeUrl"),
						astExpr("bz"),
						astExpr("any.Value"),
					),
					astDefine1(
						astExpr("err"),
						astExpr("cdc.UnmarshalBinaryAny(typeUrl,bz, &o."+field.Name+")"),
					),
					astIf(
						astExpr("err!=nil"),
						astReturn(),
					),
				)
			} else {
				// General translation.
				// TODO: ensure correctness of casting.
				body = append(body, astAssign1(
					astExpr("pb."+field.Name),
					astExpr("o."+field.Name),
				))
			}
		}
		// Body: return value.
		body = append(body, astAssign2(
			astExpr("msg"),
			astExpr("pb"),
		))
		body = append(body, astReturn())

		bindings.fromProto = astFunc("FromPBMessage",
			"o", "*"+info.Type.Name(),
			astFields("cdc", "*amino.Codec", "msg", "proto.Message"),
			astFields("err", "error"),
			astBlock(body...),
		)
	}
	return
}

func generateCommonMethodsForType(pkg *amino.Package, info *amino.TypeInfo) (decls []ast.Decl, err error) {
	return []ast.Decl{
		astFunc("GetTypeURL",
			"", info.Type.Name(),
			astFields(),
			astFields("typeURL", "string"),
			astBlock(
				astReturn(astString(info.TypeURL)),
			),
		),
	}, nil
}

//----------------------------------------
// ast convenience

func astId(name string) *ast.Ident {
	return &ast.Ident{Name: name}
}

// recvTypeName is empty if there are no receivers.
// recvTypeName cannot contain any dots.
func astFunc(name string, recvRef string, recvTypeName string, params *ast.FieldList, results *ast.FieldList, body *ast.BlockStmt) *ast.FuncDecl {
	fn := &ast.FuncDecl{
		Name: astId(name),
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: body,
	}
	if recvRef == "" {
		recvRef = "_"
	}
	if recvTypeName != "" {
		fn.Recv = &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{astId(recvRef)},
					Type:  astId(recvTypeName),
				},
			},
		}
	}
	return fn
}

// Usage: astFields("a", "int", "b", "int32", ...) and so on.
// The types get parsed by astExpr().
// Identical types are compressed into Names automatically.
// args must always be even in length.
func astFields(args ...string) *ast.FieldList {
	list := []*ast.Field{}
	names := []*ast.Ident{}
	lastType := ""
	maybePop := func() {
		if len(names) > 0 {
			list = append(list, &ast.Field{
				Names: names,
				Type:  astExpr(lastType),
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
			names = append(names, astId(name))
			continue
		} else {
			maybePop()
			names = append(names, astId(name))
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
//  * num/char (e.g. e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f')
//  * strings (e.g. "foo" or `\m\n\o`), nil, function calls
//  * square bracket indexing
//  * dot notation
//  * star expression for pointers
//  * struct construction
//  * nil
//  * type assertions, for EXPR.(EXPR) and also EXPR.(type).
// NOTE: If the implementation isn't intuitive, it doesn't belong here.
func astExpr(expr string) ast.Expr {
	if expr == "" {
		panic("astExpr requires argument")
	}
	if expr[0] == '*' {
		return &ast.StarExpr{
			X: astExpr(expr[1:]),
		}
	}
	lastChar := expr[len(expr)-1]
	switch lastChar {
	case 'l':
		if expr == "nil" {
			return astId("nil")
		}
	case 'i':
		num := astExpr(expr[:len(expr)-1]).(*ast.BasicLit)
		if num.Kind != token.INT && num.Kind != token.FLOAT {
			panic("expected int or float before 'i'")
		}
		num.Kind = token.IMAG
		return num
	case '\'':
		firstChar := expr[0]
		if firstChar != lastChar {
			panic("unmatched quote")
		}
		return &ast.BasicLit{
			Kind:  token.CHAR,
			Value: string(expr[1 : len(expr)-1]),
		}
	case '"', '`':
		firstChar := expr[0]
		if firstChar != lastChar {
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
			return astExpr(right)
		} else if left[len(left)-1] == '.' {
			// Special case, a type assert.
			var x, t ast.Expr = astExpr(left[:len(left)-1]), nil
			if right == "type" {
				t = nil
			} else {
				t = astExpr(right)
			}
			return &ast.TypeAssertExpr{
				X:    x,
				Type: t,
			}
		}

		var fn = astExpr(left)
		var args = []ast.Expr{}
		parts := strings.Split(right, ",")
		for _, part := range parts {
			// NOTE: repeated commas have no effect,
			// nor do trailing commas.
			if len(part) > 0 {
				args = append(args, astExpr(part))
			}
		}
		return &ast.CallExpr{
			Fun:  fn,
			Args: args,
		}
	case '}':
		left, _, right := chopRight(expr)
		parts := strings.Split(right, ",")
		var ty = astExpr(left)
		var elts = []ast.Expr{}
		for _, part := range parts {
			elts = append(elts, astKVExpr(part))
		}
		return &ast.CompositeLit{
			Type:       ty,
			Elts:       elts,
			Incomplete: false,
		}
	case ']':
		left, _, right := chopRight(expr)
		return &ast.IndexExpr{
			X:     astExpr(left),
			Index: astExpr(right),
		}
	}
	// Numeric int?
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
	// Numeric float?
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
	// Doesn't end with a special character.
	if idx := strings.LastIndex(expr, "."); idx != -1 {
		return &ast.SelectorExpr{
			X:   astExpr(expr[:idx]),
			Sel: astId(expr[idx+1:]),
		}
	}
	return astId(expr)
}

func astKVExpr(kv string) *ast.KeyValueExpr {
	parts := strings.Split(kv, ":")
	if len(parts) != 2 {
		panic("astKVExpr requires 1 colon")
	}
	return &ast.KeyValueExpr{
		Key:   astExpr(parts[0]),
		Value: astExpr(parts[1]),
	}
}

func astBlock(body ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: body,
	}
}

// the last argument is destructured.
func astAssign1(exprs ...ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: exprs[:len(exprs)-1],
		Tok: token.ASSIGN,
		Rhs: exprs[len(exprs)-1:],
	}
}

// even and odd arguments are paired.
func astAssign2(exprs ...ast.Expr) *ast.AssignStmt {
	even := []ast.Expr{}
	odd := []ast.Expr{}
	for i := 0; i < len(exprs); i += 2 {
		even = append(even, exprs[i])
		odd = append(odd, exprs[i+1])
	}
	return &ast.AssignStmt{
		Lhs: even,
		Tok: token.ASSIGN,
		Rhs: odd,
	}
}

// the last argument is destructured.
func astDefine1(exprs ...ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: exprs[:len(exprs)-1],
		Tok: token.DEFINE,
		Rhs: exprs[len(exprs)-1:],
	}
}

// even and odd arguments are paired.
func astDefine2(exprs ...ast.Expr) *ast.AssignStmt {
	even := []ast.Expr{}
	odd := []ast.Expr{}
	for i := 0; i < len(exprs); i += 2 {
		even = append(even, exprs[i])
		odd = append(odd, exprs[i+1])
	}
	return &ast.AssignStmt{
		Lhs: even,
		Tok: token.DEFINE,
		Rhs: odd,
	}
}

func astIf(cond ast.Expr, body ...ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: cond,
		Body: astBlock(body...),
	}
}

func astReturn(results ...ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: results,
	}
}

// Given that 'in' ends with ')', '}', or ']',
// scan the input string until the matching opener is found.
// Tok is the corresponding opening token.
func chopRight(in string) (a string, tok byte, b string) {
	var (
		curly int = 0
		round int = 0
		sqare int = 0
	)
	done := func() bool {
		return curly == 0 && round == 0 && sqare == 0
	}
	switch in[len(in)-1] {
	case '}', ')', ']':
		// good
	default:
		panic("input doesn't start with brace: " + in)
	}
	for i := len(in) - 1; i >= 0; i-- {
		results := func() (string, byte, string) {
			return in[:i], in[i], in[i+1 : len(in)-1]
		}
		var chr = in[i]
		switch chr {
		case '}':
			curly++
		case ')':
			round++
		case ']':
			sqare++
		case '{':
			curly--
			if curly < 0 {
				panic("mismatched curly: " + in)
			}
			if done() {
				return results()
			}
		case '(':
			round--
			if round < 0 {
				panic("mismatched round: " + in)
			}
			if done() {
				return results()
			}
		case '[':
			sqare--
			if sqare < 0 {
				panic("mismatched square: " + in)
			}
			if done() {
				return results()
			}
		}
	}
	panic("mismatched braces: " + in + fmt.Sprintf("<<%v,%v,%v %v>>", curly, round, sqare, in))
}

func astString(s string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(s),
	}
}
