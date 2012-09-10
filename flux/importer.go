package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
)

func srcImport(imports map[string]*ast.Object, path string) (*ast.Object, error) {
	if path == "unsafe" { return Unsafe, nil }
	if obj, ok := imports[path]; ok { return obj, nil }
	pkg, err := getPackage(path)
	if err != nil { return nil, err }
	scope := ast.NewScope(pkg.Scope.Outer)
	for name, obj := range pkg.Scope.Objects {
		if ast.IsExported(name) {
			scope.Insert(obj)
		}
	}
	obj := &ast.Object{ast.Pkg, pkg.Name, nil, scope, nil}
	imports[path] = obj
	return obj, nil
}

var pkgs = map[string]*ast.Package{}
func getPackage(path string) (*ast.Package, error) {
	if pkg, ok := pkgs[path]; ok { return pkg, nil }
	buildPkg, err := build.Import(path, "", 0)
	if err != nil { return nil, err }
	files := map[string]*ast.File{}
	fset := token.NewFileSet()
	for _, fileName := range append(buildPkg.GoFiles, buildPkg.CgoFiles...) {
		file, err := parser.ParseFile(fset, filepath.Join(buildPkg.Dir, fileName), nil, 0)
		if err != nil { return nil, err }
		files[fileName] = file
	}
	// TODO:  incorporate flux source into files (if missing)
	pkg, err := ast.NewPackage(fset, files, srcImport, Universe)
	if err != nil { return nil, err }
	pkgs[path] = pkg
	return pkg, nil
}
