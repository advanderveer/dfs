package dfs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/billziss-gh/cgofuse/examples/shared"
	"github.com/billziss-gh/cgofuse/fuse"
)

const appleResForkAttr = "com.apple.ResourceFork"

func trace(vals ...interface{}) func(vals ...interface{}) {
	uid, gid, _ := fuse.Getcontext()
	return shared.Trace(1, fmt.Sprintf("[uid=%v,gid=%v]", uid, gid), vals...)
}

func split(path string) []string {
	return strings.Split(path, "/")
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

//Memfs is an in-memory userland filesystem (FUSE) implementation that works on OSX, Linux and Windows
type Memfs struct {
	fuse.FileSystemBase
	lock  sync.Mutex //@TODO change this into an interface
	store *NodeStore //@TODO change this into an interface
}

// Statfs gets file system statistics.
func (fs *Memfs) Statfs(path string, stat *fuse.Statfs_t) int {
	return -fuse.ENOSYS
}

// Mknod creates a file node.
func (fs *Memfs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	defer fs.synchronize()()
	return fs.store.makeNode(path, mode, dev, nil)
}

// Mkdir creates a directory.
func (fs *Memfs) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer fs.synchronize()()
	return fs.store.makeNode(path, fuse.S_IFDIR|(mode&07777), 0, nil)
}

// Unlink removes a file.
func (fs *Memfs) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	defer fs.synchronize()()
	return fs.store.removeNode(path, false)
}

// Rmdir removes a directory.
func (fs *Memfs) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	defer fs.synchronize()()
	return fs.store.removeNode(path, true)
}

// Link creates a hard link to a file.
func (fs *Memfs) Link(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	defer fs.synchronize()()
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

// Symlink creates a symbolic link.
func (fs *Memfs) Symlink(target string, newpath string) (errc int) {
	defer trace(target, newpath)(&errc)
	defer fs.synchronize()()
	return fs.store.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
}

// Readlink reads the target of a symbolic link.
func (fs *Memfs) Readlink(path string) (errc int, target string) {
	defer trace(path)(&errc, &target)
	defer fs.synchronize()()
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, ""
	}
	if fuse.S_IFLNK != node.stat.Mode&fuse.S_IFMT {
		return -fuse.EINVAL, ""
	}
	return 0, string(node.data)
}

// Rename renames a file.
func (fs *Memfs) Rename(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	defer fs.synchronize()()
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

// Chmod changes the permission bits of a file.
func (fs *Memfs) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.store.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Mode = (node.stat.Mode & fuse.S_IFMT) | mode&07777
	node.stat.Ctim = fuse.Now()
	return 0
}

// Chown changes the owner and group of a file.
func (fs *Memfs) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	defer fs.synchronize()()
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

// Utimens changes the access and modification times of a file.
func (fs *Memfs) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer fs.synchronize()()
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

// Open opens a file.
func (fs *Memfs) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer fs.synchronize()()
	return fs.store.openNode(path, false)
}

// Getattr gets file attributes.
func (fs *Memfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer fs.synchronize()()
	node := fs.store.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

// Truncate changes the size of a file.
func (fs *Memfs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	defer fs.synchronize()()
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

// Read reads data from a file.
func (fs *Memfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer fs.synchronize()()
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

// Write writes data to a file.
func (fs *Memfs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer fs.synchronize()()
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

// Release closes an open file.
func (fs *Memfs) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer fs.synchronize()()
	return fs.store.closeNode(fh)
}

// Opendir opens a directory.
func (fs *Memfs) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer fs.synchronize()()
	return fs.store.openNode(path, true)
}

// Readdir reads a directory.
func (fs *Memfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer trace(path, fill, ofst, fh)(&errc)
	defer fs.synchronize()()

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

// Releasedir closes an open directory.
func (fs *Memfs) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer fs.synchronize()()
	return fs.store.closeNode(fh)
}

// Setxattr sets extended attributes.
func (fs *Memfs) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	defer fs.synchronize()()
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

// Getxattr gets extended attributes.
func (fs *Memfs) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	defer fs.synchronize()()
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

// Removexattr removes extended attributes.
func (fs *Memfs) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	defer fs.synchronize()()
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

// Listxattr lists extended attributes.
func (fs *Memfs) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	defer fs.synchronize()()
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

func (fs *Memfs) synchronize() func() {
	fs.lock.Lock()
	return fs.lock.Unlock
}

//NewMemfs sets up the filesystem
func NewMemfs() *Memfs {
	fs := Memfs{}
	defer fs.synchronize()()
	fs.store = newNodeStore()
	return &fs
}
