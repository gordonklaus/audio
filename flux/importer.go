package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"unsafe"
)

func srcImport(imports map[string]*ast.Object, path string) (*ast.Object, error) {
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

var pkgs = map[string]*ast.Package{"":&builtinAstPkg, "unsafe":&unsafePkg}
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
	pkg, err := ast.NewPackage(fset, files, srcImport, builtinAstPkg.Scope)
	if err != nil { return nil, err }
	pkgs[path] = pkg
	return pkg, nil
}

var (
	builtinAstPkg = ast.Package{Scope:ast.NewScope(nil)}
	unsafePkg = ast.Package{"unsafe", ast.NewScope(nil), nil, nil}
)

func init() {
	b := scopeBuilder{builtinAstPkg.Scope}
	b.defTypes(new(bool), new(int8), new(int16), new(int32), new(int64), new(int), new(uint8), new(uint16), new(uint32), new(uint64), new(uint), new(uintptr), new(float32), new(float64), new(complex64), new(complex128), new(string))
	b.defType("byte", new(byte))
	b.defType("rune", new(rune))
	b.define(ast.Typ, "error").Type = &NamedType{InfoBase:InfoBase{"error", nil}, underlying:&InterfaceType{methods:[]*ValueInfo{{InfoBase:InfoBase{name:"Error"}, typ:&FuncType{results:[]*ValueInfo{{typ:builtinAstPkg.Scope.Lookup("string").Type.(Type)}}}}}}}
	b.defFuncs("append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover")
	b.defConsts("false", "true", "nil", "iota")
	
	builtinPkg = &PackageInfo{}
	builtinPkg.load()
	
	unsafePkg = ast.Package{"unsafe", ast.NewScope(nil), nil, nil}
	b = scopeBuilder{unsafePkg.Scope}
	b.defTypes(new(unsafe.Pointer))
	b.defFuncs("Alignof", "Offsetof", "Sizeof")
}

type scopeBuilder struct {
	scope *ast.Scope
}

func (b scopeBuilder) defTypes(values ...interface{}) {
	for _, value := range values {
		b.defType(reflect.TypeOf(value).Elem().Name(), value)
	}
}

func (b scopeBuilder) defType(name string, value interface{}) {
	b.define(ast.Typ, name).Type = &NamedType{InfoBase:InfoBase{name, nil}, underlying:&BasicType{reflectType:reflect.TypeOf(value).Elem()}}
}

func (b scopeBuilder) defConsts(names ...string) {
	for _, name := range names {
		b.define(ast.Con, name).Data = &ValueInfo{InfoBase{name, nil}, nil, true}
	}
}

func (b scopeBuilder) defFuncs(names ...string) {
	for _, name := range names {
		b.define(ast.Fun, name).Data = &FuncInfo{InfoBase{name, nil}, nil, &FuncType{}}
	}
}

func (b scopeBuilder) define(kind ast.ObjKind, name string) *ast.Object {
	obj := ast.NewObj(kind, name)
	b.scope.Insert(obj)
	return obj
}
