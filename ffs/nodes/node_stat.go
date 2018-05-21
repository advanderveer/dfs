package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (n Node) statGetIno(tx fdb.Transaction) (ino uint64) { return n.getUint64At(tx, "ino") }

//@TODO can we reduce the number of reads if we do not need the whol state
func (n Node) Stat(tx fdb.Transaction) fuse.Stat_t {
	sta := fuse.Stat_t{}
	sta.Ino = n.statGetIno(tx)
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

func (n Node) statSetIno(tx fdb.Transaction, ino uint64)           { n.putUint64At(tx, "ino", ino) }
func (n Node) statSetDev(tx fdb.Transaction, dev uint64)           { n.putUint64At(tx, "dev", dev) }
func (n Node) StatSetMode(tx fdb.Transaction, m uint32)            { n.putUint32At(tx, "mode", m) }
func (n Node) StatSetUid(tx fdb.Transaction, uid uint32)           { n.putUint32At(tx, "uid", uid) }
func (n Node) StatSetGid(tx fdb.Transaction, gid uint32)           { n.putUint32At(tx, "gid", gid) }
func (n Node) StatSetFlags(tx fdb.Transaction, f uint32)           { n.putUint32At(tx, "flags", f) }
func (n Node) StatSetSize(tx fdb.Transaction, len int64)           { n.putInt64At(tx, "size", len) }
func (n Node) StatSetCTim(tx fdb.Transaction, t fuse.Timespec)     { n.putTimeSpec(tx, "ctim", t) }
func (n Node) StatSetMTim(tx fdb.Transaction, t fuse.Timespec)     { n.putTimeSpec(tx, "mtim", t) }
func (n Node) StatSetATim(tx fdb.Transaction, t fuse.Timespec)     { n.putTimeSpec(tx, "atim", t) }
func (n Node) StatSetBirthTim(tx fdb.Transaction, t fuse.Timespec) { n.putTimeSpec(tx, "birthtim", t) }

//@TODO use the tx.Add method instead
func (n Node) StatIncNlink(tx fdb.Transaction) {
	v := n.getUint32At(tx, "nlink")
	v++
	n.putUint32At(tx, "nlink", v)
}

//@TODO use the tx.Add method instead
func (n Node) StatDecNlink(tx fdb.Transaction) {
	v := n.getUint32At(tx, "nlink")
	if v != 0 {
		v--
	}

	n.putUint32At(tx, "nlink", v)
}
