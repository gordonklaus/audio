package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	"code.google.com/p/go.exp/go/types"
	"go/ast"
	"go/token"
	"math"
	"strconv"
	"strings"
)

type node interface {
	View
	MouseHandler
	block() *block
	setBlock(b *block)
	inputs() []*port
	outputs() []*port
	inConns() []*connection
	outConns() []*connection
}

type nodeText struct {
	*TextBase
	node *nodeBase
}
func newNodeText(node *nodeBase) *nodeText {
	t := &nodeText{node:node}
	t.TextBase = NewTextBase(t, "")
	return t
}
func (t *nodeText) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyEnter:
		t.node.TakeKeyboardFocus()
	default:
		t.TextBase.KeyPressed(event)
	}
	t.node.reform()
}

type nodeBase struct {
	*ViewBase
	self node
	AggregateMouseHandler
	blk *block
	text *nodeText
	ins []*port
	outs []*port
	focused bool
}

const (
	nodeMargin = 3
)

func newNodeBase(self node) *nodeBase {
	n := &nodeBase{self:self}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(self), NewViewDragger(self)}
	n.text = newNodeText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.ViewBase.Self = self
	return n
}

func (n *nodeBase) newInput(v *types.Var) *port {
	p := newInput(n.self, v)
	n.AddChild(p)
	n.ins = append(n.ins, p)
	n.reform()
	return p
}

func (n *nodeBase) newOutput(v *types.Var) *port {
	p := newOutput(n.self, v)
	n.AddChild(p)
	n.outs = append(n.outs, p)
	n.reform()
	return p
}

func (n *nodeBase) removePortBase(p *port) { // intentionally named to not implement interface{removePort(*port)}
	if p.out {
		SliceRemove(&n.outs, p)
	} else {
		SliceRemove(&n.ins, p)
	}
	n.RemoveChild(p)
	n.reform()
}

