package nodes

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

var store = map[uint64]nodeData{}

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

	store.Read(store.ino, store.root)
	if errc := store.Write(store.root); errc != 0 {
		return nil, fmt.Errorf("failed to write root: %v", err)
	}

	return store, nil
}

//WritePair persist one ore none of the provided nodes
func (s *Store) WritePair(nodeA *Node, nodeB *Node) int {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		//A
		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(nodeA.nodeData)
		if err != nil {
			return err
		}

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, nodeA.Ino())

		err = b.Put(key, buf.Bytes())
		if err != nil {
			return err
		}

		//A
		buf = bytes.NewBuffer(nil)
		enc = gob.NewEncoder(buf)
		err = enc.Encode(nodeA.nodeData)
		if err != nil {
			return err
		}

		key = make([]byte, 8)
		binary.BigEndian.PutUint64(key, nodeA.Ino())
		return b.Put(key, buf.Bytes())

		//@TODO serialize, put
		// store[nodeA.Ino()] = nodeA.nodeData
		// store[nodeB.Ino()] = nodeB.nodeData
		// return nil
	}); err != nil {
		s.errs <- fmt.Errorf("failed to write node pair: %v", err)
		return -fuse.EIO
	}

	return 0
}

//Write persists a single node
func (s *Store) Write(node *Node) int {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(node.nodeData)
		if err != nil {
			return err
		}

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, node.Ino())

		// defer fmt.Printf("wrote '%d', children: %#v \n", node.Ino(), node.Chld)

		b := tx.Bucket([]byte("nodes"))
		return b.Put(key, buf.Bytes())

		// store[node.Ino()] = node.nodeData
		// return nil
	}); err != nil {
		s.errs <- fmt.Errorf("failed to write node: %v", err)
		return -fuse.EIO
	}

	return 0
}

//Read updates node with persisted data
func (s *Store) Read(ino uint64, node *Node) {
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, ino)

		data := b.Get(key)
		if data == nil {
			return fmt.Errorf("couldn find node '%d'", ino)
		}

		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&node.nodeData)
		if err != nil {
			return err
		}

		// defer fmt.Printf("read '%d', children: %#v \n", node.Ino(), node.Chld)

		//get, deserialize
		// node.nodeData, _ = store[ino]
		return nil
	}); err != nil {
		s.errs <- fmt.Errorf("failed to read node: %v", err)
	}
}

//Lookup fetches a node by path
func (s *Store) Lookup(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
	rchild := func(node *Node, name string) (n *Node) {
		ino, ok := node.Chld[name]
		if ok {
			n, ok = nodes[ino]
			if ok {
				s.Read(ino, n)
			}
		}
		return n
	}

	prnt = s.root
	name = ""
	node = s.root
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c

			node = rchild(node, name)
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

//Make will create a node
func (s *Store) Make(path string, mode uint32, dev uint64, link []byte) int {
	prnt, name, node := s.Lookup(path, nil)
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
	return s.WritePair(node, prnt)
}

//Remove will remove a node
func (s *Store) Remove(path string, dir bool) int {
	prnt, name, node := s.Lookup(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.Stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.Stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR
	}

	count := 0
	node.EachChild(func(_ string, _ *Node) bool {
		count++
		return true
	})

	if 0 < count {
		return -fuse.ENOTEMPTY
	}

	node.Stat.Nlink--
	prnt.DelChild(name)
	tmsp := fuse.Now()
	node.Stat.Ctim = tmsp
	prnt.Stat.Ctim = tmsp
	prnt.Stat.Mtim = tmsp
	return s.WritePair(node, prnt)
}

//Open will setup a new node handle
func (s *Store) Open(path string, dir bool) (int, uint64) {
	_, _, node := s.Lookup(path, nil)
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

		//open a backed file
		var err error
		if node.handle, err = os.OpenFile(
			filepath.Join(s.dbdir, fmt.Sprintf("%d", node.Ino())),
			os.O_CREATE|os.O_RDWR,
			0777, //@TODO what kind of do we want for backend file permissions?
		); err != nil {
			s.errs <- err
			return -fuse.EIO, ^uint64(0)
		}

		s.openmap[node.Ino()] = node
	}
	return 0, node.Ino()
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
func (s *Store) Get(path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := s.Lookup(path, nil)
		return node
	}

	return s.openmap[fh]
}
