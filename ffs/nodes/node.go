package nodes

import (
	"encoding/binary"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/billziss-gh/cgofuse/fuse"
)

var endianess = binary.LittleEndian

func (n *Node) putTimeSpec(tx fdb.Transaction, k string, ts fuse.Timespec) {
	buf, _ := ts.Time().MarshalBinary()
	tx.Set(n.ss.Pack(tuple.Tuple{k}), buf)
}

func (n *Node) getTimeSpec(tx fdb.Transaction, k string) (ts fuse.Timespec) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	t := time.Time{}
	_ = t.UnmarshalBinary(d)
	return fuse.NewTimespec(t)
}

func (n *Node) getInt64At(tx fdb.Transaction, k string) (v int64) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	v, _ = binary.Varint(d)
	return v
}

func (n *Node) putInt64At(tx fdb.Transaction, k string, v int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, v)
	tx.Set(n.ss.Pack(tuple.Tuple{k}), buf)
}

func (n *Node) getUint64At(tx fdb.Transaction, k string) (v uint64) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	if len(d) < 8 {
		return 0
	}

	return endianess.Uint64(d)
}

func (n *Node) putUint64At(tx fdb.Transaction, k string, v uint64) {
	b := make([]byte, 8)
	endianess.PutUint64(b, v)
	tx.Set(n.ss.Pack(tuple.Tuple{k}), b)
}

func (n *Node) getUint32At(tx fdb.Transaction, k string) (v uint32) {
	d := tx.Get(n.ss.Pack(tuple.Tuple{k})).MustGet()
	if len(d) < 4 {
		return 0
	}

	return endianess.Uint32(d)
}

func (n *Node) putUint32At(tx fdb.Transaction, k string, v uint32) {
	b := make([]byte, 4)
	endianess.PutUint32(b, v)
	tx.Set(n.ss.Pack(tuple.Tuple{k}), b)
}

type Node struct {
	sss subspace.Subspace
	ss  subspace.Subspace
}

func NewNode(sss subspace.Subspace, ino uint64) *Node {
	return &Node{sss: sss, ss: sss.Sub(int64(ino))}
}

func (n *Node) Init(tx fdb.Transaction, dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) {
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

	const kB = 1024
	const MB = 1024 * kB
	n.putUint32At(tx, "csize", 4*MB)
	n.putUint32At(tx, "fanout", 64)
}
