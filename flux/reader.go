package main

import (
	"code.google.com/p/go.exp/go/types"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func loadFunc(f *funcNode) bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fluxPath(f.obj), nil, parser.ParseComments)
	if err != nil {
		return false
	}
	r := &reader{f.obj.GetPkg(), map[string]*types.Package{}, map[string][]*port{}, map[string]types.Type{}, ast.NewCommentMap(fset, file, file.Comments), map[int]node{}}
	for _, i := range file.Imports {
		path, _ := strconv.Unquote(i.Path.Value)
		pkg := r.pkg.Imports[path]
		name := pkg.Name
		if i.Name != nil {
			name = i.Name.Name
		}
		r.pkgNames[name] = pkg
	}
	t := f.obj.GetType().(*types.Signature)
	if t.Recv != nil {
		r.addVar(t.Recv.Name, f.inputsNode.newOutput(t.Recv))
	}
	for _, v := range t.Params {
		r.addVar(v.Name, f.inputsNode.newOutput(v))
		f.addPkgRef(v)
	}
	if d, ok := file.Decls[len(file.Decls)-1].(*ast.FuncDecl); ok {
		r.readBlock(f.funcblk, d.Body.List)
	}
	for _, v := range t.Results {
		r.connect(v.Name, f.outputsNode.newInput(v))
		f.addPkgRef(v)
	}
	return true
}

type reader struct {
	pkg *types.Package
	pkgNames map[string]*types.Package
	vars map[string][]*port // there is a bug here; names can be reused between disjoint blocks; vars should be passed as a param and copied, as in writer
	varTypes map[string]types.Type
	cmap ast.CommentMap
	seqNodes map[int]node
}

func (r *reader) readBlock(b *block, s []ast.Stmt) {
	for _, s := range s {
		switch s := s.(type) {
		case *ast.AssignStmt:
			if s.Tok == token.DEFINE {
				switch x := s.Rhs[0].(type) {
				case *ast.Ident, *ast.SelectorExpr:
					r.newValueNode(b, x, s.Lhs[0], false)
				case *ast.CompositeLit:
					r.newCompositeLiteralNode(b, s, x, false)
				case *ast.CallExpr:
					n := r.newCallNode(b, x)
					for i, lhs := range s.Lhs {
						r.addVar(name(lhs), n.outs[i])
					}
					r.seq(n, s)
				case *ast.IndexExpr:
					n := newIndexNode(false)
					b.addNode(n)
					r.connect(name(x.X), n.x)
					r.connect(name(x.Index), n.key)
					r.addVar(name(s.Lhs[0]), n.outVal)
					if len(s.Lhs) == 2 {
						r.addVar(name(s.Lhs[1]), n.ok)
					}
				case *ast.UnaryExpr:
					switch x.Op {
					case token.AND:
						r.newCompositeLiteralNode(b, s, x.X.(*ast.CompositeLit), true)
					}
				}
			} else {
				if x, ok := s.Lhs[0].(*ast.IndexExpr); ok {
					n := newIndexNode(true)
					b.addNode(n)
					r.connect(name(x.X), n.x)
					r.connect(name(x.Index), n.key)
					if i, ok := s.Rhs[0].(*ast.Ident); ok {
						r.connect(i.Name, n.inVal)
					}
				} else if !r.newValueNode(b, s.Lhs[0], s.Rhs[0], true) {
					for i := range s.Lhs {
						lh := name(s.Lhs[i])
						rh := name(s.Rhs[i])
						r.vars[lh] = append(r.vars[lh], r.vars[rh]...)
						// the types of lhs and rhs are not necessarily the same.
						// varType is set under DeclStmt.
						// until go/types is complete, also setting varType here, which will work in most cases but fail in a few
						r.varTypes[lh] = r.varTypes[rh]
					}
				}
			}
		case *ast.DeclStmt:
			decl := s.Decl.(*ast.GenDecl)
			v := decl.Specs[0].(*ast.ValueSpec)
			switch decl.Tok {
			case token.VAR:
				// this only handles named types; when will go/types do it all for me?
				var t types.Type
				switch x := v.Type.(type) {
				case *ast.Ident:
					t = r.pkg.Scope.Lookup(x.Name).GetType()
				case *ast.SelectorExpr:
					t = r.pkgNames[name(x.X)].Scope.Lookup(x.Sel.Name).GetType()
				}
				if t == nil {
					fmt.Printf("DeclStmt: not fully implemented: %#v\n", v.Type)
				} else {
					r.varTypes[v.Names[0].Name] = t
				}
			case token.CONST:
				lit := v.Values[0].(*ast.BasicLit)
				n := newBasicLiteralNode(lit.Kind)
				b.addNode(n)
				switch lit.Kind {
				case token.INT, token.FLOAT:
					n.text.SetText(lit.Value)
				case token.IMAG:
					// TODO
				case token.STRING, token.CHAR:
					text, _ := strconv.Unquote(lit.Value)
					n.text.SetText(text)
				}
				r.addVar(name(v.Names[0]), n.outs[0])
			}
		case *ast.ForStmt:
			n := newLoopNode()
			b.addNode(n)
			if s.Cond != nil {
				r.connect(name(s.Cond.(*ast.BinaryExpr).Y), n.input)
			}
			if s.Init != nil {
				r.addVar(name(s.Init.(*ast.AssignStmt).Lhs[0]), n.inputsNode.outs[0])
			}
			r.readBlock(n.loopblk, s.Body.List)
		case *ast.IfStmt:
			n := newIfNode()
			b.addNode(n)
			r.connect(name(s.Cond), n.input)
			r.readBlock(n.trueblk, s.Body.List)
			if s.Else != nil {
				r.readBlock(n.falseblk, s.Else.(*ast.BlockStmt).List)
			}
		case *ast.RangeStmt:
			n := newLoopNode()
			b.addNode(n)
			r.connect(name(s.X), n.input)
			r.addVar(name(s.Key), n.inputsNode.outs[0])
			if s.Value != nil {
				r.addVar(name(s.Value), n.inputsNode.outs[1])
			}
			r.readBlock(n.loopblk, s.Body.List)
		case *ast.ExprStmt:
			switch x := s.X.(type) {
			case *ast.CallExpr:
				r.seq(r.newCallNode(b, x), s)
			}
		}
	}
}

