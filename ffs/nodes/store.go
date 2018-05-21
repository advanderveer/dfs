package nodes

import (
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	fdbdir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/billziss-gh/cgofuse/fuse"
)

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
		ino:  1,
	}
}

func (store *Store) IncIno(tx fdb.Transaction) {
	store.ino++ //@TODO protect by tr
}

func (store *Store) Ino(tx fdb.Transaction) uint64 {
	return store.ino //@TODO protect by tr
}

func (store *Store) Root(tx fdb.Transaction) *NodeT {
	return store.root //@TODO ptoject by tr
}

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
