package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/billziss-gh/cgofuse/fuse"
)

type node struct {
	stat    fuse.Stat_t
	xatr    map[string][]byte
	chld    map[string]*NodeT
	data    []byte
	opencnt int
}

type NodeT struct {
	ss subspace.Subspace
	no node
}

func (n *NodeT) Init(tx fdb.Transaction, dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) {
	tmsp := fuse.Now()
	n.statSetDev(tx, dev)
	n.statSetIno(tx, ino)
	n.StatIncNlink(tx) //to 1
	n.StatSetMode(tx, mode)
	n.StatSetUid(tx, uid)
	n.StatSetGid(tx, gid)
	n.StatSetATim(tx, tmsp)
	n.StatSetMTim(tx, tmsp)
	n.StatSetCTim(tx, tmsp)
	n.StatSetBirthTim(tx, tmsp)
	n.StatSetFlags(tx, 0)
	if fuse.S_IFDIR == mode&fuse.S_IFMT {
		n.no.chld = map[string]*NodeT{}
	}
}

func (n *NodeT) SetChld(tx fdb.Transaction, name string, nn *NodeT) { n.no.chld[name] = nn }
func (n *NodeT) GetChld(tx fdb.Transaction, name string) *NodeT     { return n.no.chld[name] }

func (n *NodeT) DelChld(tx fdb.Transaction, name string) { delete(n.no.chld, name) }
func (n *NodeT) CountChld(tx fdb.Transaction) int64      { return int64(len(n.no.chld)) }
func (n *NodeT) ChldEach(tx fdb.Transaction, f func(name string, n *NodeT) bool) {
	for name, n := range n.no.chld {
		stop := f(name, n)
		if stop {
			return
		}
	}
}

func (n *NodeT) XAtrGet(tx fdb.Transaction, name string) (a []byte, ok bool) {
	a, ok = n.no.xatr[name]
	return
}
func (n *NodeT) XAtrDel(tx fdb.Transaction, name string) { delete(n.no.xatr, name) }
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

func (n *NodeT) Data(tx fdb.Transaction) []byte        { return n.no.data }
func (n *NodeT) SetData(tx fdb.Transaction, d []byte)  { n.no.data = d }
func (n *NodeT) CopyData(tx fdb.Transaction, d []byte) { copy(n.no.data, d) }

func (n *NodeT) Opencnt(tx fdb.Transaction) int { return n.no.opencnt }
func (n *NodeT) IncOpencnt(tx fdb.Transaction)  { n.no.opencnt++ }
func (n *NodeT) DecOpencnt(tx fdb.Transaction)  { n.no.opencnt-- }
