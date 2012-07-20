package main

import (
	."code.google.com/p/gordon-go/util"
	"go/build"
	"go/token"
	"go/parser"
	"go/ast"
	"os"
	."log"
	"path"
	"path/filepath"
	"unsafe"
)

var packageInfo chan *PackageInfo = make(chan *PackageInfo)

func getPackageInfo(pathStr string) *PackageInfo {
	packageInfo := newPackageInfo(filepath.Base(pathStr))
	
	pkg, err := build.ImportDir(pathStr, build.FindOnly & build.AllowBinary)
	if err == nil && !pkg.IsCommand() {
		packageInfo.buildPackage = *pkg
	} else if err != nil {
		if _, ok := err.(*build.NoGoError); !ok { Println(err) }
	}
	
	if file, err := os.Open(pathStr); err == nil {
		if fileInfos, err := file.Readdir(-1); err == nil {
			for _, fileInfo := range fileInfos {
				if fileInfo.IsDir() {
					subPackageInfo := getPackageInfo(filepath.Join(pathStr, fileInfo.Name()))
					if len(subPackageInfo.subPackages) > 0 || len(subPackageInfo.buildPackage.ImportPath) > 0 {
						subPackageInfo.parent = packageInfo
						packageInfo.subPackages = append(packageInfo.subPackages, subPackageInfo)
					}
				}
			}
		}
	}
	
	if len(packageInfo.buildPackage.ImportPath) == 0 && len(packageInfo.subPackages) == 1 && len(packageInfo.subPackages[0].buildPackage.ImportPath) == 0 {
		subPackageInfo := packageInfo.subPackages[0]
		subPackageInfo.name = path.Join(packageInfo.name, subPackageInfo.name)
		packageInfo = subPackageInfo
	}
	
	return packageInfo
}

func init() {
	go func() {
		rootPackageInfo := newPackageInfo("")
		rootPackageInfo.types = append(rootPackageInfo.types, BoolTypeInfo{newTypeInfoBase("bool", rootPackageInfo)},
															IntTypeInfo{newTypeInfoBase("int", rootPackageInfo), true, 8*int(unsafe.Sizeof(int(0)))},
															IntTypeInfo{newTypeInfoBase("int8", rootPackageInfo), true, 8},
															IntTypeInfo{newTypeInfoBase("int16", rootPackageInfo), true, 16},
															IntTypeInfo{newTypeInfoBase("int32", rootPackageInfo), true, 32},
															IntTypeInfo{newTypeInfoBase("int64", rootPackageInfo), true, 64},
															IntTypeInfo{newTypeInfoBase("uint", rootPackageInfo), false, 8*int(unsafe.Sizeof(uint(0)))},
															IntTypeInfo{newTypeInfoBase("uint8", rootPackageInfo), false, 8},
															IntTypeInfo{newTypeInfoBase("uint16", rootPackageInfo), false, 16},
															IntTypeInfo{newTypeInfoBase("uint32", rootPackageInfo), false, 32},
															IntTypeInfo{newTypeInfoBase("uint64", rootPackageInfo), false, 64},
															IntTypeInfo{newTypeInfoBase("uintptr", rootPackageInfo), false, 8*int(unsafe.Sizeof(uintptr(0)))},
															FloatTypeInfo{newTypeInfoBase("float32", rootPackageInfo), 32},
															FloatTypeInfo{newTypeInfoBase("float64", rootPackageInfo), 64},
															ComplexTypeInfo{newTypeInfoBase("complex64", rootPackageInfo), 64},
															ComplexTypeInfo{newTypeInfoBase("complex128", rootPackageInfo), 128},
															IntTypeInfo{newTypeInfoBase("byte", rootPackageInfo), false, 8},
															IntTypeInfo{newTypeInfoBase("rune", rootPackageInfo), true, 32},
															StringTypeInfo{newTypeInfoBase("string", rootPackageInfo)},
										)
		for _, srcDir := range build.Default.SrcDirs() {
			subPackageInfo := getPackageInfo(srcDir)
			for _, p := range subPackageInfo.subPackages {
				p.parent = rootPackageInfo
			}
			rootPackageInfo.subPackages = append(rootPackageInfo.subPackages, subPackageInfo.subPackages...)
		}
		for { packageInfo <- rootPackageInfo }
	}()
}

