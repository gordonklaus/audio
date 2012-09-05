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
	."strings"
	."strconv"
	"unicode"
)

func GetPackageInfo() *PackageInfo {
	return <-packageInfo
}

func FindPackageInfo(path string) *PackageInfo {
	return GetPackageInfo().FindPackageInfo(path)
}

var packageInfo chan *PackageInfo = make(chan *PackageInfo)

func getPackageInfo(pathStr string) *PackageInfo {
	packageInfo := newPackageInfo(filepath.Base(pathStr))
	
	pkg, err := build.ImportDir(pathStr, 0)
	packageInfo.buildPackage.Dir = pathStr
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok { Println(err) }
	} else if !pkg.IsCommand() {
		packageInfo.buildPackage = *pkg
	}
	
	if file, err := os.Open(pathStr); err == nil {
		if fileInfos, err := file.Readdir(-1); err == nil {
			for _, fileInfo := range fileInfos {
				if fileInfo.IsDir() && unicode.IsLetter(([]rune(fileInfo.Name()))[0]) && !HasSuffix(fileInfo.Name(), methodDirSuffix) {
					subPackageInfo := getPackageInfo(filepath.Join(pathStr, fileInfo.Name()))
					subPackageInfo.parent = packageInfo
					packageInfo.subPackages = append(packageInfo.subPackages, subPackageInfo)
				}
			}
		}
	}
	
	return packageInfo
}

func init() {
	go func() {
		rootPackageInfo := newPackageInfo("")
		for _, srcDir := range build.Default.SrcDirs() {
			subPackageInfo := getPackageInfo(srcDir)
			for _, p := range subPackageInfo.subPackages {
				p.parent = rootPackageInfo
			}
			rootPackageInfo.subPackages = append(rootPackageInfo.subPackages, subPackageInfo.subPackages...)
		}
		Sort(rootPackageInfo.subPackages, "Name")
		for { packageInfo <- rootPackageInfo }
	}()
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
	buildPackage build.Package
	types []*NamedType
	functions []*FuncInfo
	variables []*ValueInfo
	constants []*ValueInfo
	subPackages []*PackageInfo
	loaded bool
}

func newPackageInfo(name string) *PackageInfo { return &PackageInfo{InfoBase:InfoBase{name:name}} }

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
	index := 0
	switch info := info.(type) {
	case *PackageInfo:
		for i, pkg := range p.subPackages { if pkg.name > info.name { index = i; break } }
		p.subPackages = append(p.subPackages[:index], append([]*PackageInfo{info}, p.subPackages[index:]...)...)
	case *FuncInfo:
		for i, f := range p.functions { if f.name > info.name { index = i; break } }
		p.functions = append(p.functions[:index], append([]*FuncInfo{info}, p.functions[index:]...)...)
	default:
		panic("not yet implemented")
	}
}
func (p PackageInfo) FluxSourcePath() string { return p.buildPackage.Dir }

