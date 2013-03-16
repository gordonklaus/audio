package main

import (
	."fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	."strconv"
	."strings"
)

func savePackageName(p *Package) {
	path := filepath.Join(FluxSourcePath(p), "package.flux")
	if p.pkgName == p.name {
		os.Remove(path)
	} else {
		ioutil.WriteFile(path, []byte(p.pkgName), 0777)
	}
}

func saveType(t *NamedType) {
	w := newWriter(t)
	defer w.close()
	
	u := t.underlying
	walkType(u, func(tt *NamedType) {
		if p := tt.parent.(*Package); p != t.parent.(*Package) && p != builtinPkg {
			w.pkgNames[p] = w.newName(p.pkgName)
		}
	})
	
	tStr := w.typeString(u)
	w.flux.WriteString(tStr)
	w.writeImports()
	
	Fprintf(w.go_, "type %s %s", t.name, tStr)
}

func saveFunc(f funcNode) {
	w := newWriter(f.info)
	defer w.close()
	
	for p := range f.pkgRefs {
		w.pkgNames[p] = w.newName(p.pkgName)
	}
	
	w.flux.WriteString(w.typeString(f.info.typeWithReceiver()))
	w.writeImports()
	w.writeBlockFlux(f.funcblk, 0)
	
	w.go_.WriteString("func ")
	vars := map[*input]string{}
	params := []string{}
	for _, p := range f.inputsNode.outputs() {
		name := w.newName(p.val.name)
		if len(p.conns) > 0 {
			vars[p.conns[0].dst] = name
		}
		params = append(params, name + " " + w.typeString(p.val.typ))
	}
	if f.info.receiver != nil {
		Fprintf(w.go_, "(%s) ", params[0])
		params = params[1:]
	}
	results := []string{}
	for _, p := range f.outputsNode.inputs() {
		name := w.newName(p.val.name)
		vars[p] = name
		results = append(results, name + " " + w.typeString(p.val.typ))
	}
	Fprintf(w.go_, "%s(%s) ", f.info.Name(), Join(params, ", "))
	if len(results) > 0 {
		Fprintf(w.go_, "(%s) ", Join(results, ", "))
	}
	w.go_.WriteString("{\n")
	w.writeBlockGo(f.funcblk, 0, vars)
	if len(results) > 0 {
		w.go_.WriteString("\treturn\n")
	}
	w.go_.WriteString("}")
}

type writer struct {
	flux, go_ *os.File
	pkgNames map[*Package]string
	names map[string]int
	nodeID int
	nodeIDs map[node]int
}

func newWriter(info Info) *writer {
	chk := func(err error) { if err != nil { panic(err) }}
	
	var typ *NamedType
	i := info.Parent()
	if t, ok := i.(*NamedType); ok { typ, i = t, t.parent }
	pkg := i.(*Package)
	
	var err error
	w := &writer{nil, nil, map[*Package]string{}, map[string]int{}, 0, map[node]int{}}
	fluxPath := FluxSourcePath(info)
	chk(os.MkdirAll(filepath.Dir(fluxPath), 0777))
	w.flux, err = os.Create(fluxPath); chk(err)
	name := info.Name(); if typ != nil { name = typ.Name() + "." + name }
	w.go_, err = os.Create(Sprintf("%s/%s.flux.go", FluxSourcePath(pkg), name)); chk(err)
	
	for _, i := range append(Children(builtinPkg), Children(info.Parent())...) {
		if _, ok := i.(*Package); !ok {
			w.newName(i.Name())
		}
	}
	
	Fprintf(w.go_, "package %s\n\n", pkg.pkgName)
	
	return w
}

func (w *writer) writeImports() {
	if len(w.pkgNames) > 0 {
		w.go_.WriteString("import (\n")
	}
	for p, n := range w.pkgNames {
		Fprintf(w.flux, "\n%s", p.importPath)
		w.go_.WriteString("\t")
		if n != p.pkgName {
			Fprintf(w.flux, " %s", n)
			Fprintf(w.go_, "%s ", n)
		}
		Fprintf(w.go_, "\"%s\"\n", p.importPath)
	}
	if len(w.pkgNames) > 0 {
		w.go_.WriteString(")\n\n")
	}
}