func GetPackageInfo() *PackageInfo {
	return <-packageInfo
}

type Info interface {
	Name() string
	Parent() Info
	Children() []Info
}

type InfoBase struct {
	name string
	parent Info
}
func (i InfoBase) Name() string { return i.name }
func (i InfoBase) Parent() Info { return i.parent }
func (i InfoBase) Children() []Info { return nil }

type PackageInfo struct {
	InfoBase
	buildPackage build.Package
	types []TypeInfo
	functions []*FunctionInfo
	variables []ValueInfo
	constants []ValueInfo
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

func typeName(expr ast.Expr) string {
	switch typ := expr.(type) {
	case nil:
		return "<untyped>"
	case *ast.Ident:
		return typ.Name
	case *ast.SelectorExpr:
		return typeName(typ.X) + typ.Sel.Name
	case *ast.StarExpr:
		return "*" + typeName(typ.X)
	case *ast.ArrayType:
		return "[]" + typeName(typ.Elt)
	case *ast.Ellipsis:
		return "..." + typeName(typ.Elt)
	case *ast.StructType:
		return "struct{with some fields}"
	case *ast.InterfaceType:
		return "interface{with some methods}"
	case *ast.MapType:
		return "map[" + typeName(typ.Key) + "]" + typeName(typ.Value)
	case *ast.ChanType:
		s := ""
		switch typ.Dir {
		case ast.SEND: s = "chan<- "
		case ast.RECV: s = "<-chan "
		case ast.SEND | ast.RECV: s = "chan "
		}
		return s + typeName(typ.Value)
	case *ast.FuncType:
		return "func(with) args"
	}
	Panicf("other type!:  %#v", expr)
	return ""
}

func (p *PackageInfo) load() {
	if p.loaded { return }
	p.loaded = true
	
	if len(p.buildPackage.Dir) == 0 { return }
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, p.buildPackage.Dir, func(fileInfo os.FileInfo) bool {
		for _, file := range append(p.buildPackage.GoFiles, p.buildPackage.CgoFiles...) {
			if fileInfo.Name() == file {
				return true
			}
		}
		return false
	}, parser.ParseComments)
	if err != nil { panic(err) }
	if len(pkgs) != 1 { Panicf("%v packages found in %v; expected 1.", len(pkgs), p.buildPackage.Dir) }
	
	pkg := pkgs[p.buildPackage.Name]
	if !ast.PackageExports(pkg) { return }
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					switch genDecl.Tok {
					case token.TYPE:
						p.types = append(p.types, NewTypeInfo(spec.(*ast.TypeSpec), p))
					}
				}
			}
		}
	}
	Sort(p.types, "Name")
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch decl.Tok {
					case token.CONST:
						v := spec.(*ast.ValueSpec)
						for _, name := range v.Names {
							p.constants = append(p.constants, ValueInfo{InfoBase{name.Name, p}, typeName(v.Type), true})
						}
					case token.VAR:
						v := spec.(*ast.ValueSpec)
						for _, name := range v.Names {
							p.variables = append(p.variables, ValueInfo{InfoBase{name.Name, p}, typeName(v.Type), false})
						}
					}
				}
			case *ast.FuncDecl:
				functionInfo := &FunctionInfo{InfoBase:InfoBase{decl.Name.Name, nil}}
				for _, field := range decl.Type.Params.List {
					for _, name := range field.Names {
						functionInfo.parameters = append(functionInfo.parameters, ValueInfo{InfoBase{name.Name, nil}, typeName(field.Type), false})
					}
				}
				if results := decl.Type.Results; results != nil {
					for _, field := range results.List {
						if field.Names == nil {
							functionInfo.results = append(functionInfo.results, ValueInfo{InfoBase{"", nil}, typeName(field.Type), false})
						} else {
							for _, name := range field.Names {
								functionInfo.results = append(functionInfo.results, ValueInfo{InfoBase{name.Name, nil}, typeName(field.Type), false})
							}
						}
					}
				}
				if recv := decl.Recv; recv != nil {
					if typeInfo := p.findTypeInfo(recv); typeInfo != nil {
						functionInfo.parent = typeInfo
						*typeInfo.Methods() = append(*typeInfo.Methods(), functionInfo)
					} else {
						// exported method on an unexported type
						// TODO:  expose the type (and its exported methods) but don't allow reference to the type name alone
					}
				} else {
					functionInfo.parent = p
					p.functions = append(p.functions, functionInfo)
				}
			}
		}
	}
	Sort(p.constants, "Name")
	Sort(p.variables, "Name")
	for _, typeInfo := range p.types {
		Sort(*typeInfo.Methods(), "Name")
	}
	Sort(p.functions, "Name")
}

