// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

// Flux code is stored in a subset of the Go language because this makes it fully interoperable with pure Go projects.  If instead it were stored in a custom Flux file format then, in order to share Flux code with non-Flux projects, generated Go files would also have to be saved in addition to Flux files, whereas in the normal build process these Go files would be temporary artifacts.

func loadFunc(obj types.Object) *funcNode {
	f := newFuncNode(obj, nil)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fluxPath(obj), nil, parser.ParseComments)
	if err == nil {
		r := &reader{fset, obj.GetPkg(), types.NewScope(obj.GetPkg().Scope()), map[string]*port{}, map[string][]*connection{}, map[string]*localVar{}, ast.NewCommentMap(fset, file, file.Comments), map[int]node{}}
		for _, i := range file.Imports {
			path, _ := strconv.Unquote(i.Path.Value)
			pkg, err := getPackage(path)
			if err != nil {
				panic(err)
			}
			name := pkg.Name
			if i.Name != nil {
				name = i.Name.Name
			}
			r.scope.Insert(types.NewPkgName(0, pkg, name))
		}
		decl := file.Decls[len(file.Decls)-1].(*ast.FuncDecl) // get param and result var names from the source, as the obj names might not match
		if decl.Recv != nil {
			r.out(decl.Recv.List[0].Names[0], f.inputsNode.newOutput(obj.GetType().(*types.Signature).Recv))
		}
		r.fun(f, decl.Type, decl.Body)
	} else {
		// this is a new func; save it
		if isMethod(obj) {
			f.inputsNode.newOutput(obj.GetType().(*types.Signature).Recv)
		}
		saveFunc(f)
	}
	return f
}

type reader struct {
	fset      *token.FileSet
	pkg       *types.Package
	scope     *types.Scope
	ports     map[string]*port
	conns     map[string][]*connection
	localVars map[string]*localVar
	cmap      ast.CommentMap
	seqNodes  map[int]node
}

func (r *reader) fun(n *funcNode, typ *ast.FuncType, body *ast.BlockStmt) {
	obj := n.obj
	f := n
	if obj == nil {
		obj = n.output.obj
		f = n.blk.func_()
	}
	sig := obj.GetType().(*types.Signature)

	for i, p := range typ.Params.List {
		v := sig.Params[i]
		r.out(p.Names[0], n.inputsNode.newOutput(v))
		f.addPkgRef(v.Type)
	}
	if sig.IsVariadic {
		n.inputsNode.outs[len(n.inputsNode.outs)-1].valView.ellipsis = true
	}
	var results []*ast.Field
	if r := typ.Results; r != nil {
		results = r.List
	}
	for i, p := range results {
		r.conns[p.Names[0].Name] = []*connection{}
		f.addPkgRef(sig.Results[i].Type)
	}
	r.block(n.funcblk, body.List)
	for i, p := range results {
		r.in(p.Names[0], n.outputsNode.newInput(sig.Results[i]))
	}
}

