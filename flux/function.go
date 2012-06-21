package main

import (
	."io/ioutil"
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."fmt"
	."strconv"
	."strings"
)

type Function struct {
	*ViewBase
	block *Block
}

func NewFunction() *Function {
	f := &Function{}
	f.ViewBase = NewView(f)
	f.block = NewBlock(nil)
	f.AddChild(f.block)
	return f
}

func nodeOrder(nodes map[Node]bool) (order []Node, ok bool) {
	visited := Set{}
	var insertInOrder func(node Node, visitedThisCall Set) bool
	insertInOrder = func(node Node, visitedThisCall Set) bool {
		if visitedThisCall[node] { return false }
		visitedThisCall[node] = true
		
		if !visited[node] {
			visited[node] = true
			switch node := node.(type) {
			default:
				for _, input := range node.Inputs() {
					for _, conn := range input.connections {
						// if block.connections[conn] {
							srcNode := conn.src.node
						// 	for !block.nodes[srcNode] { srcNode = srcNode.block.node }
							if !insertInOrder(srcNode, visitedThisCall.Copy()) { return false }
						// } else {
						// 	blockInputs[conn] = true
						// }
					}
				}
			case *IfNode:
				// order, falseInputs := node.falseBlock.nodeOrder()
				// order, trueInputs := node.trueBlock.nodeOrder()
				// for _, input := range append(falseInputs, node.input, trueInputs...) {
				// 	// same as above
				// }
			}
			order = append(order, node)
		}
		return true
	}
	
	endNodes := []Node{}
	for node := range nodes {
		switch node := node.(type) {
		default:
			for _, output := range node.Outputs() {
				if len(output.connections) > 0 { continue }
			}
		case *IfNode:
			if node.falseBlock.HasOutputConnections() || node.trueBlock.HasOutputConnections() { continue }
		}
		endNodes = append(endNodes, node)
	}
	if len(endNodes) == 0 && len(nodes) > 0 { return }
	
	for _, node := range endNodes {
		if !insertInOrder(node, Set{}) { return }
	}
	ok = true
	return
}

func (f Function) Save() {
	order, ok := nodeOrder(f.block.nodes)
	if !ok {
		Println("cyclic!")
		return
	}
	s := "package main\n\nimport (\n"
	s += "\t.\"fmt\"\n"
	s += ")\n\nfunc main() {\n"
	i := 0; newName := func() string { i++; return "v" + Itoa(i) }
	vars := map[*Output]string{}
	for _, node := range order {
		inputs := []string{}
		for _, input := range node.Inputs() {
			name := ""
			switch len(input.connections) {
			case 0:
				// *new(typeName) instead?
				name = newName()
				s += Sprintf("\tvar %v %v\n", name, input.info.typeName)
			case 1:
				name = vars[input.connections[0].src]
			default:
				panic("more than one input connection not yet supported")
			}
			inputs = append(inputs, name)
		}
		outputs := []string{}
		anyOutputConnections := false
		for _, output := range node.Outputs() {
			name := "_"
			if len(output.connections) > 0 {
				anyOutputConnections = true
				name = newName()
				vars[output] = name
			}
			outputs = append(outputs, name)
		}
		assignment := ""
		if anyOutputConnections {
			assignment = Join(outputs, ", ") + " := "
		}
		s += Sprintf("\t%v%v\n", assignment, node.GoCode(Join(inputs, ", ")))
	}
	s += "}"
	WriteFile("../main.go", []byte(s), 0644)
}

func (f *Function) Resize(w, h float64) {
	f.ViewBase.Resize(w, h)
	f.block.Resize(w, h)
}

func (f *Function) TookKeyboardFocus() { f.block.TakeKeyboardFocus() }

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		f.Save()
	default:
		f.ViewBase.KeyPressed(event)
	}
}

func (Function) Paint() {}
