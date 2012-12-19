package main

import (
	."fmt"
	"path/filepath"
	"os"
	"reflect"
	."strconv"
	."strings"
)

func saveType(t *NamedType) {
	w := newWriter(t)
	defer w.close()
	
	u := t.underlying
	walkType(u, func(tt *NamedType) {
		if p := tt.parent.(*PackageInfo); p != t.parent.(*PackageInfo) && p != builtinPkg {
			w.pkgNames[p] = w.newName(p.Name())
		}
	})
	
	tStr := w.typeString(u)
	w.flux.WriteString("type " + tStr)
	w.writeImports()
	
	Fprintf(w.go_, "type %s %s", t.name, tStr)
}

func saveFunc(f FuncNode) {
	w := newWriter(f.info)
	defer w.close()
	
	for p := range f.pkgRefs {
		w.pkgNames[p] = w.newName(p.Name())
	}
	
	w.flux.WriteString(w.typeString(f.info.typeWithReceiver()))
	w.writeImports()
	w.writeBlockFlux(f.funcBlock, 0)
	
	w.go_.WriteString("func ")
	vars := map[*Input]string{}
	params := []string{}
	for _, output := range f.inputNode.Outputs() {
		name := ""
		if len(output.connections) > 0 {
			name = w.newName(output.info.name)
			vars[output.connections[0].dst] = name
		} else {
			name = "_"
		}
		params = append(params, name + " " + w.typeString(output.info.typ))
	}
	if f.info.receiver != nil {
		Fprintf(w.go_, "(%s) ", params[0])
		params = params[1:]
	}
	Fprintf(w.go_, "%s(%s) {\n", f.info.Name(), Join(params, ", "))
	w.writeBlockGo(f.funcBlock, 0, vars)
	w.go_.WriteString("}")
}

type writer struct {
	flux, go_ *os.File
	pkgNames map[*PackageInfo]string
	names map[string]int
	nodeID int
	nodeIDs map[Node]int
}

func newWriter(info Info) *writer {
	chk := func(err error) { if err != nil { panic(err) }}
	
	parent := info.Parent()
	i := parent
	if t, ok := i.(*NamedType); ok { i = t.parent }
	pkg := i.(*PackageInfo)
	
	var err error
	w := &writer{nil, nil, map[*PackageInfo]string{}, map[string]int{}, 0, map[Node]int{}}
	fluxPath := info.FluxSourcePath()
	chk(os.MkdirAll(filepath.Dir(fluxPath), 0777))
	w.flux, err = os.Create(fluxPath); chk(err)
	w.go_, err = os.Create(Sprintf("%s/%s.go", parent.FluxSourcePath(), info.Name())); chk(err)
	
	for _, i := range append(builtinPkg.Children(), info.Parent().Children()...) {
		if _, ok := i.(*PackageInfo); !ok {
			w.newName(i.Name())
		}
	}
	
	Fprintf(w.go_, "package %s\n\n", pkg.Name())
	
	return w
}

func (w *writer) writeImports() {
	if len(w.pkgNames) > 0 {
		w.go_.WriteString("import (\n")
	}
	for p, n := range w.pkgNames {
		Fprintf(w.flux, "\n%s", p.importPath)
		w.go_.WriteString("\t")
		if n != p.name {
			Fprintf(w.flux, " %s", n)
			Fprintf(w.go_, "%s ", n)
		}
		Fprintf(w.go_, "\"%s\"\n", p.importPath)
	}
	if len(w.pkgNames) > 0 {
		w.go_.WriteString(")\n\n")
	}
}

func (w *writer) writeBlockFlux(b *Block, indent int) {
	order, ok := b.nodeOrder()
	if !ok {
		Println("cyclic!")
		return
	}
	
	Fprintf(w.flux, "\n%s\\", tabs(indent))
	indent++
	for _, node := range order {
		w.nodeID++
		w.nodeIDs[node] = w.nodeID
		Fprintf(w.flux, "\n%s%d ", tabs(indent), w.nodeIDs[node])
		switch n := node.(type) {
		case *CallNode:
			w.flux.WriteString(w.qualifiedName(n.info))
		case *ConstantNode:
			w.flux.WriteString(Quote(n.text.GetText()))
		case *IfNode:
			w.flux.WriteString("if")
			w.writeBlockFlux(n.trueBlock, indent)
			w.writeBlockFlux(n.falseBlock, indent)
		case *LoopNode:
			w.flux.WriteString("loop")
			w.writeBlockFlux(n.loopBlock, indent)
		case *InOutNode:
			w.flux.WriteString("\\in")
		}
	}
	for conn := range b.connections {
		iSrc := -1; for i, src := range conn.src.node.Outputs() { if src == conn.src { iSrc = i; break } }
		iDst := -1; for i, dst := range conn.dst.node.Inputs() { if dst == conn.dst { iDst = i; break } }
		Fprintf(w.flux, "\n%s- %d %d %d %d", tabs(indent), w.nodeIDs[conn.src.node], iSrc, w.nodeIDs[conn.dst.node], iDst)
	}
}

