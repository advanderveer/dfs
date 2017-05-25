package ddfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
)

const (
	xattrAppleResourceFork = "com.apple.ResourceFork"
)

//FS is an in-memory store
type FS struct {
	fuse.FileSystemBase
	lock    sync.Mutex
	ino     uint64
	root    *Node
	openmap map[uint64]*Node
	errs    chan error
	dbdir   string
}

//NewFS creates a new filesystem
func NewFS(dbdir string, errw io.Writer) (fs *FS, err error) {
	fs = &FS{}
	defer fs.synchronize()()
	fs.ino++
	fs.dbdir = dbdir
	fs.root = newNode(0, fs.ino, fuse.S_IFDIR|00777, 0, 0)
	fs.openmap = map[uint64]*Node{}
	fs.errs = make(chan error, 10)
	go func() {
		for err := range fs.errs {
			fmt.Fprintf(errw, "FSError: %v\n", err)
		}
	}()

	return fs, nil
}

//Mknod creates a file node
func (fs *FS) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	defer fs.synchronize()()
	return fs.makeNode(path, mode, dev, nil)
}

// Mkdir creates a directory.
func (fs *FS) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer fs.synchronize()()
	return fs.makeNode(path, fuse.S_IFDIR|(mode&07777), 0, nil)
}

// Unlink removes a file.
func (fs *FS) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	defer fs.synchronize()()
	return fs.removeNode(path, false)
}

// Rmdir removes a directory.
func (fs *FS) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	defer fs.synchronize()()
	return fs.removeNode(path, true)
}

// Link creates a hard link to a file.
func (fs *FS) Link(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	defer fs.synchronize()()
	_, _, oldnode := fs.lookupNode(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.lookupNode(newpath, nil)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if nil != newnode {
		return -fuse.EEXIST
	}

	oldnode.stat.Nlink++
	newprnt.PutChild(newname, oldnode)
	tmsp := fuse.Now()
	oldnode.stat.Ctim = tmsp
	newprnt.stat.Ctim = tmsp
	newprnt.stat.Mtim = tmsp
	return fs.writeNodePair(oldnode, newprnt)
}

// Symlink creates a symbolic link.
func (fs *FS) Symlink(target string, newpath string) (errc int) {
	defer trace(target, newpath)(&errc)
	defer fs.synchronize()()
	return fs.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
}

// Readlink reads the target of a symbolic link.
func (fs *FS) Readlink(path string) (errc int, target string) {
	defer trace(path)(&errc, &target)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, ""
	}
	if fuse.S_IFLNK != node.stat.Mode&fuse.S_IFMT {
		return -fuse.EINVAL, ""
	}

	return 0, string(node.link)
}

// Rename renames a file.
func (fs *FS) Rename(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	defer fs.synchronize()()
	oldprnt, oldname, oldnode := fs.lookupNode(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.lookupNode(newpath, oldnode)
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
		errc = fs.removeNode(newpath, fuse.S_IFDIR == oldnode.stat.Mode&fuse.S_IFMT)
		if 0 != errc {
			return errc
		}
	}

	oldprnt.DelChild(oldname)
	newprnt.PutChild(newname, oldnode)
	return fs.writeNodePair(oldprnt, newprnt)
}

// Chmod changes the permission bits of a file.
func (fs *FS) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Mode = (node.stat.Mode & fuse.S_IFMT) | mode&07777
	node.stat.Ctim = fuse.Now()
	return fs.writeNode(node)
}

// Chown changes the owner and group of a file.
func (fs *FS) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
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
	return fs.writeNode(node)
}

// Utimens changes the access and modification times of a file.
func (fs *FS) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
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
	return fs.writeNode(node)
}

// Open opens a file.
func (fs *FS) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer fs.synchronize()()
	return fs.openNode(path, false)
}

// Getattr gets file attributes.
func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer fs.synchronize()()
	node := fs.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

// Truncate changes the size of a file.
func (fs *FS) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	defer fs.synchronize()()
	node := fs.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}

	var err error
	if fh == ^uint64(0) {
		err = os.Truncate(filepath.Join(fs.dbdir, fmt.Sprintf("%d", node.stat.Ino)), size)
	} else {
		err = node.Truncate(size)
	}

	if err != nil {
		fs.errs <- fmt.Errorf("failed to Truncate: %v", err)
		return -fuse.EIO
	}

	node.stat.Size = size
	tmsp := fuse.Now()
	node.stat.Ctim = tmsp
	node.stat.Mtim = tmsp
	return fs.writeNode(node)
}

// Read reads data from a file.
func (fs *FS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer fs.synchronize()()
	node := fs.getNode(path, fh)
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

	n, err := node.ReadAt(buff, ofst)
	if err != nil && err != io.EOF {
		fs.errs <- fmt.Errorf("failed to ReadAt(path: %s, ofst: %d, fh: %d): %v", path, ofst, fh, err)
		return -fuse.EIO
	}

	node.stat.Atim = fuse.Now()
	if werrc := fs.writeNode(node); werrc != 0 {
		return werrc
	}

	return n
}

// Write writes data to a file.
func (fs *FS) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer fs.synchronize()()
	node := fs.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}

	n, err := node.WriteAt(buff, ofst)
	if err != nil {
		fs.errs <- fmt.Errorf("failed to WriteAt: %v", err)
		return -fuse.EIO
	}

	endofst := ofst + int64(len(buff))
	if n > 0 && endofst > node.stat.Size {
		node.stat.Size = endofst
	}

	tmsp := fuse.Now()
	node.stat.Ctim = tmsp
	node.stat.Mtim = tmsp
	if werrc := fs.writeNode(node); werrc != 0 {
		return werrc
	}

	return
}

// Release closes an open file.
func (fs *FS) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer fs.synchronize()()
	return fs.closeNode(fh)
}

// Opendir opens a directory.
func (fs *FS) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer fs.synchronize()()
	return fs.openNode(path, true)
}

// Readdir reads a directory.
func (fs *FS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer trace(path, fill, ofst, fh)(&errc)
	defer fs.synchronize()()
	node := fs.openmap[fh]
	fill(".", &node.stat, 0)
	fill("..", nil, 0)

	node.EachChild(func(name string, chld *Node) bool {
		if !fill(name, &chld.stat, 0) {
			return false
		}
		return true
	})

	return 0
}

// Releasedir closes an open directory.
func (fs *FS) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer fs.synchronize()()
	return fs.closeNode(fh)
}

// Setxattr sets extended attributes.
func (fs *FS) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if xattrAppleResourceFork == name {
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
	return fs.writeNode(node)
}

// Getxattr gets extended attributes.
func (fs *FS) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, nil
	}
	if xattrAppleResourceFork == name {
		return -fuse.ENOTSUP, nil
	}
	xatr, ok := node.xatr[name]
	if !ok {
		return -fuse.ENOATTR, nil
	}
	return 0, xatr
}

// Removexattr removes extended attributes.
func (fs *FS) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if xattrAppleResourceFork == name {
		return -fuse.ENOTSUP
	}
	if _, ok := node.xatr[name]; !ok {
		return -fuse.ENOATTR
	}
	delete(node.xatr, name)
	return fs.writeNode(node)
}

// Listxattr lists extended attributes.
func (fs *FS) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.lookupNode(path, nil)
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
