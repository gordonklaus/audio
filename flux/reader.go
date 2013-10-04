package main

import (
	"code.google.com/p/go.exp/go/types"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func loadFunc(obj types.Object) *funcNode {
	f := newFuncNode()
	f.obj = obj
	f.pkgRefs = map[*types.Package]int{}
	f.awaken = make(chan struct{}, 1)
	f.stop = make(chan struct{}, 1)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fluxPath(obj), nil, parser.ParseComments)
	if err == nil {
		r := &reader{obj.GetPkg(), map[string]*types.Package{}, map[string][]*port{}, map[string]types.Type{}, ast.NewCommentMap(fset, file, file.Comments), map[int]node{}}
		for _, i := range file.Imports {
			path, _ := strconv.Unquote(i.Path.Value)
			pkg := r.pkg.Imports[path]
			name := pkg.Name
			if i.Name != nil {
				name = i.Name.Name
			}
			r.pkgNames[name] = pkg
		}
		decl := file.Decls[len(file.Decls)-1].(*ast.FuncDecl) // get param and result var names from the source, as the obj names might not match
		if decl.Recv != nil {
			r.addVar(decl.Recv.List[0].Names[0].Name, f.inputsNode.newOutput(obj.GetType().(*types.Signature).Recv))
		}
		r.fun(f, decl.Type, decl.Body)
	} else {
		// this is a new func; save it
		if m, ok := obj.(method); ok {
			f.inputsNode.newOutput(m.Type.Recv)
		}
		saveFunc(f)
	}

	return f
}

type reader struct {
	pkg      *types.Package
	pkgNames map[string]*types.Package
	vars     map[string][]*port // there is a bug here; names can be reused between disjoint blocks; vars should be passed as a param and copied, as in writer
	varTypes map[string]types.Type
	cmap     ast.CommentMap
	seqNodes map[int]node
}

func (r *reader) fun(n *funcNode, typ *ast.FuncType, body *ast.BlockStmt) {
	obj := n.obj
	if obj == nil {
		obj = n.output.obj
	}
	sig := obj.GetType().(*types.Signature)

	for i, p := range typ.Params.List {
		v := sig.Params[i]
		r.addVar(p.Names[0].Name, n.inputsNode.newOutput(v))
		n.addPkgRef(v.Type)
	}
	var results []*ast.Field
	if r := typ.Results; r != nil {
		results = r.List
	}
	for i, p := range results {
		r.vars[p.Names[0].Name] = nil
		n.addPkgRef(sig.Results[i].Type)
	}
	r.block(n.funcblk, body.List)
	for i, p := range results {
		r.connect(p.Names[0].Name, n.outputsNode.newInput(sig.Results[i]))
	}
}

