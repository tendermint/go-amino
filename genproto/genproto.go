package genproto

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Represents a single Amino type to be transcoded to proto3 schema.
type transcodeTask struct {
	// Fields general to package and file.
	pkg      *packages.Package // package in which this is defined.
	imports  map[string]string // short name to full package name translation.
	filenode *ast.File         // file in which this is defined.

	// Fields specific to the amino structure.
	directive *struct{} // amino directive is initially binary, no options.
	name      string    // name of the amino type.
	typeExpr  ast.Expr  // type expression of the amino type.
}

// All dependent modules not listed are assumed to already have corresponding proto schemas generated.
// Package must match pattern spec as defined here: https://godoc.org/golang.org/x/tools/go/packages#Package
// NOTE: the `packages` module handles typechecking and a lot more, so it's the easiest way to make this work.
// Ergo, unless decided, depend on the packages module as the main entrypoint for this tool.
// This also means that we assume that projects must compile correctly before genproto can be used.
func GenerateProtoForPatterns(patterns ...string) {
	cfg := &packages.Config{Mode: packages.NeedName |
		packages.NeedFiles |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedImports |
		packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		GenerateProtoForPackage(pkg)
	}
}

func GenerateProtoForPackage(pkg *packages.Package) {
	for _, filenode := range pkg.Syntax {
		GenerateProtoForFilenode(pkg, filenode)
	}
}

func GenerateProtoForFilenode(pkg *packages.Package, filenode *ast.File) {

	// First, create a map of imports.
	// We will need this map later.
	var imports = map[string]string{} // short name -> full package path
	fmt.Printf("Getting imports from %v\n", filenode.Name)
	ast.Inspect(filenode, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ImportSpec:
			pkgpath, err := strconv.Unquote(node.Path.Value)
			if node.Name == nil {
				// implicit name, figure out the name from the module.
				if err != nil {
					panic(err)
				}
				pkgname := pkg.Imports[pkgpath].Name
				if pkgname == "" {
					panic("pkg.Name is empty, config should include NeedName")
				}
				imports[pkgname] = pkgpath
			} else {
				// explicit name, figure out the name from the import spec.
				imports[node.Name.Name] = pkgpath
			}
		}
		return true
	})

	// Then, discover the amino structs to transcode.
	var transcodeTasks []transcodeTask
	fmt.Printf("Getting Amino structs from %v\n", filenode.Name)
	ast.Inspect(filenode, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.GenDecl:
			decl := n.(*ast.GenDecl)
			comments := decl.Doc

			// currenty just a bool, but in the future may be extended with options.
			var declDirective *struct{} // at the group level?
			var typeDirective *struct{} // at the line level?

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
							transcodeTasks = append(transcodeTasks, transcodeTask{
								pkg:       pkg,
								filenode:  filenode,
								imports:   imports,
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

		default:
			// nothing
		}

		return true
	})

	// P3Doc is what we will print out.
	var p3doc = P3Doc{}

	// For each amino type, run directives.
	for _, task := range transcodeTasks {
		p3msg := TranscodeToP3Message(task)
		p3doc.Messages = append(p3doc.Messages, p3msg)
	}

	// Print the final proto3 schema file.
	fmt.Println(p3doc.Print())
}

func TranscodeToP3Message(task transcodeTask) P3Message {
	var pkg = task.pkg
	var imports = task.imports
	var p3fields []P3Field

	fmt.Println("Generating proto schema for:", task.name)
	switch atype := task.typeExpr.(type) {
	case *ast.StructType:
		// Transcode a struct type expr into the appropriate proto shema.
		// For each struct field, discover the value type.
		for _, field := range atype.Fields.List {
			switch se := field.Type.(type) {
			case *ast.SelectorExpr:
				// se.X should be an *ast.Ident for the package.
				// se.Sel.Name should be exported in above package.
				importPkgName := se.X.(*ast.Ident).Name
				importPkgPath, ok := imports[importPkgName]
				if !ok {
					panic(fmt.Sprintf("unrecognized package name %v (was it imported?)", importPkgName))
				}
				importObjName := se.Sel.Name
				importPkg, ok := pkg.Imports[importPkgPath]
				if !ok {
					panic(fmt.Sprintf("unrecognized package identifier %v", importPkgPath))
				}
				if importPkg.Types == nil {
					panic("pkg.Types nil, config should include NeedTypes")
				}
				importObj := importPkg.Types.Scope().Lookup(importObjName)
				if importObj.Type() == nil {
					panic(fmt.Sprintf("unexpected nil Type for %v.%v", importPkgName, importObjName))
				}
				importNamed, ok := importObj.Type().(*types.Named)
				if !ok {
					panic(fmt.Sprintf("imported object isn't named?!"))
				}
				fieldType := importNamed.Underlying() // This is what ultimately gets encoded
				fmt.Printf("Found field %v of underlying type %v\n", field.Names[0].Name, fieldType)
				// XXX This is not complete, complete it.
				p3fields = append(p3fields, P3Field{
					Type:     "fixme",
					Name:     field.Names[0].Name,
					Number:   0,
					Repeated: false,
				})

			case *ast.Ident:
				// A basic identifier for the field name.
				p3fields = append(p3fields, P3Field{
					Type:     P3FieldType(fmt.Sprintf("fixme(%v)", se.Name)),
					Name:     field.Names[0].Name,
					Number:   0,
					Repeated: false,
				})

			default:
				panic(fmt.Sprintf("unexpected field AST type %v", reflect.TypeOf(field.Type)))
			}
		}
	default:
		// TODO what more types do we need to support?
		panic(fmt.Sprintf("amino directive unsupported for %v", atype))
	}

	return P3Message{
		Name:   task.name,
		Fields: p3fields,
	}
}

//----------------------------------------

// TODO actually define directive options and merge them.
func mergeDirectives(a *struct{}, b *struct{}) *struct{} {
	if a != nil {
		return a
	}
	return b
}
