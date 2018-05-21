package nodes

import "github.com/apple/foundationdb/bindings/go/src/fdb"

func (n *Node) Opencnt(tx fdb.Transaction) int64 {
	return n.getInt64At(tx, "opencnt")
}

//@TODO use the tx.Add method
func (n *Node) IncOpencnt(tx fdb.Transaction) {
	cnt := n.getInt64At(tx, "opencnt")
	cnt++
	n.putInt64At(tx, "opencnt", cnt)
}

//@TODO use the tx.Add method
func (n *Node) DecOpencnt(tx fdb.Transaction) {
	cnt := n.getInt64At(tx, "opencnt")
	cnt--
	n.putInt64At(tx, "opencnt", cnt)
}
