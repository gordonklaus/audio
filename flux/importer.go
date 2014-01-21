// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

var pkgs = map[string]*types.Package{"unsafe": types.Unsafe}

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
	cfg := types.Config{IgnoreFuncBodies: true, FakeImportC: true, Import: srcImport}
	pkg, err := cfg.Check(path, fset, files, nil)
	if err != nil {
		return nil, err
	}
	pkg.Path = buildPkg.ImportPath
	pkgs[path] = pkg
	
	for _, file := range fluxFiles {
		n := strings.Split(file, ".")
		for i := range n {
			n[i] = strings.TrimRight(n[i], "-")
		}
		if len(n) == 1 {
			fluxObjs[pkg.Scope().Lookup(n[0])] = true
		} else {
			for _, m := range pkg.Scope().Lookup(n[0]).GetType().(*types.Named).Methods {
				if m.Name == n[1] {
					fluxObjs[m] = true
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
