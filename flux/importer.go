package main

import (
	"code.google.com/p/go.exp/go/types"
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
		origErr := err
		pkg, err = types.GcImport(pkgs, path)
		if err != nil {
			return nil, origErr
		}
	}
	pkg.Path = buildPkg.ImportPath
	pkgs[path] = pkg
	
	for _, file := range fluxFiles {
		n := strings.Split(file, ".")
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
