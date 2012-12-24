package main

import "C"

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
		if ast.IsExported(name) || path == "C" {
			scope.Insert(obj)
		}
	}
	obj := &ast.Object{ast.Pkg, pkg.Name, nil, scope, nil}
	imports[path] = obj
	return obj, nil
}

var pkgs = map[string]*ast.Package{"":&builtinAstPkg, "unsafe":&unsafePkg, "C":&cAstPkg}
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
	cAstPkg = ast.Package{"C", ast.NewScope(nil), nil, nil}
)

func init() {
	b := scopeBuilder{builtinAstPkg.Scope}
	b.defTypes(new(bool), new(int8), new(int16), new(int32), new(int64), new(int), new(uint8), new(uint16), new(uint32), new(uint64), new(uint), new(uintptr), new(float32), new(float64), new(complex64), new(complex128), new(string))
	b.defType("byte", new(byte))
	b.defType("rune", new(rune))
	b.define(ast.Typ, "error").Type = &NamedType{InfoBase:InfoBase{"error", nil}, underlying:&InterfaceType{methods:[]*Value{{InfoBase:InfoBase{name:"Error"}, typ:&FuncType{results:[]*Value{{typ:builtinAstPkg.Scope.Lookup("string").Type.(Type)}}}}}}}
	b.defFuncs("append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover")
	b.defConsts("false", "true", "nil", "iota")
	builtinPkg.load()
	
	b = scopeBuilder{unsafePkg.Scope}
	b.defTypes(new(unsafe.Pointer))
	b.defFuncs("Alignof", "Offsetof", "Sizeof")
	
	b = scopeBuilder{cAstPkg.Scope}
	b.defType("char", new(C.char))
	b.defType("schar", new(C.schar))
	b.defType("uchar", new(C.uchar))
	b.defType("short", new(C.short))
	b.defType("ushort", new(C.ushort))
	b.defType("int", new(C.int))
	b.defType("uint", new(C.uint))
	b.defType("long", new(C.long))
	b.defType("ulong", new(C.ulong))
	b.defType("longlong", new(C.longlong))
	b.defType("ulonglong", new(C.ulonglong))
	b.defType("float", new(C.float))
	b.defType("double", new(C.double))
}

type scopeBuilder struct {
	scope *ast.Scope
}

func (b scopeBuilder) defTypes(values ...interface{}) {
	for _, v := range values {
		b.defType(reflect.TypeOf(v).Elem().Name(), v)
	}
}

func (b scopeBuilder) defType(name string, value interface{}) {
	b.define(ast.Typ, name).Type = &NamedType{InfoBase:InfoBase{name, nil}, underlying:&BasicType{reflectType:reflect.TypeOf(value).Elem()}}
}

func (b scopeBuilder) defConsts(names ...string) {
	for _, n := range names {
		b.define(ast.Con, n).Data = &Value{InfoBase{n, nil}, nil, true}
	}
}

func (b scopeBuilder) defFuncs(names ...string) {
	for _, n := range names {
		b.define(ast.Fun, n).Data = &Func{InfoBase{n, nil}, nil, &FuncType{}}
	}
}

func (b scopeBuilder) define(kind ast.ObjKind, name string) *ast.Object {
	obj := ast.NewObj(kind, name)
	b.scope.Insert(obj)
	return obj
}
