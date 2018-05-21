package nodes

import (
	"encoding/binary"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (n *NodeT) putTimeSpec(tx fdb.Transaction, k string, ts fuse.Timespec) {
	buf, _ := ts.Time().MarshalBinary()
	tx.Set(n.ss.Pack(tuple.Tuple{k}), buf)
}

func (n *NodeT) getTimeSpec(tx fdb.Transaction, k string) (ts fuse.Timespec) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	t := time.Time{}
	_ = t.UnmarshalBinary(d)
	return fuse.NewTimespec(t)
}

func (n *NodeT) getInt64At(tx fdb.Transaction, k string) (v int64) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	v, _ = binary.Varint(d)
	return v
}

func (n *NodeT) putInt64At(tx fdb.Transaction, k string, v int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, v)
	tx.Set(n.ss.Pack(tuple.Tuple{k}), buf)
}

func (n *NodeT) getUint64At(tx fdb.Transaction, k string) (v uint64) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	if len(d) < 8 {
		return 0
	}

	return binary.LittleEndian.Uint64(d)
}

func (n *NodeT) putUint64At(tx fdb.Transaction, k string, v uint64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	tx.Set(n.ss.Pack(tuple.Tuple{k}), b)
}

func (n *NodeT) getUint32At(tx fdb.Transaction, k string) (v uint32) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	if len(d) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(d)
}

func (n *NodeT) putUint32At(tx fdb.Transaction, k string, v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	tx.Set(n.ss.Pack(tuple.Tuple{k}), b)
}

type node struct {
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

func (n *NodeT) Data(tx fdb.Transaction) []byte        { return n.no.data }
func (n *NodeT) SetData(tx fdb.Transaction, d []byte)  { n.no.data = d }
func (n *NodeT) CopyData(tx fdb.Transaction, d []byte) { copy(n.no.data, d) }

func (n *NodeT) Opencnt(tx fdb.Transaction) int { return n.no.opencnt }
func (n *NodeT) IncOpencnt(tx fdb.Transaction)  { n.no.opencnt++ }
func (n *NodeT) DecOpencnt(tx fdb.Transaction)  { n.no.opencnt-- }
