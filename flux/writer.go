package main

import (
	"code.google.com/p/go.exp/go/types"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func savePackageName(p *build.Package) {
	for _, name := range append(append(append(p.GoFiles, p.IgnoredGoFiles...), p.CgoFiles...), p.TestGoFiles...) {
		path := filepath.Join(p.Dir, name)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		src := string(b)
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, "", src, parser.PackageClauseOnly)
		if err != nil {
			panic(err)
		}
		oldName := astFile.Name
		i := fset.Position(oldName.Pos()).Offset
		src = src[:i] + p.Name + src[i+len(oldName.Name):]
		if err := ioutil.WriteFile(path, []byte(src), 0666); err != nil {
			panic(err)
		}
	}
	
	if pkg, ok := pkgs[p.ImportPath]; ok {
		pkg.Name = p.Name
	}
	
	// TODO: update all uses?  could get messy with name conflicts.  not that everything has work perfectly.
}

func saveType(t *types.NamedType) {
	w := newWriter(t.Obj)
	defer w.write()
	
	u := t.Underlying
	walkType(u, func(tt *types.NamedType) {
		if p := tt.Obj.Pkg; p != t.Obj.Pkg {
			w.pkgIds[p] = w.id(p.Name)
		}
	})
	w.imports()
	
	w.file.Decls = append(w.file.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{&ast.TypeSpec{
			Name: id(t.Obj.Name),
			Type: w.typ(u),
		}},
	})
}

func saveFunc(f funcNode) {
	w := newWriter(f.obj)
	defer w.write()
	
	for p := range f.pkgRefs {
		w.pkgIds[p] = w.id(p.Name)
	}
	w.imports()
	
	var recv *ast.FieldList
	t := w.typ(f.obj.GetType()).(*ast.FuncType)
	vars := map[*port]*ast.Ident{}
	i := 0
	if m, ok := f.obj.(method); ok {
		recv = &ast.FieldList{List:[]*ast.Field{&ast.Field{Type:w.typ(m.Type.Recv.Type)}}}
		i = -1
	}
	for _, p := range f.inputsNode.outputs() {
		name := w.id(p.obj.GetName())
		if len(p.conns) > 0 {
			vars[p.conns[0].dst] = name
		}
		if i < 0 {
			recv.List[0].Names = []*ast.Ident{name}
		} else {
			t.Params.List[i].Names[0] = name
		}
		i++
	}
	for i, p := range f.outputsNode.inputs() {
		name := w.id(p.obj.GetName())
		vars[p] = name
		t.Results.List[i].Names[0] = name
	}
	stmts := w.block(f.funcblk, vars)
	if t.Results.NumFields() > 0 {
		stmts = append(stmts, &ast.ReturnStmt{})
	}
	
	w.file.Decls = append(w.file.Decls, &ast.FuncDecl{
		Name: id(f.obj.GetName()),
		Recv: recv,
		Type: t,
		Body: &ast.BlockStmt{List:stmts},
	})
}

type writer struct {
	file *ast.File
	pkgIds map[*types.Package]*ast.Ident
	names map[string]int
	write func()
}

func newWriter(obj types.Object) *writer {
	pkg := obj.GetPkg()
	w := &writer{&ast.File{}, map[*types.Package]*ast.Ident{}, map[string]int{}, nil}
	w.file.Name = id(pkg.Name)
	for _, obj := range append(types.Universe.Entries, pkg.Scope.Entries...) {
		w.id(obj.GetName())
	}
	
	w.write = func() {
		bp, err := build.Import(pkg.Path, "", build.FindOnly)
		if err != nil {
			panic(err)
		}
		
		name := obj.GetName()
		if m, ok := obj.(method); ok {
			t := m.Type.Recv.Type
			if p, ok := t.(*types.Pointer); ok {
				t = p.Base
			}
			name = t.(*types.NamedType).Obj.Name + "." + name
		}
		src, err := os.Create(filepath.Join(bp.Dir, name + ".flux.go"))
		if err != nil {
			panic(err)
		}
		
		if err := format.Node(src, token.NewFileSet(), w.file); err != nil {
			panic(err)
		}
		fluxObjs[obj] = w.file
	}
	
	return w
}

