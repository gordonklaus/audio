package util

type Signal struct {
	conns []*SignalConnection
}

func NewSignal() *Signal { return &Signal{[]*SignalConnection{}} }

func (s *Signal) Connect(f func(...interface{})) *SignalConnection {
	c := &SignalConnection{s, true, f}
	s.conns = append(s.conns, c)
	return c
}

func (s *Signal) ConnectSingleShot(f func(...interface{})) *SignalConnection {
	var c *SignalConnection
	c = s.Connect(func(...interface{}) {
		f()
		c.Disconnect()
	})
	return c
}

func (s *Signal) Emit(args ...interface{}) {
	for _, c := range s.conns {
		if c.active {
			c.callback(args...)
		}
	}
}


type SignalConnection struct {
	signal *Signal
	active bool
	callback func(...interface{})
}

func (c *SignalConnection) Disconnect() {
	s := c.signal
	for i, conn := range s.conns {
		if c == conn {
			s.conns = append(s.conns[:i], s.conns[i+1:]...)
			return
		}
	}
}

func (c *SignalConnection) Block() { c.active = false }
func (c *SignalConnection) Unblock() { c.active = true }
