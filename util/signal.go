package util

type Signal struct {
	connections []*SignalConnection
}

func NewSignal() *Signal { return &Signal{[]*SignalConnection{}} }

func (s *Signal) Connect(c func(...interface{})) *SignalConnection {
	connection := &SignalConnection{s, true, c}
	s.connections = append(s.connections, connection)
	return connection
}

func (s *Signal) ConnectSingleShot(c func(...interface{})) *SignalConnection {
	var connection *SignalConnection
	connection = s.Connect(func(...interface{}){ c(); connection.Disconnect() })
	return connection
}

func (s *Signal) Emit(args ...interface{}) {
	for _, connection := range s.connections {
		if connection.active {
			connection.callback(args...)
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
	for i, connection := range s.connections { if c == connection { s.connections = append(s.connections[:i], s.connections[i+1:]...); return } }
}

func (c *SignalConnection) Block() { c.active = false }
func (c *SignalConnection) Unblock() { c.active = true }
