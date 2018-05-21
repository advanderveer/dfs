package blocks

import (
	"crypto/rand"
	"path/filepath"

	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/boltdb/bolt"
)

type Store struct {
	blocks map[uint64][]byte
	db     *bolt.DB
	bucket []byte
}

func NewStore(dir string, ns string) (store *Store, err error) {
	store = &Store{
		blocks: map[uint64][]byte{},
	}

	store.bucket = []byte(ns)
	if len(store.bucket) < 1 {
		store.bucket = make([]byte, 6)
		_, err = rand.Read(store.bucket)
		if err != nil {
			return nil, err
		}
	}

	store.db, err = bolt.Open(filepath.Join(dir, "blocks.bolt"), 0600, nil)
	if err != nil {
		return nil, err
	}

	return store, store.db.Update(func(btx *bolt.Tx) error {
		_, err := btx.CreateBucketIfNotExists(store.bucket)
		if err != nil {
			return err
		}

		return nil
	})
}

func (store *Store) ReadData(n *nodes.Node, tx fdb.Transaction) (d []byte) {
	// if err := store.db.View(func(btx *bolt.Tx) (err error) {
	// 	ino := n.StatGetIno(tx)
	//
	// 	k := make([]byte, 8)
	// 	binary.LittleEndian.PutUint64(k, ino)
	//
	// 	v := btx.Bucket(store.bucket).Get(k)
	// 	d = make([]byte, len(v))
	// 	copy(d, v)
	//
	// 	return
	// }); err != nil {
	// 	fmt.Print("blocks: failed to read data:", err)
	// }

	// return

	return store.blocks[n.StatGetIno(tx)]
}

func (store *Store) WriteData(n *nodes.Node, tx fdb.Transaction, d []byte) {
	// if err := store.db.Update(func(btx *bolt.Tx) (err error) {
	// 	b, _ := btx.CreateBucketIfNotExists(store.bucket)
	// 	ino := n.StatGetIno(tx)
	//
	// 	k := make([]byte, 8)
	// 	binary.LittleEndian.PutUint64(k, ino)
	// 	return b.Put(k, d)
	// }); err != nil {
	// 	fmt.Print("blocks: failed to read data:", err)
	// }

	store.blocks[n.StatGetIno(tx)] = d
}

func (store *Store) CopyData(n *nodes.Node, tx fdb.Transaction, d []byte) {
	d2 := store.ReadData(n, tx)
	copy(d2, d)
	// store.WriteData(n, tx, d2)
}
