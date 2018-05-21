package dfs

import (
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

func (fs *FS) doMkNod(tx *bolt.Tx, path string, mode uint32, dev uint64) (errc int) {
	return fs.store.MakeNode(tx, path, mode, dev, nil)
}

func (fs *FS) doMkdir(tx *bolt.Tx, path string, mode uint32) (errc int) {
	return fs.store.MakeNode(tx, path, fuse.S_IFDIR|(mode&07777), 0, nil)
}

func (fs *FS) doUnlink(tx *bolt.Tx, path string) (errc int) {
	return fs.store.RemoveNode(tx, path, false)
}

func (fs *FS) doRmdir(tx *bolt.Tx, path string) (errc int) {
	return fs.store.RemoveNode(tx, path, true)
}

func (fs *FS) doLink(tx *bolt.Tx, oldpath string, newpath string) (errc int) {
	_, _, oldnode := fs.store.LookupNode(tx, oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.store.LookupNode(tx, newpath, nil)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if nil != newnode {
		return -fuse.EEXIST
	}

	//node manipulation
	oldnode.Stat.Nlink++
	newprnt.PutChld(tx, newname, oldnode)
	tmsp := fuse.Now()
	oldnode.Stat.Ctim = tmsp
	newprnt.Stat.Ctim = tmsp
	newprnt.Stat.Mtim = tmsp

	newprnt.Persist(tx)
	oldnode.Persist(tx)
	return 0
}

func (fs *FS) doSymlink(tx *bolt.Tx, target string, newpath string) (errc int) {
	return fs.store.MakeNode(tx, newpath, fuse.S_IFLNK|00777, 0, []byte(target))
}

func (fs *FS) doReadlink(tx *bolt.Tx, path string) (errc int, target string) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT, ""
	}
	if fuse.S_IFLNK != node.Stat.Mode&fuse.S_IFMT {
		return -fuse.EINVAL, ""
	}
	return 0, string(node.ReadData())
}

func (fs *FS) doRename(tx *bolt.Tx, oldpath string, newpath string) (errc int) {
	oldprnt, oldname, oldnode := fs.store.LookupNode(tx, oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.store.LookupNode(tx, newpath, oldnode)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if "" == newname {
		return -fuse.EINVAL // guard against directory loop creation
	}
	if oldprnt == newprnt && oldname == newname {
		return 0
	}

	//node manipulation
	if nil != newnode {
		errc = fs.store.RemoveNode(tx, newpath, fuse.S_IFDIR == oldnode.Stat.Mode&fuse.S_IFMT)
		if 0 != errc {
			return errc
		}
	}
	oldprnt.DelChld(tx, oldname)
	newprnt.PutChld(tx, newname, oldnode)

	newprnt.Persist(tx)
	oldprnt.Persist(tx)
	return 0
}

func (fs *FS) doChmod(tx *bolt.Tx, path string, mode uint32) (errc int) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}

	//node manipulation
	node.Stat.Mode = (node.Stat.Mode & fuse.S_IFMT) | mode&07777
	node.Stat.Ctim = fuse.Now()

	node.Persist(tx)
	return 0
}

func (fs *FS) doChown(tx *bolt.Tx, path string, uid uint32, gid uint32) (errc int) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if ^uint32(0) != uid {
		node.Stat.Uid = uid
	}
	if ^uint32(0) != gid {
		node.Stat.Gid = gid
	}
	node.Stat.Ctim = fuse.Now()

	node.Persist(tx)
	return 0
}

func (fs *FS) doUtimens(tx *bolt.Tx, path string, tmsp []fuse.Timespec) (errc int) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if nil == tmsp {
		tmsp0 := fuse.Now()
		tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
		tmsp = tmsa[:]
	}
	node.Stat.Atim = tmsp[0]
	node.Stat.Mtim = tmsp[1]

	node.Persist(tx)
	return 0
}

func (fs *FS) doOpen(tx *bolt.Tx, path string, flags int) (errc int, fh uint64) {
	return fs.store.OpenNode(tx, path, false)
}

