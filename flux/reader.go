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
	f *FuncNode
	pkgNames map[string]*PackageInfo
	nodes map[int]Node
}

func loadFunc(f *FuncNode) bool {
	r := &reader{"", f, map[string]*PackageInfo{}, map[int]Node{}}
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
	for _, parameter := range f.info.typ.parameters {
		p := NewOutput(f.inputNode, parameter)
		f.inputNode.AddChild(p)
		f.inputNode.outputs = append(f.inputNode.outputs, p)
		f.AddPackageRef(parameter.typ)
	}
	f.inputNode.reform()
	r.readBlock(f.funcBlock, 0)
	
	return true
}

func (r *reader) readBlock(b *Block, indent int) {
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
			srcNodeID, _ := Atoi(x[1])
			iSrcPut, _ := Atoi(x[2])
			dstNodeID, _ := Atoi(x[3])
			iDstPut, _ := Atoi(x[4])
			conn := b.NewConnection(ZP)
			r.nodes[srcNodeID].Outputs()[iSrcPut].ConnectTo(conn)
			r.nodes[dstNodeID].Inputs()[iDstPut].ConnectTo(conn)
		} else {
			r.readNode(b, indent)
		}
	}
}

func (r *reader) readNode(b *Block, indent int) {
	var node Node
	line := ""
	line, r.s = Split2(r.s, "\n")
	fields := Fields(line)
	switch f := fields[1]; f {
 	case "\\in":
		for n := range b.nodes {
			if _, ok := n.(*InOutNode); ok {
				node = n
			}
		}
	case "if":
		n := NewIfNode(b)
		r.readBlock(n.trueBlock, indent)
		r.readBlock(n.falseBlock, indent)
		node = n
	case "loop":
		n := NewLoopNode(b)
		r.readBlock(n.loopBlock, indent)
		node = n
	default:
		if f[0] == '"' {
			strNode := NewStringConstantNode(b)
			text, _ := Unquote(fields[1])
			strNode.text.SetText(text)
			node = strNode
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
					node = NewCallNode(info, b)
				default:
					panic("not yet implemented")
				}
			}
		}
	}
	nodeID, _ := Atoi(fields[0])
	r.nodes[nodeID] = node
	b.AddNode(node)
}
