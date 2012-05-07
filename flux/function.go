package main

import (
	."io/ioutil"
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."fmt"
	."strconv"
	."strings"
)

type Function struct {
	ViewBase
	AggregateMouseHandler
	nodes []*Node
	connections []*Connection
}

func NewFunction() *Function {
	f := &Function{}
	f.ViewBase = *NewView(f)
	f.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(f), NewViewPanner(f)}
	return f
}

func (f *Function) AddNode(node *Node) {
	node.function = f
	f.AddChild(node)
	f.nodes = append(f.nodes, node)
}

func (f *Function) NewConnection(pt Point) *Connection {
	conn := NewConnection(f, pt)
	f.AddChild(conn)
	conn.Lower()
	f.connections = append(f.connections, conn)
	return conn
}

func (f *Function) DeleteConnection(connection *Connection) {
	for i, conn := range f.connections {
		if conn == connection {
			f.connections = append(f.connections[:i], f.connections[i+1:]...)
			f.RemoveChild(connection)
			connection.Disconnect()
			return
		}
	}
}

func (f Function) nodeOrder() (order []*Node, ok bool) {
	visited := Set{}
	var insertInOrder func(node *Node, visitedThisCall Set) bool
	insertInOrder = func(node *Node, visitedThisCall Set) bool {
		if visitedThisCall[node] { return false }
		visitedThisCall[node] = true
		
		if !visited[node] {
			visited[node] = true
			for _, input := range node.inputs {
				for _, conn := range input.connections {
					if !insertInOrder(conn.src.node, visitedThisCall.Copy()) { return false }
				}
			}
			order = append(order, node)
		}
		return true
	}
	
	endNodes := []*Node{}
	for _, node := range f.nodes {
		for _, output := range node.outputs {
			if len(output.connections) > 0 { continue }
		}
		endNodes = append(endNodes, node)
	}
	if len(endNodes) == 0 && len(f.nodes) > 0 { return }
	
	for _, node := range endNodes {
		if !insertInOrder(node, Set{}) { return }
	}
	ok = true
	return
}

func (f Function) Save() {
	order, ok := f.nodeOrder()
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
		inputNames := []string{}
		for _, input := range node.inputs {
			name := ""
			switch len(input.connections) {
			case 0:
				name = newName()
				s += Sprintf("\tvar %v %v\n", name, input.info.typeName)
			case 1:
				name = vars[input.connections[0].src]
			default:
				panic("more than one input connection not yet supported")
			}
			inputNames = append(inputNames, name)
		}
		outputNames := []string{}
		anyConnections := false
		for _, output := range node.outputs {
			name := "_"
			if len(output.connections) > 0 {
				anyConnections = true
				name = newName()
				vars[output] = name
			}
			outputNames = append(outputNames, name)
		}
		assignment := ""
		if anyConnections {
			assignment = Join(outputNames, ", ") + " := "
		}
		s += Sprintf("\t%v%v(%v)\n", assignment, node.name.GetText(), Join(inputNames, ", "))
	}
	s += "}"
	WriteFile("../main.go", []byte(s), 0644)
}

func (f *Function) GetNearestView(views []View, point Point, directionKey int) (nearest View) {
	dir := map[int]Point{glfw.KeyLeft:{-1, 0}, glfw.KeyRight:{1, 0}, glfw.KeyUp:{0, 1}, glfw.KeyDown:{0, -1}}[directionKey]
	bestScore := 0.0
	for _, view := range views {
		d := f.GetViewCenter(view).Sub(point)
		score := (dir.X * d.X + dir.Y * d.Y) / (d.X * d.X + d.Y * d.Y);
		if (score > bestScore) {
			bestScore = score
			nearest = view
		}
	}
	return
}

func (f *Function) FocusNearestView(v View, directionKey int) {
	views := []View{}
	for _, node := range f.nodes {
		views = append(views, node)
		views = append(views, node.Getputs()...)
	}
	for _, connection := range f.connections {
		views = append(views, connection)
	}
	nearest := f.GetNearestView(views, f.GetViewCenter(v), directionKey)
	if nearest != nil { nearest.TakeKeyboardFocus() }
}

func (f *Function) GetViewCenter(v View) Point {
	center := v.Center()
	for v != f && v != nil {
		center = v.MapToParent(center);
		v = v.Parent()
	}
	return center
}

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		f.FocusNearestView(f, event.Key)
	case glfw.KeyF1:
		f.Save()
	}
	if len(event.Text) > 0 {
		if event.Text == "\"" {
			creator := NewStringLiteralCreator(f)
			creator.Move(f.Center())
			creator.created.Connect(func(n ...interface{}) {
				node := n[0].(*Node)
				f.AddNode(node)
				node.MoveCenter(f.Center())
				node.TakeKeyboardFocus()
			})
			creator.canceled.Connect(func(...interface{}) { f.TakeKeyboardFocus() })
		} else {
			creator := NewNodeCreator(f)
			creator.Move(f.Center())
			creator.created.Connect(func(n ...interface{}) {
				node := n[0].(*Node)
				f.AddNode(node)
				node.MoveCenter(f.Center())
				node.TakeKeyboardFocus()
			})
			creator.canceled.Connect(func(...interface{}) { f.TakeKeyboardFocus() })
			creator.text.KeyPressed(event)
		}
	}
}

func (Function) Paint() {}
