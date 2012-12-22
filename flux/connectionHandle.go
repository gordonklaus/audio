package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type ConnectionHandle struct {
	*ViewBase
	spec ConnectionHandleSpecializer
	connection *Connection
	
	focused bool
	editing bool
}

const connectionHandleSize = portSize - 2

type ConnectionHandleSpecializer interface {
	View
	SaveConnection()
	RestoreSavedConnection()
	UpdateConnection(p Point)
	MoveToNearestConnectableput(key int)
}

func NewConnectionHandle(spec ConnectionHandleSpecializer, conn *Connection) *ConnectionHandle {
	h := &ConnectionHandle{}
	h.ViewBase = NewView(h)
	h.spec = spec
	h.connection = conn
	h.Resize(connectionHandleSize, connectionHandleSize)
	h.Self = spec
	return h
}

func (h *ConnectionHandle) StartEditing() {
	h.spec.SaveConnection()
	h.TakeKeyboardFocus()
	h.editing = true
	h.connection.reform()
}

func (h *ConnectionHandle) CancelEditing() {
	h.spec.RestoreSavedConnection()
	h.StopEditing()
}

func (h *ConnectionHandle) StopEditing() {
	if h.editing {
		h.editing = false
		if h.connection.Connected() {
			h.connection.reform()
		} else {
			h.connection.block.DeleteConnection(h.connection)
			h.connection.block.TakeKeyboardFocus()
		}
	}
}

func (h *ConnectionHandle) TookKeyboardFocus() { h.focused = true; h.Repaint() }
func (h *ConnectionHandle) LostKeyboardFocus() { h.focused = false; h.StopEditing(); h.Repaint() }

func (h *ConnectionHandle) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		if h.editing {
			h.spec.MoveToNearestConnectableput(event.Key)
		} else {
			h.connection.block.Outermost().FocusNearestView(h.spec, event.Key)
		}
	case glfw.KeyEnter:
		if h.editing {
			h.StopEditing()
		} else {
			h.StartEditing()
		}
	case glfw.KeyEsc:
		if h.editing {
			h.CancelEditing()
		} else {
			h.connection.TakeKeyboardFocus()
		}
	default:
		h.ViewBase.KeyPressed(event)
	}
}

func (h *ConnectionHandle) MousePressed(button int, p Point) {
	h.StartEditing()
	h.spec.UpdateConnection(p)
}
func (h *ConnectionHandle) MouseDragged(button int, p Point) {
	if h.editing { h.spec.UpdateConnection(p) }
}
func (h *ConnectionHandle) MouseReleased(button int, p Point) {
	if h.editing { h.spec.UpdateConnection(p) }
	h.StopEditing()
}

func (h ConnectionHandle) Paint() {
	SetColor(map[bool]Color{true:{1, .5, 0, .5}, false:map[bool]Color{true:{.4, .4, 1, .4}, false:{0, 0, 0, .5}}[h.focused]}[h.editing])
	SetPointSize(connectionHandleSize)
	DrawPoint(h.Center())
}


type ConnectionSourceHandle struct {
	*ConnectionHandle
	savedConnection *Output
}

func NewConnectionSourceHandle(conn *Connection) *ConnectionSourceHandle {
	h := &ConnectionSourceHandle{}
	h.ConnectionHandle = NewConnectionHandle(h, conn)
	return h
}

func (h *ConnectionSourceHandle) SaveConnection() { h.savedConnection = h.connection.src }
func (h *ConnectionSourceHandle) RestoreSavedConnection() { h.connection.SetSource(h.savedConnection) }

func (h ConnectionSourceHandle) UpdateConnection(p Point) {
	outermost := h.connection.block.Outermost()
	if output, ok := outermost.ViewAt(h.MapTo(p, outermost)).(*Output); ok && h.connection.dst.CanConnect(output) {
		h.connection.SetSource(output)
	} else {
		h.connection.DisconnectSource(h.MapTo(p, h.connection.block))
	}
}

func (h *ConnectionSourceHandle) MoveToNearestConnectableput(key int) {
	block := h.connection.block.Outermost()
	connectableputs := []View{}
	for _, node := range block.AllNodes() {
		for _, output := range node.Outputs() {
			if h.connection.dst.CanConnect(output) { connectableputs = append(connectableputs, output) }
		}
	}
	
	view := nearestView(block, connectableputs, h.connection.srcPt, key)
	if p, ok := view.(*Output); ok {
		h.connection.SetSource(p)
	}
}

func (h *ConnectionSourceHandle) KeyPressed(event KeyEvent) {
	if h.editing {
		h.ConnectionHandle.KeyPressed(event)
		return
	}
	
	if event.Key == glfw.KeyDown && h.connection.src != nil {
		h.connection.src.TakeKeyboardFocus()
	} else if event.Key == glfw.KeyUp {
		h.connection.dstHandle.TakeKeyboardFocus()
	} else {
		h.ConnectionHandle.KeyPressed(event)
	}
}


type ConnectionDestinationHandle struct {
	*ConnectionHandle
	savedConnection *Input
}

func NewConnectionDestinationHandle(conn *Connection) *ConnectionDestinationHandle {
	h := &ConnectionDestinationHandle{}
	h.ConnectionHandle = NewConnectionHandle(h, conn)
	return h
}

func (h *ConnectionDestinationHandle) SaveConnection() { h.savedConnection = h.connection.dst }
func (h *ConnectionDestinationHandle) RestoreSavedConnection() { h.connection.SetDestination(h.savedConnection) }

func (h ConnectionDestinationHandle) UpdateConnection(p Point) {
	outermost := h.connection.block.Outermost()
	if input, ok := outermost.ViewAt(h.MapTo(p, outermost)).(*Input); ok && input.CanConnect(h.connection.src) {
		h.connection.SetDestination(input)
	} else {
		h.connection.DisconnectDestination(h.MapTo(p, h.connection.block))
	}
}

func (h *ConnectionDestinationHandle) MoveToNearestConnectableput(key int) {
	block := h.connection.block.Outermost()
	connectableputs := []View{}
	for _, node := range block.AllNodes() {
		for _, input := range node.Inputs() {
			if input.CanConnect(h.connection.src) { connectableputs = append(connectableputs, input) }
		}
	}
	
	view := nearestView(block, connectableputs, h.connection.dstPt, key)
	if p, ok := view.(*Input); ok {
		h.connection.SetDestination(p)
	}
}

func (h *ConnectionDestinationHandle) KeyPressed(event KeyEvent) {
	if h.editing {
		h.ConnectionHandle.KeyPressed(event)
		return
	}
	
	if event.Key == glfw.KeyDown {
		h.connection.srcHandle.TakeKeyboardFocus()
	} else if event.Key == glfw.KeyUp && h.connection.dst != nil {
		h.connection.dst.TakeKeyboardFocus()
	} else {
		h.ConnectionHandle.KeyPressed(event)
	}
}
