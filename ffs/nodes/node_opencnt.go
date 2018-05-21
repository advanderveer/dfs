package nodes

import "github.com/apple/foundationdb/bindings/go/src/fdb"

func (n *NodeT) Opencnt(tx fdb.Transaction) int64 {
	return n.getInt64At(tx, "opencnt")
}

//@TODO use the tx.Add method
func (n *NodeT) IncOpencnt(tx fdb.Transaction) {
	cnt := n.getInt64At(tx, "opencnt")
	cnt++
	n.putInt64At(tx, "opencnt", cnt)
}

//@TODO use the tx.Add method
func (n *NodeT) DecOpencnt(tx fdb.Transaction) {
	cnt := n.getInt64At(tx, "opencnt")
	cnt--
	n.putInt64At(tx, "opencnt", cnt)
}
