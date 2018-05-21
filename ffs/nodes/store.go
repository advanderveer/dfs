package nodes

import (
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
)

type NodeT struct {
	stat    fuse.Stat_t
	xatr    map[string][]byte
	chld    map[string]*NodeT
	data    []byte
	opencnt int
}

func (n *NodeT) Stat() fuse.Stat_t               { return n.stat }
func (n *NodeT) StatIncNlink()                   { n.stat.Nlink++ }
func (n *NodeT) StatDecNlink()                   { n.stat.Nlink-- }
func (n *NodeT) StatSetCTim(t fuse.Timespec)     { n.stat.Ctim = t }
func (n *NodeT) StatSetMTim(t fuse.Timespec)     { n.stat.Mtim = t }
func (n *NodeT) StatSetATim(t fuse.Timespec)     { n.stat.Atim = t }
func (n *NodeT) StatSetBirthTim(t fuse.Timespec) { n.stat.Birthtim = t }
func (n *NodeT) StatSetMode(m uint32)            { n.stat.Mode = m }
func (n *NodeT) StatSetUid(uid uint32)           { n.stat.Uid = uid }
func (n *NodeT) StatSetGid(gid uint32)           { n.stat.Gid = gid }
func (n *NodeT) StatSetSize(len int64)           { n.stat.Size = len }
func (n *NodeT) StatSetFlags(f uint32)           { n.stat.Flags = f }

func (n *NodeT) SetChld(name string, nn *NodeT) { n.chld[name] = nn }
func (n *NodeT) DelChld(name string)            { delete(n.chld, name) }

func (n *NodeT) XAtrGet(name string) (a []byte, ok bool) { a, ok = n.xatr[name]; return }
func (n *NodeT) XAtrDel(name string)                     { delete(n.xatr, name) }
func (n *NodeT) XAtrSet(name string, xatr []byte) {
	if nil == n.xatr {
		n.xatr = map[string][]byte{}
	}
	n.xatr[name] = xatr
}

func (n *NodeT) XAtrEach(f func(name string) int) (errc int) {
	for name := range n.xatr {
		errc = f(name)
		if errc != 0 {
			return errc
		}
	}

	return 0
}

func (n *NodeT) CountChld() int64           { return int64(len(n.chld)) }
func (n *NodeT) GetChld(name string) *NodeT { return n.chld[name] }
func (n *NodeT) ChldEach(f func(name string, n *NodeT) bool) {
	for name, n := range n.chld {
		stop := f(name, n)
		if stop {
			return
		}
	}
}

func (n *NodeT) Data() []byte      { return n.data }
func (n *NodeT) SetData(d []byte)  { n.data = d }
func (n *NodeT) CopyData(d []byte) { copy(n.data, d) }

func (n *NodeT) Opencnt() int { return n.opencnt }
func (n *NodeT) IncOpencnt()  { n.opencnt++ }
func (n *NodeT) DecOpencnt()  { n.opencnt-- }

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

func NewStore(root *NodeT) *Store {
	return &Store{root: root}
}

func (store *Store) Transact() func() {
	store.lock.Lock()
	return func() {
		store.lock.Unlock()
	}
}

type Store struct {
	ino  uint64
	root *NodeT
	lock sync.Mutex
}

func (store *Store) IncIno()      { store.ino++ }       //@TODO protect by tr
func (store *Store) Ino() uint64  { return store.ino }  //@TODO protect by tr
func (store *Store) Root() *NodeT { return store.root } //@TODO ptoject by tr
