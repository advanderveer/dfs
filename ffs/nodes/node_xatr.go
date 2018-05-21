package nodes

import "github.com/apple/foundationdb/bindings/go/src/fdb"

func (n *NodeT) XAtrDel(tx fdb.Transaction, name string) { delete(n.no.xatr, name) }

func (n *NodeT) XAtrGet(tx fdb.Transaction, name string) (a []byte, ok bool) {
	a, ok = n.no.xatr[name]
	return
}

func (n *NodeT) XAtrSet(tx fdb.Transaction, name string, xatr []byte) {
	if nil == n.no.xatr {
		n.no.xatr = map[string][]byte{}
	}
	n.no.xatr[name] = xatr
}

func (n *NodeT) XAtrEach(tx fdb.Transaction, f func(name string) int) (errc int) {
	for name := range n.no.xatr {
		errc = f(name)
		if errc != 0 {
			return errc
		}
	}

	return 0
}
