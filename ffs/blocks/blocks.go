package blocks

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/boltdb/bolt"
)

type Store struct {
	dir    string
	blocks map[uint64][]byte //@TODO add lock
	db     *bolt.DB
	bucket []byte
}

func NewStore(dir string, ns string) (store *Store, err error) {
	store = &Store{
		dir:    dir,
		blocks: map[uint64][]byte{},
	}

	f, err := os.OpenFile(filepath.Join(store.dir, "blocks.gob"), os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	dec := gob.NewDecoder(f)
	err = dec.Decode(&store.blocks)
	if err != nil && err != io.EOF {
		return nil, err
	}

	fmt.Printf("loaded %d blocks\n", len(store.blocks))
	return store, nil
}

func (store *Store) Close() error {
	f, err := os.Create(filepath.Join(store.dir, "blocks.gob"))
	if err != nil {
		return err
	}

	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(store.blocks)
}

func (store *Store) ReadData(n *nodes.Node, tx fdb.Transaction) (d []byte) {
	return store.blocks[n.StatGetIno(tx)]
}

func (store *Store) WriteData(n *nodes.Node, tx fdb.Transaction, d []byte) {
	store.blocks[n.StatGetIno(tx)] = d
}

func (store *Store) CopyData(n *nodes.Node, tx fdb.Transaction, d []byte) {
	d2 := store.ReadData(n, tx)
	copy(d2, d)
}
