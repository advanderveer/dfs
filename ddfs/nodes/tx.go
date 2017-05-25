package nodes

import "github.com/billziss-gh/cgofuse/fuse"

//Tx is a read-only node transaction
type Tx struct {
	root    *Node
	openmap map[uint64]*Node //@TODO move this out of the tx (to the fs)
}

//TxRW is a read-only node transaction
type TxRW struct {
	Tx
}

//Get will lookup or get an open node
//@TODO move this out of the tx, to the fs
func (tx *Tx) Get(path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := tx.Lookup(path, nil)
		return node
	}

	return tx.openmap[fh]
}

//Lookup a node by its path
func (tx *Tx) Lookup(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
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