func (n *nodeBase) reform() {
	var ins, outs []*port
	for _, p := range n.ins {
		if p.obj.Type != seqType {
			ins = append(ins, p)
		}
	}
	for _, p := range n.outs {
		if p.obj.Type != seqType {
			outs = append(outs, p)
		}
	}
	
	numIn := float64(len(ins))
	numOut := float64(len(outs))
	rx, ry := 2.0 * portSize, (math.Max(numIn, numOut) + 1) * portSize / 2
	
	rect := ZR
	for i, p := range ins {
		y := -portSize * (float64(i) - (numIn - 1) / 2)
		p.MoveCenter(Pt(-8 - rx * math.Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(p.MapRectToParent(p.Rect()))
	}
	for i, p := range outs {
		y := -portSize * (float64(i) - (numOut - 1) / 2)
		p.MoveCenter(Pt(8 + rx * math.Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(p.MapRectToParent(p.Rect()))
	}

	n.text.MoveCenter(Pt(0, rect.Max.Y + n.text.Height() / 2))
	n.Pan(rect.Min)
	n.Resize(rect.Dx(), rect.Dy())
}

func (n nodeBase) block() *block { return n.blk }
func (n *nodeBase) setBlock(b *block) { n.blk = b }
func (n nodeBase) inputs() []*port { return n.ins }
func (n nodeBase) outputs() []*port { return n.outs }

func (n nodeBase) inConns() (conns []*connection) {
	for _, p := range n.inputs() {
		for _, c := range p.conns {
			conns = append(conns, c)
		}
	}
	return
}

func (n nodeBase) outConns() (conns []*connection) {
	for _, p := range n.outputs() {
		for _, c := range p.conns {
			conns = append(conns, c)
		}
	}
	return
}

func (n *nodeBase) Move(p Point) {
	n.ViewBase.Move(p)
	for _, p := range append(n.ins, n.outs...) {
		for _, c := range p.conns {
			c.reform()
		}
	}
}

func (n nodeBase) Center() Point { return ZP }

func (n *nodeBase) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *nodeBase) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *nodeBase) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft:
		if p := seqIn(n); p != nil {
			p.TakeKeyboardFocus()
		} else {
			n.blk.outermost().focusNearestView(n, event.Key)
		}
	case KeyRight:
		if p := seqOut(n); p != nil {
			p.TakeKeyboardFocus()
		} else {
			n.blk.outermost().focusNearestView(n, event.Key)
		}
	case KeyUp, KeyDown:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyEscape:
		n.blk.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n nodeBase) Paint() {
	const DX = 8.0
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	for _, p := range append(n.ins, n.outs...) {
		pt := p.MapToParent(p.Center())
		dx := -DX
		if p.out {
			dx = DX
		}
		x := (pt.X - dx) / 2 + dx
		DrawCubic([4]Point{Pt(dx, 0), Pt(x, 0), Pt(x, pt.Y), pt}, int(pt.Len() / 2))
	}
	in, out := len(n.ins) > 0, len(n.outs) > 0
	if in {
		DrawLine(Pt(-DX, 0), ZP)
	}
	if out {
		DrawLine(ZP, Pt(DX, 0))
	}
	if !in || !out {
		DrawLine(Pt(0, -3), Pt(0, 3))
	}
}

func newNode(obj types.Object) node {
	switch obj := obj.(type) {
	case special:
		switch obj.name {
		case "[]":                        return newIndexNode(false)
		case "[]=":                       return newIndexNode(true)
		case "if":                        return newIfNode()
		case "loop":                      return newLoopNode()
		}
	case *types.Func, method:             return newCallNode(obj)
	case *types.Var, *types.Const, field: return newValueNode(obj, false)
	}
	return nil
}

var seqType = struct { types.Type }{}

func seqIn(n node) *port {
	for _, in := range n.inputs() {
		if in.obj.Type == seqType {
			return in
		}
	}
	return nil
}

func seqOut(n node) *port {
	for _, out := range n.outputs() {
		if out.obj.Type == seqType {
			return out
		}
	}
	return nil
}

type callNode struct {
	*nodeBase
	obj types.Object
}
func newCallNode(obj types.Object) *callNode {
	n := &callNode{obj:obj}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(obj.GetName())
	switch t := obj.GetType().(type) {
	case *types.Signature:
		if t.Recv != nil { n.newInput(t.Recv) }
		for _, v := range t.Params { n.newInput(v) }
		for _, v := range t.Results { n.newOutput(v) }
	default: // builtin -- handled specially elsewhere?
	}
	seqIn := n.newInput(&types.Var{Name: "seq", Type: seqType})
	seqIn.MoveCenter(Pt(-8, 0))
	seqOut := n.newOutput(&types.Var{Name: "seq", Type: seqType})
	seqOut.MoveCenter(Pt(8, 0))
	
	return n
}

type basicLiteralNode struct {
	*nodeBase
	kind token.Token
}
func newBasicLiteralNode(kind token.Token) *basicLiteralNode {
	n := &basicLiteralNode{kind: kind}
	n.nodeBase = newNodeBase(n)
	out := n.newOutput(&types.Var{})
	switch kind {
	case token.INT, token.FLOAT:
		if kind == token.INT {
			out.setType(types.Typ[types.UntypedInt])
		} else {
			out.setType(types.Typ[types.UntypedFloat])
		}
		n.text.SetValidator(func(s *string) bool {
			*s = strings.TrimLeft(*s, "0")
			if *s == "" || *s == "-" {
				*s = "0"
			}
			if (*s)[0] == '.' {
				*s = "0" + *s
			}
			if l := len(*s); (*s)[l - 1] == '-' {
				if (*s)[0] == '-' {
					*s = (*s)[1:l - 1]
				} else {
					*s = "-" + (*s)[:l - 1]
				}
			}
			if _, err := strconv.ParseInt(*s, 10, 64); err == nil {
				n.kind = token.INT
				out.setType(types.Typ[types.UntypedInt])
			} else {
				if _, err := strconv.ParseFloat(*s, 4096); err == nil {
					n.kind = token.FLOAT
					out.setType(types.Typ[types.UntypedFloat])
				} else {
					return false
				}
			}
			return true
		})
	case token.IMAG:
		// TODO
	case token.STRING:
		out.setType(types.Typ[types.UntypedString])
	case token.CHAR:
		out.setType(types.Typ[types.UntypedRune])
		n.text.SetValidator(func(s *string) bool {
			if *s == "" {
				return false
			}
			*s = (*s)[len(*s) - 1:]
			return true
		})
	}
	return n
}

type compositeLiteralNode struct {
	*nodeBase
	typ *typeView
}
func newCompositeLiteralNode() *compositeLiteralNode {
	n := &compositeLiteralNode{}
	n.nodeBase = newNodeBase(n)
	v := &types.Var{}
	n.newOutput(v)
	n.typ = newTypeView(&v.Type)
	n.typ.mode = compositeOrPtrType
	n.AddChild(n.typ)
	return n
}
func (n *compositeLiteralNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			n.blk.TakeKeyboardFocus()
		}
	})
}
func (n *compositeLiteralNode) setType(t types.Type) {
	n.typ.setType(t)
	n.outs[0].setType(t)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		switch t, _ = indirect(t); t := t.(type) {
		case *types.NamedType:
			for _, f := range t.Underlying.(*types.Struct).Fields {
				if t.Obj.Pkg == n.blk.func_().pkg() || fieldIsExported(f) {
					n.newInput(&types.Var{Pkg: f.Pkg , Name: f.Name, Type: f.Type})
				}
			}
		case *types.Struct:
			for _, f := range t.Fields {
				n.newInput(&types.Var{Pkg: f.Pkg , Name: f.Name, Type: f.Type})
			}
		case *types.Slice:
			// TODO: variable number of inputs? (same can be achieved using append.)  variable number of index/value input pairs?
		case *types.Map:
			// TODO: variable number of key/value input pairs?
		}
		n.typ.MoveCenter(Pt(0, n.Rect().Max.Y + n.typ.Height() / 2))
		n.TakeKeyboardFocus()
	}
}

func fieldIsExported(f *types.Field) bool {
	name := f.Name
	if name == "" {
		t := f.Type
		t, _ = indirect(t)
		name = t.(*types.NamedType).Obj.Name
	}
	return ast.IsExported(name)
}
