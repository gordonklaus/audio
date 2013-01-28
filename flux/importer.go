package main

import "C"

import (
	."code.google.com/p/gordon-go/util"
	"bufio"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	for _, fileName := range append(append(fluxFiles(buildPkg), buildPkg.GoFiles...), buildPkg.CgoFiles...) {
		if strings.HasSuffix(fileName, ".flux.go") { continue }
		file, err := parser.ParseFile(fset, filepath.Join(buildPkg.Dir, fileName), fluxSrc(buildPkg, fileName), 0)
		if err != nil { return nil, err }
		files[fileName] = file
	}
	pkg, err := ast.NewPackage(fset, files, srcImport, builtinAstPkg.Scope)
	if err != nil { return nil, err }
	
	pkgs[path] = pkg
	return pkg, nil
}

func fluxFiles(pkg *build.Package) (files []string) {
	fluxPaths, _ := filepath.Glob(filepath.Join(pkg.Dir, "*.flux"))
	for _, fluxPath := range fluxPaths {
		name := filepath.Base(fluxPath)
		if name == "package.flux" {
			pkgName, err := ioutil.ReadFile(fluxPath)
			if err != nil {
				// TODO
			}
			pkg.Name = string(pkgName)
			continue
		}
		if fluxTypePaths, _ := filepath.Glob(filepath.Join(fluxPath, "*.flux")); fluxTypePaths != nil {
			for _, fluxTypePath := range fluxTypePaths {
				fileName, _ := filepath.Rel(pkg.Dir, fluxTypePath)
				files = append(files, fileName)
			}
			continue
		}
		files = append(files, name)
	}
	return
}

func fluxSrc(pkg *build.Package, fileName string) interface{} {
	if !strings.HasSuffix(fileName, ".flux") { return nil }
	file, err := os.Open(filepath.Join(pkg.Dir, fileName))
	if err != nil {
		// TODO
	}
	defer file.Close()
	var imports, decl string
	r := bufio.NewReader(file)
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	funcName := filepath.Base(fileName[:len(fileName)-5])
	if typeName := filepath.Dir(fileName); typeName != "." {
		typeName = typeName[:len(typeName)-5]
		if funcName == "type" {
			decl = fmt.Sprintf("type %s %s", typeName, line)
		} else {
			i := strings.Index(line, ",")
			j, k := i + 2, len(line)-1
			if i < 0 { i, j = k, k }
			decl = fmt.Sprintf("func (%s) %s(%s)", line[5:i], funcName, line[j:k])
		}
	} else {
		decl = fmt.Sprintf("func %s%s", funcName, line[4:])
	}
	for {
		line, _ := r.ReadString('\n')
		if line == "" || line[0] == '\\' { break }
		importPath, name := Split2(strings.TrimSpace(line), " ")
		imports += fmt.Sprintf("%s\"%s\";", name, importPath)
	}
	return fmt.Sprintf("package %s;import(%s);%s", pkg.Name, imports, decl)
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
