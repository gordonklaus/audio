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

var pkgs = map[string]*types.Package{"unsafe":types.Unsafe}

var fluxObjs = map[types.Object]bool{}

func getPackage(path string) (*types.Package, error) {
	if pkg, ok := pkgs[path]; ok {
		return pkg, nil
	}
	
	buildPkg, err := build.Import(path, "", 0)
	if err != nil {
		return nil, err
	}
	
	var fluxNames []string
	
	files := []*ast.File{}
	fset := token.NewFileSet()
	for _, fileName := range append(buildPkg.GoFiles, buildPkg.CgoFiles...) {
		if strings.HasSuffix(fileName, ".flux.go") {
			fluxNames = append(fluxNames, fileName[:len(fileName) - 8])
		}
		file, err := parser.ParseFile(fset, filepath.Join(buildPkg.Dir, fileName), nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
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
	
fluxNames:
	for _, n := range fluxNames {
		i := strings.Index(n, ".")
		if i < 0 {
			if obj := pkg.Scope.Lookup(n); obj != nil {
				fluxObjs[obj] = true
				continue
			}
		} else {
			t, f := n[:i], n[i+1:]
			if t, ok := pkg.Scope.Lookup(t).(*types.TypeName); ok {
				if t, ok := t.Type.(*types.NamedType); ok {
					for _, m := range t.Methods {
						if m.Name == f {
							fluxObjs[method{nil, m}] = true
							continue fluxNames
						}
					}
				}
			}
		}
		fmt.Printf("flux object %s not found in package %s\n", n, path)
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