func (r *reader) newValueNode(b *block, x, y ast.Expr, set bool) bool {
	var obj types.Object
	switch x := x.(type) {
	case *ast.Ident:        obj = r.pkg.Scope.Lookup(x.Name)
	case *ast.SelectorExpr: obj = r.pkgNames[name(x.X)].Scope.Lookup(x.Sel.Name)
	default:                return false
	}
	n := newValueNode(obj, set)
	b.addNode(n)
	if set {
		r.connect(name(y), n.in)
	} else {
		r.addVar(name(y), n.out)
	}
	return true
}

func (r *reader) newCallNode(b *block, x *ast.CallExpr) (n *callNode) {
	var recvExpr ast.Expr
	switch f := x.Fun.(type) {
	case *ast.Ident:
		n = newCallNode(r.pkg.Scope.Lookup(f.Name))
	case *ast.SelectorExpr:
		n1 := name(f.X)
		n2 := f.Sel.Name
		if pkg, ok := r.pkgNames[n1]; ok {
			n = newCallNode(pkg.Scope.Lookup(n2))
		} else {
			recv, _ := indirect(r.varTypes[n1])
			for _, m := range recv.(*types.NamedType).Methods {
				if m.Name == n2 {
					n = newCallNode(method{nil, m})
					break
				}
			}
			recvExpr = f.X
		}
	}
	b.addNode(n)
	args := x.Args
	if recvExpr != nil {
		args = append([]ast.Expr{recvExpr}, args...)
	}
	for i, arg := range args {
		if arg, ok := arg.(*ast.Ident); ok {
			r.connect(arg.Name, n.ins[i])
		}
	}
	return
}

func (r *reader) newCompositeLiteralNode(b *block, s *ast.AssignStmt, x *ast.CompositeLit, ptr bool) {
	// this only handles named types; when will go/types do it all for me?
	var t types.Type
	switch x := x.Type.(type) {
	case *ast.Ident:
		t = r.pkg.Scope.Lookup(x.Name).(*types.TypeName).Type
	case *ast.SelectorExpr:
		t = r.pkgNames[name(x.X)].Scope.Lookup(x.Sel.Name).(*types.TypeName).Type
	}
	if t == nil {
		fmt.Printf("CompositeLit: not fully implemented: %#v\n", x.Type)
		return
	}
	if ptr {
		t = &types.Pointer{Base: t}
	}
	n := newCompositeLiteralNode()
	b.addNode(n)
	n.setType(t)
elts:
	for _, elt := range x.Elts {
		elt := elt.(*ast.KeyValueExpr)
		field := name(elt.Key)
		val := name(elt.Value)
		for _, in := range n.ins {
			if in.obj.GetName() == field {
				r.connect(val, in)
				continue elts
			}
		}
		panic("no field matching " + field)
	}
	r.addVar(name(s.Lhs[0]), n.outs[0])
}

func (r *reader) connect(name string, in *port) {
	for _, out := range r.vars[name] {
		c := newConnection()
		c.setSrc(out)
		c.setDst(in)
	}
}

func (r *reader) addVar(name string, out *port) {
	if name != "" && name != "_" {
		r.vars[name] = append(r.vars[name], out)
		r.varTypes[name] = out.obj.Type
	}
}

func (r *reader) seq(n node, an ast.Node) {
	if c, ok := r.cmap[an]; ok {
		txt := c[0].Text()
		s := strings.Split(txt[:len(txt)-1], ";")
		seqIn := seqIn(n)
		for _, s := range strings.Split(s[0], ",") {
			if id, err := strconv.Atoi(s); err == nil {
				c := newConnection()
				c.setSrc(seqOut(r.seqNodes[id]))
				c.setDst(seqIn)
			}
		}
		if id, err := strconv.Atoi(s[1]); err == nil {
			r.seqNodes[id] = n
		}
	}
}

func name(x ast.Expr) string {
	return x.(*ast.Ident).Name
}
