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

var pkgs = map[string]*types.Package{"unsafe":types.Unsafe}

var fluxObjs = map[types.Object]*ast.File{}

func getPackage(path string) (*types.Package, error) {
	if pkg, ok := pkgs[path]; ok {
		return pkg, nil
	}
	
	buildPkg, err := build.Import(path, "", 0)
	if err != nil {
		return nil, err
	}
	
	var fluxFiles []*ast.File
	
	files := []*ast.File{}
	fset := token.NewFileSet()
	for _, fileName := range append(buildPkg.GoFiles, buildPkg.CgoFiles...) {
		file, err := parser.ParseFile(fset, filepath.Join(buildPkg.Dir, fileName), nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		if strings.HasSuffix(fileName, ".flux.go") {
			fluxFiles = append(fluxFiles, file)
		}
	}
	ctx := types.Context{Import:srcImport}
	pkg, err := ctx.Check(fset, files)
	if err != nil {
		pkg, err = types.GcImport(pkgs, path)
		if err != nil {
			return nil, err
		}
	}
	pkg.Path = buildPkg.ImportPath
	pkgs[path] = pkg
	
	for _, file := range fluxFiles {
		for _, d := range file.Decls {
			// go/types doesn't set the Objects yet, so we have to look them up
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil {
					t := d.Recv.List[0].Type
					if p, ok := t.(*ast.StarExpr); ok {
						t = p.X
					}
					for _, m := range pkg.Scope.Lookup(t.(*ast.Ident).Name).GetType().(*types.NamedType).Methods {
						if m.Name == d.Name.Name {
							fluxObjs[method{nil, m}] = file
						}
					}
				} else {
					fluxObjs[pkg.Scope.Lookup(d.Name.Name)] = file
				}
			case *ast.GenDecl:
				if s, ok := d.Specs[0].(*ast.TypeSpec); ok {
					fluxObjs[pkg.Scope.Lookup(s.Name.Name)] = file
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
