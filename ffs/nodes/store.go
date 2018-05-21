package nodes

import (
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	fdbdir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
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
func (n *NodeT) GetChld(name string) *NodeT     { return n.chld[name] }
func (n *NodeT) DelChld(name string)            { delete(n.chld, name) }
func (n *NodeT) CountChld() int64               { return int64(len(n.chld)) }
func (n *NodeT) ChldEach(f func(name string, n *NodeT) bool) {
	for name, n := range n.chld {
		stop := f(name, n)
		if stop {
			return
		}
	}
}

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

type Store struct {
	tr   fdb.Transactor
	ss   fdbdir.DirectorySubspace
	ino  uint64
	root *NodeT
	lock sync.Mutex
}

func NewStore(tr fdb.Transactor, ss fdbdir.DirectorySubspace) *Store {
	return &Store{
		root: NewNode(0, 0, fuse.S_IFDIR|00777, 0, 0), //@TODO check if its ok if the root always has ino0, tr: tr, ss: ss}
		tr:   tr,
		ss:   ss,
		ino:  1, //@TODO root is 0
	}
}

func (store *Store) IncIno()      { store.ino++ }       //@TODO protect by tr
func (store *Store) Ino() uint64  { return store.ino }  //@TODO protect by tr
func (store *Store) Root() *NodeT { return store.root } //@TODO ptoject by tr

func (store *Store) TxWithInt(f func(tx fdb.Transaction) (n int)) (n int) {
	store.lock.Lock()
	defer store.lock.Unlock()
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		n = f(tx)
		return
	}); err != nil {
		return 0 //@TODO log somewhere that the tx failed
	}

	return
}

func (store *Store) TxWithErrcBytes(f func(tx fdb.Transaction) (errc int, d []byte)) (errc int, d []byte) {
	store.lock.Lock()
	defer store.lock.Unlock()
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc, d = f(tx)
		return
	}); err != nil {
		return -fuse.EIO, nil
	}

	return
}

func (store *Store) TxWithErrcUint64(f func(tx fdb.Transaction) (errc int, n uint64)) (errc int, n uint64) {
	store.lock.Lock()
	defer store.lock.Unlock()
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc, n = f(tx)
		return
	}); err != nil {
		return -fuse.EIO, 0
	}

	return
}

func (store *Store) TxWithErrcStr(f func(tx fdb.Transaction) (errc int, str string)) (errc int, str string) {
	store.lock.Lock()
	defer store.lock.Unlock()
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc, str = f(tx)
		return
	}); err != nil {
		return -fuse.EIO, ""
	}

	return
}

func (store *Store) TxWithErrc(f func(tx fdb.Transaction) (errc int)) (errc int) {
	store.lock.Lock()
	defer store.lock.Unlock()
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc = f(tx)
		return
	}); err != nil {
		return -fuse.EIO
	}

	return
}
