package dfs

import (
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

func (fs *FS) doMkNod(tx *bolt.Tx, path string, mode uint32, dev uint64) (errc int) {
	return fs.store.makeNode(path, mode, dev, nil)
}

func (fs *FS) doMkdir(tx *bolt.Tx, path string, mode uint32) (errc int) {
	return fs.store.makeNode(path, fuse.S_IFDIR|(mode&07777), 0, nil)
}

func (fs *FS) doUnlink(tx *bolt.Tx, path string) (errc int) {
	return fs.store.removeNode(path, false)
}

func (fs *FS) doRmdir(tx *bolt.Tx, path string) (errc int) {
	return fs.store.removeNode(path, true)
}

func (fs *FS) doLink(tx *bolt.Tx, oldpath string, newpath string) (errc int) {
	_, _, oldnode := fs.store.lookupNode(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.store.lookupNode(newpath, nil)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if nil != newnode {
		return -fuse.EEXIST
	}
	oldnode.stat.Nlink++
	newprnt.chld[newname] = oldnode
	tmsp := fuse.Now()
	oldnode.stat.Ctim = tmsp
	newprnt.stat.Ctim = tmsp
	newprnt.stat.Mtim = tmsp
	return 0
}

func (fs *FS) doSymlink(tx *bolt.Tx, target string, newpath string) (errc int) {
	return fs.store.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
}

func (fs *FS) doReadlink(tx *bolt.Tx, path string) (errc int, target string) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, ""
	}
	if fuse.S_IFLNK != node.stat.Mode&fuse.S_IFMT {
		return -fuse.EINVAL, ""
	}
	return 0, string(node.data)
}

func (fs *FS) doRename(tx *bolt.Tx, oldpath string, newpath string) (errc int) {
	oldprnt, oldname, oldnode := fs.store.lookupNode(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.store.lookupNode(newpath, oldnode)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if "" == newname {
		// guard against directory loop creation
		return -fuse.EINVAL
	}
	if oldprnt == newprnt && oldname == newname {
		return 0
	}
	if nil != newnode {
		errc = fs.store.removeNode(newpath, fuse.S_IFDIR == oldnode.stat.Mode&fuse.S_IFMT)
		if 0 != errc {
			return errc
		}
	}
	delete(oldprnt.chld, oldname)
	newprnt.chld[newname] = oldnode
	return 0
}

func (fs *FS) doChmod(tx *bolt.Tx, path string, mode uint32) (errc int) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Mode = (node.stat.Mode & fuse.S_IFMT) | mode&07777
	node.stat.Ctim = fuse.Now()
	return 0
}

func (fs *FS) doChown(tx *bolt.Tx, path string, uid uint32, gid uint32) (errc int) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if ^uint32(0) != uid {
		node.stat.Uid = uid
	}
	if ^uint32(0) != gid {
		node.stat.Gid = gid
	}
	node.stat.Ctim = fuse.Now()
	return 0
}

func (fs *FS) doUtimens(tx *bolt.Tx, path string, tmsp []fuse.Timespec) (errc int) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if nil == tmsp {
		tmsp0 := fuse.Now()
		tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
		tmsp = tmsa[:]
	}
	node.stat.Atim = tmsp[0]
	node.stat.Mtim = tmsp[1]
	return 0
}

func (fs *FS) doOpen(tx *bolt.Tx, path string, flags int) (errc int, fh uint64) {
	return fs.store.openNode(path, false)
}

func (fs *FS) doGetattr(tx *bolt.Tx, path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	node := fs.store.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

func (fs *FS) doTruncate(tx *bolt.Tx, path string, size int64, fh uint64) (errc int) {
	node := fs.store.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	node.data = resize(node.data, size, true)
	node.stat.Size = size
	tmsp := fuse.Now()
	node.stat.Ctim = tmsp
	node.stat.Mtim = tmsp
	return 0
}

func (fs *FS) doRead(tx *bolt.Tx, path string, buff []byte, ofst int64, fh uint64) (n int) {
	node := fs.store.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	endofst := ofst + int64(len(buff))
	if endofst > node.stat.Size {
		endofst = node.stat.Size
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, node.data[ofst:endofst])
	node.stat.Atim = fuse.Now()
	return
}

func (fs *FS) doWrite(tx *bolt.Tx, path string, buff []byte, ofst int64, fh uint64) (n int) {
	node := fs.store.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	endofst := ofst + int64(len(buff))
	if endofst > node.stat.Size {
		node.data = resize(node.data, endofst, true)
		node.stat.Size = endofst
	}
	n = copy(node.data[ofst:endofst], buff)
	tmsp := fuse.Now()
	node.stat.Ctim = tmsp
	node.stat.Mtim = tmsp
	return
}

func (fs *FS) doRelease(tx *bolt.Tx, path string, fh uint64) (errc int) {
	return fs.store.closeNode(fh)
}

func (fs *FS) doOpendir(tx *bolt.Tx, path string) (errc int, fh uint64) {
	return fs.store.openNode(path, true)
}
func (fs *FS) doReaddir(tx *bolt.Tx, path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	node := fs.store.getNode(path, fh)
	fill(".", &node.stat, 0)
	fill("..", nil, 0)
	for name, chld := range node.chld {
		if !fill(name, &chld.stat, 0) {
			break
		}
	}
	return 0
}

func (fs *FS) doReleasedir(tx *bolt.Tx, path string, fh uint64) (errc int) {
	return fs.store.closeNode(fh)
}

func (fs *FS) doSetxattr(tx *bolt.Tx, path string, name string, value []byte, flags int) (errc int) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if appleResForkAttr == name {
		return -fuse.ENOTSUP
	}
	if fuse.XATTR_CREATE == flags {
		if _, ok := node.xatr[name]; ok {
			return -fuse.EEXIST
		}
	} else if fuse.XATTR_REPLACE == flags {
		if _, ok := node.xatr[name]; !ok {
			return -fuse.ENOATTR
		}
	}
	xatr := make([]byte, len(value))
	copy(xatr, value)
	if nil == node.xatr {
		node.xatr = map[string][]byte{}
	}
	node.xatr[name] = xatr
	return 0
}

func (fs *FS) doGetxattr(tx *bolt.Tx, path string, name string) (errc int, xatr []byte) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, nil
	}
	if appleResForkAttr == name {
		return -fuse.ENOTSUP, nil
	}
	xatr, ok := node.xatr[name]
	if !ok {
		return -fuse.ENOATTR, nil
	}
	return 0, xatr
}

func (fs *FS) doRemovexattr(tx *bolt.Tx, path string, name string) (errc int) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if appleResForkAttr == name {
		return -fuse.ENOTSUP
	}
	if _, ok := node.xatr[name]; !ok {
		return -fuse.ENOATTR
	}
	delete(node.xatr, name)
	return 0
}

func (fs *FS) doListxattr(tx *bolt.Tx, path string, fill func(name string) bool) (errc int) {
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	for name := range node.xatr {
		if !fill(name) {
			return -fuse.ERANGE
		}
	}
	return 0
}
