package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (n *NodeT) Stat(tx fdb.Transaction) fuse.Stat_t                 { return n.no.stat }
func (n *NodeT) statSetIno(tx fdb.Transaction, ino uint64)           { n.no.stat.Ino = ino }
func (n *NodeT) statSetDev(tx fdb.Transaction, dev uint64)           { n.no.stat.Dev = dev }
func (n *NodeT) StatIncNlink(tx fdb.Transaction)                     { n.no.stat.Nlink++ }
func (n *NodeT) StatDecNlink(tx fdb.Transaction)                     { n.no.stat.Nlink-- }
func (n *NodeT) StatSetCTim(tx fdb.Transaction, t fuse.Timespec)     { n.no.stat.Ctim = t }
func (n *NodeT) StatSetMTim(tx fdb.Transaction, t fuse.Timespec)     { n.no.stat.Mtim = t }
func (n *NodeT) StatSetATim(tx fdb.Transaction, t fuse.Timespec)     { n.no.stat.Atim = t }
func (n *NodeT) StatSetBirthTim(tx fdb.Transaction, t fuse.Timespec) { n.no.stat.Birthtim = t }
func (n *NodeT) StatSetMode(tx fdb.Transaction, m uint32)            { n.no.stat.Mode = m }
func (n *NodeT) StatSetUid(tx fdb.Transaction, uid uint32)           { n.no.stat.Uid = uid }
func (n *NodeT) StatSetGid(tx fdb.Transaction, gid uint32)           { n.no.stat.Gid = gid }
func (n *NodeT) StatSetSize(tx fdb.Transaction, len int64)           { n.no.stat.Size = len }
func (n *NodeT) StatSetFlags(tx fdb.Transaction, f uint32)           { n.no.stat.Flags = f }
