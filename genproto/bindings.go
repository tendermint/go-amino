package genproto

import (
	"go/ast"
	"reflect"
	"strings"

	"github.com/tendermint/go-amino"
)

// Given genproto generated schema files for Go objects, generate mappers to
// and from pb messages.  The purpose of this is to let Amino use
// already-optimized probuf logic for serialization.

func GenerateProtoBindings(pi *amino.Package, genpkg string) (file *ast.File, err error) {
	file = &ast.File{
		Name:  astId(pi.GoPkg),
		Decls: nil,
	}
	for _, type_ := range pi.Types {
		bindings, err := GenerateProtoBindingsForType(pi, type_, genpkg)
		if err != nil {
			return nil, err
		}
		file.Decls = append(file.Decls, bindings.toProto)
		file.Decls = append(file.Decls, bindings.fromProto)
	}
	return file, nil
}

type protoBindings struct {
	// func (Obj) ToPBMessage() (proto.Message, error)
	toProto *ast.FuncDecl
	// func (*Obj) FromPBMessage(proto.Message) (error)
	fromProto *ast.FuncDecl
}

func GenerateProtoBindingsForType(pi *amino.Package, type_ reflect.Type, genpkg string) (protoBindings, error) {
	toProto := astFunc("ToPBMessage",
		"o", type_.Name(),
		astFields(),
		astFields("msg", "proto.Message", "err", "error"),
		nil,
	)
	fromProto := astFunc("FromPBMessage",
		"o", type_.Name(),
		astFields("msg", "proto.Message"),
		astFields("err", "error"),
		nil,
	)
	return protoBindings{toProto, fromProto}, nil
}

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
// Useful for parsing strings to ast nodes, like
// foo.bar["qwe"] or *bytes.Buffer
func astExpr(expr string) ast.Expr {
	if expr == "" {
		panic("astExpr requires argument")
	}
	if expr[0] == '*' {
		return &ast.StarExpr{
			X: astExpr(expr[1:]),
		}
	}
	if expr[len(expr)-1] == ']' {
		idx := strings.LastIndex(expr, "[")
		return &ast.IndexExpr{
			X:     astExpr(expr[:idx]),
			Index: astExpr(expr[idx+1 : len(expr)-1]),
		}
	}
	if sel := strings.LastIndex(expr, "."); sel != -1 {
		return &ast.SelectorExpr{
			X:   astExpr(expr[:sel]),
			Sel: astId(expr[sel+1:]),
		}
	}
	return astId(expr)
}
