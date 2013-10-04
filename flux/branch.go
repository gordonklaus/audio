package main

type branchNode struct {
	*nodeBase
}

func newBranchNode(kind string) *branchNode {
	n := &branchNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(kind)
	n.addSeqPorts()
	return n
}
