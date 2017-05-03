package dfs

import (
	"github.com/billziss-gh/cgofuse/fuse"
)

//NodeStore handles low-level node manipulation
type NodeStore struct {
	ino     uint64            //number of nodes in the tree
	root    *nodeT            //root of the node tree
	openmap map[uint64]*nodeT //open nodes
}

func newNodeStore() (store *NodeStore) {
	store = &NodeStore{}
	store.ino++
	store.root = newNode(0, store.ino, fuse.S_IFDIR|00777, 0, 0)
	store.openmap = map[uint64]*nodeT{}
	return store
}

func (store *NodeStore) lookupNode(path string, ancestor *nodeT) (prnt *nodeT, name string, node *nodeT) {
	prnt = store.root
	name = ""
	node = store.root
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c
			node = node.chld[c]
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

func (store *NodeStore) makeNode(path string, mode uint32, dev uint64, data []byte) int {
	prnt, name, node := store.lookupNode(path, nil)
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
		node.data = make([]byte, len(data))
		node.stat.Size = int64(len(data))
		copy(node.data, data)
	}
	prnt.chld[name] = node
	prnt.stat.Ctim = node.stat.Ctim
	prnt.stat.Mtim = node.stat.Ctim
	return 0
}

func (store *NodeStore) removeNode(path string, dir bool) int {
	prnt, name, node := store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR
	}
	if 0 < len(node.chld) {
		return -fuse.ENOTEMPTY
	}
	node.stat.Nlink--
	delete(prnt.chld, name)
	tmsp := fuse.Now()
	node.stat.Ctim = tmsp
	prnt.stat.Ctim = tmsp
	prnt.stat.Mtim = tmsp
	return 0
}

func (store *NodeStore) openNode(path string, dir bool) (int, uint64) {
	_, _, node := store.lookupNode(path, nil)
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
		store.openmap[node.stat.Ino] = node
	}
	return 0, node.stat.Ino
}

func (store *NodeStore) closeNode(fh uint64) int {
	node := store.openmap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		delete(store.openmap, node.stat.Ino)
	}
	return 0
}

func (store *NodeStore) getNode(path string, fh uint64) *nodeT {
	if ^uint64(0) == fh {
		_, _, node := store.lookupNode(path, nil)
		return node
	}

	return store.openmap[fh]
}
