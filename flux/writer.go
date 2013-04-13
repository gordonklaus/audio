package main

import (
	"code.google.com/p/go.exp/go/types"
	."fmt"
	"go/ast"
	"go/build"
	"os"
	"path/filepath"
	."strconv"
	."strings"
)

func savePackageName(p *types.Package) {
	// replace package name in all .go files
}

func saveType(t *types.NamedType) {
	w := newWriter(t.Obj)
	defer w.close()
	
	u := t.Underlying
	walkType(u, func(tt *types.NamedType) {
		if p := tt.Obj.Pkg; p != t.Obj.Pkg {
			w.pkgNames[p] = w.newName(p.Name)
		}
	})
	
	tStr := w.typeString(u)
	w.writeImports()
	
	Fprintf(w.src, "type %s %s", t.Obj.Name, tStr)
}

func saveFunc(f funcNode) {
	w := newWriter(f.obj)
	defer w.close()
	
	for p := range f.pkgRefs {
		w.pkgNames[p] = w.newName(p.Name)
	}
	
	w.writeImports()
	
	w.src.WriteString("func ")
	vars := map[*input]string{}
	params := []string{}
	for _, p := range f.inputsNode.outputs() {
		name := w.newName(p.obj.GetName())
		if len(p.conns) > 0 {
			vars[p.conns[0].dst] = name
		}
		params = append(params, name + " " + w.typeString(p.obj.GetType()))
	}
	if _, ok := f.obj.(method); ok {
		Fprintf(w.src, "(%s) ", params[0])
		params = params[1:]
	}
	results := []string{}
	for _, p := range f.outputsNode.inputs() {
		name := w.newName(p.obj.GetName())
		vars[p] = name
		results = append(results, name + " " + w.typeString(p.obj.GetType()))
	}
	Fprintf(w.src, "%s(%s) ", f.obj.GetName(), Join(params, ", "))
	if len(results) > 0 {
		Fprintf(w.src, "(%s) ", Join(results, ", "))
	}
	w.src.WriteString("{\n")
	w.writeBlockGo(f.funcblk, 0, vars)
	if len(results) > 0 {
		w.src.WriteString("\treturn\n")
	}
	w.src.WriteString("}")
}

type writer struct {
	src *os.File
	pkgNames map[*types.Package]string
	names map[string]int
	nodeID int
	nodeIDs map[node]int
}

func newWriter(obj types.Object) *writer {
	chk := func(err error) { if err != nil { panic(err) }}
	
	pkg := obj.GetPkg()
	bp, err := build.Import(pkg.Path, "", build.FindOnly)
	chk(err)
	dir := bp.Dir
	
	w := &writer{nil, map[*types.Package]string{}, map[string]int{}, 0, map[node]int{}}
	name := obj.GetName()
	if m, ok := obj.(method); ok {
		t := m.Type.Recv.Type
		if p, ok := t.(*types.Pointer); ok {
			t = p.Base
		}
		name = t.(*types.NamedType).Obj.Name + "." + name
	}
	w.src, err = os.Create(filepath.Join(dir, name + ".flux.go"))
	chk(err)
	
	for _, obj := range append(types.Universe.Entries, pkg.Scope.Entries...) {
		w.newName(obj.GetName())
	}
	
	Fprintf(w.src, "package %s\n\n", pkg.Name)
	
	return w
}

func (w *writer) writeImports() {
	if len(w.pkgNames) > 0 {
		w.src.WriteString("import (\n")
	}
	for p, n := range w.pkgNames {
		w.src.WriteString("\t")
		if n != p.Name {
			Fprintf(w.src, "%s ", n)
		}
		Fprintf(w.src, "\"%s\"\n", p.Path)
	}
	if len(w.pkgNames) > 0 {
		w.src.WriteString(")\n\n")
	}
}

func (w *writer) writeBlockGo(b *block, indent int, vars map[*input]string) {
	order, ok := b.nodeOrder()
	if !ok {
		Println("cyclic!")
		return
	}
	
	indent++
	vars, varsCopy := map[*input]string{}, vars
	for k, v := range varsCopy { vars[k] = v }

	for c := range b.conns {
		if _, ok := vars[c.dst]; !ok && c.src.node.block() != b {
			name := w.newName("v")
			Fprintf(w.src, "%svar %s %s\n", tabs(indent), name, w.typeString(c.dst.obj.GetType()))
			vars[c.dst] = name
		}
	}
	for _, n := range order {
		switch n := n.(type) {
		default:
			inputs := []string{}
			for _, input := range n.inputs() {
				name := ""
				if len(input.conns) > 0 {
					name = vars[input.conns[0].dst]
				} else {
					name = w.zeroLiteral(input.obj.GetType())
				}
				inputs = append(inputs, name)
			}
			out, assignExisting := w.outputNames(n, vars)
			if len(out) > 0 {
				out += " := "
			}
			Fprintf(w.src, "%s%s", tabs(indent), out)
			switch n := n.(type) {
			case *callNode:
				Fprintf(w.src, "%s(%s)", w.qualifiedName(n.obj), Join(inputs, ", "))
			case *indexNode:
				if n.set {
					Fprintf(w.src, "%s[%s] = %s", inputs[0], inputs[1], inputs[2])
				} else {
					Fprintf(w.src, "%s[%s]", inputs[0], inputs[1])
				}
			case *constantNode:
				w.src.WriteString(Quote(n.text.GetText()))
			}
			w.src.WriteString("\n")
			w.assignExisting(assignExisting, indent)
		case *portsNode:
		case *ifNode:
			cond := "false"
			if len(n.input.conns) > 0 {
				cond = vars[n.input]
			}
			Fprintf(w.src, "%sif %s {\n", tabs(indent), cond)
			w.writeBlockGo(n.trueblk, indent, vars)
			if len(n.falseblk.nodes) > 0 {
				Fprintf(w.src, "%s} else {\n", tabs(indent))
				w.writeBlockGo(n.falseblk, indent, vars)
			}
			Fprintf(w.src, "%s}\n", tabs(indent))
		case *loopNode:
			Fprintf(w.src, "%sfor ", tabs(indent))
			out, assignExisting := w.outputNames(n.inputsNode, vars)
			if conns := n.input.conns; len(conns) > 0 {
				switch conns[0].src.obj.GetType().(type) {
				case *types.NamedType:
					if out == "" {
						out = w.newName("")
					}
					Fprintf(w.src, "%s := 0; %s < %s; %s++ ", out, out, vars[n.input], out)
				case *types.Array, *types.Slice, *types.Map, *types.Chan:
					if len(out) > 0 {
						Fprintf(w.src, "%s := ", out)
					}
					Fprintf(w.src, "range %s ", vars[n.input])
				}
			} else if len(out) > 0 {
				Fprintf(w.src, "%s := 0;; %s++ ", out, out)
			}
			w.src.WriteString("{\n")
			w.assignExisting(assignExisting, indent + 1)
			w.writeBlockGo(n.loopblk, indent, vars)
			Fprintf(w.src, "%s}\n", tabs(indent))
		}
	}
}

