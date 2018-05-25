package nodes

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	fdbdir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/billziss-gh/cgofuse/fuse"
)

type Store struct {
	tr   fdb.Transactor
	ss   fdbdir.DirectorySubspace
	root *Node
}

func NewStore(tr fdb.Transactor, ss fdbdir.DirectorySubspace) *Store {
	store := &Store{
		tr: tr,
		ss: ss,
	}

	if _, err := tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		store.root = store.NewNode(tx, 0, 0, fuse.S_IFDIR|00777, 0, 0) //@TODO check if its ok if the root always has ino0, tr: tr, ss: ss}
		ino := store.getIno(tx)
		if ino == 0 {
			store.setIno(tx, 1)
		}

		return
	}); err != nil {
		panic("ffs: failed to create root node")
	}

	return store
}

func (store *Store) getIno(tx fdb.Transaction) (ino uint64) {
	d := tx.Get(store.ss.Pack(tuple.Tuple{"ino"})).MustGet()
	if len(d) < 8 {
		return 0
	}

	return endianess.Uint64(d)
}

func (store *Store) setIno(tx fdb.Transaction, ino uint64) {
	b := make([]byte, 8)
	endianess.PutUint64(b, ino)
	tx.Set(store.ss.Pack(tuple.Tuple{"ino"}), b)
}

func (store *Store) NewNode(tx fdb.Transaction, dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *Node {
	node := Node{
		sss: store.ss,
		ss:  store.ss.Sub(int64(ino)),
	}

	node.Init(tx, dev, ino, mode, uid, gid)
	return &node
}

func (store *Store) IncIno(tx fdb.Transaction) {
	ino := store.getIno(tx)
	ino++
	store.setIno(tx, ino)
}

func (store *Store) Ino(tx fdb.Transaction) uint64 {
	return store.getIno(tx)
}

func (store *Store) Root(tx fdb.Transaction) *Node {
	return store.root
}

func (store *Store) TxWithInt(f func(tx fdb.Transaction) (n int)) (n int) {
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		n = f(tx)
		return
	}); err != nil {
		return 0 //@TODO log somewhere that the tx failed
	}

	return
}

func (store *Store) TxWithErrcBytes(f func(tx fdb.Transaction) (errc int, d []byte)) (errc int, d []byte) {
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc, d = f(tx)
		return
	}); err != nil {
		return -fuse.EIO, nil
	}

	return
}

func (store *Store) TxWithErrcUint64(f func(tx fdb.Transaction) (errc int, n uint64)) (errc int, n uint64) {
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc, n = f(tx)
		return
	}); err != nil {
		return -fuse.EIO, 0
	}

	return
}

func (store *Store) TxWithErrcStr(f func(tx fdb.Transaction) (errc int, str string)) (errc int, str string) {
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc, str = f(tx)
		return
	}); err != nil {
		return -fuse.EIO, ""
	}

	return
}

func (store *Store) TxWithErrc(f func(tx fdb.Transaction) (errc int)) (errc int) {
	if _, err := store.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		errc = f(tx)
		return
	}); err != nil {
		return -fuse.EIO
	}

	return
}
