package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
)

type Context struct {
	srcDir string
	bctx   build.Context
	fset   *token.FileSet
	pkgs   map[string]*ast.Package
}

func NewContext(srcDir string) *Context {
	return &Context{
		srcDir: srcDir,
		bctx:   build.Default,
		fset:   token.NewFileSet(),
		pkgs:   make(map[string]*ast.Package),
	}
}

func (ctx *Context) ResolveModule(path string) (dir string, err error) {
	pkg, err := ctx.bctx.Import(path, ctx.srcDir, build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.Dir, nil
}

func (ctx *Context) ImportModule(path string) (pkg *ast.Package, err error) {
	if pkg, ok := ctx.pkgs[path]; ok {
		return pkg, nil
	}
	dir, err := ctx.ResolveModule(path)
	if err != nil {
		return
	}
	pkgs, err := parser.ParseDir(ctx.fset, dir, nil, parser.ParseComments)
	if err != nil {
		return
	}
	names := []string{}
	for name, _ := range pkgs {
		names = append(names, name)
	}
	if len(names) > 1 {
		err = fmt.Errorf("Conflicting module names in dir %v: %v", dir, names)
		return
	}
	pkg = pkgs[names[0]]
	ctx.pkgs[path] = pkg
	return pkg, nil
}
