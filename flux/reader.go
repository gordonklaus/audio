package main

import (
	"code.google.com/p/go.exp/go/types"
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
	for _, v := range t.Results {
		r.vars[v.Name] = nil
		f.addPkgRef(v)
	}
	r.readBlock(f.funcblk, file.Decls[len(file.Decls)-1].(*ast.FuncDecl).Body.List)
	for _, v := range t.Results {
		r.connect(v.Name, f.outputsNode.newInput(v))
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
				case *ast.Ident, *ast.SelectorExpr, *ast.StarExpr:
					r.newValueNode(b, x, s.Lhs[0], false)
				case *ast.CompositeLit:
					r.newCompositeLiteralNode(b, s, x, false)
				case *ast.CallExpr:
					n := r.newCallNode(b, x)
					for i, lhs := range s.Lhs {
						r.addVar(name(lhs), n.outputs()[i])
					}
					r.seq(n, s)
				case *ast.IndexExpr:
					n := newIndexNode(false)
					b.addNode(n)
					r.connect(name(x.X), n.x)
					if arg, ok := x.Index.(*ast.Ident); ok {
						r.connect(arg.Name, n.key)
					}
					r.addVar(name(s.Lhs[0]), n.outVal)
					if len(s.Lhs) == 2 {
						r.addVar(name(s.Lhs[1]), n.ok)
					}
				case *ast.UnaryExpr:
					switch x.Op {
					case token.AND:
						r.newCompositeLiteralNode(b, s, x.X.(*ast.CompositeLit), true)
					}
				case *ast.BinaryExpr:
					n := newOperatorNode(findOp(x.Op.String()))
					b.addNode(n)
					r.connect(name(x.X), n.ins[0])
					r.connect(name(x.Y), n.ins[1])
					r.addVar(name(s.Lhs[0]), n.outs[0])
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
					break
				}
				if id, ok := s.Lhs[0].(*ast.Ident); ok {
					if _, ok := r.vars[id.Name]; ok {
						for i := range s.Lhs {
							lh := name(s.Lhs[i])
							rh := name(s.Rhs[i])
							r.vars[lh] = append(r.vars[lh], r.vars[rh]...)
						}
						break
					}
				}
				r.newValueNode(b, s.Lhs[0], s.Rhs[0], true)
			}
		case *ast.DeclStmt:
			decl := s.Decl.(*ast.GenDecl)
			v := decl.Specs[0].(*ast.ValueSpec)
			switch decl.Tok {
			case token.VAR:
				name := v.Names[0].Name
				r.vars[name] = nil
				r.varTypes[name] = r.typ(v.Type)
			case token.CONST:
				switch x := v.Values[0].(type) {
				case *ast.BasicLit:
					n := newBasicLiteralNode(x.Kind)
					b.addNode(n)
					switch x.Kind {
					case token.INT, token.FLOAT:
						n.text.SetText(x.Value)
					case token.IMAG:
						// TODO
					case token.STRING, token.CHAR:
						text, _ := strconv.Unquote(x.Value)
						n.text.SetText(text)
					}
					r.addVar(name(v.Names[0]), n.outs[0])
				case *ast.Ident, *ast.SelectorExpr:
					r.newValueNode(b, x, v.Names[0], false)
				}
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

func (r *reader) newValueNode(b *block, x, y ast.Expr, set bool) {
	indirect := false
	if s, ok := x.(*ast.StarExpr); ok {
		x = s.X
		indirect = true
	}
	n := newValueNode(r.obj(x), indirect, set)
	b.addNode(n)
	if n.val != nil {
		if s, ok := x.(*ast.SelectorExpr); ok {
			r.connect(name(s.X), n.val)
		} else {
			r.connect(name(x), n.val)
		}
	}
	if set {
		r.connect(name(y), n.in)
	} else {
		r.addVar(name(y), n.out)
	}
}

func (r *reader) newCallNode(b *block, x *ast.CallExpr) (n node) {
	obj := r.obj(x.Fun)
	n = newCallNode(obj)
	b.addNode(n)
	args := x.Args
	if _, ok := obj.(method); ok {
		recv := x.Fun.(*ast.SelectorExpr).X
		args = append([]ast.Expr{recv}, args...)
	}
	switch n := n.(type) {
	case *makeNode:
		n.setType(r.typ(args[0]))
		args = args[1:]
	}
	for i, arg := range args {
		if arg, ok := arg.(*ast.Ident); ok {
			r.connect(arg.Name, n.inputs()[i])
		}
	}
	return
}

func (r *reader) newCompositeLiteralNode(b *block, s *ast.AssignStmt, x *ast.CompositeLit, ptr bool) {
	t := r.typ(x.Type)
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

func (r *reader) obj(x ast.Expr) types.Object {
	// TODO: shouldn't go/types be able to do this for me?
	switch x := x.(type) {
	case *ast.Ident:
		for s := r.pkg.Scope; s != nil; s = s.Outer {
			 if obj := s.Lookup(x.Name); obj != nil {
				 return obj
			 }
		 }
	case *ast.SelectorExpr:
		// TODO: Type.Method and pkg.Type.Method
		n1 := name(x.X)
		n2 := x.Sel.Name
		if pkg, ok := r.pkgNames[n1]; ok {
			return pkg.Scope.Lookup(n2)
		}
		// TODO: use types.LookupFieldOrMethod()
		t, _ := indirect(r.varTypes[n1])
		recv := t.(*types.NamedType)
		for _, m := range recv.Methods {
			if m.Name == n2 {
				return method{nil, m}
			}
		}
		if st, ok := recv.Underlying.(*types.Struct); ok {
			for _, f := range st.Fields {
				if f.Name == n2 {
					return field{nil, f, recv}
				}
			}
		}
	}
	return nil
}

func (r *reader) typ(x ast.Expr) types.Type {
	// TODO: replace with types.EvalNode()
	switch x := x.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return r.obj(x).GetType()
	case *ast.ArrayType:
		if x.Len == nil {
			return &types.Slice{r.typ(x.Elt)}
		}
	}
	panic("not yet implemented")
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