func (w *writer) writeBlockFlux(b *block, indent int) {
	order, ok := b.nodeOrder()
	if !ok {
		Println("cyclic!")
		return
	}
	
	Fprintf(w.flux, "\n%s\\", tabs(indent))
	indent++
	for _, n := range order {
		w.nodeID++
		w.nodeIDs[n] = w.nodeID
		Fprintf(w.flux, "\n%s%d ", tabs(indent), w.nodeID)
		switch n := n.(type) {
		case *callNode:
			w.flux.WriteString(w.qualifiedName(n.info))
		case *constantNode:
			w.flux.WriteString(Quote(n.text.GetText()))
		case *compositeLiteralNode:
			w.flux.WriteString(w.typeString(*n.typ.typ))
		case *indexNode:
			if n.set {
				w.flux.WriteString("[]=")
			} else {
				w.flux.WriteString("[]")
			}
		case *ifNode:
			w.flux.WriteString("if")
		case *loopNode:
			w.flux.WriteString("loop")
		case *portsNode:
			if n.out {
				w.flux.WriteString("\\out")
			} else {
				w.flux.WriteString("\\in")
			}
		}
		for iDst, p := range n.inputs() {
			for _, c := range p.conns {
				iSrc := -1
				for i, src := range c.src.node.outputs() {
					if src == c.src {
						iSrc = i
						break
					}
				}
				Fprintf(w.flux, " %d.%d-%d", w.nodeIDs[c.src.node], iSrc, iDst)
			}
		}
		switch n := n.(type) {
		case *ifNode:
			w.writeBlockFlux(n.trueblk, indent)
			w.writeBlockFlux(n.falseblk, indent)
		case *loopNode:
			w.writeBlockFlux(n.loopblk, indent)
		}
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
			Fprintf(w.go_, "%svar %s %s\n", tabs(indent), name, w.typeString(c.dst.val.typ))
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
					name = w.zeroLiteral(input.val.typ)
				}
				inputs = append(inputs, name)
			}
			out, assignExisting := w.outputNames(n, vars)
			if len(out) > 0 {
				out += " := "
			}
			Fprintf(w.go_, "%s%s", tabs(indent), out)
			switch n := n.(type) {
			case *callNode:
				Fprintf(w.go_, "%s(%s)", w.qualifiedName(n.info), Join(inputs, ", "))
			case *indexNode:
				if n.set {
					Fprintf(w.go_, "%s[%s] = %s", inputs[0], inputs[1], inputs[2])
				} else {
					Fprintf(w.go_, "%s[%s]", inputs[0], inputs[1])
				}
			case *constantNode:
				w.go_.WriteString(Quote(n.text.GetText()))
			}
			w.go_.WriteString("\n")
			w.assignExisting(assignExisting, indent)
		case *portsNode:
		case *ifNode:
			cond := "false"
			if len(n.input.conns) > 0 {
				cond = vars[n.input]
			}
			Fprintf(w.go_, "%sif %s {\n", tabs(indent), cond)
			w.writeBlockGo(n.trueblk, indent, vars)
			if len(n.falseblk.nodes) > 0 {
				Fprintf(w.go_, "%s} else {\n", tabs(indent))
				w.writeBlockGo(n.falseblk, indent, vars)
			}
			Fprintf(w.go_, "%s}\n", tabs(indent))
		case *loopNode:
			Fprintf(w.go_, "%sfor ", tabs(indent))
			out, assignExisting := w.outputNames(n.inputsNode, vars)
			if conns := n.input.conns; len(conns) > 0 {
				switch conns[0].src.val.typ.(type) {
				case *NamedType:
					if out == "" {
						out = w.newName("")
					}
					Fprintf(w.go_, "%s := 0; %s < %s; %s++ ", out, out, vars[n.input], out)
				case *ArrayType, *SliceType, *MapType, *ChanType:
					if len(out) > 0 {
						Fprintf(w.go_, "%s := ", out)
					}
					Fprintf(w.go_, "range %s ", vars[n.input])
				}
			} else if len(out) > 0 {
				Fprintf(w.go_, "%s := 0;; %s++ ", out, out)
			}
			w.go_.WriteString("{\n")
			w.assignExisting(assignExisting, indent + 1)
			w.writeBlockGo(n.loopblk, indent, vars)
			Fprintf(w.go_, "%s}\n", tabs(indent))
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
			name = w.newName(p.val.name)
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
		Fprintf(w.go_, "%s%s = %s\n", tabs(indent), Join(existingNames, ", "), Join(sourceNames, ", "))
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

