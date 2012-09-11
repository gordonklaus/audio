package main

import (
	"go/ast"
	"reflect"
	"unsafe"
)

var (
	Universe *ast.Scope
	Unsafe *ast.Object
)

func init() {
	Universe = ast.NewScope(nil)
	b := scopeBuilder{Universe}
	b.defTypes(new(bool), new(int8), new(int16), new(int32), new(int64), new(int), new(uint16), new(uint32), new(uint64), new(uint), new(uintptr), new(byte), new(float32), new(float64), new(complex64), new(complex128), new(string), new(error))
	b.defType("uint8", new(uint8))
	b.defType("rune", new(rune))
	Universe.Lookup("error").Type.(*NamedType).underlying = InterfaceType{methods:[]FuncInfo{{InfoBase:InfoBase{name:"Error"}, typ:FuncType{results:[]ValueInfo{{typ:Universe.Lookup("string").Type.(Type)}}}}}}
	b.defConsts("false", "true", "nil", "iota")
	b.defFuncs("append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover")
	
	Unsafe = ast.NewObj(ast.Pkg, "unsafe")
	unsafeScope := ast.NewScope(nil)
	Unsafe.Data = unsafeScope
	b = scopeBuilder{unsafeScope}
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
	obj := b.define(ast.Typ, name)
	typ := newNamedType(name, nil)
	typ.underlying = &BasicType{reflectType:reflect.TypeOf(value).Elem()}
	obj.Type = typ
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
