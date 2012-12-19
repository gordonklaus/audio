package main

import (
	."code.google.com/p/gordon-go/util"
	"go/ast"
	"go/build"
	"go/token"
	"os"
	"fmt"
	."log"
	"path/filepath"
	"reflect"
	."strings"
	."strconv"
	"unicode"
)

var (
	rootPackageInfo = &PackageInfo{loaded:true}
	builtinPkg = &PackageInfo{}
	cPkg = &PackageInfo{InfoBase:InfoBase{"C", rootPackageInfo}, importPath:"C"}
)

func newPackageInfo(parent *PackageInfo, name string) *PackageInfo {
	p := &PackageInfo{InfoBase:InfoBase{name, parent}, importPath:filepath.Join(parent.importPath, name), fullPath:filepath.Join(parent.fullPath, name)}
	
	if file, err := os.Open(p.fullPath); err == nil {
		if fileInfos, err := file.Readdir(-1); err == nil {
			for _, fileInfo := range fileInfos {
				name := fileInfo.Name()
				if fileInfo.IsDir() && unicode.IsLetter(([]rune(name))[0]) && name != "testdata" && !HasSuffix(name, ".fluxmethods") {
					p2 := newPackageInfo(p, name)
					p.subPackages = append(p.subPackages, p2)
				}
			}
		}
	}
	
	return p
}

func init() {
	rootPackageInfo = &PackageInfo{loaded:true}
	for _, srcDir := range build.Default.SrcDirs() {
		rootPackageInfo.fullPath = srcDir
		srcPackageInfo := newPackageInfo(rootPackageInfo, "")
		for _, p := range srcPackageInfo.subPackages {
			p.parent = rootPackageInfo
		}
		rootPackageInfo.subPackages = append(rootPackageInfo.subPackages, srcPackageInfo.subPackages...)
	}
	rootPackageInfo.subPackages = append(rootPackageInfo.subPackages, cPkg)
	Sort(rootPackageInfo.subPackages, "Name")
}

func findPackageInfo(path string) *PackageInfo {
	return rootPackageInfo.findPackageInfo(path)
}

type Info interface {
	Name() string
	SetName(name string)
	Parent() Info
	SetParent(info Info)
	Children() []Info
	AddChild(info Info)
	FluxSourcePath() string
}

type InfoBase struct {
	name string
	parent Info
}
func (i InfoBase) Name() string { return i.name }
func (i *InfoBase) SetName(name string) { i.name = name }
func (i InfoBase) Parent() Info { return i.parent }
func (i *InfoBase) SetParent(info Info) { i.parent = info }
func (i InfoBase) Children() []Info { return nil }
func (i InfoBase) AddChild(info Info) {}
func (i InfoBase) FluxSourcePath() string { return fmt.Sprintf("%v/%v.flux", i.parent.FluxSourcePath(), i.name) }

type PackageInfo struct {
	InfoBase
	importPath string
	fullPath string
	types []*NamedType
	functions []*FuncInfo
	variables []*ValueInfo
	constants []*ValueInfo
	subPackages []*PackageInfo
	loaded bool
}

func (p *PackageInfo) Children() []Info {
	p.load()
	var children []Info
	for _, t := range p.types { children = append(children, t) }
	for _, f := range p.functions { children = append(children, f) }
	for _, v := range p.variables { children = append(children, v) }
	for _, c := range p.constants { children = append(children, c) }
	for _, p := range p.subPackages { children = append(children, p) }
	return children
}
func (p *PackageInfo) AddChild(info Info) {
	info.SetParent(p)
	switch info := info.(type) {
	case *PackageInfo:
		p.subPackages = append(p.subPackages, info)
		Sort(p.subPackages, "Name")
	case *NamedType:
		p.types = append(p.types, info)
		Sort(p.types, "Name")
	case *FuncInfo:
		p.functions = append(p.functions, info)
		Sort(p.functions, "Name")
	default:
		panic("not yet implemented")
	}
}
func (p PackageInfo) FluxSourcePath() string { return p.fullPath }

