package nodes

import "github.com/billziss-gh/cgofuse/fuse"

//Tx allows for atomic interaction with nodes
type Tx interface {
	Lookup(path string, ancestor *Node) (prnt *Node, name string, node *Node)
}

//TxR is a read-only node transaction
type TxR struct {
	root *Node
}

//TxRW is a read-only node transaction
type TxRW struct {
	TxR
}

//Lookup a node by its path, it walks the in-memory tree structure of inodes and loads persistent information from disk
func (tx *TxR) Lookup(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
	prnt = tx.root
	name = ""
	node = tx.root
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}

			prnt, name = node, c
			node = node.chlds[c]

			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return prnt, name, node
			}
		}
	}

	return prnt, name, node
}
