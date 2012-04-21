package flux

type Node struct {
	function *Function
}

func NewNode(function *Function) *Node {
	return &Node{function}
}