func (p *PackageInfo) load() {
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
		findPackageInfo(id).load()
	}
	
	// create NamedTypes first (to support recursive types); also, parent and add existing types, funcs, consts (special support for builtinPkg and unsafePkg)
	for name, obj := range pkg.Scope.Objects {
		if obj.Kind == ast.Typ {
			n, ok := obj.Type.(*NamedType)
			if !ok {
				n = &NamedType{InfoBase:InfoBase{name, p}}
				obj.Type = n
			}
			n.parent = p
			p.types = append(p.types, n)
		}
		if f, ok := obj.Data.(*FuncInfo); ok {
			f.parent = p
			p.functions = append(p.functions, f)
		}
		if v, ok := obj.Data.(*ValueInfo); ok {
			v.parent = p
			p.constants = append(p.constants, v)
		}
	}
	
	for _, obj := range pkg.Scope.Objects {
		if t, ok := obj.Decl.(*ast.TypeSpec); ok {
			obj.Type.(*NamedType).underlying = specUnderlyingType(t.Type, map[string]bool{})
		}
	}
	Sort(p.types, "Name")
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
						p.constants = append(p.constants, &ValueInfo{InfoBase{name.Name, p}, typ, true})
					}
				}
			}
		}
	}
	Sort(p.constants, "Name")
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
							p.variables = append(p.variables, &ValueInfo{InfoBase{name.Name, p}, typ, false})
						}
					}
				}
			case *ast.FuncDecl:
				f := &FuncInfo{InfoBase{decl.Name.Name, nil}, nil, specType(decl.Type).(*FuncType)}
				if r := decl.Recv; r != nil {
					// TODO:  FuncType of a method should include the receiver as first argument
					recv := specType(r.List[0].Type)
					f.receiver = recv
					if r, ok := recv.(*PointerType); ok {
						recv = r.element
					}
					nr := recv.(*NamedType)
					f.parent = nr
					nr.methods = append(nr.methods, f)
				} else {
					f.parent = p
					p.functions = append(p.functions, f)
				}
			}
		}
	}
	Sort(p.variables, "Name")
	for _, t := range p.types {
		Sort(t.methods, "Name")
	}
	Sort(p.functions, "Name")
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
				t.methods = append(t.methods, &ValueInfo{InfoBase{m.Names[0].Name, nil}, specType(field).(*FuncType), false})
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

func getFields(f *ast.FieldList) (v []*ValueInfo) {
	if f == nil { return }
	for _, field := range f.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				v = append(v, &ValueInfo{InfoBase{name.Name, nil}, specType(field.Type), false})
			}
		} else {
			v = append(v, &ValueInfo{InfoBase{"", nil}, specType(field.Type), false})
		}
	}
	return
}

func (p *PackageInfo) findPackageInfo(path string) *PackageInfo {
	if path == p.importPath { return p }
	for _, pkg := range p.subPackages {
		if info := pkg.findPackageInfo(path); info != nil {
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
	parameters []*ValueInfo
	results []*ValueInfo
}
type InterfaceType struct {
	implementsType
	methods []*ValueInfo
}
type StructType struct {
	implementsType
	fields []*ValueInfo
}
type NamedType struct {
	implementsType
	InfoBase
	underlying Type
	methods []*FuncInfo
}

func (t NamedType) Children() []Info {
	var children []Info
	for _, m := range t.methods { children = append(children, m) }
	return children
}
func (t *NamedType) AddChild(info Info) {
	info.SetParent(t)
	if info, ok := info.(*FuncInfo); ok {
		index := 0
		for i, f := range t.methods { if f.name > info.name { index = i; break } }
		t.methods = append(t.methods[:index], append([]*FuncInfo{info}, t.methods[index:]...)...)
	} else {
		panic("types can only contain functions")
	}
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

type FuncInfo struct {
	InfoBase
	receiver Type
	typ *FuncType
}
func (FuncInfo) AddChild(info Info) { panic("functions can't have children") }
func (f FuncInfo) FluxSourcePath() string {
	if _, ok := f.parent.(*NamedType); ok {
		return fmt.Sprintf("%smethods/%s.flux", f.parent.FluxSourcePath(), f.name)
	}
	return f.InfoBase.FluxSourcePath()
}
func (f FuncInfo) typeWithReceiver() *FuncType {
	t := *f.typ
	if r := f.receiver; r != nil {
		param := &ValueInfo{typ:r}
		if p, ok := r.(*PointerType); ok { r = p.element }
		param.name = ToLower(r.(*NamedType).Name()[:1])
		t.parameters = append([]*ValueInfo{param}, t.parameters...)
	}
	return &t
}

type ValueInfo struct {
	InfoBase
	typ Type
	constant bool
}

type SpecialInfo struct { *InfoBase }
func special(name string) SpecialInfo { return SpecialInfo{&InfoBase{name, nil}} }
