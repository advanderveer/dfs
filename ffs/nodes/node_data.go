package nodes

import "github.com/apple/foundationdb/bindings/go/src/fdb"

//@TODO store somewhere presistent
var blockStore = map[uint64][]byte{}

func (n Node) Data(tx fdb.Transaction) []byte {
	return blockStore[n.statGetIno(tx)]
}

func (n Node) SetData(tx fdb.Transaction, d []byte) {
	blockStore[n.statGetIno(tx)] = d
}

func (n Node) CopyData(tx fdb.Transaction, d []byte) {
	d2 := n.Data(tx)
	copy(d2, d)
}