func (r *reader) block(b *block, s []ast.Stmt) {
	for _, s := range s {
		switch s := s.(type) {
		case *ast.AssignStmt:
			if s.Tok == token.DEFINE {
				switch x := s.Rhs[0].(type) {
				case *ast.BinaryExpr:
					n := newOperatorNode(types.NewFunc(0, nil, x.Op.String(), nil))
					b.addNode(n)
					r.in(x.X, n.ins[0])
					r.in(x.Y, n.ins[1])
					r.out(s.Lhs[0], n.outs[0])
				case *ast.CallExpr:
					if p, ok := x.Fun.(*ast.ParenExpr); ok { // writer puts conversions in parens for easy recognition
						n := newConvertNode()
						b.addNode(n)
						n.setType(r.typ(p.X))
						r.in(x.Args[0], n.ins[0])
						r.out(s.Lhs[0], n.outs[0])
					} else {
						n := r.call(b, x, "", s)
						for i, p := range outs(n) {
							r.out(s.Lhs[i], p)
						}
					}
				case *ast.CompositeLit:
					r.compositeLit(b, x, false, s)
				case *ast.FuncLit:
					n := newFuncNode(nil, b.childArranged)
					b.addNode(n)
					n.output.setType(r.typ(x.Type))
					r.out(s.Lhs[0], n.output)
					r.fun(n, x.Type, x.Body)
				case *ast.Ident, *ast.SelectorExpr, *ast.StarExpr:
					r.value(b, x, s.Lhs[0], false, s)
				case *ast.IndexExpr:
					r.index(b, x, s.Lhs[0], false, s)
				case *ast.SliceExpr:
					n := newSliceNode()
					b.addNode(n)
					r.in(x.X, n.x)
					r.in(x.Low, n.low)
					if x.High == nil {
						n.removePortBase(n.high)
						n.high = nil
					} else {
						r.in(x.High, n.high)
					}
					if x.Max != nil {
						n.max = n.newInput(newVar("max", types.Typ[types.Int]))
						r.in(x.Max, n.max)
					}
					r.out(s.Lhs[0], n.y)
				case *ast.TypeAssertExpr:
					n := newTypeAssertNode()
					b.addNode(n)
					n.setType(r.typ(x.Type))
					r.in(x.X, n.ins[0])
					r.out(s.Lhs[0], n.outs[0])
					r.out(s.Lhs[1], n.outs[1])
				case *ast.UnaryExpr:
					switch x.Op {
					case token.AND:
						switch y := x.X.(type) {
						case *ast.CompositeLit:
							r.compositeLit(b, y, true, s)
						case *ast.IndexExpr:
							r.index(b, y, s.Lhs[0], false, s)
						default:
							r.value(b, x, s.Lhs[0], false, s)
						}
					case token.NOT:
						n := newOperatorNode(types.NewFunc(0, nil, x.Op.String(), nil))
						b.addNode(n)
						r.in(x.X, n.ins[0])
						r.out(s.Lhs[0], n.outs[0])
					case token.ARROW:
						n := r.sendrecv(b, x.X, nil, s)
						r.out(s.Lhs[0], n.elem)
						r.out(s.Lhs[1], n.ok)
					}
				}
			} else {
				lh := s.Lhs[0]
				rh := s.Rhs[0]
				if x, ok := lh.(*ast.IndexExpr); ok {
					r.index(b, x, rh, true, s)
				} else if id, ok := lh.(*ast.Ident); !ok || r.conns[id.Name] == nil {
					r.value(b, lh, rh, true, s)
				} else {
					c := newConnection()
					lh := name(lh)
					rh := name(rh)
					c.setSrc(r.ports[rh])
					if cmt, ok := r.cmap[s]; ok {
						c.src.conntxt.SetText(cmt[0].List[0].Text[2:])
						c.toggleHidden()
					}
					if p, ok := r.ports[lh]; ok {
						c.feedback = true
						c.setDst(p)
					} else {
						r.conns[lh] = append(r.conns[lh], c)
					}
				}
			}
		case *ast.BranchStmt:
			n := newBranchNode(s.Tok.String())
			b.addNode(n)
			r.seq(n, s)
		case *ast.DeclStmt:
			decl := s.Decl.(*ast.GenDecl)
			v := decl.Specs[0].(*ast.ValueSpec)
			switch decl.Tok {
			case token.VAR:
				name := v.Names[0].Name
				if r.cmap[s] != nil {
					lv := &localVar{}
					lv.Name = name
					lv.Type = r.typ(v.Type)
					lv.refs = map[*valueNode]bool{}
					r.localVars[name] = lv
				} else if v.Type != nil {
					r.scope.Insert(newVar(name, r.typ(v.Type))) // temporary local var has nil Pkg
					r.conns[name] = []*connection{}
				} else {
					r.out(v.Names[0], b.node.(*loopNode).inputsNode.outs[1])
				}
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
					r.out(v.Names[0], n.outs[0])
				case *ast.Ident, *ast.SelectorExpr:
					r.value(b, x, v.Names[0], false, s)
				}
			}
		case *ast.DeferStmt:
			r.call(b, s.Call, "defer ", s)
		case *ast.ExprStmt:
			switch x := s.X.(type) {
			case *ast.CallExpr:
				r.call(b, x, "", s)
			case *ast.UnaryExpr:
				r.sendrecv(b, x.X, nil, s)
			}
		case *ast.ForStmt:
			n := newLoopNode(b.childArranged)
			b.addNode(n)
			if s.Cond != nil {
				r.in(s.Cond.(*ast.BinaryExpr).Y, n.input)
			}
			if s.Init != nil {
				r.out(s.Init.(*ast.AssignStmt).Lhs[0], n.inputsNode.outs[0])
			}
			r.block(n.loopblk, s.Body.List)
			r.seq(n, s)
		case *ast.GoStmt:
			r.call(b, s.Call, "go ", s)
		case *ast.IfStmt:
			n := newIfNode(b.childArranged)
			b.addNode(n)
			for s := ast.Stmt(s); s != nil; {
				b, cond := n.newBlock()
				switch s2 := s.(type) {
				case *ast.IfStmt:
					r.in(s2.Cond, cond)
					r.block(b, s2.Body.List)
					s = s2.Else
				case *ast.BlockStmt:
					r.block(b, s2.List)
					s = nil
				}
			}
			r.seq(n, s)
		case *ast.RangeStmt:
			n := newLoopNode(b.childArranged)
			b.addNode(n)
			r.in(s.X, n.input)
			r.out(s.Key, n.inputsNode.outs[0])
			if s.Value != nil {
				r.out(s.Value, n.inputsNode.outs[1])
			}
			r.block(n.loopblk, s.Body.List)
			r.seq(n, s)
		case *ast.SelectStmt:
			n := newSelectNode(b.childArranged)
			b.addNode(n)
			for _, s := range s.Body.List {
				s := s.(*ast.CommClause)
				c := n.newCase()
				switch s := s.Comm.(type) {
				case *ast.AssignStmt:
					c.send = false
					r.in(s.Rhs[0].(*ast.UnaryExpr).X, c.ch)
					r.out(s.Lhs[0], c.elemOk.outs[0])
					r.out(s.Lhs[1], c.elemOk.outs[1])
				case *ast.ExprStmt:
					c.send = false
					r.in(s.X.(*ast.UnaryExpr).X, c.ch)
				case *ast.SendStmt:
					r.in(s.Chan, c.ch)
					r.in(s.Value, c.elem)
				case nil:
					c.setDefault()
				}
				r.block(c.blk, s.Body)
			}
			r.seq(n, s)
		case *ast.SendStmt:
			r.sendrecv(b, s.Chan, s.Value, s)
		}
	}
}

