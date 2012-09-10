package main

import "go/ast"

var (
	Universe *ast.Scope
	Unsafe *ast.Object
)

func init() {
	Universe = ast.NewScope(nil)
	b := scopeBuilder{Universe}
	b.defTypes("bool", "int8", "int16", "int32", "int64", "int", "uint8", "uint16", "uint32", "uint64", "uint", "uintptr", "byte", "rune", "float32", "float64", "complex64", "complex128", "string", "error")
	Universe.Lookup("error").Type.(*NamedType).underlying = InterfaceType{methods:[]FuncInfo{{InfoBase:InfoBase{name:"Error"}, typ:FuncType{results:[]ValueInfo{{typ:Universe.Lookup("string").Type.(Type)}}}}}}
	b.defConsts("false", "true", "nil", "iota")
	b.defFuncs("append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover")
	
	Unsafe = ast.NewObj(ast.Pkg, "unsafe")
	unsafeScope := ast.NewScope(nil)
	Unsafe.Data = unsafeScope
	b = scopeBuilder{unsafeScope}
	b.defTypes("Pointer")
	b.defFuncs("Alignof", "Offsetof", "Sizeof")
}

type scopeBuilder struct {
	scope *ast.Scope
}

func (b scopeBuilder) defTypes(names ...string) {
	for _, name := range names {
		obj := b.define(ast.Typ, name)
		typ := newNamedType(name, nil)
		typ.underlying = &BasicType{}
		obj.Type = typ
	}
}

func (b scopeBuilder) defConsts(names ...string) {
	for _, name := range names {
		b.define(ast.Con, name)
	}
}

func (b scopeBuilder) defFuncs(names ...string) {
	for _, name := range names {
		b.define(ast.Fun, name)
	}
}

func (b scopeBuilder) define(kind ast.ObjKind, name string) *ast.Object {
	obj := ast.NewObj(kind, name)
	b.scope.Insert(obj)
	return obj
}
