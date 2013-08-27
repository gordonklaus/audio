package main

import (
	"code.google.com/p/go.exp/go/types"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

var pkgs = map[string]*types.Package{"unsafe": types.Unsafe, "C": &types.Package{Name: "C", Path: "C", Scope: &types.Scope{}}}

var fluxObjs = map[types.Object]bool{}

func getPackage(path string) (*types.Package, error) {
	if pkg, ok := pkgs[path]; ok {
		return pkg, nil
	}
	
	buildPkg, err := build.Import(path, "", 0)
	if err != nil {
		return nil, err
	}
	
	var fluxFiles []string
	
	files := []*ast.File{}
	fset := token.NewFileSet()
	for _, fileName := range append(buildPkg.GoFiles, buildPkg.CgoFiles...) {
		file, err := parser.ParseFile(fset, filepath.Join(buildPkg.Dir, fileName), nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		if strings.HasSuffix(fileName, ".flux.go") {
			fluxFiles = append(fluxFiles, fileName[:len(fileName)-8])
		}
	}
	ctx := types.Context{Import:srcImport}
	pkg, err := ctx.Check(fset, files)
	if err != nil {
		fmt.Printf("Error typechecking package %s:\n%v\n", path, err)
		// types.Check doesn't handle cgo packages, so we fall back to GcImport here.
		// TODO:  Is there any other reason to do GcImport if the source doesn't check?
		// If not, we should only do GcImport specifically if there are only cgo related
		// errors, otherwise we risk importing a stale binary.
		origErr := err
		pkg, err = types.GcImport(pkgs, path)
		if err != nil {
			fmt.Printf("Error GcImporting package %s:\n%v.\n", path, err)
			return nil, origErr
		}
	}
	// go/types forgot to set the receiver on interface methods
	for _, obj := range pkg.Scope.Entries {
		if t, ok := obj.(*types.TypeName); ok {
			if it, ok := t.Type.(*types.NamedType).Underlying.(*types.Interface); ok {
				for _, m := range it.Methods {
					if m.Type.Recv == nil {
						m.Type.Recv = &types.Var{Type: t.Type}
					}
				}
			}
		}
	}
	pkg.Path = buildPkg.ImportPath
	pkgs[path] = pkg
	
	for _, file := range fluxFiles {
		n := strings.Split(file, ".")
		for i := range n {
			n[i] = strings.TrimRight(n[i], "-")
		}
		if len(n) == 1 {
			fluxObjs[pkg.Scope.Lookup(n[0])] = true
		} else {
			for _, m := range pkg.Scope.Lookup(n[0]).GetType().(*types.NamedType).Methods {
				if m.Name == n[1] {
					fluxObjs[method{nil, m}] = true
				}
			}
		}
	}
	
	return pkg, nil
}

func srcImport(imports map[string]*types.Package, path string) (*types.Package, error) {
	if pkg, ok := imports[path]; ok {
		return pkg, nil
	}
	
	pkg, err := getPackage(path)
	if err != nil {
		return nil, err
	}
	
	imports[path] = pkg
	return pkg, nil
}