func (w *writer) imports() {
	if len(w.pkgIds) == 0 {
		return
	}
	s := []ast.Spec{}
	for p, n := range w.pkgIds {
		i := &ast.ImportSpec{Path:&ast.BasicLit{Kind:token.STRING, Value:strconv.Quote(p.Path)}}
		if n.Name != p.Name {
			i.Name = n
		}
		s = append(s, i)
		w.file.Imports = append(w.file.Imports, i)
	}
	d := &ast.GenDecl{Tok:token.IMPORT, Specs:s}
	if len(s) > 1 {
		d.Lparen = 1
	}
	w.file.Decls = append(w.file.Decls, d)
}

func (w *writer) block(b *block, vars map[*port]*ast.Ident) (s []ast.Stmt) {
	order, ok := b.nodeOrder()
	if !ok {
		fmt.Println("cyclic!")
		return
	}
	
	vars, varsCopy := map[*port]*ast.Ident{}, vars
	for k, v := range varsCopy { vars[k] = v }
	
	for c := range b.conns {
		if _, ok := vars[c.dst]; !ok && c.src.node.block() != b {
			name := w.id("v")
			s = append(s, &ast.DeclStmt{Decl:&ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{
					Names: []*ast.Ident{name},
					Type: w.typ(c.dst.obj.GetType())},
				},
			}})
			vars[c.dst] = name
		}
	}
	for _, n := range order {
		switch n := n.(type) {
		default:
			args := make([]ast.Expr, len(n.inputs()))
			for i, in := range n.inputs() {
				if len(in.conns) > 0 {
					args[i] = vars[in.conns[0].dst]
				} else {
					args[i] = w.zero(in.obj.GetType())
				}
			}
			results, assignExisting := w.results(n, vars)
			switch n := n.(type) {
			case *callNode:
				var f ast.Expr
				if m, ok := n.obj.(method); ok {
					f = &ast.SelectorExpr{X:args[0], Sel:id(m.Name)}
					args = args[1:]
				} else {
					f = w.qualifiedId(n.obj)
				}
				c := &ast.CallExpr{Fun:f, Args:args}
				if len(results) > 0 {
					s = append(s, &ast.AssignStmt{
						Tok: token.DEFINE,
						Lhs: results,
						Rhs: []ast.Expr{c},
					})
				} else {
					s = append(s, &ast.ExprStmt{X:c})
				}
			case *indexNode:
				i := &ast.IndexExpr{X:args[0], Index:args[1]}
				if n.set {
					s = append(s, &ast.AssignStmt{
						Tok: token.ASSIGN,
						Lhs: []ast.Expr{i},
						Rhs: []ast.Expr{args[2]},
					})
				} else if len(results) > 0 {
					s = append(s, &ast.AssignStmt{
						Tok: token.DEFINE,
						Lhs: results,
						Rhs: []ast.Expr{i},
					})
				}
			case *basicLiteralNode:
				if len(results) > 0 {
					val := n.text.GetText()
					switch n.kind {
					case token.STRING:
						val = strconv.Quote(val)
					case token.CHAR:
						val = strconv.QuoteRune([]rune(val)[0])
					}
					s = append(s, &ast.DeclStmt{Decl:&ast.GenDecl{
						Tok: token.CONST,
						Specs: []ast.Spec{&ast.ValueSpec{
							Names: []*ast.Ident{results[0].(*ast.Ident)},
							Values: []ast.Expr{&ast.BasicLit{Kind: n.kind, Value: val}},
						}},
					}})
				}
			}
			if assignExisting != nil {
				s = append(s, assignExisting)
			}
		case *compositeLiteralNode:
			results, assignExisting := w.results(n, vars)
			if len(results) > 0 {
				cl := &ast.CompositeLit{Type: w.typ(*n.typ.typ)}
				for _, in := range n.inputs() {
					if len(in.conns) > 0 {
						cl.Elts = append(cl.Elts, &ast.KeyValueExpr{Key: id(in.obj.GetName()), Value: vars[in.conns[0].dst]})
					}
				}
				s = append(s, &ast.AssignStmt{
					Tok: token.DEFINE,
					Lhs: results,
					Rhs: []ast.Expr{cl},
				})
				if assignExisting != nil {
					s = append(s, assignExisting)
				}
			}
		case *portsNode:
		case *ifNode:
			cond := id("false")
			if len(n.input.conns) > 0 {
				cond = vars[n.input]
			}
			is := &ast.IfStmt{Cond:cond, Body:&ast.BlockStmt{}}
			is.Body.List = w.block(n.trueblk, vars)
			if len(n.falseblk.nodes) > 0 {
				is.Else = &ast.BlockStmt{List:w.block(n.falseblk, vars)}
			}
			s = append(s, is)
		case *loopNode:
			body := &ast.BlockStmt{}
			results, assignExisting := w.results(n.inputsNode, vars)
			if conns := n.input.conns; len(conns) > 0 {
				switch conns[0].src.obj.GetType().(type) {
				case *types.Basic, *types.NamedType:
					if len(results) == 0 {
						results = []ast.Expr{w.id("")}
					}
					s = append(s, &ast.ForStmt{
						Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: results, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
						Cond: &ast.BinaryExpr{Op: token.LSS, X: results[0], Y: vars[n.input]},
						Post: &ast.IncDecStmt{Tok: token.INC, X: results[0]},
						Body: body,
					})
				case *types.Array, *types.Slice, *types.Map, *types.Chan:
					rs := &ast.RangeStmt{X: vars[n.input], Body: body}
					if len(results) == 0 {
						rs.Key = id("_")
						rs.Tok = token.ASSIGN
					} else {
						rs.Key = results[0]
						if len(results) == 2 {
							rs.Value = results[1]
						}
						rs.Tok = token.DEFINE
					}
					s = append(s, rs)
				}
			} else if len(results) > 0 {
				s = append(s, &ast.ForStmt{
					Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: results, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
					Post: &ast.IncDecStmt{Tok: token.INC, X: results[0]},
					Body: body,
				})
			} else {
				s = append(s, &ast.ForStmt{Body: body})
			}
			if assignExisting != nil {
				body.List = []ast.Stmt{assignExisting}
			}
			body.List = append(body.List, w.block(n.loopblk, vars)...)
		}
	}
	return
}

