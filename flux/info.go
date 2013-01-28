package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"io/ioutil"
	."log"
	"os"
	"path/filepath"
	"reflect"
	."strconv"
	."strings"
	"unicode"
)

var (
	rootPkg = &Package{loaded:true}
	builtinPkg = &Package{}
	cPkg = &Package{InfoBase:InfoBase{"C", rootPkg}, pkgName:"C", importPath:"C"}
)

func init() {
	for _, srcDir := range build.Default.SrcDirs() {
		rootPkg.fullPath = srcDir
		srcPackage := NewPackage(rootPkg, "")
		for _, p := range srcPackage.subPkgs {
			p.parent = rootPkg
			AddInfo(rootPkg, p)
		}
	}
	AddInfo(rootPkg, cPkg)
}

func FindPackage(path string) *Package {
	return rootPkg.FindPackage(path)
}

type Info interface {
	Name() string
	SetName(name string)
	Parent() Info
	SetParent(info Info)
}

func Children(i Info) (c []Info) {
	switch i := i.(type) {
	case *Package:
		i.load()
		for _, x := range i.types { c = append(c, x) }
		for _, x := range i.funcs { c = append(c, x) }
		for _, x := range i.vars { c = append(c, x) }
		for _, x := range i.consts { c = append(c, x) }
		for _, x := range i.subPkgs { c = append(c, x) }
	case *NamedType:
		for _, x := range i.methods { c = append(c, x) }
	}
	return
}

func AddInfo(p, c Info) {
	c.SetParent(p)
	index := 0
	switch p := p.(type) {
	case *Package:
		switch c := c.(type) {
		case *Package:
			for i, x := range p.subPkgs { if x.name > c.name { index = i; break } }
			p.subPkgs = append(p.subPkgs[:index], append([]*Package{c}, p.subPkgs[index:]...)...)
		case *NamedType:
			for i, x := range p.types { if x.name > c.name { index = i; break } }
			p.types = append(p.types[:index], append([]*NamedType{c}, p.types[index:]...)...)
		case *Func:
			for i, x := range p.funcs { if x.name > c.name { index = i; break } }
			p.funcs = append(p.funcs[:index], append([]*Func{c}, p.funcs[index:]...)...)
		case *Value:
			if c.constant {
				for i, x := range p.consts { if x.name > c.name { index = i; break } }
				p.consts = append(p.consts[:index], append([]*Value{c}, p.consts[index:]...)...)
			} else {
				for i, x := range p.vars { if x.name > c.name { index = i; break } }
				p.vars = append(p.vars[:index], append([]*Value{c}, p.vars[index:]...)...)
			}
		}
	case *NamedType:
		if f, ok := c.(*Func); ok {
			for i, m := range p.methods { if m.name > f.name { index = i; break } }
			p.methods = append(p.methods[:index], append([]*Func{f}, p.methods[index:]...)...)
		} else {
			panic("types can only contain funcs")
		}
	default:
		Panicf("a %T cannot have children", p)
	}
}

func FluxSourcePath(i Info) string {
	switch i := i.(type) {
	case *Package:
		return i.fullPath
	case *NamedType:
		return fmt.Sprintf("%s/%s.flux/type.flux", FluxSourcePath(i.parent), i.name)
	case *Func:
		s := FluxSourcePath(i.parent)
		if _, ok := i.parent.(*NamedType); ok {
			s = filepath.Dir(s)
		}
		return fmt.Sprintf("%s/%s.flux", s, i.name)
	case *Value:
		return fmt.Sprintf("%s/%s.flux", FluxSourcePath(i.parent), i.name)
	default:
		Panicf("unknown Info %#v", i)
	}
	return ""
}

type InfoBase struct {
	name string
	parent Info
}
func (i InfoBase) Name() string { return i.name }
func (i *InfoBase) SetName(name string) { i.name = name }
func (i InfoBase) Parent() Info { return i.parent }
func (i *InfoBase) SetParent(info Info) { i.parent = info }

type Package struct {
	InfoBase
	pkgName string
	importPath string
	fullPath string
	types []*NamedType
	funcs []*Func
	vars []*Value
	consts []*Value
	subPkgs []*Package
	loaded bool
}

func NewPackage(parent *Package, name string) *Package {
	p := &Package{InfoBase:InfoBase{name, parent}, importPath:filepath.Join(parent.importPath, name), fullPath:filepath.Join(parent.fullPath, name)}
	if b, err := ioutil.ReadFile(filepath.Join(p.fullPath, "package.flux")); err == nil {
		p.pkgName = string(b)
	} else if pkg, err := build.Import(p.importPath, "", 0); err == nil {
		p.pkgName = pkg.Name
	}
	
	if file, err := os.Open(p.fullPath); err == nil {
		if fileInfos, err := file.Readdir(-1); err == nil {
			for _, fileInfo := range fileInfos {
				name := fileInfo.Name()
				if fileInfo.IsDir() && unicode.IsLetter(([]rune(name))[0]) && name != "testdata" && !HasSuffix(name, ".flux") {
					AddInfo(p, NewPackage(p, name))
				}
			}
		}
	}
	
	return p
}

