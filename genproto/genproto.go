package genproto

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type aminoType struct {
	directive *struct{} // amino directive is initially binary, no options
	name      string
	typeExpr  ast.Expr
}

func RunExample() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "example/main.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {

		switch n.(type) {
		case *ast.GenDecl:
			decl := n.(*ast.GenDecl)
			comments := decl.Doc

			// currenty just a bool, but in the future may be extended with options.
			var declDirective *struct{} // at the group level?
			var typeDirective *struct{} // at the line level?
			var aminoTypes []aminoType

			// look for amino directives in comment for GenDecl.
			if comments != nil {
				for _, comment := range comments.List {
					if strings.HasPrefix(comment.Text, "// amino") {
						declDirective = &struct{}{}
						break // TODO: instead of breaking, merge decl directives
					}
				}
			}

			// look for amino directives in comment for TypeSpec.
			for _, spec := range decl.Specs {
				typeDirective = nil
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if comments := spec.Doc; comments != nil {
						for _, comment := range comments.List {
							if strings.HasPrefix(comment.Text, "// amino") {
								typeDirective = &struct{}{}
								break // TODO: instead of breaking, merge type directives
							}
						}
					}
					if comments := spec.Comment; comments != nil {
						for _, comment := range comments.List {
							if strings.HasPrefix(comment.Text, "// amino") {
								typeDirective = &struct{}{}
								break // TODO: instead of breaking, merge type directives
							}
						}
					}

					// We found a type spec to apply amino directives to.
					directive := mergeDirectives(declDirective, typeDirective)
					if directive != nil {
						switch stype := spec.Type.(type) {
						case *ast.StructType:
							aminoTypes = append(aminoTypes, aminoType{
								directive: directive,
								name:      spec.Name.Name,
								typeExpr:  spec.Type,
							})
						default:
							panic(fmt.Sprintf("amino directive unsupported for %v", stype))
						}
					}
				}
			}

			// for each amino type, run directives.
			for _, aminoType := range aminoTypes {
				fmt.Println("Generating proto schema for:", aminoType.name)
				switch atype := aminoType.typeExpr.(type) {
				case *ast.StructType:
					// Transcode a struct type expr into the appropriate proto shema.
				default:
					panic(fmt.Sprintf("amino directive unsupported for %v", atype))
				}
			}

		default:

			// nothing
		}

		return true
	})
}

// TODO actually define directive options and merge them.
func mergeDirectives(a *struct{}, b *struct{}) *struct{} {
	if a != nil {
		return a
	}
	return b
}
