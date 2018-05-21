package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

func (n *NodeT) CountChld(tx fdb.Transaction) (nc int64) {
	n.ChldEach(tx, func(name string, n *NodeT) (stop bool) {
		nc++
		return
	})

	// len := int64(len(n.no.chld))
	// fmt.Println(len)
	return nc
}

func (n *NodeT) DelChld(tx fdb.Transaction, name string) {
	delete(n.no.chld, name)
	tx.Clear(n.ss.Pack(tuple.Tuple{"chldr", name}))
}

func (n *NodeT) GetChld(tx fdb.Transaction, name string) *NodeT {
	// k := n.ss.Pack(tuple.Tuple{"chldr", name})
	// d, _ := tx.Get(k).Get()
	// if len(d) != 8 {
	// 	return nil
	// }

	rn, ok := n.no.chld[name]
	if !ok {
		return nil
	}

	return rn

	// ino := endianess.Uint64(d)
	// node := &NodeT{
	// 	sss: n.sss,
	// 	ss:  n.sss.Sub(ino),
	// }
	//
	// node.no = rn.no
	// fmt.Println(node)
	// t, err := n.ss.Unpack(nil)
	// fmt.Println(t[:len(t)-2])

	// node := NodeT{ss: store.ss.Sub(ino)}
	// fmt.Println(node)

	// return node
}

func (n *NodeT) SetChld(tx fdb.Transaction, name string, nn *NodeT) {
	n.no.chld[name] = nn
	//
	// b := make([]byte, 8)
	// endianess.PutUint64(b, nn.statGetIno(tx))
	// tx.Set(n.ss.Pack(tuple.Tuple{"chldr", name}), b) //ref
}

func (n *NodeT) ChldEach(tx fdb.Transaction, f func(name string, n *NodeT) bool) {
	// iter := tx.GetRange(n.ss.Sub(tuple.Tuple{"chldr"}), fdb.RangeOptions{}).Iterator()
	// for iter.Advance() {
	// 	fmt.Println(iter.Get())
	// }

	for name, n := range n.no.chld {
		stop := f(name, n)
		if stop {
			return
		}
	}
}
