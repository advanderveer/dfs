package node

import (
	"fmt"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

var (
	//BucketNameNodes defines what database bucket is used
	BucketNameNodes = []byte(`nodes`)
)

//Store handles low-level node manipulation
type Store struct {
	db      *bolt.DB
	ino     uint64        //number of nodes in the tree
	root    *N            //root of the node tree
	openmap map[uint64]*N //open nodes
}

//NewStore creates a new node store
func NewStore(db *bolt.DB) (store *Store) {
	store = &Store{}
	store.ino++
	store.root = newNode(0, store.ino, fuse.S_IFDIR|00777, 0, 0)

	if err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketNameNodes)
		if b == nil || err != nil {
			return fmt.Errorf("failed to create nodes bucket: %+v", err)
		}

		return nil
	}); err != nil {
		panic("unable to setup node storage")
	}

	store.openmap = map[uint64]*N{}
	return store
}

//LookupNode looks up a new Node
func (store *Store) LookupNode(tx *bolt.Tx, path string, ancestor *N) (prnt *N, name string, node *N) {
	prnt = store.root
	name = ""
	node = store.root
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c

			node = node.ListChld(tx)[c]
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

//MakeNode creates a new node
func (store *Store) MakeNode(tx *bolt.Tx, path string, mode uint32, dev uint64, data []byte) int {
	prnt, name, node := store.LookupNode(tx, path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}
	store.ino++
	uid, gid, _ := fuse.Getcontext()
	node = newNode(dev, store.ino, mode, uid, gid)
	if nil != data {
		node.Data = make([]byte, len(data))
		node.Stat.Size = int64(len(data))
		copy(node.Data, data)
	}
	prnt.PutChld(tx, name, node)
	prnt.Stat.Ctim = node.Stat.Ctim
	prnt.Stat.Mtim = node.Stat.Ctim

	//prnt.Put()
	//node.Put()

	return 0
}

//RemoveNode removes a node
func (store *Store) RemoveNode(tx *bolt.Tx, path string, dir bool) int {
	prnt, name, node := store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.Stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.Stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR
	}
	if 0 < len(node.ListChld(tx)) {
		return -fuse.ENOTEMPTY
	}
	node.Stat.Nlink--
	prnt.DelChld(tx, name)
	tmsp := fuse.Now()
	node.Stat.Ctim = tmsp
	prnt.Stat.Ctim = tmsp
	prnt.Stat.Mtim = tmsp

	//node.Delete()
	//prnt.Put()

	return 0
}

//OpenNode opens a new node handle
func (store *Store) OpenNode(tx *bolt.Tx, path string, dir bool) (int, uint64) {
	_, _, node := store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT, ^uint64(0)
	}
	if !dir && fuse.S_IFDIR == node.Stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR, ^uint64(0)
	}
	if dir && fuse.S_IFDIR != node.Stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR, ^uint64(0)
	}
	node.opencnt++
	if 1 == node.opencnt {
		store.openmap[node.Stat.Ino] = node
	}
	return 0, node.Stat.Ino
}

//CloseNode closes a node handle
func (store *Store) CloseNode(tx *bolt.Tx, fh uint64) int {
	node := store.openmap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		delete(store.openmap, node.Stat.Ino)
	}
	return 0
}

//GetNode returns an open node handle
func (store *Store) GetNode(tx *bolt.Tx, path string, fh uint64) *N {
	if ^uint64(0) == fh {
		_, _, node := store.LookupNode(tx, path, nil)
		return node
	}

	return store.openmap[fh]
}
