package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

type NodeT struct {
	stat    fuse.Stat_t
	xatr    map[string][]byte
	chld    map[string]*NodeT
	data    []byte
	opencnt int
}

func (n *NodeT) Stat(tx fdb.Transaction) fuse.Stat_t                 { return n.stat }
func (n *NodeT) StatIncNlink(tx fdb.Transaction)                     { n.stat.Nlink++ }
func (n *NodeT) StatDecNlink(tx fdb.Transaction)                     { n.stat.Nlink-- }
func (n *NodeT) StatSetCTim(tx fdb.Transaction, t fuse.Timespec)     { n.stat.Ctim = t }
func (n *NodeT) StatSetMTim(tx fdb.Transaction, t fuse.Timespec)     { n.stat.Mtim = t }
func (n *NodeT) StatSetATim(tx fdb.Transaction, t fuse.Timespec)     { n.stat.Atim = t }
func (n *NodeT) StatSetBirthTim(tx fdb.Transaction, t fuse.Timespec) { n.stat.Birthtim = t }
func (n *NodeT) StatSetMode(tx fdb.Transaction, m uint32)            { n.stat.Mode = m }
func (n *NodeT) StatSetUid(tx fdb.Transaction, uid uint32)           { n.stat.Uid = uid }
func (n *NodeT) StatSetGid(tx fdb.Transaction, gid uint32)           { n.stat.Gid = gid }
func (n *NodeT) StatSetSize(tx fdb.Transaction, len int64)           { n.stat.Size = len }
func (n *NodeT) StatSetFlags(tx fdb.Transaction, f uint32)           { n.stat.Flags = f }

func (n *NodeT) SetChld(tx fdb.Transaction, name string, nn *NodeT) { n.chld[name] = nn }
func (n *NodeT) GetChld(tx fdb.Transaction, name string) *NodeT     { return n.chld[name] }
func (n *NodeT) DelChld(tx fdb.Transaction, name string)            { delete(n.chld, name) }
func (n *NodeT) CountChld(tx fdb.Transaction) int64                 { return int64(len(n.chld)) }
func (n *NodeT) ChldEach(tx fdb.Transaction, f func(name string, n *NodeT) bool) {
	for name, n := range n.chld {
		stop := f(name, n)
		if stop {
			return
		}
	}
}

func (n *NodeT) XAtrGet(tx fdb.Transaction, name string) (a []byte, ok bool) {
	a, ok = n.xatr[name]
	return
}
func (n *NodeT) XAtrDel(tx fdb.Transaction, name string) { delete(n.xatr, name) }
func (n *NodeT) XAtrSet(tx fdb.Transaction, name string, xatr []byte) {
	if nil == n.xatr {
		n.xatr = map[string][]byte{}
	}
	n.xatr[name] = xatr
}

func (n *NodeT) XAtrEach(tx fdb.Transaction, f func(name string) int) (errc int) {
	for name := range n.xatr {
		errc = f(name)
		if errc != 0 {
			return errc
		}
	}

	return 0
}

func (n *NodeT) Data(tx fdb.Transaction) []byte        { return n.data }
func (n *NodeT) SetData(tx fdb.Transaction, d []byte)  { n.data = d }
func (n *NodeT) CopyData(tx fdb.Transaction, d []byte) { copy(n.data, d) }

func (n *NodeT) Opencnt(tx fdb.Transaction) int { return n.opencnt }
func (n *NodeT) IncOpencnt(tx fdb.Transaction)  { n.opencnt++ }
func (n *NodeT) DecOpencnt(tx fdb.Transaction)  { n.opencnt-- }

func NewNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *NodeT {
	tmsp := fuse.Now()
	self := NodeT{
		fuse.Stat_t{
			Dev:      dev,
			Ino:      ino,
			Mode:     mode,
			Nlink:    1,
			Uid:      uid,
			Gid:      gid,
			Atim:     tmsp,
			Mtim:     tmsp,
			Ctim:     tmsp,
			Birthtim: tmsp,
			Flags:    0,
		},
		nil,
		nil,
		nil,
		0}
	if fuse.S_IFDIR == self.stat.Mode&fuse.S_IFMT {
		self.chld = map[string]*NodeT{}
	}
	return &self
}