func (p *Package) load() {
	if p.loaded { return }
	p.loaded = true
	
	pkg, err := getPackage(p.importPath)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			Println(err)
		}
		return
	}
	
	for id := range pkg.Imports {
		FindPackage(id).load()
	}
	
	// create NamedTypes first (to support recursive types); also, parent and add existing types, funcs, consts (special support for builtinPkg and unsafePkg)
	for name, obj := range pkg.Scope.Objects {
		if obj.Kind == ast.Typ {
			n, ok := obj.Type.(*NamedType)
			if !ok {
				n = &NamedType{InfoBase:InfoBase{name, p}}
				obj.Type = n
			}
			AddInfo(p, n)
		}
		if f, ok := obj.Data.(*Func); ok {
			AddInfo(p, f)
		}
		if v, ok := obj.Data.(*Value); ok {
			AddInfo(p, v)
		}
	}
	
	for _, obj := range pkg.Scope.Objects {
		if t, ok := obj.Decl.(*ast.TypeSpec); ok {
			obj.Type.(*NamedType).underlying = specUnderlyingType(t.Type, map[string]bool{})
		}
	}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.CONST {
				for iota, s := range decl.Specs {
					spec := s.(*ast.ValueSpec)
					var typ Type
					if spec.Type != nil {
						typ = specType(spec.Type)
					}
					for _, name := range spec.Names {
						if typ == nil {
							// TODO:  compute value (may be needed for the length of an ArrayType var) and type
							_ = iota
							typ = builtinAstPkg.Scope.Lookup("int").Type.(Type)
						}
						AddInfo(p, &Value{InfoBase{name.Name, p}, typ, true})
					}
				}
			}
		}
	}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok == token.VAR {
					for _, s := range decl.Specs {
						spec := s.(*ast.ValueSpec)
						var typ Type
						if spec.Type != nil {
							typ = specType(spec.Type)
						}
						for _, name := range spec.Names {
							if typ == nil {
								// TODO:  compute type
								// typ = exprType(spec.Values[i])
								typ = builtinAstPkg.Scope.Lookup("int").Type.(Type)
							}
							AddInfo(p, &Value{InfoBase{name.Name, p}, typ, false})
						}
					}
				}
			case *ast.FuncDecl:
				f := &Func{InfoBase{decl.Name.Name, nil}, nil, specType(decl.Type).(*FuncType)}
				parent := Info(p)
				if r := decl.Recv; r != nil {
					// TODO:  FuncType of a method should include the receiver as first argument
					recv := specType(r.List[0].Type)
					f.receiver = recv
					if r, ok := recv.(*PointerType); ok {
						recv = r.element
					}
					parent = recv.(*NamedType)
				}
				AddInfo(parent, f)
			}
		}
	}
}

func specUnderlyingType(x ast.Expr, visitedNames map[string]bool) Type {
	switch x := x.(type) {
	case *ast.ParenExpr:
		return specUnderlyingType(x.X, visitedNames)
	case *ast.Ident:
		if visitedNames[x.Name] {
			Printf("invalid recursive type %s\n", x.Name)
			return nil
		}
		visitedNames[x.Name] = true
		
		if spec, ok := x.Obj.Decl.(*ast.TypeSpec); ok {
			return specUnderlyingType(spec.Type, visitedNames)
		}
	case *ast.SelectorExpr:
		pkgScope := x.X.(*ast.Ident).Obj.Data.(*ast.Scope)
		objName := x.Sel.Name
		if o, ok := pkgScope.Objects[objName]; ok {
			if spec, ok := o.Decl.(*ast.TypeSpec); ok {
				return specUnderlyingType(spec.Type, visitedNames)
			}
		}
	}
	return specType(x)
}