func (w *writer) results(n node, vars map[*port]*ast.Ident) (results []ast.Expr, ass *ast.AssignStmt) {
	any := false
	for _, p := range n.outputs() {
		id := id("_")
		if len(p.conns) > 0 {
			any = true
			id = w.id(p.obj.GetName())
			for _, c := range p.conns {
				if existing, ok := vars[c.dst]; ok {
					if ass == nil {
						ass = &ast.AssignStmt{Tok:token.ASSIGN}
					}
					ass.Lhs = append(ass.Lhs, existing)
					ass.Rhs = append(ass.Rhs, id)
				} else {
					vars[c.dst] = id
				}
			}
		}
		results = append(results, id)
	}
	if !any {
		return nil, nil
	}
	return
}

func (w writer) id(s string) *ast.Ident {
	if s == "" || s == "_" { s = "x" }
	if i, ok := w.names[s]; ok {
		w.names[s]++
		return w.id(s + strconv.Itoa(i))
	}
	w.names[s] = 2
	return id(s)
}

func (w writer) qualifiedId(obj types.Object) ast.Expr {
	i := id(obj.GetName())
	if n, ok := w.pkgIds[obj.GetPkg()]; ok {
		return &ast.SelectorExpr{X:n, Sel:i}
	}
	return i
}

