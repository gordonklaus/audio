package flux

type Function struct {
	nodes []Node
}

func NewFunction() *Function {
	return &Function{[]Node{}}
}
