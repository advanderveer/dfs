package nodes

import (
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

//Tx allows for atomic interaction with nodes
type Tx interface {
	Lookup(path string, ancestor *Node) (prnt *Node, name string, node *Node)
	Iterate(node *Node, next func(name string, chld *Node) bool)
	Save(node ...*Node)
}

//TxR is a read-only node transaction
type TxR struct {
	root   *Node
	bucket *bolt.Bucket
}

//Save will persist the node at the end of the trans
func (tx *TxR) Save(node ...*Node) {
	panic("called save in read-only transaction")
}

//Iterate walks direct children of a node
func (tx *TxR) Iterate(node *Node, next func(name string, chld *Node) bool) {
	for name := range node.Chld {
		chld, err := loadChild(tx.bucket, node, name)
		if err != nil {
			//@TODO handle error
		}

		ok := next(name, chld)
		if !ok {
			break
		}
	}
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

			var err error
			node, err = loadChild(tx.bucket, prnt, name)
			if err != nil {
				_ = err //@TODO handle error
			}

			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return prnt, name, node
			}
		}
	}

	return prnt, name, node
}

//TxRW is a read-only node transaction
type TxRW struct {
	TxR
	saves []*Node //@TODO how about concurrent access?
}

//Save will persist the node at the end of the trans
func (tx *TxRW) Save(node ...*Node) {
	tx.saves = append(tx.saves, node...)
}