func (r *reader) value(b *block, x, y ast.Expr, set bool, an ast.Node) {
	if x2, ok := x.(*ast.UnaryExpr); ok {
		x = x2.X
	}
	n := newValueNode(r.obj(x), set) // r.obj returns nil for a *ast.StarExpr, which is what valueNode expects for an assignment (*x = y)
	b.addNode(n)
	switch x := x.(type) {
	case *ast.SelectorExpr:
		r.in(x.X, n.x)
	case *ast.StarExpr:
		r.in(x.X, n.x)
	}
	if set {
		r.in(y, n.y)
	} else {
		r.out(y, n.y)
	}
	r.seq(n, an)
}

func (r *reader) call(b *block, x *ast.CallExpr, godefer string, s ast.Stmt) node {
	obj := r.obj(x.Fun)
	n := newCallNode(obj, godefer)
	b.addNode(n)
	args := x.Args
	switch {
	case isMethod(obj):
		recv := x.Fun.(*ast.SelectorExpr).X
		args = append([]ast.Expr{recv}, args...)
	case obj == nil: // func value call
		args = append([]ast.Expr{x.Fun}, args...)
	}
	if n, ok := n.(interface {
		setType(types.Type)
	}); ok {
		n.setType(r.typ(args[0]))
		args = args[1:]
	}
	for i, arg := range args {
		if i >= len(ins(n)) {
			var newInput func(*types.Var) *port
			var v *types.Var
			switch n := n.(type) {
			case *callNode:
				newInput = n.newInput
				_, v = n.variadic()
			case *appendNode:
				newInput = n.newInput
				v = ins(n)[0].obj
			}
			if x.Ellipsis == 0 {
				v = newVar(v.Name, v.Type.(*types.Slice).Elem)
			}
			in := newInput(v)
			if x.Ellipsis != 0 {
				in.valView.ellipsis = true
			}
		}
		r.in(arg, ins(n)[i])
	}
	r.seq(n, s)
	return n
}