func specType(x ast.Expr) Type {
	switch x := x.(type) {
	case *ast.ParenExpr:
		return specType(x.X)
	case *ast.Ident:
		return x.Obj.Type.(Type)
	case *ast.SelectorExpr:
		pkgID := x.X.(*ast.Ident)
		pkgScope := pkgID.Obj.Data.(*ast.Scope)
		objName := x.Sel.Name
		if o, ok := pkgScope.Objects[objName]; ok {
			return o.Type.(Type)
		} else if pkgID.Name == "C" {
			// unknown C types not supported
			return nil
		}
	case *ast.StarExpr:
		return &PointerType{element:specType(x.X)}
	case *ast.ArrayType:
		if x.Len != nil {
			return &ArrayType{size:intConstValue(x.Len), element:specType(x.Elt)}
		}
		return &SliceType{element:specType(x.Elt)}
	case *ast.Ellipsis:
		return &SliceType{element:specType(x.Elt)}
	case *ast.MapType:
		return &MapType{key:specType(x.Key), value:specType(x.Value)}
	case *ast.ChanType:
		return &ChanType{send:x.Dir&ast.SEND != 0, recv:x.Dir&ast.RECV != 0, element:specType(x.Value)}
	case *ast.FuncType:
		return &FuncType{parameters:getFields(x.Params), results:getFields(x.Results)}
	case *ast.InterfaceType:
		t := &InterfaceType{}
		for _, m := range x.Methods.List {
			switch field := m.Type.(type) {
			case *ast.FuncType:
				t.methods = append(t.methods, &Value{InfoBase{m.Names[0].Name, nil}, specType(field).(*FuncType), false})
			case *ast.Ident:
				emb, ok := field.Obj.Type.(*NamedType).underlying.(*InterfaceType)
				if !ok {
					emb = specType(field.Obj.Decl.(*ast.TypeSpec).Type).(*InterfaceType)
				}
				t.methods = append(t.methods, emb.methods...)
			case *ast.SelectorExpr:
				pkgScope := field.X.(*ast.Ident).Obj.Data.(*ast.Scope)
				objName := field.Sel.Name
				t.methods = append(t.methods, specType(pkgScope.Objects[objName].Decl.(*ast.TypeSpec).Type).(*InterfaceType).methods...)
			default:
				Panicf("unknown interface field:  %#v\n", field)
			}
		}
		return t
	case *ast.StructType:
		return &StructType{fields:getFields(x.Fields)}
	}
	Panicf("unknown type:  %#v\n", x)
	return nil
}

func intConstValue(x ast.Expr) int {
	switch x := x.(type) {
	case *ast.ParenExpr:
		return intConstValue(x.X)
	case *ast.BasicLit:
		l, _ := Atoi(x.Value)
		return l
	case *ast.Ident:
		// x.Obj.Data?
	case *ast.BinaryExpr:
	case *ast.UnaryExpr:
	case *ast.CallExpr:
		// len(...) or Type(...)
	}
	Printf("unknown const value:  %#v\n", x)
	return -1
}

func getFields(f *ast.FieldList) (v []*Value) {
	if f == nil { return }
	for _, field := range f.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				v = append(v, &Value{InfoBase{name.Name, nil}, specType(field.Type), false})
			}
		} else {
			v = append(v, &Value{InfoBase{"", nil}, specType(field.Type), false})
		}
	}
	return
}

func (p *Package) FindPackage(path string) *Package {
	if path == p.importPath { return p }
	for _, pkg := range p.subPkgs {
		if info := pkg.FindPackage(path); info != nil {
			return info
		}
	}
	return nil
}

type Type interface { isType() }
type implementsType struct {}
func (implementsType) isType() {}

type BasicType struct {
	implementsType
	reflectType reflect.Type
}
type PointerType struct {
	implementsType
	element Type
}
type ArrayType struct {
	implementsType
	size int
	element Type
}
type SliceType struct {
	implementsType
	element Type
}
type MapType struct {
	implementsType
	key Type
	value Type
}
type ChanType struct {
	implementsType
	send bool
	recv bool
	element Type
}
type FuncType struct {
	implementsType
	parameters []*Value
	results []*Value
}
type InterfaceType struct {
	implementsType
	methods []*Value
}
type StructType struct {
	implementsType
	fields []*Value
}
type NamedType struct {
	implementsType
	InfoBase
	underlying Type
	methods []*Func
}

func walkType(t Type, op func(*NamedType)) {
	switch t := t.(type) {
	case *PointerType:
		walkType(t.element, op)
	case *ArrayType:
		walkType(t.element, op)
	case *SliceType:
		walkType(t.element, op)
	case *MapType:
		walkType(t.key, op)
		walkType(t.value, op)
	case *ChanType:
		walkType(t.element, op)
	case *FuncType:
		for _, v := range append(t.parameters, t.results...) { walkType(v.typ, op) }
	case *InterfaceType:
		for _, m := range t.methods { walkType(m.typ, op) }
	case *StructType:
		for _, v := range t.fields { walkType(v.typ, op) }
	case *NamedType:
		op(t)
	default:
		panic(fmt.Sprintf("unexpected type %#v\n", t))
	}
}

type Func struct {
	InfoBase
	receiver Type
	typ *FuncType
}
func (f Func) typeWithReceiver() *FuncType {
	t := *f.typ
	if r := f.receiver; r != nil {
		param := &Value{typ:r}
		if p, ok := r.(*PointerType); ok { r = p.element }
		param.name = ToLower(r.(*NamedType).Name()[:1])
		t.parameters = append([]*Value{param}, t.parameters...)
	}
	return &t
}

type Value struct {
	InfoBase
	typ Type
	constant bool
}

type Special struct { *InfoBase }
func NewSpecial(name string) Special { return Special{&InfoBase{name:name}} }