func (r *reader) block(b *block, s []ast.Stmt) {
	for _, s := range s {
		switch s := s.(type) {
		case *ast.AssignStmt:
			if s.Tok == token.DEFINE {
				switch x := s.Rhs[0].(type) {
				case *ast.Ident, *ast.SelectorExpr, *ast.StarExpr:
					r.value(b, x, s.Lhs[0], false, s)
				case *ast.CompositeLit:
					r.compositeLit(b, s, x, false)
				case *ast.CallExpr:
					n := r.callOrConvert(b, x)
					outs := outs(n)
					for i, lhs := range s.Lhs {
						r.addVar(name(lhs), outs[i])
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
					r.seq(n, s)
				case *ast.UnaryExpr:
					switch x.Op {
					case token.AND:
						switch y := x.X.(type) {
						case *ast.CompositeLit:
							r.compositeLit(b, s, y, true)
						default:
							r.value(b, x, s.Lhs[0], false, s)
						}
					case token.NOT:
						n := newOperatorNode(&types.Func{Name: x.Op.String()})
						b.addNode(n)
						if arg, ok := x.X.(*ast.Ident); ok {
							r.connect(arg.Name, n.ins[0])
						}
						r.addVar(name(s.Lhs[0]), n.outs[0])
					}
				case *ast.BinaryExpr:
					n := newOperatorNode(&types.Func{Name: x.Op.String()})
					b.addNode(n)
					if arg, ok := x.X.(*ast.Ident); ok {
						r.connect(arg.Name, n.ins[0])
					}
					if arg, ok := x.Y.(*ast.Ident); ok {
						r.connect(arg.Name, n.ins[1])
					}
					r.addVar(name(s.Lhs[0]), n.outs[0])
				case *ast.TypeAssertExpr:
					n := newTypeAssertNode()
					b.addNode(n)
					n.setType(r.typ(x.Type))
					r.connect(name(x.X), n.ins[0])
					r.addVar(name(s.Lhs[0]), n.outs[0])
					r.addVar(name(s.Lhs[1]), n.outs[1])
				case *ast.FuncLit:
					n := newFuncLiteralNode()
					b.addNode(n)
					n.output.setType(r.typ(x.Type))
					r.addVar(name(s.Lhs[0]), n.output)
					r.fun(n, x.Type, x.Body)
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
					r.seq(n, s)
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
				r.value(b, s.Lhs[0], s.Rhs[0], true, s)
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
					r.value(b, x, v.Names[0], false, s)
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
			r.block(n.loopblk, s.Body.List)
			r.seq(n, s)
		case *ast.IfStmt:
			n := newIfNode()
			b.addNode(n)
			r.connect(name(s.Cond), n.input)
			r.block(n.trueblk, s.Body.List)
			if s.Else != nil {
				r.block(n.falseblk, s.Else.(*ast.BlockStmt).List)
			}
			r.seq(n, s)
		case *ast.RangeStmt:
			n := newLoopNode()
			b.addNode(n)
			r.connect(name(s.X), n.input)
			r.addVar(name(s.Key), n.inputsNode.outs[0])
			if s.Value != nil {
				r.addVar(name(s.Value), n.inputsNode.outs[1])
			}
			r.block(n.loopblk, s.Body.List)
			r.seq(n, s)
		case *ast.ExprStmt:
			switch x := s.X.(type) {
			case *ast.CallExpr:
				r.seq(r.callOrConvert(b, x), s)
			}
		}
	}
}

func (r *reader) value(b *block, x, y ast.Expr, set bool, an ast.Node) {
	var addr, indirect bool
	switch x2 := x.(type) {
	case *ast.UnaryExpr:
		x = x2.X
		addr = true
	case *ast.StarExpr:
		x = x2.X
		indirect = true
	}
	n := newValueNode(r.obj(x), addr, indirect, set)
	b.addNode(n)
	if n.x != nil {
		if s, ok := x.(*ast.SelectorExpr); ok {
			r.connect(name(s.X), n.x)
		} else {
			r.connect(name(x), n.x)
		}
	}
	if set {
		r.connect(name(y), n.y)
	} else {
		r.addVar(name(y), n.y)
	}
	r.seq(n, an)
}

func (r *reader) callOrConvert(b *block, x *ast.CallExpr) (n node) {
	if p, ok := x.Fun.(*ast.ParenExpr); ok { // writer puts conversions in parens for easy recognition
		n := newConvertNode()
		b.addNode(n)
		n.setType(r.typ(p.X))
		r.connect(x.Args[0].(*ast.Ident).Name, n.ins[0])
		return n
	}

	obj := r.obj(x.Fun)
	n = newCallNode(obj)
	b.addNode(n)
	args := x.Args
	switch obj.(type) {
	case method:
		recv := x.Fun.(*ast.SelectorExpr).X
		args = append([]ast.Expr{recv}, args...)
	case nil: // func value call
		f := x.Fun.(*ast.Ident)
		args = append([]ast.Expr{f}, args...)
	}
	switch n := n.(type) {
	case *makeNode:
		n.setType(r.typ(args[0]))
		args = args[1:]
	}
	for i, arg := range args {
		if arg, ok := arg.(*ast.Ident); ok {
			// ins(n) must be called on each iteration because making a connection may
			// cause inputs to change, in particular in the case of calling a func value
			r.connect(arg.Name, ins(n)[i])
		}
	}
	return
}

func (r *reader) compositeLit(b *block, s *ast.AssignStmt, x *ast.CompositeLit, ptr bool) {
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
		if it, ok := recv.Underlying.(*types.Interface); ok {
			for _, m := range it.Methods {
				if m.Name == n2 {
					return method{nil, m}
				}
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
	case *ast.StarExpr:
		return &types.Pointer{r.typ(x.X)}
	case *ast.ArrayType:
		if x.Len == nil {
			return &types.Slice{r.typ(x.Elt)}
		}
	case *ast.MapType:
		return &types.Map{r.typ(x.Key), r.typ(x.Value)}
	case *ast.StructType:
		return &types.Struct{}
	case *ast.InterfaceType:
		return &types.Interface{}
	case *ast.FuncType:
		t := &types.Signature{}
		if x.Params != nil {
			for _, f := range x.Params.List {
				t.Params = append(t.Params, &types.Var{Name: f.Names[0].Name, Type: r.typ(f.Type)})
			}
		}
		if x.Results != nil {
			for _, f := range x.Results.List {
				t.Results = append(t.Results, &types.Var{Name: f.Names[0].Name, Type: r.typ(f.Type)})
			}
		}
		return t
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