func (r *reader) compositeLit(b *block, x *ast.CompositeLit, ptr bool, s *ast.AssignStmt) {
	t := r.typ(x.Type)
	if ptr {
		t = &types.Pointer{Elem: t}
	}
	n := newCompositeLiteralNode()
	b.addNode(n)
	n.setType(t)
elts:
	for _, elt := range x.Elts {
		elt := elt.(*ast.KeyValueExpr)
		field := name(elt.Key)
		for _, in := range n.ins {
			if in.obj.GetName() == field {
				r.in(elt.Value, in)
				continue elts
			}
		}
		panic("no field matching " + field)
	}
	r.out(s.Lhs[0], n.outs[0])
}

func (r *reader) index(b *block, x *ast.IndexExpr, y ast.Expr, set bool, s *ast.AssignStmt) {
	n := newIndexNode(set)
	b.addNode(n)
	r.in(x.X, n.x)
	r.in(x.Index, n.key)
	if set {
		r.in(y, n.elem)
	} else {
		r.out(y, n.elem)
	}
	if len(s.Lhs) == 2 {
		r.out(s.Lhs[1], n.ok)
	}
	r.seq(n, s)
}

func (r *reader) sendrecv(b *block, ch, elem ast.Expr, s ast.Stmt) *chanNode {
	n := newChanNode(elem != nil)
	b.addNode(n)
	r.in(ch, n.ch)
	if n.send {
		r.in(elem, n.elem)
	}
	r.seq(n, s)
	return n
}

func (r *reader) obj(x ast.Expr) types.Object {
	// TODO: go/types should be able to do this for me.  see http://golang.org/issue/7151
	switch x := x.(type) {
	case *ast.Ident:
		if v, ok := r.localVars[x.Name]; ok {
			return v
		}
		if obj := r.scope.LookupParent(x.Name); obj != nil {
			if v, ok := obj.(*types.Var); ok && v.Pkg == nil { // ignore temporary local vars
				return nil
			}
			return obj
		}
	case *ast.SelectorExpr:
		// TODO: Type.Method and pkg.Type.Method
		n1 := name(x.X)
		n2 := x.Sel.Name
		switch obj := r.scope.LookupParent(n1).(type) {
		case *types.PkgName:
			return obj.Pkg.Scope().Lookup(n2)
		case *types.Var:
			t := obj.Type
			fm, _, addr := types.LookupFieldOrMethod(t, r.pkg, n2)
			switch fm := fm.(type) {
			case *types.Func:
				sig := fm.Type.(*types.Signature)
				return types.NewFunc(0, r.pkg, n2, types.NewSignature(nil, newVar("", t), sig.Params, sig.Results, sig.IsVariadic))
			case *types.Var:
				return field{fm, t, addr}
			}
		}
	}
	return nil
}

func (r *reader) typ(x ast.Expr) types.Type {
	t, _, err := types.EvalNode(r.fset, x, r.pkg, r.scope)
	if err != nil {
		panic(err)
	}
	return t
}

func (r *reader) out(x ast.Expr, out *port) {
	r.ports[name(x)] = out
}

func (r *reader) in(x ast.Expr, in *port) {
	name := name(x)
	for _, c := range r.conns[name] {
		if !c.connectable(c.src, in) {
			println("not connectable:")
			printf("%#v\n", c.src.node)
			printf("%#v\n\n", in.node)
		}
		c.setDst(in)
	}
	r.ports[name] = in //for feedback conns
}

func (r *reader) seq(n node, an ast.Node) {
	if c, ok := r.cmap[an]; ok {
		s := strings.Split(c[0].List[0].Text[2:], ";")
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
	switch x := x.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return name(x.X)
	}
	return ""
}