func (w *writer) writeBlockGo(b *Block, indent int, vars map[*Input]string) {
	order, ok := b.nodeOrder()
	if !ok {
		Println("cyclic!")
		return
	}
	
	indent++
	vars, varsCopy := map[*Input]string{}, vars
	for k, v := range varsCopy { vars[k] = v }

	for conn := range b.connections {
		if _, ok := vars[conn.dst]; !ok && conn.src.node.Block() != b {
			name := w.newName("v")
			Fprintf(w.go_, "%svar %s %s\n", tabs(indent), name, w.typeString(conn.dst.info.typ))
			vars[conn.dst] = name
		}
	}
	for _, node := range order {
		switch node := node.(type) {
		default:
			inputs := []string{}
			for _, input := range node.Inputs() {
				name := ""
				if len(input.connections) > 0 {
					name = vars[input.connections[0].dst]
				} else {
					name = w.zeroLiteral(input.info.typ)
				}
				inputs = append(inputs, name)
			}
			outputs, anyOutputConnections, assignExisting := w.outputNames(node, vars)
			assignment := ""
			if anyOutputConnections {
				assignment = Join(outputs, ", ") + " := "
			}
			w.go_.WriteString(tabs(indent))
			w.go_.WriteString(assignment)
			switch n := node.(type) {
			case *CallNode:
				Fprintf(w.go_, "%s(%s)", w.qualifiedName(n.info), Join(inputs, ", "))
			case *ConstantNode:
				w.go_.WriteString(Quote(n.text.GetText()))
			}
			w.go_.WriteString("\n")
			w.assignExisting(assignExisting, indent)
		case *InOutNode:
		case *IfNode:
			cond := "false"
			if len(node.input.connections) > 0 {
				cond = vars[node.input]
			}
			Fprintf(w.go_, "%sif %s {\n", tabs(indent), cond)
			w.writeBlockGo(node.trueBlock, indent, vars)
			if len(node.falseBlock.nodes) > 0 {
				Fprintf(w.go_, "%s} else {\n", tabs(indent))
				w.writeBlockGo(node.falseBlock, indent, vars)
			}
			Fprintf(w.go_, "%s}\n", tabs(indent))
		case *LoopNode:
			Fprintf(w.go_, "%sfor ", tabs(indent))
			outputs, anyOutputConnections, assignExisting := w.outputNames(node.inputNode, vars)
			if conns := node.input.connections; len(conns) > 0 {
				switch conns[0].src.info.typ.(type) {
				case *NamedType:
					i := outputs[0]
					Fprintf(w.go_, "%s := 0; %s < %s; %s++ ", i, i, vars[node.input], i)
				case *ArrayType, *SliceType, *MapType, *ChanType:
					if anyOutputConnections {
						Fprintf(w.go_, "%s := ", Join(outputs, ", "))
					}
					Fprintf(w.go_, "range %s ", vars[node.input])
				}
			} else if anyOutputConnections {
				i := outputs[0]
				Fprintf(w.go_, "%s := 0;; %s++ ", i, i)
			}
			w.go_.WriteString("{\n")
			w.assignExisting(assignExisting, indent + 1)
			w.writeBlockGo(node.loopBlock, indent, vars)
			Fprintf(w.go_, "%s}\n", tabs(indent))
		}
	}
}

func (w *writer) outputNames(node Node, vars map[*Input]string) ([]string, bool, map[string]string) {
	outputs := []string{}
	anyOutputConnections := false
	assignExisting := map[string]string{}
	for _, output := range node.Outputs() {
		name := "_"
		if len(output.connections) > 0 {
			anyOutputConnections = true
			name = w.newName(output.info.name)
			for _, conn := range output.connections {
				if existingName, ok := vars[conn.dst]; ok {
					assignExisting[existingName] = name
				} else {
					vars[conn.dst] = name
				}
			}
		}
		outputs = append(outputs, name)
	}
	return outputs, anyOutputConnections, assignExisting
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
	if n, ok := w.pkgNames[i.Parent().(*PackageInfo)]; ok {
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

func (w writer) paramsString(params []*ValueInfo) string {
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
