package nodes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/billziss-gh/cgofuse/fuse"
)

var store = map[uint64]nodeData{}

//Store manages nodes
type Store struct {
	ino     uint64
	root    *Node
	openmap map[uint64]*Node
	errs    chan<- error
	dbdir   string
}

//NewStore sets up a new store
func NewStore(dbdir string, errs chan<- error) (store *Store, err error) {
	store = &Store{
		dbdir:   dbdir,
		errs:    errs,
		openmap: map[uint64]*Node{},
	}

	store.ino++
	store.root = newNode(0, store.ino, fuse.S_IFDIR|00777, 0, 0)
	return store, nil
}

//WriteNodePair persist one ore none of the provided nodes
func (fs *Store) WriteNodePair(nodeA *Node, nodeB *Node) int {
	store[nodeA.Ino()] = nodeA.nodeData
	store[nodeB.Ino()] = nodeB.nodeData
	return 0
}

//WriteNode persists a single node
func (fs *Store) WriteNode(node *Node) int {
	store[node.Ino()] = node.nodeData
	return 0
}

//LookupNode fetches a node by path
func (fs *Store) LookupNode(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
	prnt = fs.root
	name = ""
	node = fs.root
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c
			node = node.GetChild(c)
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

//MakeNode will create a node
func (fs *Store) MakeNode(path string, mode uint32, dev uint64, link []byte) int {
	prnt, name, node := fs.LookupNode(path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}
	fs.ino++
	uid, gid, _ := fuse.Getcontext()
	node = newNode(dev, fs.ino, mode, uid, gid)
	if nil != link {
		node.Link = make([]byte, len(link))
		node.Stat.Size = int64(len(link))
		copy(node.Link, link)
	}
	prnt.PutChild(name, node)
	prnt.Stat.Ctim = node.Stat.Ctim
	prnt.Stat.Mtim = node.Stat.Ctim
	return fs.WriteNodePair(node, prnt)
}

//RemoveNode will remove a node
func (fs *Store) RemoveNode(path string, dir bool) int {
	prnt, name, node := fs.LookupNode(path, nil)
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
	return fs.WriteNodePair(node, prnt)
}

//OpenNode will setup a new node handle
func (fs *Store) OpenNode(path string, dir bool) (int, uint64) {
	_, _, node := fs.LookupNode(path, nil)
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
			filepath.Join(fs.dbdir, fmt.Sprintf("%d", node.Ino())),
			os.O_CREATE|os.O_RDWR,
			0777, //@TODO what kind of do we want for backend file permissions?
		); err != nil {
			fs.errs <- err
			return -fuse.EIO, ^uint64(0)
		}

		fs.openmap[node.Ino()] = node
	}
	return 0, node.Ino()
}

//CloseNode will close a node
func (fs *Store) CloseNode(fh uint64) int {
	node := fs.openmap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		delete(fs.openmap, node.Ino())

		if node.handle == nil {
			fs.errs <- fmt.Errorf("node '%d' has no file handle upon closing", fh)
			return -fuse.EIO
		}

		if err := node.handle.Close(); err != nil {
			fs.errs <- fmt.Errorf("failed to close node handle: %v", err)
			return -fuse.EIO
		}

		node.handle = nil
	}
	return 0
}

//GetNode will lookup or get an open node
func (fs *Store) GetNode(path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := fs.LookupNode(path, nil)
		return node
	}

	return fs.openmap[fh]
}
