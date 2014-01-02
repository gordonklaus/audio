// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
