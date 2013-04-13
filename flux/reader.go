package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	"code.google.com/p/go.exp/go/types"
	// ."io/ioutil"
	"strings"
	."strconv"
)

type reader struct {
	s string
	f *funcNode
	pkgNames map[string]*types.Package
	nodes map[int]node
}

func loadFunc(f *funcNode) bool {
	// r := &reader{"", f, map[string]*types.Package{}, map[int]node{}}
	// if b, err := ReadFile(FluxSourcePath(f.obj)); err != nil {
	// 	return false
	// } else {
	// 	r.s = string(b)
	// }
	// 
	// line := ""
	// line, r.s = Split2(r.s, "\n")
	// for r.s[0] != '\\' {
	// 	line, r.s = Split2(r.s, "\n")
	// 	importPath, name := Split2(line, " ")
	// 	pkg := FindPackage(importPath)
	// 	if name == "" {
	// 		name = pkg.pkgName
	// 	}
	// 	r.pkgNames[name] = pkg
	// }
	// for _, v := range f.obj.Params {
	// 	f.inputsNode.newOutput(v)
	// 	f.addPkgRef(v.obj)
	// }
	// for _, v := range f.obj.Results {
	// 	f.outputsNode.newInput(v)
	// 	f.addPkgRef(v.obj)
	// }
	// r.readBlock(f.funcblk, 0)
	
	return false
}

func (r *reader) readBlock(b *block, indent int) {
	_, r.s = Split2(r.s, "\n")
	indent++
	for len(r.s) > 0 {
		i := 0
		for r.s[i] == '\t' { i++ }
		if i < indent { return }
		r.readNode(b, indent)
	}
}

func (r *reader) readNode(b *block, indent int) {
	var node node
	line := ""
	line, r.s = Split2(r.s, "\n")
	fields := strings.Fields(line)
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
	case "[]":
		node = newIndexNode(b, false)
	case "[]=":
		node = newIndexNode(b, true)
	case "if":
		n := newIfNode(b)
		node = n
	case "loop":
		n := newLoopNode(b)
		node = n
	default:
		if f[0] == '"' {
			n := newStringConstantNode(b)
			text, _ := Unquote(fields[1])
			n.text.SetText(text)
			node = n
		} else {
			pkgName, name := Split2(fields[1], ".")
			var pkg *types.Package
			if name == "" {
				name = pkgName
				pkg = r.f.pkg()
			} else {
				pkg = r.pkgNames[pkgName]
			}
			for _, obj := range pkg.Scope.Entries {
				if obj.GetName() != name { continue }
				switch obj := obj.(type) {
				case *types.Func:
					node = newCallNode(obj, b)
				default:
					panic("not yet implemented")
				}
			}
		}
	}
	id, _ := Atoi(fields[0])
	r.nodes[id] = node
	b.addNode(node)
	for _, f := range fields[2:] {
		i := strings.Index(f, ".")
		j := strings.Index(f, "-")
		srcID, _ := Atoi(f[:i])
		srcPort, _ := Atoi(f[i+1:j])
		dstPort, _ := Atoi(f[j+1:])
		c := newConnection(b, ZP)
		c.setSrc(r.nodes[srcID].outputs()[srcPort])
		c.setDst(node.inputs()[dstPort])
	}
	switch n := node.(type) {
	case *ifNode:
		r.readBlock(n.trueblk, indent)
		r.readBlock(n.falseblk, indent)
	case *loopNode:
		r.readBlock(n.loopblk, indent)
	}
}
