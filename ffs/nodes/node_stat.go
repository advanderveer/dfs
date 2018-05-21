package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (n *NodeT) Stat(tx fdb.Transaction) fuse.Stat_t {
	sta := fuse.Stat_t{}
	sta.Ino = n.getUint64At(tx, "ino")
	sta.Dev = n.getUint64At(tx, "dev")
	sta.Mode = n.getUint32At(tx, "mode")
	sta.Uid = n.getUint32At(tx, "uid")
	sta.Gid = n.getUint32At(tx, "gid")
	sta.Flags = n.getUint32At(tx, "flags")
	sta.Size = n.getInt64At(tx, "size")
	sta.Ctim = n.getTimeSpec(tx, "ctim")
	sta.Mtim = n.getTimeSpec(tx, "mtim")
	sta.Atim = n.getTimeSpec(tx, "atim")
	sta.Birthtim = n.getTimeSpec(tx, "btim")
	sta.Nlink = n.getUint32At(tx, "nlink")
	return sta
}

func (n *NodeT) statSetIno(tx fdb.Transaction, ino uint64)           { n.putUint64At(tx, "ino", ino) }
func (n *NodeT) statSetDev(tx fdb.Transaction, dev uint64)           { n.putUint64At(tx, "dev", dev) }
func (n *NodeT) StatSetMode(tx fdb.Transaction, m uint32)            { n.putUint32At(tx, "mode", m) }
func (n *NodeT) StatSetUid(tx fdb.Transaction, uid uint32)           { n.putUint32At(tx, "uid", uid) }
func (n *NodeT) StatSetGid(tx fdb.Transaction, gid uint32)           { n.putUint32At(tx, "gid", gid) }
func (n *NodeT) StatSetFlags(tx fdb.Transaction, f uint32)           { n.putUint32At(tx, "flags", f) }
func (n *NodeT) StatSetSize(tx fdb.Transaction, len int64)           { n.putInt64At(tx, "size", len) }
func (n *NodeT) StatSetCTim(tx fdb.Transaction, t fuse.Timespec)     { n.putTimeSpec(tx, "ctim", t) }
func (n *NodeT) StatSetMTim(tx fdb.Transaction, t fuse.Timespec)     { n.putTimeSpec(tx, "mtim", t) }
func (n *NodeT) StatSetATim(tx fdb.Transaction, t fuse.Timespec)     { n.putTimeSpec(tx, "atim", t) }
func (n *NodeT) StatSetBirthTim(tx fdb.Transaction, t fuse.Timespec) { n.putTimeSpec(tx, "birthtim", t) }

//@TODO use the tx.Add method instead
func (n *NodeT) StatIncNlink(tx fdb.Transaction) {
	v := n.getUint32At(tx, "nlink")
	v++
	n.putUint32At(tx, "nlink", v)
}

//@TODO use the tx.Add method instead
func (n *NodeT) StatDecNlink(tx fdb.Transaction) {
	v := n.getUint32At(tx, "nlink")
	if v != 0 {
		v--
	}

	n.putUint32At(tx, "nlink", v)
}
