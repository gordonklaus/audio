package flux

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
)

var packageInfo chan *PackageInfo = make(chan *PackageInfo)

func getPackageInfo(pathStr string) *PackageInfo {
	packageInfo := newPackageInfo(filepath.Base(pathStr))
	
	pkg, err := build.ImportDir(pathStr, build.FindOnly & build.AllowBinary)
	if err == nil && !pkg.IsCommand() {
		packageInfo.buildPackage = pkg
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
		subPackageInfo.buildPackage.Name = path.Join(packageInfo.Name(), subPackageInfo.Name())
		packageInfo = subPackageInfo
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

type PackageInfo struct {
	buildPackage *build.Package
	parent *PackageInfo
	subPackages []*PackageInfo
	types []*TypeInfo
	functions []FunctionInfo
	variables []ValueInfo
	constants []ValueInfo
	loaded bool
}

func newPackageInfo(name string) *PackageInfo { return &PackageInfo{buildPackage:&build.Package{Name:name}} }

func (p PackageInfo) Name() string { return p.buildPackage.Name }
func (p PackageInfo) Parent() Info { if p.parent == nil { return nil }; return p.parent }
func (p *PackageInfo) Children() []Info {
	p.load()
	var children []Info
	for _, p := range p.subPackages { children = append(children, p) }
	for _, t := range p.types { children = append(children, t) }
	for _, f := range p.functions { children = append(children, f) }
	for _, v := range p.variables { children = append(children, v) }
	for _, c := range p.constants { children = append(children, c) }
	return children
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
					case token.CONST:
						v := spec.(*ast.ValueSpec)
						for _, name := range v.Names {
							p.constants = append(p.constants, ValueInfo{name.Name, p, true})
						}
					case token.VAR:
						v := spec.(*ast.ValueSpec)
						for _, name := range v.Names {
							p.variables = append(p.variables, ValueInfo{name.Name, p, false})
						}
					case token.TYPE:
						p.types = append(p.types, &TypeInfo{spec.(*ast.TypeSpec).Name.Name, p, nil, nil})
						// TODO:  handle interface types differently.  spec.(*ast.TypeSpec).Type.(*ast.InterfaceType)
					}
				}
			}
		}
	}
	Sort(p.constants, "Name")
	Sort(p.variables, "Name")
	Sort(p.types, "Name")
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				functionInfo := FunctionInfo{name:funcDecl.Name.Name}
				if recv := funcDecl.Recv; recv != nil {
					if typeInfo := p.findTypeInfo(recv); typeInfo != nil {
						functionInfo.parent = typeInfo
						typeInfo.methods = append(typeInfo.methods, functionInfo)
					} else {
						// exported method on an unexported type
						// TODO:  expose the type (and its exported methods) but don't allow reference to the type name alone
					}
				} else if typeInfo := p.findTypeInfo(funcDecl.Type.Results); typeInfo != nil {
					functionInfo.parent = typeInfo
					typeInfo.functions = append(typeInfo.functions, functionInfo)
				} else {
					functionInfo.parent = p
					p.functions = append(p.functions, functionInfo)
				}
			}
		}
	}
	for _, typeInfo := range p.types {
		Sort(typeInfo.functions, "Name")
		Sort(typeInfo.methods, "Name")
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
func (p *PackageInfo) findTypeInfo(fields *ast.FieldList) *TypeInfo {
	if fields.NumFields() == 0 { return nil }
	if typeID := exprTypeID(fields.List[0].Type); typeID != nil {
		for _, typeInfo := range p.types {
			if typeInfo.name == typeID.Name {
				return typeInfo
			}
		}
	}
	return nil
}

type ValueInfo struct {
	name string
	parent *PackageInfo
	constant bool
}
func (p ValueInfo) Name() string { return p.name }
func (p ValueInfo) Parent() Info { return p.parent }
func (p ValueInfo) Children() []Info { return nil }

type TypeInfo struct {
	name string
	parent *PackageInfo
	functions []FunctionInfo
	methods []FunctionInfo
}
func (t TypeInfo) Name() string { return t.name }
func (t TypeInfo) Parent() Info { return t.parent }
func (t TypeInfo) Children() []Info {
	var children []Info
	for _, f := range t.functions { children = append(children, f) }
	for _, m := range t.methods { children = append(children, m) }
	return children
}

type FunctionInfo struct {
	name string
	parent Info
}
func (f FunctionInfo) Name() string { return f.name }
func (f FunctionInfo) Parent() Info { return f.parent }
func (f FunctionInfo) Children() []Info { return nil }
