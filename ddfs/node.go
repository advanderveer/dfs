package ddfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/billziss-gh/cgofuse/fuse"
)

var store = map[uint64]*Node{}

//NodeData is the persisting part of a Node
type NodeData struct {
	stat fuse.Stat_t
	xatr map[string][]byte
	link []byte
	chld map[string]uint64
}

//Node represents a filesystem node
type Node struct {
	NodeData

	opencnt int
	handle  *os.File
}

func newNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *Node {
	tmsp := fuse.Now()
	fs := Node{
		NodeData{
			fuse.Stat_t{
				Dev:      dev,
				Ino:      ino,
				Mode:     mode,
				Nlink:    1,
				Uid:      uid,
				Gid:      gid,
				Atim:     tmsp,
				Mtim:     tmsp,
				Ctim:     tmsp,
				Birthtim: tmsp,
			},
			nil,
			nil,
			nil,
		},
		0,
		nil}
	if fuse.S_IFDIR == fs.stat.Mode&fuse.S_IFMT {
		fs.chld = map[string]uint64{}
	}
	return &fs
}

//Update starts a writable transaction on the node
func (node *Node) Update(fn func(n *NodeData) int) (errc int) {
	return fn(&node.NodeData)
}

//Read starts a readable transaction on the node
func (node *Node) Read(fn func(n NodeData) int) (errc int) {
	return fn(node.NodeData)
}

//EachChild calls next for each child, if it returns false it will stop
func (node *NodeData) EachChild(next func(name string, n *Node) bool) {
	for name, ino := range node.chld {
		node, _ := store[ino]
		ok := next(name, node)
		if !ok {
			break
		}
	}
}

//GetChild gets a child node by name or returns nil if it doesn't exist
func (node *NodeData) GetChild(name string) (n *Node) {
	ino, ok := node.chld[name]
	if ok {
		n, _ = store[ino]
	}

	return
}

//PutChild sets a child node by name
func (node *NodeData) PutChild(name string, n *Node) {
	node.chld[name] = n.stat.Ino
	store[n.stat.Ino] = n
}

//DelChild deletes a child node by name
func (node *NodeData) DelChild(name string) {
	ino, ok := node.chld[name]
	if ok {
		delete(node.chld, name)
		delete(store, ino)
	}
}

//ReadAt implements: https://godoc.org/os#File.ReadAt
func (node *Node) ReadAt(b []byte, off int64) (n int, err error) {
	return node.handle.ReadAt(b, off)
}

//WriteAt implements: https://godoc.org/os#File.WriteAt
func (node *Node) WriteAt(b []byte, off int64) (n int, err error) {
	return node.handle.WriteAt(b, off)
}

//Truncate implements: https://godoc.org/os#File.Truncate
func (node *Node) Truncate(size int64) error {
	return node.handle.Truncate(size)
}

func (fs *FS) lookupNode(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
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

func (fs *FS) makeNode(path string, mode uint32, dev uint64, link []byte) int {
	prnt, name, node := fs.lookupNode(path, nil)
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
		node.link = make([]byte, len(link))
		node.stat.Size = int64(len(link))
		copy(node.link, link)
	}
	prnt.PutChild(name, node)
	prnt.stat.Ctim = node.stat.Ctim
	prnt.stat.Mtim = node.stat.Ctim
	return 0
}

func (fs *FS) removeNode(path string, dir bool) int {
	prnt, name, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
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

	node.stat.Nlink--
	prnt.DelChild(name)
	tmsp := fuse.Now()
	node.stat.Ctim = tmsp
	prnt.stat.Ctim = tmsp
	prnt.stat.Mtim = tmsp
	return 0
}

func (fs *FS) openNode(path string, dir bool) (int, uint64) {
	_, _, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, ^uint64(0)
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR, ^uint64(0)
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR, ^uint64(0)
	}
	node.opencnt++
	if 1 == node.opencnt {

		//open a backed file
		var err error
		if node.handle, err = os.OpenFile(
			filepath.Join(fs.dbdir, fmt.Sprintf("%d", node.stat.Ino)),
			os.O_CREATE|os.O_RDWR,
			0777, //@TODO what kind of do we want for backend file permissions?
		); err != nil {
			fs.errs <- err
			return -fuse.EIO, ^uint64(0)
		}

		fs.openmap[node.stat.Ino] = node
	}
	return 0, node.stat.Ino
}

func (fs *FS) closeNode(fh uint64) int {
	node := fs.openmap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		delete(fs.openmap, node.stat.Ino)

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

func (fs *FS) getNode(path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := fs.lookupNode(path, nil)
		return node
	}

	return fs.openmap[fh]
}
