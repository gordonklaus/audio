package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."io/ioutil"
	."strings"
	."strconv"
)

type reader struct {
	s string
	f *funcNode
	pkgNames map[string]*PackageInfo
	nodes map[int]node
}

func loadFunc(f *funcNode) bool {
	r := &reader{"", f, map[string]*PackageInfo{}, map[int]node{}}
	if b, err := ReadFile(f.info.FluxSourcePath()); err != nil {
		return false
	} else {
		r.s = string(b)
	}
	
	line := ""
	line, r.s = Split2(r.s, "\n")
	for r.s[0] != '\\' {
		line, r.s = Split2(r.s, "\n")
		importPath, name := Split2(line, " ")
		pkg := findPackageInfo(importPath)
		if name == "" {
			name = pkg.name
		}
		r.pkgNames[name] = pkg
	}
	for _, v := range f.info.typ.parameters {
		f.inputsNode.newOutput(v)
		f.addPackageRef(v.typ)
	}
	for _, v := range f.info.typ.results {
		f.outputsNode.newInput(v)
		f.addPackageRef(v.typ)
	}
	r.readBlock(f.funcblk, 0)
	
	return true
}

func (r *reader) readBlock(b *block, indent int) {
	_, r.s = Split2(r.s, "\n")
	indent++
	for len(r.s) > 0 {
		i := 0
		for r.s[i] == '\t' { i++ }
		if i < indent { return }
		if r.s[i] == '-' {
			var line string
			line, r.s = Split2(r.s, "\n")
			x := Fields(line)
			srcID, _ := Atoi(x[1])
			srcPort, _ := Atoi(x[2])
			dstID, _ := Atoi(x[3])
			dstPort, _ := Atoi(x[4])
			c := newConnection(b, ZP)
			c.setSource(r.nodes[srcID].outputs()[srcPort])
			c.setDestination(r.nodes[dstID].inputs()[dstPort])
		} else {
			r.readNode(b, indent)
		}
	}
}

func (r *reader) readNode(b *block, indent int) {
	var node node
	line := ""
	line, r.s = Split2(r.s, "\n")
	fields := Fields(line)
	switch f := fields[1]; f {
 	case "\\in":
		for n := range b.nodes {
			if n, ok := n.(*portsNode); ok && !n.out {
				node = n
			}
		}
 	case "\\out":
		for n := range b.nodes {
			if n, ok := n.(*portsNode); ok && n.out {
				node = n
			}
		}
	case "if":
		n := newIfNode(b)
		r.readBlock(n.trueblk, indent)
		r.readBlock(n.falseblk, indent)
		node = n
	case "loop":
		n := newLoopNode(b)
		r.readBlock(n.loopblk, indent)
		node = n
	default:
		if f[0] == '"' {
			n := newStringConstantNode(b)
			text, _ := Unquote(fields[1])
			n.text.SetText(text)
			node = n
		} else {
			pkgName, name := Split2(fields[1], ".")
			var pkg *PackageInfo
			if name == "" {
				name = pkgName
				pkg = r.f.pkg()
			} else {
				pkg = r.pkgNames[pkgName]
			}
			for _, info := range pkg.Children() {
				if info.Name() != name { continue }
				switch info := info.(type) {
				case *FuncInfo:
					node = newCallNode(info, b)
				default:
					panic("not yet implemented")
				}
			}
		}
	}
	id, _ := Atoi(fields[0])
	r.nodes[id] = node
	b.addNode(node)
}