func (w writer) qualifiedName(i Info) string {
	s := ""
	if n, ok := w.pkgNames[i.Parent().(*Package)]; ok {
		s = n + "."
	}
	return s + i.Name()
}

func (w writer) typeString(t Type) string {
	switch t := t.(type) {
	case *PointerType:
		return "*" + w.typeString(t.element)
	case *ArrayType:
		return Sprintf("[%d]%s", t.size, w.typeString(t.element))
	case *SliceType:
		return "[]" + w.typeString(t.element)
	case *MapType:
		return Sprintf("map[%s]%s", w.typeString(t.key), w.typeString(t.value))
	case *ChanType:
		s := ""
		switch {
		case t.send && t.recv:
			s = "chan "
		case t.send:
			s = "chan<- "
		case t.recv:
			s = "<-chan "
		}
		return s + w.typeString(t.element)
	case *FuncType:
		return "func" + w.signature(t)
	case *InterfaceType:
		s := "interface{"
		for i, m := range t.methods {
			if i > 0 {
				s += "; "
			}
			s += m.name + w.signature(m.typ.(*FuncType))
		}
		return s + "}"
	case *StructType:
		s := "struct{"
		for i, f := range t.fields {
			if i > 0 {
				s += "; "
			}
			if f.name != "" {
				s += f.name + " "
			}
			s += w.typeString(f.typ)
		}
		return s + "}"
	case *NamedType:
		return w.qualifiedName(t)
	}
	panic(Sprintf("no string for type %#v\n", t))
	return ""
}

func (w writer) signature(f *FuncType) string {
	s := w.paramsString(f.parameters)
	if len(f.results) > 0 {
		s += " "
		if len(f.results) == 1 && f.results[0].name == "" {
			return s + w.typeString(f.results[0].typ)
		}
		return s + w.paramsString(f.results)
	}
	return s
}

func (w writer) paramsString(params []*Value) string {
	s := "("
	for i, p := range params {
		if i > 0 {
			s += ", "
		}
		name := p.name
		if name == "" { name = "_" }
		s += name + " "
		s += w.typeString(p.typ)
	}
	return s + ")"
}

func (w writer) zeroLiteral(t Type) string {
	switch t := t.(type) {
	case *BasicType:
		switch t.reflectType.Kind() {
		case reflect.Bool:
			return "false"
		case reflect.String:
			return `""`
		case reflect.UnsafePointer:
			return "nil"
		}
		return "0"
	case *PointerType, *SliceType, *MapType, *ChanType, *FuncType, *InterfaceType:
		return "nil"
	case *ArrayType, *StructType:
		return w.typeString(t) + "{}"
	case *NamedType:
		switch t.underlying.(type) {
		case *ArrayType, *StructType:
			return w.typeString(t) + "{}"
		}
		return w.zeroLiteral(t.underlying)
	}
	panic(Sprintf("no zero literal for type %#v\n", t))
	return ""
}

func (w *writer) close() {
	w.flux.Close()
	w.go_.Close()
}

func tabs(n int) string { return Repeat("\t", n) }
