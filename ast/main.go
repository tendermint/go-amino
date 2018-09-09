package main

import (
	"fmt"
	"os"
	"strconv"
)

func unquote(val string) string {
	s, err := strconv.Unquote(val)
	if err != nil {
		panic(err)
	}
	return s
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	ctx := NewContext(pwd)
	pkg, err := ctx.ImportModule("./example/foo")
	if err != nil {
		fmt.Println("failed to import:", err)
		return
	}
	fmt.Println("found package", pkg, err)

	for name, file := range pkg.Files {
		fmt.Printf(">> File %v\n      %#v\n", name, file)
		for _, imp := range file.Imports {
			fmt.Printf(">>>> Import %#v\n", imp)
			fmt.Printf(">>>>        %#v\n", unquote(imp.Path.Value))
		}
		for _, decl := range file.Decls {
			fmt.Printf(">>>> Declaration %#v\n", decl)
		}
	}
	return

	fmt.Println("########### Manual Iteration ###########")
	/*

		fmt.Println("Imports:")
		for _, i := range node.Imports {
			path, err := strconv.Unquote(i.Path.Value)
			if err != nil {
				panic(err)
			}
			pkg, err := build.Default.Import(path, "", build.FindOnly)
			fmt.Println(path, pkg, err)
		}

		fmt.Println("Comments:")
		for _, c := range node.Comments {
			fmt.Print(c.Text())
		}

		fmt.Println("Functions:")
		for _, f := range node.Decls {
			fn, ok := f.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fmt.Println(fn.Name.Name)
		}

		fmt.Println("########### Inspect ###########")
		ast.Inspect(node, func(n ast.Node) bool {
			// Find Return Statements
			ret, ok := n.(*ast.ReturnStmt)
			if ok {
				fmt.Printf("return statement found on line %d:\n\t", fset.Position(ret.Pos()).Line)
				printer.Fprint(os.Stdout, fset, ret)
				return true
			}
			// Find Functions
			fn, ok := n.(*ast.FuncDecl)
			if ok {
				var exported string
				if fn.Name.IsExported() {
					exported = "exported "
				}
				fmt.Printf("%sfunction declaration found on line %d: \n\t%s\n", exported, fset.Position(fn.Pos()).Line, fn.Name.Name)
				return true
			}
			return true
		})
		fmt.Println()
	*/
}
