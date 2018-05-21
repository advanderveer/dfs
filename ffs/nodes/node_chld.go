package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

func (n *Node) CountChld(tx fdb.Transaction) (nc int64) {
	n.ChldEach(tx, func(name string, n *Node) (stop bool) {
		nc++
		return
	})

	return nc
}

func (n *Node) DelChld(tx fdb.Transaction, name string) {
	tx.Clear(n.ss.Pack(tuple.Tuple{"chldr", name}))
}

func (n *Node) GetChld(tx fdb.Transaction, name string) *Node {
	k := n.ss.Pack(tuple.Tuple{"chldr", name})
	d, _ := tx.Get(k).Get()
	if len(d) != 8 {
		return nil
	}

	ino := endianess.Uint64(d)
	return &Node{sss: n.sss, ss: n.sss.Sub(int64(ino))}
}

func (n *Node) SetChld(tx fdb.Transaction, name string, nn *Node) {
	b := make([]byte, 8)
	endianess.PutUint64(b, nn.StatGetIno(tx))
	tx.Set(n.ss.Pack(tuple.Tuple{"chldr", name}), b) //ref
}

func (n *Node) ChldEach(tx fdb.Transaction, f func(name string, n *Node) bool) {
	rng := n.ss.Sub("chldr")
	iter := tx.GetRange(rng, fdb.RangeOptions{}).Iterator()
	for iter.Advance() {
		kv := iter.MustGet()
		t, _ := rng.Unpack(kv.Key)
		if len(t) != 1 {
			break
		}

		name, ok := t[0].(string)
		if !ok {
			break
		}

		ino := endianess.Uint64(kv.Value)
		chld := &Node{sss: n.sss, ss: n.sss.Sub(int64(ino))}
		stop := f(name, chld)
		if stop {
			return
		}
	}
}
