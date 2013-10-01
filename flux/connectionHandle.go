package main

import (
	. "code.google.com/p/gordon-go/gui"
)

type connectionHandle struct {
	*ViewBase
	spec             connectionHandleSpecializer
	conn             *connection
	focused, editing bool
}

const connectionHandleSize = portSize - 2

type connectionHandleSpecializer interface {
	View
	saveConnection()
	restoreSavedConnection()
	updateConnection(p Point)
	moveToNearestConnectablePort(dirKey int)
}

func newConnectionHandle(spec connectionHandleSpecializer, c *connection) *connectionHandle {
	h := &connectionHandle{}
	h.ViewBase = NewView(h)
	h.spec = spec
	h.conn = c
	Resize(h, Pt(connectionHandleSize, connectionHandleSize))
	h.Self = spec
	return h
}

func (h *connectionHandle) startEditing() {
	h.spec.saveConnection()
	SetKeyFocus(h)
	h.editing = true
	h.conn.reform()
}

func (h *connectionHandle) cancelEditing() {
	h.spec.restoreSavedConnection()
	h.stopEditing()
}

func (h *connectionHandle) stopEditing() {
	if h.editing {
		h.editing = false
		if h.conn.connected() {
			h.conn.reform()
		} else {
			h.conn.blk.removeConn(h.conn)
			SetKeyFocus(h.conn.blk)
		}
	}
}

func (h *connectionHandle) TookKeyFocus() { h.focused = true; Repaint(h) }
func (h *connectionHandle) LostKeyFocus() { h.focused = false; h.stopEditing(); Repaint(h) }

func (h *connectionHandle) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		if h.editing {
			h.spec.moveToNearestConnectablePort(event.Key)
		} else {
			h.ViewBase.KeyPress(event)
		}
	case KeyEnter:
		if h.editing {
			h.stopEditing()
		} else {
			h.startEditing()
		}
	case KeyEscape:
		if h.editing {
			h.cancelEditing()
		} else {
			SetKeyFocus(h.conn)
		}
	default:
		h.ViewBase.KeyPress(event)
	}
}

func (h *connectionHandle) Mouse(m MouseEvent) {
	switch {
	case m.Press:
		h.startEditing()
		h.spec.updateConnection(m.Pos)
	case m.Drag:
		if h.editing {
			h.spec.updateConnection(m.Pos)
		}
	case m.Release:
		if h.editing {
			h.spec.updateConnection(m.Pos)
		}
		h.stopEditing()
	}
}

func (h connectionHandle) Paint() {
	// SetColor(map[bool]Color{true:{1, .5, 0, .5}, false:map[bool]Color{true:{.4, .4, 1, .4}, false:{0, 0, 0, .5}}[h.focused]}[h.editing])
	// SetPointSize(connectionHandleSize)
	// DrawPoint(h.Center())
}

type connectionSourceHandle struct {
	*connectionHandle
	savedConnection *port
}

func newConnectionSourceHandle(conn *connection) *connectionSourceHandle {
	h := &connectionSourceHandle{}
	h.connectionHandle = newConnectionHandle(h, conn)
	return h
}

func (h *connectionSourceHandle) saveConnection()         { h.savedConnection = h.conn.src }
func (h *connectionSourceHandle) restoreSavedConnection() { h.conn.setSrc(h.savedConnection) }

func (h *connectionSourceHandle) updateConnection(pt Point) {
	b := h.conn.blk.outermost()
	if p, ok := ViewAt(b, MapTo(h, pt, b)).(*port); ok && h.conn.dst != nil && h.conn.dst.canConnect(p) {
		h.conn.setSrc(p)
	} else {
		srcPt := MapToParent(h, pt)
		h.conn.setSrc(nil)
		h.conn.srcPt = srcPt
		Repaint(h)
	}
}

func (h *connectionSourceHandle) moveToNearestConnectablePort(dirKey int) {
	b := h.conn.blk.outermost()
	ports := []View{}
	for _, n := range b.allNodes() {
		for _, p := range n.outputs() {
			if h.conn.dst.canConnect(p) {
				ports = append(ports, p)
			}
		}
	}

	v := nearestView(b, ports, MapTo(h.conn, h.conn.srcPt, b), dirKey)
	if p, ok := v.(*port); ok {
		h.conn.setSrc(p)
	}
}

func (h *connectionSourceHandle) KeyPress(event KeyEvent) {
	if h.editing {
		h.connectionHandle.KeyPress(event)
		return
	}

	if event.Key == KeyDown && h.conn.src != nil {
		SetKeyFocus(h.conn.src)
	} else if event.Key == KeyUp {
		SetKeyFocus(h.conn.dstHandle)
	} else {
		h.connectionHandle.KeyPress(event)
	}
}

type connectionDestinationHandle struct {
	*connectionHandle
	savedConnection *port
}

func newConnectionDestinationHandle(conn *connection) *connectionDestinationHandle {
	h := &connectionDestinationHandle{}
	h.connectionHandle = newConnectionHandle(h, conn)
	return h
}

func (h *connectionDestinationHandle) saveConnection()         { h.savedConnection = h.conn.dst }
func (h *connectionDestinationHandle) restoreSavedConnection() { h.conn.setDst(h.savedConnection) }

func (h *connectionDestinationHandle) updateConnection(pt Point) {
	b := h.conn.blk.outermost()
	if p, ok := ViewAt(b, MapTo(h, pt, b)).(*port); ok && h.conn.src != nil && p.canConnect(h.conn.src) {
		h.conn.setDst(p)
	} else {
		dstPt := MapToParent(h, pt)
		h.conn.setDst(nil)
		h.conn.dstPt = dstPt
		Repaint(h)
	}
}

func (h *connectionDestinationHandle) moveToNearestConnectablePort(dirKey int) {
	b := h.conn.blk.outermost()
	ports := []View{}
	for _, n := range b.allNodes() {
		for _, p := range n.inputs() {
			if p.canConnect(h.conn.src) {
				ports = append(ports, p)
			}
		}
	}

	v := nearestView(b, ports, MapTo(h.conn, h.conn.dstPt, b), dirKey)
	if p, ok := v.(*port); ok {
		h.conn.setDst(p)
	}
}

func (h *connectionDestinationHandle) KeyPress(event KeyEvent) {
	if h.editing {
		h.connectionHandle.KeyPress(event)
		return
	}

	if event.Key == KeyDown {
		SetKeyFocus(h.conn.srcHandle)
	} else if event.Key == KeyUp && h.conn.dst != nil {
		SetKeyFocus(h.conn.dst)
	} else {
		h.connectionHandle.KeyPress(event)
	}
}
