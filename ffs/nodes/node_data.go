package nodes

import "github.com/apple/foundationdb/bindings/go/src/fdb"

//@TODO store somewhere presistent
var blockStore = map[uint64][]byte{}

func (n *NodeT) Data(tx fdb.Transaction) []byte {
	return blockStore[n.statGetIno(tx)]
}

func (n *NodeT) SetData(tx fdb.Transaction, d []byte) {
	blockStore[n.statGetIno(tx)] = d
}

func (n *NodeT) CopyData(tx fdb.Transaction, d []byte) {
	d2 := n.Data(tx)
	copy(d2, d)
}