func (p *PackageInfo) load() {
	if p.loaded { return }
	p.loaded = true
	
	// "builtin" contains invalid Go code.  (move this line earlier, so builtin doesn't even show up in the list)
	if p.buildPackage.ImportPath == "builtin" { return }
	
	pkg, err := getPackage(p.buildPackage.ImportPath)
	if err != nil { Println(err); return }
	
	for id := range pkg.Imports {
		FindPackageInfo(id).load()
	}
	
	for name, obj := range pkg.Scope.Objects {
		if _, ok := obj.Decl.(*ast.TypeSpec); ok {
			obj.Type = newNamedType(name, p)
		}
	}
	for _, obj := range pkg.Scope.Objects {
		if t, ok := obj.Decl.(*ast.TypeSpec); ok {
			n := obj.Type.(*NamedType)
			n.underlying = specUnderlyingType(t.Type)
			p.types = append(p.types, n)
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
							typ = Universe.Lookup("int").Type.(Type)
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
								typ = Universe.Lookup("int").Type.(Type)
							}
							p.variables = append(p.variables, &ValueInfo{InfoBase{name.Name, p}, typ, false})
						}
					}
				}
			case *ast.FuncDecl:
				var recv Type
				if r := decl.Recv; r != nil {
					recv = specType(r.List[0].Type)
				}
				tp := decl.Type
				f := &FuncInfo{InfoBase{decl.Name.Name, nil}, FuncType{implementsType{}, recv, getFields(tp.Params), getFields(tp.Results)}}
				if recv != nil {
					if r, ok := recv.(PointerType); ok {
						recv = r.element
					}
					r := recv.(*NamedType)
					f.parent = r
					r.methods = append(r.methods, f)
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

func specUnderlyingType(x ast.Expr) Type {
	switch x := x.(type) {
	case *ast.ParenExpr:
		return specUnderlyingType(x.X)
	case *ast.Ident:
		if spec, ok := x.Obj.Decl.(*ast.TypeSpec); ok {
			return specUnderlyingType(spec.Type)
		}
	case *ast.SelectorExpr:
		pkgScope := x.X.(*ast.Ident).Obj.Data.(*ast.Scope)
		objName := x.Sel.Name
		if spec, ok := pkgScope.Objects[objName].Decl.(*ast.TypeSpec); ok {
			return specUnderlyingType(spec.Type)
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
		pkgScope := x.X.(*ast.Ident).Obj.Data.(*ast.Scope)
		objName := x.Sel.Name
		return pkgScope.Objects[objName].Type.(Type)
	case *ast.StarExpr:
		return PointerType{element:specType(x.X)}
	case *ast.ArrayType:
		if x.Len != nil {
			return ArrayType{size:intConstValue(x.Len), element:specType(x.Elt)}
		}
		return SliceType{element:specType(x.Elt)}
	case *ast.Ellipsis:
		return SliceType{element:specType(x.Elt)}
	case *ast.MapType:
		return MapType{key:specType(x.Key), value:specType(x.Value)}
	case *ast.ChanType:
		return ChanType{send:x.Dir&ast.SEND != 0, recv:x.Dir&ast.RECV != 0, element:specType(x.Value)}
	case *ast.FuncType:
		return FuncType{parameters:getFields(x.Params), results:getFields(x.Results)}
	case *ast.InterfaceType:
		t := InterfaceType{}
		for _, m := range x.Methods.List {
			switch field := m.Type.(type) {
			case *ast.FuncType:
				t.methods = append(t.methods, FuncInfo{InfoBase{m.Names[0].Name, nil}, FuncType{implementsType{}, nil, getFields(field.Params), getFields(field.Results)}})
			case *ast.Ident:
				emb, ok := field.Obj.Type.(*NamedType).underlying.(InterfaceType)
				if !ok {
					emb = specType(field.Obj.Decl.(*ast.TypeSpec).Type).(InterfaceType)
				}
				t.methods = append(t.methods, emb.methods...)
			case *ast.SelectorExpr:
				pkgScope := field.X.(*ast.Ident).Obj.Data.(*ast.Scope)
				objName := field.Sel.Name
				t.methods = append(t.methods, specType(pkgScope.Objects[objName].Decl.(*ast.TypeSpec).Type).(InterfaceType).methods...)
			default:
				Panicf("unknown interface field:  %#v\n", field)
			}
		}
		return t
	case *ast.StructType:
		// TODO
		return StructType{}
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

func getFields(f *ast.FieldList) (v []ValueInfo) {
	if f == nil { return }
	for _, field := range f.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				v = append(v, ValueInfo{InfoBase{name.Name, nil}, specType(field.Type), false})
			}
		} else {
			v = append(v, ValueInfo{InfoBase{"", nil}, specType(field.Type), false})
		}
	}
	return
}

func (p *PackageInfo) FindPackageInfo(path string) *PackageInfo {
	if path == p.buildPackage.ImportPath { return p }
	for _, pkg := range p.subPackages {
		if info := pkg.FindPackageInfo(path); info != nil {
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
	receiver Type
	parameters []ValueInfo
	results []ValueInfo
}
type InterfaceType struct {
	implementsType
	methods []FuncInfo
}
type StructType struct {
	implementsType
	fields []ValueInfo
}
type NamedType struct {
	implementsType
	InfoBase
	underlying Type
	methods []*FuncInfo
}

func newNamedType(name string, parent Info) *NamedType { return &NamedType{InfoBase:InfoBase{name, parent}} }
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

func typeString(t Type) string {
	switch t := t.(type) {
	case PointerType:
		return "*" + typeString(t.element)
	case ArrayType:
		return fmt.Sprintf("[%d]%s", t.size, typeString(t.element))
	case SliceType:
		return "[]" + typeString(t.element)
	case MapType:
		return fmt.Sprintf("[%s]%s", typeString(t.key), typeString(t.value))
	case ChanType:
		s := ""
		switch {
		case t.send && t.recv:
			s = "chan "
		case t.send:
			s = "chan<- "
		case t.recv:
			s = "<-chan "
		}
		return s + typeString(t.element)
	case FuncType:
		return "func" + signature(t)
	case InterfaceType:
		s := "interface{"
		for i, m := range t.methods {
			if i > 0 {
				s += "; "
			}
			s += m.name + signature(m.typ)
		}
		return s + "}"
	case StructType:
		s := "struct{"
		for i, f := range t.fields {
			if i > 0 {
				s += "; "
			}
			if f.name != "" {
				s += f.name + " "
			}
			s += typeString(f.typ)
		}
		return s + "}"
	case *NamedType:
		return t.name
	}
	panic(fmt.Sprintf("no string for type %#v\n", t))
	return ""
}

func signature(f FuncType) string {
	s := paramsString(f.parameters) + " "
	if len(f.results) == 1 && f.results[0].name == "" {
		return s + typeString(f.results[0].typ)
	}
	return s + paramsString(f.results)
}

func paramsString(params []ValueInfo) string {
	s := "("
	for i, p := range params {
		if i > 0 {
			s += ", "
		}
		if p.name != "" {
			s += p.name + " "
		}
		s += typeString(p.typ)
	}
	return s + ")"
}

type FuncInfo struct {
	InfoBase
	typ FuncType
}
func (FuncInfo) AddChild(info Info) { panic("functions can't have children") }
const methodDirSuffix = ".fluxmethods"
func (f FuncInfo) FluxSourcePath() string {
	if _, ok := f.parent.(*NamedType); ok {
		return fmt.Sprintf("%s%s/%s.flux", f.parent.FluxSourcePath(), methodDirSuffix, f.name)
	}
	return f.InfoBase.FluxSourcePath()
}

type ValueInfo struct {
	InfoBase
	typ Type
	constant bool
}

func qualifiedName(x interface{}, pkg *PackageInfo) string {
	if x, ok := x.(Info); ok {
		if p := x.Parent(); p != nil {
			s := ""
			println(p, pkg)
			if p != pkg {
				s = p.Name() + "."
			}
			return s + x.Name()
		}
	}
	if x, ok := x.(Type); ok {
		// TODO:  pass pkg to typeString -- nested names may need to be qualified
		return typeString(x)
	}
	panic(fmt.Sprintf("can't qualify %#v\n", x))
	return ""
}
