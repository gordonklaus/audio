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
	packageInfo := &PackageInfo{&build.Package{Name:filepath.Base(pathStr)}, nil, []*PackageInfo{}, []ValueInfo{}, []ValueInfo{}, []*TypeInfo{}, []FunctionInfo{}}
	
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
		rootPackageInfo := &PackageInfo{&build.Package{}, nil, []*PackageInfo{}, []ValueInfo{}, []ValueInfo{}, []*TypeInfo{}, []FunctionInfo{}}
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
	constants []ValueInfo
	variables []ValueInfo
	types []*TypeInfo
	functions []FunctionInfo
}

func (p PackageInfo) Name() string { return p.buildPackage.Name }
func (p PackageInfo) Parent() Info { if p.parent == nil { return nil }; return p.parent }
func (p PackageInfo) Children() []Info {
	children := []Info{}
	for _, p := range p.subPackages { children = append(children, p) }
	for _, c := range p.constants { children = append(children, c) }
	for _, v := range p.variables { children = append(children, v) }
	for _, t := range p.types { children = append(children, t) }
	for _, f := range p.functions { children = append(children, f) }
	return children
}

func (p *PackageInfo) Load() {
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
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch d.Tok {
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
						p.types = append(p.types, &TypeInfo{spec.(*ast.TypeSpec).Name.Name, p})
					}
				}
			case *ast.FuncDecl:
				p.functions = append(p.functions, FunctionInfo{d.Name.Name, p})
			}
		}
	}
	
	Sort(p.constants, "Name")
	Sort(p.variables, "Name")
	Sort(p.types, "Name")
	Sort(p.functions, "Name")
}

type ValueInfo struct {
	name string
	parent *PackageInfo
	constant bool
}
func (p ValueInfo) Name() string { return p.name }
func (p ValueInfo) Parent() Info { return p.parent }
func (p ValueInfo) Children() []Info { return []Info{} }

type TypeInfo struct {
	name string
	parent *PackageInfo
}
func (t TypeInfo) Name() string { return t.name }
func (t TypeInfo) Parent() Info { return t.parent }
func (t TypeInfo) Children() []Info { return []Info{} }

type FunctionInfo struct {
	name string
	parent *PackageInfo
}
func (f FunctionInfo) Name() string { return f.name }
func (f FunctionInfo) Parent() Info { return f.parent }
func (f FunctionInfo) Children() []Info { return []Info{} }
