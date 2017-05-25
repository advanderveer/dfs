package ddfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/advanderveer/dfs/ddfs/nodes"
	"github.com/billziss-gh/cgofuse/fuse"
)

const (
	xattrAppleResourceFork = "com.apple.ResourceFork"
)

//FS is an in-memory store
type FS struct {
	fuse.FileSystemBase
	lock sync.Mutex
	// ino  uint64
	// root    *Node
	// openmap map[uint64]*Node
	errs  chan error
	dbdir string
	nodes *nodes.Store
}

//NewFS creates a new filesystem
func NewFS(dbdir string, errw io.Writer) (fs *FS, err error) {
	fs = &FS{}
	defer fs.synchronize()()
	// fs.ino++
	fs.dbdir = dbdir
	// fs.root = newNode(0, fs.ino, fuse.S_IFDIR|00777, 0, 0)
	// fs.openmap = map[uint64]*Node{}
	fs.errs = make(chan error, 10)
	go func() {
		for err := range fs.errs {
			fmt.Fprintf(errw, "FSError: %v\n", err)
		}
	}()

	fs.nodes, err = nodes.NewStore(dbdir, fs.errs)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

//Mknod creates a file node
func (fs *FS) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	defer fs.synchronize()()
	return fs.nodes.Make(path, mode, dev, nil)
}

// Mkdir creates a directory.
func (fs *FS) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer fs.synchronize()()
	return fs.nodes.Make(path, fuse.S_IFDIR|(mode&07777), 0, nil)
}

// Unlink removes a file.
func (fs *FS) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	defer fs.synchronize()()
	return fs.nodes.Remove(path, false)
}

// Rmdir removes a directory.
func (fs *FS) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	defer fs.synchronize()()
	return fs.nodes.Remove(path, true)
}

// Link creates a hard link to a file.
func (fs *FS) Link(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	defer fs.synchronize()()
	_, _, oldnode := fs.nodes.Lookup(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.nodes.Lookup(newpath, nil)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if nil != newnode {
		return -fuse.EEXIST
	}

	oldnode.Stat.Nlink++
	newprnt.PutChild(newname, oldnode)
	tmsp := fuse.Now()
	oldnode.Stat.Ctim = tmsp
	newprnt.Stat.Ctim = tmsp
	newprnt.Stat.Mtim = tmsp
	return fs.nodes.WritePair(oldnode, newprnt)
}

// Symlink creates a symbolic link.
func (fs *FS) Symlink(target string, newpath string) (errc int) {
	defer trace(target, newpath)(&errc)
	defer fs.synchronize()()
	return fs.nodes.Make(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
}

// Readlink reads the target of a symbolic link.
func (fs *FS) Readlink(path string) (errc int, target string) {
	defer trace(path)(&errc, &target)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
	if nil == node {
		return -fuse.ENOENT, ""
	}
	if fuse.S_IFLNK != node.Stat.Mode&fuse.S_IFMT {
		return -fuse.EINVAL, ""
	}

	return 0, string(node.Link)
}

// Rename renames a file.
func (fs *FS) Rename(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	defer fs.synchronize()()
	oldprnt, oldname, oldnode := fs.nodes.Lookup(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := fs.nodes.Lookup(newpath, oldnode)
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
		errc = fs.nodes.Remove(newpath, fuse.S_IFDIR == oldnode.Stat.Mode&fuse.S_IFMT)
		if 0 != errc {
			return errc
		}
	}

	oldprnt.DelChild(oldname)
	newprnt.PutChild(newname, oldnode)
	return fs.nodes.WritePair(oldprnt, newprnt)
}

// Chmod changes the permission bits of a file.
func (fs *FS) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.Stat.Mode = (node.Stat.Mode & fuse.S_IFMT) | mode&07777
	node.Stat.Ctim = fuse.Now()
	return fs.nodes.Write(node)
}

// Chown changes the owner and group of a file.
func (fs *FS) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
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
	return fs.nodes.Write(node)
}

