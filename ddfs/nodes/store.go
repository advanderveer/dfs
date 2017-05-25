package nodes

import (
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

	store.ino++
	store.root = newNode(0, store.ino, fuse.S_IFDIR|00777, 0, 0)
	if err = store.db.Update(func(tx *bolt.Tx) error {
		_, txerr := tx.CreateBucketIfNotExists([]byte("nodes"))
		if txerr != nil {
			return txerr
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to setup nodes bucket: %v", err)
	}

	// if errc := store.Read(store.ino, store.root); errc != 0 {
	// 	if errc != -fuse.ENOENT {
	// 		return nil, fmt.Errorf("failed to read root: %v", errc)
	// 	}
	//
	// 	if errc := store.Write(store.root); errc != 0 {
	// 		return nil, fmt.Errorf("failed to write root: %v", errc)
	// 	}
	// }

	return store, nil
}

//WritePair persist one ore none of the provided nodes
// func (s *Store) WritePair(nodeA *Node, nodeB *Node) int {
// 	if err := s.db.Update(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte("nodes"))
//
// 		//A
// 		buf := bytes.NewBuffer(nil)
// 		enc := gob.NewEncoder(buf)
// 		err := enc.Encode(nodeA.nodeData)
// 		if err != nil {
// 			return err
// 		}
//
// 		key := make([]byte, 8)
// 		binary.BigEndian.PutUint64(key, nodeA.Ino())
//
// 		err = b.Put(key, buf.Bytes())
// 		if err != nil {
// 			return err
// 		}
//
// 		//B
// 		buf = bytes.NewBuffer(nil)
// 		enc = gob.NewEncoder(buf)
// 		err = enc.Encode(nodeA.nodeData)
// 		if err != nil {
// 			return err
// 		}
//
// 		key = make([]byte, 8)
// 		binary.BigEndian.PutUint64(key, nodeA.Ino())
// 		return b.Put(key, buf.Bytes())
// 	}); err != nil {
// 		s.errs <- fmt.Errorf("failed to put pair of nodes: %v", err)
// 		return -fuse.EIO
// 	}
//
// 	return 0
// }

//Write persists a single node
// func (s *Store) Write(node *Node) int {
// 	if err := s.db.Update(func(tx *bolt.Tx) error {
// 		buf := bytes.NewBuffer(nil)
// 		enc := gob.NewEncoder(buf)
// 		err := enc.Encode(node.nodeData)
// 		if err != nil {
// 			return err
// 		}
//
// 		key := make([]byte, 8)
// 		binary.BigEndian.PutUint64(key, node.Ino())
//
// 		b := tx.Bucket([]byte("nodes"))
// 		return b.Put(key, buf.Bytes())
// 	}); err != nil {
// 		s.errs <- fmt.Errorf("failed to write node: %v", err)
// 		return -fuse.EIO
// 	}
//
// 	return 0
// }

//Read updates node with persisted data
// func (s *Store) Read(ino uint64, node *Node) int {
// 	ErrNodeNotExist := fmt.Errorf("no such node")
//
// 	if err := s.db.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte("nodes"))
//
// 		key := make([]byte, 8)
// 		binary.BigEndian.PutUint64(key, ino)
//
// 		data := b.Get(key)
// 		if data == nil {
// 			return ErrNodeNotExist
// 		}
//
// 		buf := bytes.NewBuffer(data)
// 		dec := gob.NewDecoder(buf)
// 		err := dec.Decode(&node.nodeData)
// 		if err != nil {
// 			return err
// 		}
//
// 		return nil
// 	}); err != nil {
// 		if err == ErrNodeNotExist {
// 			return -fuse.ENOENT
// 		}
//
// 		s.errs <- fmt.Errorf("failed to read node: %v", err)
// 		return -fuse.EIO
// 	}
//
// 	return 0
// }

//Iterate over children of a node
// func (s *Store) Iterate(node *Node, next func(name string, chld *Node) bool) {
// 	for name := range node.Chld {
// 		child := node.chlds[name]
// 		ok := next(name, child)
// 		if !ok {
// 			break
// 		}
// 	}
// }

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

		//@TODO serialize and save
		fmt.Println(tx.saves)

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
