package refactor

import (
	_ "code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/loader"
	"code.google.com/p/go.tools/go/types"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Rename(importPath, recv, name, newName string) error {
	paths, err := importers(importPath)
	if err != nil {
		return err
	}
	config := loader.Config{
		ParserMode:          parser.ParseComments,
		TypeCheckFuncBodies: func(path string) bool { return paths[path] },
		AllowErrors:         true,
	}
	for p := range paths {
		config.Import(p)
	}
	prog, err := config.Load()
	if err != nil {
		return err
	}
	pkg := prog.Imported[importPath]
	for _, err := range pkg.Errors {
		fmt.Println(err)
	}
	obj := pkg.Pkg.Scope().Lookup(name)
	if recv != "" {
		tn, ok := pkg.Pkg.Scope().Lookup(recv).(*types.TypeName)
		if !ok {
			return fmt.Errorf("type %s.%s not found", importPath, recv)
		}
		t := tn.Type().(*types.Named)
		for i := 0; i < t.NumMethods(); i++ {
			m := t.Method(i)
			if m.Name() == name {
				obj = m
				break
			}
		}
		if t, ok := t.Underlying().(*types.Struct); ok && obj == nil {
			for i := 0; i < t.NumFields(); i++ {
				f := t.Field(i)
				if f.Name() == name {
					obj = f
					break
				}
			}
		}
		if obj == nil {
			return fmt.Errorf("field or method %s.%s.%s not found", importPath, recv, name)
		}
	}
	if obj == nil {
		return fmt.Errorf("object %s.%s not found", importPath, name)
	}
	for path := range paths {
		pkg := prog.Imported[path]
		for _, f := range pkg.Files {
			modified := false
			ast.Inspect(f, func(node ast.Node) bool {
				if id, ok := node.(*ast.Ident); ok && pkg.ObjectOf(id) == obj {
					id.Name = newName
					modified = true
				}
				return true
			})
			if !modified {
				continue
			}
			file, err := os.Create(prog.FilePath(f))
			if err != nil {
				return err
			}
			printer.Fprint(file, prog.Fset, f)
			file.Close()
		}
	}
	return nil
}

func importers(importPath string) (map[string]bool, error) {
	paths := map[string]bool{importPath: true}
	for _, srcDir := range build.Default.SrcDirs() {
		if err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
			p, err := build.ImportDir(path, 0)
			if err != nil {
				return nil
			}
			i := sort.SearchStrings(p.Imports, importPath)
			if i < len(p.Imports) && p.Imports[i] == importPath {
				paths[p.ImportPath] = true
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

// MovePackage moves the package from the old import path to the new one.  Any
// packages in subdirectories are also moved.  All uses of the old import path
// in the Go environment are updated to the new import path.
//
func MovePackage(old, new string) error {
	p, err := build.Import(old, "", build.FindOnly)
	if err != nil {
		return err
	}
	oldFullPath := p.Dir
	if p, err := build.Import(new, "", build.FindOnly); err == nil {
		return fmt.Errorf("package %s already exists at %s", new, p.Dir)
	}
	newFullPath := filepath.Join(p.SrcRoot, new)

	for _, srcDir := range build.Default.SrcDirs() {
		if err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
			if c := filepath.Base(path)[0]; c == '_' || c == '.' {
				if f.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if f.Mode().IsRegular() && filepath.Ext(path) == ".go" {
				astFile, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
				if err != nil {
					return nil
				}
				for _, imp := range astFile.Imports {
					importPath := imp.Path.Value[1 : len(imp.Path.Value)-1]
					if importPath == old || strings.HasPrefix(importPath, old+"/") {
						b, err := ioutil.ReadFile(path)
						if err != nil {
							return nil
						}
						i := int(imp.Path.Pos())
						s := string(b[:i]) + new + string(b[i+len(old):])
						ioutil.WriteFile(path, []byte(s), f.Mode())
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return os.Rename(oldFullPath, newFullPath)
}

func ReportShadowedPackages() {
	for _, srcDir := range build.Default.SrcDirs() {
		filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
			if p, err := build.ImportDir(path, build.AllowBinary); err == nil && p.ConflictDir != "" {
				fmt.Println("WARNING: package at " + p.Dir)
				fmt.Println("     is shadowed by " + p.ConflictDir)
			}
			return nil
		})
	}
}