func (fs *FS) doGetattr(tx *bolt.Tx, path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	node := fs.store.GetNode(tx, path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.Stat
	return 0
}

func (fs *FS) doTruncate(tx *bolt.Tx, path string, size int64, fh uint64) (errc int) {
	node := fs.store.GetNode(tx, path, fh)
	if nil == node {
		return -fuse.ENOENT
	}

	node.WriteData(fs.resize(node.ReadData(), size, true))
	node.Stat.Size = size
	tmsp := fuse.Now()
	node.Stat.Ctim = tmsp
	node.Stat.Mtim = tmsp

	node.Persist(tx)
	return 0
}

func (fs *FS) doRead(tx *bolt.Tx, path string, buff []byte, ofst int64, fh uint64) (n int) {
	node := fs.store.GetNode(tx, path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	endofst := ofst + int64(len(buff))
	if endofst > node.Stat.Size {
		endofst = node.Stat.Size
	}
	if endofst < ofst {
		return 0
	}

	n = copy(buff, node.ReadDataAt(ofst, endofst))
	node.Stat.Atim = fuse.Now()

	node.Persist(tx)
	return
}

func (fs *FS) doWrite(tx *bolt.Tx, path string, buff []byte, ofst int64, fh uint64) (n int) {
	node := fs.store.GetNode(tx, path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	endofst := ofst + int64(len(buff))
	if endofst > node.Stat.Size {
		node.WriteData(fs.resize(node.ReadData(), endofst, true))
		node.Stat.Size = endofst
	}

	n = copy(node.ReadDataAt(ofst, endofst), buff)
	tmsp := fuse.Now()
	node.Stat.Ctim = tmsp
	node.Stat.Mtim = tmsp

	node.Persist(tx)
	return
}

func (fs *FS) doRelease(tx *bolt.Tx, path string, fh uint64) (errc int) {
	return fs.store.CloseNode(tx, fh)
}

func (fs *FS) doOpendir(tx *bolt.Tx, path string) (errc int, fh uint64) {
	return fs.store.OpenNode(tx, path, true)
}
func (fs *FS) doReaddir(tx *bolt.Tx, path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	node := fs.store.GetNode(tx, path, fh)
	fill(".", &node.Stat, 0)
	fill("..", nil, 0)
	for name, chld := range node.ListChld(tx) {
		if !fill(name, &chld.Stat, 0) {
			break
		}
	}
	return 0
}

func (fs *FS) doReleasedir(tx *bolt.Tx, path string, fh uint64) (errc int) {
	return fs.store.CloseNode(tx, fh)
}

func (fs *FS) doSetxattr(tx *bolt.Tx, path string, name string, value []byte, flags int) (errc int) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if appleResForkAttr == name {
		return -fuse.ENOTSUP
	}
	if fuse.XATTR_CREATE == flags {
		if _, ok := node.Xatr[name]; ok {
			return -fuse.EEXIST
		}
	} else if fuse.XATTR_REPLACE == flags {
		if _, ok := node.Xatr[name]; !ok {
			return -fuse.ENOATTR
		}
	}
	xatr := make([]byte, len(value))
	copy(xatr, value)
	if nil == node.Xatr {
		node.Xatr = map[string][]byte{}
	}
	node.Xatr[name] = xatr

	node.Persist(tx)
	return 0
}

func (fs *FS) doGetxattr(tx *bolt.Tx, path string, name string) (errc int, xatr []byte) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT, nil
	}
	if appleResForkAttr == name {
		return -fuse.ENOTSUP, nil
	}
	xatr, ok := node.Xatr[name]
	if !ok {
		return -fuse.ENOATTR, nil
	}
	return 0, xatr
}

func (fs *FS) doRemovexattr(tx *bolt.Tx, path string, name string) (errc int) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if appleResForkAttr == name {
		return -fuse.ENOTSUP
	}
	if _, ok := node.Xatr[name]; !ok {
		return -fuse.ENOATTR
	}
	delete(node.Xatr, name)

	node.Persist(tx)
	return 0
}

func (fs *FS) doListxattr(tx *bolt.Tx, path string, fill func(name string) bool) (errc int) {
	_, _, node := fs.store.LookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	for name := range node.Xatr {
		if !fill(name) {
			return -fuse.ERANGE
		}
	}
	return 0
}
