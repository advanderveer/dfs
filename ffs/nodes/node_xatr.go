package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (n *Node) XAtrDel(tx fdb.Transaction, name string) {
	tx.Clear(n.ss.Pack(tuple.Tuple{"xatr", name}))
}

func (n *Node) XAtrGet(tx fdb.Transaction, name string) (a []byte, ok bool) {
	d, err := tx.Get(n.ss.Pack(tuple.Tuple{"xatr", name})).Get()
	if err != nil { //@TODO handle other errors
		return nil, false
	}

	return d, true
}

func (n *Node) XAtrSet(tx fdb.Transaction, name string, xatr []byte) {
	if len(xatr) > 100*1000 {
		panic("ffs: tried to set xattr that is larger then 100KB reached") //@TODO handle more graceflly
	}

	tx.Set(n.ss.Pack(tuple.Tuple{"xatr", name}), xatr)
}

func (n *Node) XAtrEach(tx fdb.Transaction, f func(name string) int) (errc int) {
	rng := n.ss.Sub("xatr")
	iter := tx.GetRange(rng, fdb.RangeOptions{Limit: 20}).Iterator()
	for iter.Advance() {
		kv := iter.MustGet()
		t, _ := rng.Unpack(kv.Key)
		if len(t) != 1 {
			return -fuse.EIO //@TODO think of a better error
		}

		name, ok := t[0].(string)
		if !ok {
			return -fuse.EIO //@TODO think of a better error
		}

		errc = f(name)
		if errc != 0 {
			return errc
		}
	}

	return 0
}