func (w *writer) outputNames(n node, vars map[*input]string) (string, map[string]string) {
	names := []string{}
	any := false
	assignExisting := map[string]string{}
	for _, p := range n.outputs() {
		name := "_"
		if len(p.conns) > 0 {
			any = true
			name = w.newName(p.obj.GetName())
			for _, c := range p.conns {
				if existingName, ok := vars[c.dst]; ok {
					assignExisting[existingName] = name
				} else {
					vars[c.dst] = name
				}
			}
		}
		names = append(names, name)
	}
	if any {
		return Join(names, ", "), assignExisting
	}
	return "", nil
}

func (w writer) assignExisting(m map[string]string, indent int) {
	if len(m) > 0 {
		var existingNames, sourceNames []string
		for v1, v2 := range m {
			existingNames = append(existingNames, v1)
			sourceNames = append(sourceNames, v2)
		}
		Fprintf(w.src, "%s%s = %s\n", tabs(indent), Join(existingNames, ", "), Join(sourceNames, ", "))
	}
}

func (w writer) newName(s string) string {
	if s == "" || s == "_" { s = "x" }
	if i, ok := w.names[s]; ok {
		w.names[s]++
		return w.newName(s + Itoa(i))
	}
	w.names[s] = 2
	return s
}

func (w writer) qualifiedName(obj types.Object) string {
	s := ""
	if n, ok := w.pkgNames[obj.GetPkg()]; ok {
		s = n + "."
	}
	return s + obj.GetName()
}

func (w writer) typeString(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		return t.Name
	case *types.NamedType:
		return w.qualifiedName(t.Obj)
	case *types.Pointer:
		return "*" + w.typeString(t.Base)
	case *types.Array:
		return Sprintf("[%d]%s", t.Len, w.typeString(t.Elt))
	case *types.Slice:
		return "[]" + w.typeString(t.Elt)
	case *types.Map:
		return Sprintf("map[%s]%s", w.typeString(t.Key), w.typeString(t.Elt))
	case *types.Chan:
		s := ""
		switch t.Dir {
		case ast.SEND:
			s = "chan<- "
		case ast.RECV:
			s = "<-chan "
		default:
			s = "chan "
		}
		return s + w.typeString(t.Elt)
	case *types.Signature:
		return "func" + w.signature(t)
	case *types.Interface:
		s := "interface{"
		for i, m := range t.Methods {
			if i > 0 {
				s += "; "
			}
			s += m.Name + w.signature(m.Type)
		}
		return s + "}"
	case *types.Struct:
		s := "struct{"
		for i, f := range t.Fields {
			if i > 0 {
				s += "; "
			}
			if f.Name != "" {
				s += f.Name + " "
			}
			s += w.typeString(f.Type)
		}
		return s + "}"
	}
	panic(Sprintf("no string for type %#v\n", t))
	return ""
}

func (w writer) signature(f *types.Signature) string {
	s := w.paramsString(f.Params)
	if len(f.Results) > 0 {
		s += " "
		if len(f.Results) == 1 && f.Results[0].Name == "" {
			return s + w.typeString(f.Results[0].Type)
		}
		return s + w.paramsString(f.Results)
	}
	return s
}

func (w writer) paramsString(params []*types.Var) string {
	s := "("
	for i, p := range params {
		if i > 0 {
			s += ", "
		}
		name := p.Name
		if name == "" { name = "_" }
		s += name + " "
		s += w.typeString(p.Type)
	}
	return s + ")"
}

func (w writer) zeroLiteral(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind {
		case types.Bool:
			return "false"
		case types.String:
			return `""`
		case types.UnsafePointer:
			return "nil"
		}
		return "0"
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return "nil"
	case *types.Array, *types.Struct:
		return w.typeString(t) + "{}"
	case *types.NamedType:
		switch t.Underlying.(type) {
		case *types.Array, *types.Struct:
			return w.typeString(t) + "{}"
		}
		return w.zeroLiteral(t.Underlying)
	}
	panic(Sprintf("no zero literal for type %#v\n", t))
	return ""
}

func (w *writer) close() {
	w.src.Close()
}

func tabs(n int) string { return Repeat("\t", n) }

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
		panic(Sprintf("unexpected type %#v\n", t))
	}
}