func exprTypeID(expr ast.Expr) *ast.Ident {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return exprTypeID(e.X)
	case *ast.Ident:
		return e
	}
	return nil
}
func (p *PackageInfo) findTypeInfo(fields *ast.FieldList) TypeInfo {
	if fields.NumFields() == 0 { return nil }
	if typeID := exprTypeID(fields.List[0].Type); typeID != nil {
		for _, typeInfo := range p.types {
			if typeInfo.Name() == typeID.Name {
				return typeInfo
			}
		}
	}
	return nil
}

type TypeInfo interface {
	Info
	Methods() *[]*FunctionInfo
}
type TypeInfoBase struct {
	InfoBase
	methods []*FunctionInfo
}
func newTypeInfoBase(name string, parent Info) *TypeInfoBase { return &TypeInfoBase{InfoBase{name, parent}, nil} }
func NewTypeInfo(spec *ast.TypeSpec, parent Info) TypeInfo {
	switch typ := spec.Type.(type) {
	// case *ast.Ident:
	// 	return typ.Name
	// case *ast.SelectorExpr:
	// 	return typeName(typ.X) + typ.Sel.Name
	// case *ast.StarExpr:
	// 	return "*" + typeName(typ.X)
	// case *ast.ArrayType:
	// 	return "[]" + typeName(typ.Elt)
	// case *ast.Ellipsis:
	// 	return "..." + typeName(typ.Elt)
	// case *ast.StructType:
	// 	return "struct{with some fields}"
	// case *ast.InterfaceType:
	// 	return "interface{with some methods}"
	// case *ast.MapType:
	// 	return "map[" + typeName(typ.Key) + "]" + typeName(typ.Value)
	// case *ast.ChanType:
	// 	s := ""
	// 	switch typ.Dir {
	// 	case ast.SEND: s = "chan<- "
	// 	case ast.RECV: s = "<-chan "
	// 	case ast.SEND | ast.RECV: s = "chan "
	// 	}
	// 	return s + typeName(typ.Value)
	// case *ast.FuncType:
	// 	return "func(with) args"
	}
	return newTypeInfoBase(spec.Name.Name, parent)
	Panicf("other type!:  %#v", spec)
	return nil
}
func (t TypeInfoBase) Children() []Info {
	var children []Info
	for _, m := range t.methods { children = append(children, m) }
	return children
}
func (t *TypeInfoBase) Methods() *[]*FunctionInfo { return &t.methods }

type BoolTypeInfo struct { *TypeInfoBase }
type IntTypeInfo struct {
	*TypeInfoBase
	signed bool
	bits int
}
type FloatTypeInfo struct {
	*TypeInfoBase
	bits int
}
type ComplexTypeInfo struct {
	*TypeInfoBase
	bits int
}
type StringTypeInfo struct { *TypeInfoBase }
type ArrayTypeInfo struct {
	*TypeInfoBase
	size int
	element TypeInfo
}
type ChanTypeInfo struct {
	*TypeInfoBase
	send bool
	recv bool
	element TypeInfo
}
type FunctionTypeInfo struct {
	*TypeInfoBase
	receiver TypeInfo
	inputs []TypeInfo
	outputs []TypeInfo
}
type InterfaceTypeInfo struct {
	*TypeInfoBase
	methods []FunctionTypeInfo
}
type MapTypeInfo struct {
	*TypeInfoBase
	key TypeInfo
	value TypeInfo
}
type PointerTypeInfo struct {
	*TypeInfoBase
	element TypeInfo
}
type SliceTypeInfo struct {
	*TypeInfoBase
	element TypeInfo
}
type StructTypeInfo struct {
	*TypeInfoBase
	fields []ValueInfo
}

type FunctionInfo struct {
	InfoBase
	tmp string
	parameters []ValueInfo
	results []ValueInfo
}

type ValueInfo struct {
	InfoBase
	typeName string
	constant bool
}
