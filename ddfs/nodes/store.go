package nodes

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

//Store manages nodes
type Store struct {
	ino     uint64
	root    *Node
	openmap map[uint64]*Node
	errs    chan<- error
	dbdir   string
	db      *bolt.DB
}

//NewStore sets up a new store
func NewStore(dbdir string, errs chan<- error) (store *Store, err error) {
	store = &Store{
		dbdir:   dbdir,
		errs:    errs,
		openmap: map[uint64]*Node{},
	}

	store.db, err = bolt.Open(filepath.Join(dbdir, "meta.db"), 0777, nil)
	if err != nil {
		return nil, err
	}

	if err = store.db.Update(func(tx *bolt.Tx) error {
		b, terr := tx.CreateBucketIfNotExists([]byte("nodes"))
		if terr != nil {
			return terr
		}

		//load root node
		store.root, terr = loadNode(b, 1)
		if terr != nil {
			if terr != errNodeNotExist {
				return terr
			}

			//or write a new one
			store.root = newNode(0, 1, fuse.S_IFDIR|00777, 0, 0)
			if terr = saveNode(b, store.root); terr != nil {
				return terr
			}
		}

		//find the last ino and continue from there
		b.ForEach(func(k, v []byte) error {
			store.ino = binary.BigEndian.Uint64(k)
			return nil
		})

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to setup nodes bucket: %v", err)
	}

	return store, nil
}

//View will start an read-only transaction
func (s *Store) View(fn func(tx Tx) int) (errc int) {
	if err := s.db.View(func(btx *bolt.Tx) error {
		errc = fn(&TxR{
			root:   s.root,
			bucket: btx.Bucket([]byte("nodes")),
		})
		return nil //@TODO test if we want fuse errors to rollback the tx?
	}); err != nil {
		s.errs <- fmt.Errorf("view tx failed: %v", err)
		return -fuse.EIO
	}

	return errc
}

//Update will start an writeable transaction
func (s *Store) Update(fn func(tx Tx) int) (errc int) {
	if err := s.db.Update(func(btx *bolt.Tx) error {
		tx := &TxRW{TxR: TxR{
			root:   s.root,
			bucket: btx.Bucket([]byte("nodes")),
		}}

		errc = fn(tx)
		for _, node := range tx.saves {
			err := saveNode(tx.bucket, node)
			if err != nil {
				return fmt.Errorf("failed to put node: '%d'", node.Ino())
			}
		}

		return nil //@TODO test if we want fuse errors to rollback the tx?
	}); err != nil {
		s.errs <- fmt.Errorf("update tx failed: %v", err)
		return -fuse.EIO
	}

	return errc
}

//Make will create a node
func (s *Store) Make(path string, mode uint32, dev uint64, link []byte) int {
	return s.Update(func(tx Tx) int {
		prnt, name, node := tx.Lookup(path, nil)
		if nil == prnt {
			return -fuse.ENOENT
		}
		if nil != node {
			return -fuse.EEXIST
		}

		s.ino++
		uid, gid, _ := fuse.Getcontext()
		node = newNode(dev, s.ino, mode, uid, gid)
		if nil != link {
			node.Link = make([]byte, len(link))
			node.Stat.Size = int64(len(link))
			copy(node.Link, link)
		}
		prnt.PutChild(name, node)
		prnt.Stat.Ctim = node.Stat.Ctim
		prnt.Stat.Mtim = node.Stat.Ctim

		tx.Save(prnt, node)
		return 0
	})
}

//Remove will remove a node
func (s *Store) Remove(path string, dir bool) int {
	return s.Update(func(tx Tx) int {
		prnt, name, node := tx.Lookup(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if !dir && fuse.S_IFDIR == node.Stat.Mode&fuse.S_IFMT {
			return -fuse.EISDIR
		}
		if dir && fuse.S_IFDIR != node.Stat.Mode&fuse.S_IFMT {
			return -fuse.ENOTDIR
		}

		if 0 < len(node.chlds) {
			return -fuse.ENOTEMPTY
		}

		node.Stat.Nlink--
		prnt.DelChild(name)
		tmsp := fuse.Now()
		node.Stat.Ctim = tmsp
		prnt.Stat.Ctim = tmsp
		prnt.Stat.Mtim = tmsp

		tx.Save(node, prnt)
		return 0
	})
}

//Open will setup a new node handle
func (s *Store) Open(path string, dir bool) (errc int, fh uint64) {
	errc = s.View(func(tx Tx) int {
		_, _, node := tx.Lookup(path, nil)
		if nil == node {
			fh = ^uint64(0)
			return -fuse.ENOENT
		}
		if !dir && fuse.S_IFDIR == node.Stat.Mode&fuse.S_IFMT {
			fh = ^uint64(0)
			return -fuse.EISDIR
		}
		if dir && fuse.S_IFDIR != node.Stat.Mode&fuse.S_IFMT {
			fh = ^uint64(0)
			return -fuse.ENOTDIR
		}
		node.opencnt++
		if 1 == node.opencnt {

			//open a backed file
			var err error
			if node.handle, err = os.OpenFile(
				filepath.Join(s.dbdir, fmt.Sprintf("%d", node.Ino())),
				os.O_CREATE|os.O_RDWR,
				0777, //@TODO what kind of do we want for backend file permissions?
			); err != nil {
				s.errs <- err
				fh = ^uint64(0)
				return -fuse.EIO
			}

			s.openmap[node.Ino()] = node
		}

		fh = node.Ino()
		return 0
	})

	return errc, fh
}

//Close will close a node
func (s *Store) Close(fh uint64) int {
	node := s.openmap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		delete(s.openmap, node.Ino())

		if node.handle == nil {
			s.errs <- fmt.Errorf("node '%d' has no file handle upon closing", fh)
			return -fuse.EIO
		}

		if err := node.handle.Close(); err != nil {
			s.errs <- fmt.Errorf("failed to close node handle: %v", err)
			return -fuse.EIO
		}

		node.handle = nil
	}

	return 0
}

//Get will lookup or get an open node
func (s *Store) Get(tx Tx, path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := tx.Lookup(path, nil)
		return node
	}

	return s.openmap[fh]
}