func (w writer) typ(t types.Type) ast.Expr {
	switch t := t.(type) {
	case *types.Basic:
		return id(t.Name)
	case *types.NamedType:
		return w.qualifiedId(t.Obj)
	case *types.Pointer:
		return &ast.StarExpr{X:w.typ(t.Base)}
	case *types.Array:
		return &ast.ArrayType{Len:&ast.BasicLit{Kind:token.INT, Value:strconv.FormatInt(t.Len, 10)}, Elt:w.typ(t.Elt)}
	case *types.Slice:
		return &ast.ArrayType{Elt:w.typ(t.Elt)}
	case *types.Map:
		return &ast.MapType{Key:w.typ(t.Key), Value:w.typ(t.Elt)}
	case *types.Chan:
		return &ast.ChanType{Dir:t.Dir, Value:w.typ(t.Elt)}
	case *types.Signature:
		var p, r ast.FieldList
		for _, v := range t.Params {
			p.List = append(p.List, w.field(v.Name, v.Type))
		}
		for _, v := range t.Results {
			r.List = append(r.List, w.field(v.Name, v.Type))
		}
		return &ast.FuncType{Params:&p, Results:&r}
	case *types.Interface:
		var f []*ast.Field
		for _, m := range t.Methods {
			f = append(f, w.field(m.Name, m.Type))
		}
		return &ast.InterfaceType{Methods:&ast.FieldList{Opening:1, List:f, Closing:1}}
	case *types.Struct:
		var f []*ast.Field
		for _, x := range t.Fields {
			f = append(f, w.field(x.Name, x.Type))
		}
		return &ast.StructType{Fields:&ast.FieldList{Opening:1, List:f, Closing:1}}
	}
	panic(fmt.Sprintf("unexpected type %#v\n", t))
}

func (w writer) field(n string, t types.Type) *ast.Field {
	return &ast.Field{Names:[]*ast.Ident{id(n)}, Type:w.typ(t)}
}

func (w writer) zero(t types.Type) ast.Expr {
	switch t := t.(type) {
	case *types.Basic:
		switch {
		case t.Info & types.IsBoolean != 0:
			return id("false")
		case t.Info & types.IsNumeric != 0:
			return &ast.BasicLit{Kind:token.INT, Value:"0"}
		case t.Info & types.IsString != 0:
			return &ast.BasicLit{Kind:token.STRING, Value:`""`}
		default:
			return id("nil")
		}
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return id("nil")
	case *types.Array, *types.Struct:
		return &ast.CompositeLit{Type:w.typ(t)}
	case *types.NamedType:
		switch t.Underlying.(type) {
		case *types.Array, *types.Struct:
			return &ast.CompositeLit{Type:w.typ(t)}
		}
		return w.zero(t.Underlying)
	}
	panic(fmt.Sprintf("unexpected type %#v\n", t))
}

func walkType(t types.Type, op func(*types.NamedType)) {
	switch t := t.(type) {
	case *types.Basic:
	case *types.NamedType:
		op(t)
	case *types.Pointer:
		walkType(t.Base, op)
	case *types.Array:
		walkType(t.Elt, op)
	case *types.Slice:
		walkType(t.Elt, op)
	case *types.Map:
		walkType(t.Key, op)
		walkType(t.Elt, op)
	case *types.Chan:
		walkType(t.Elt, op)
	case *types.Signature:
		for _, v := range append(t.Params, t.Results...) { walkType(v.Type, op) }
	case *types.Interface:
		for _, m := range t.Methods { walkType(m.Type, op) }
	case *types.Struct:
		for _, v := range t.Fields { walkType(v.Type, op) }
	default:
		panic(fmt.Sprintf("unexpected type %#v\n", t))
	}
}

func id(s string) *ast.Ident {
	return ast.NewIdent(s)
}