// Utimens changes the access and modification times of a file.
func (fs *FS) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
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
	return fs.nodes.Write(node)
}

// Open opens a file.
func (fs *FS) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer fs.synchronize()()
	return fs.nodes.Open(path, false)
}

// Getattr gets file attributes.
func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer fs.synchronize()()
	node := fs.nodes.Get(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.Stat
	return 0
}

// Truncate changes the size of a file.
func (fs *FS) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	defer fs.synchronize()()
	node := fs.nodes.Get(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}

	var err error
	if fh == ^uint64(0) {
		err = os.Truncate(filepath.Join(fs.dbdir, fmt.Sprintf("%d", node.Stat.Ino)), size)
	} else {
		err = node.Truncate(size)
	}

	if err != nil {
		fs.errs <- fmt.Errorf("failed to Truncate: %v", err)
		return -fuse.EIO
	}

	node.Stat.Size = size
	tmsp := fuse.Now()
	node.Stat.Ctim = tmsp
	node.Stat.Mtim = tmsp
	return fs.nodes.Write(node)
}

// Read reads data from a file.
func (fs *FS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer fs.synchronize()()
	node := fs.nodes.Get(path, fh)
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

	n, err := node.ReadAt(buff, ofst)
	if err != nil && err != io.EOF {
		fs.errs <- fmt.Errorf("failed to ReadAt(path: %s, ofst: %d, fh: %d): %v", path, ofst, fh, err)
		return -fuse.EIO
	}

	node.Stat.Atim = fuse.Now()
	if werrc := fs.nodes.Write(node); werrc != 0 {
		return werrc
	}

	return n
}

// Write writes data to a file.
func (fs *FS) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer fs.synchronize()()
	node := fs.nodes.Get(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}

	n, err := node.WriteAt(buff, ofst)
	if err != nil {
		fs.errs <- fmt.Errorf("failed to WriteAt: %v", err)
		return -fuse.EIO
	}

	endofst := ofst + int64(len(buff))
	if n > 0 && endofst > node.Stat.Size {
		node.Stat.Size = endofst
	}

	tmsp := fuse.Now()
	node.Stat.Ctim = tmsp
	node.Stat.Mtim = tmsp
	if werrc := fs.nodes.Write(node); werrc != 0 {
		return werrc
	}

	return
}

// Release closes an open file.
func (fs *FS) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer fs.synchronize()()
	return fs.nodes.Close(fh)
}

// Opendir opens a directory.
func (fs *FS) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer fs.synchronize()()
	return fs.nodes.Open(path, true)
}

// Readdir reads a directory.
func (fs *FS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer trace(path, fill, ofst, fh)(&errc)
	defer fs.synchronize()()

	node := fs.nodes.Get(path, fh)
	fill(".", &node.Stat, 0)
	fill("..", nil, 0)

	node.EachChild(func(name string, chld *nodes.Node) bool {
		if !fill(name, &chld.Stat, 0) {
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
	return fs.nodes.Close(fh)
}

// Setxattr sets extended attributes.
func (fs *FS) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if xattrAppleResourceFork == name {
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
	return fs.nodes.Write(node)
}

// Getxattr gets extended attributes.
func (fs *FS) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
	if nil == node {
		return -fuse.ENOENT, nil
	}
	if xattrAppleResourceFork == name {
		return -fuse.ENOTSUP, nil
	}
	xatr, ok := node.Xatr[name]
	if !ok {
		return -fuse.ENOATTR, nil
	}
	return 0, xatr
}

// Removexattr removes extended attributes.
func (fs *FS) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if xattrAppleResourceFork == name {
		return -fuse.ENOTSUP
	}
	if _, ok := node.Xatr[name]; !ok {
		return -fuse.ENOATTR
	}
	delete(node.Xatr, name)
	return fs.nodes.Write(node)
}

// Listxattr lists extended attributes.
func (fs *FS) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	defer fs.synchronize()()
	_, _, node := fs.nodes.Lookup(path, nil)
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
