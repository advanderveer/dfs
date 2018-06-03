package blocks

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

type Store struct {
	dir    string
	data   map[uint64][]byte //@TODO add lock
	db     *bolt.DB
	bucket []byte
}

func NewStore(dir string, ns string) (store *Store, err error) {
	store = &Store{
		dir:  dir,
		data: map[uint64][]byte{},
	}

	f, err := os.OpenFile(filepath.Join(store.dir, "data.gob"), os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	dec := gob.NewDecoder(f)
	err = dec.Decode(&store.data)
	if err != nil && err != io.EOF {
		return nil, err
	}

	fmt.Printf("loaded %d data\n", len(store.data))
	return store, nil
}

func (store *Store) Close() error {
	f, err := os.Create(filepath.Join(store.dir, "data.gob"))
	if err != nil {
		return err
	}

	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(store.data)
}

func (store *Store) Truncate(tx fdb.Transaction, node *nodes.Node, size int64) (errc int) {
	ino := node.StatGetIno(tx)
	store.data[ino] = resize(store.data[ino], size, true)
	return
}

func (store *Store) ReadAt(tx fdb.Transaction, node *nodes.Node, buff []byte, ofst int64) (n int) {
	endofst := ofst + int64(len(buff))
	if endofst > node.Stat(tx).Size {
		endofst = node.Stat(tx).Size
	}

	ino := node.StatGetIno(tx)
	n = copy(buff, store.data[ino][ofst:endofst])
	return
}

func (store *Store) WriteAt(tx fdb.Transaction, node *nodes.Node, buff []byte, ofst int64) (n int) {
	ino := node.StatGetIno(tx)
	endofst := ofst + int64(len(buff))
	if endofst > node.Stat(tx).Size {
		store.data[ino] = resize(store.data[ino], endofst, true)
		node.StatSetSize(tx, endofst)
	}

	n = copy(store.data[ino][ofst:endofst], buff)

	return
}

func resize(slice []byte, size int64, zeroinit bool) []byte {
	const allocunit = 64 * 1024
	allocsize := (size + allocunit - 1) / allocunit * allocunit
	if cap(slice) != int(allocsize) {
		var newslice []byte
		{
			defer func() {
				if r := recover(); nil != r {
					panic(fuse.Error(-fuse.ENOSPC))
				}
			}()
			newslice = make([]byte, size, allocsize)
		}
		copy(newslice, slice)
		slice = newslice
	} else if zeroinit {
		i := len(slice)
		slice = slice[:size]
		for ; len(slice) > i; i++ {
			slice[i] = 0
		}
	}
	return slice
}
