package dfs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/billziss-gh/cgofuse/examples/shared"
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
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

//FS is an in-memory userland filesystem (FUSE) implementation that works on OSX, Linux and Windows
type FS struct {
	fuse.FileSystemBase
	db    *bolt.DB
	lock  sync.Mutex //@TODO change this into an interface
	store *NodeStore //@TODO change this into an interface
}

func endTx(tx *bolt.Tx, perrc *int) {
	errc := *perrc
	if errc >= 0 {
		if err := tx.Commit(); err != nil {
			errc = -fuse.ENXIO //commit failed, we're now in an incosistent state
		}
	} else if errc < 0 {
		if err := tx.Rollback(); err != nil {
			errc = -fuse.EFAULT //rollback failed, we're now in an incosistent state
		}
	}
	*perrc = errc
}

// Statfs gets file system statistics.
func (fs *FS) Statfs(path string, stat *fuse.Statfs_t) int {
	return -fuse.ENOSYS
}

// Mknod creates a file node.
func (fs *FS) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doMkNod(tx, path, mode, dev)
}

// Mkdir creates a directory.
func (fs *FS) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doMkdir(tx, path, mode)
}

// Unlink removes a file.
func (fs *FS) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doUnlink(tx, path)
}

// Rmdir removes a directory.
func (fs *FS) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doRmdir(tx, path)
}

// Link creates a hard link to a file.
func (fs *FS) Link(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doLink(tx, oldpath, newpath)
}

// Symlink creates a symbolic link.
func (fs *FS) Symlink(target string, newpath string) (errc int) {
	defer trace(target, newpath)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doSymlink(tx, target, newpath)
}

// Readlink reads the target of a symbolic link.
func (fs *FS) Readlink(path string) (errc int, target string) {
	defer trace(path)(&errc, &target)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO, "" //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doReadlink(tx, path)
}

// Rename renames a file.
func (fs *FS) Rename(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doRename(tx, oldpath, newpath)
}

// Chmod changes the permission bits of a file.
func (fs *FS) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doChmod(tx, path, mode)
}

// Chown changes the owner and group of a file.
func (fs *FS) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doChown(tx, path, uid, gid)
}

// Utimens changes the access and modification times of a file.
func (fs *FS) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doUtimens(tx, path, tmsp)
}

// Open opens a file.
func (fs *FS) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO, ^uint64(0) //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doOpen(tx, path, flags)
}

// Getattr gets file attributes.
func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doGetattr(tx, path, stat, fh)
}

// Truncate changes the size of a file.
func (fs *FS) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doTruncate(tx, path, size, fh)
}

// Read reads data from a file.
func (fs *FS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &n)
	return fs.doRead(tx, path, buff, ofst, fh)
}

// Write writes data to a file.
func (fs *FS) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &n)
	return fs.doWrite(tx, path, buff, ofst, fh)
}

// Release closes an open file.
func (fs *FS) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doRelease(tx, path, fh)
}

// Opendir opens a directory.
func (fs *FS) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO, ^uint64(0)
	}

	defer endTx(tx, &errc)
	return fs.doOpendir(tx, path)
}

// Readdir reads a directory.
func (fs *FS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer trace(path, fill, ofst, fh)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doReaddir(tx, path, fill, ofst, fh)
}

// Releasedir closes an open directory.
func (fs *FS) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doReleasedir(tx, path, fh)
}

// Setxattr sets extended attributes.
func (fs *FS) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doSetxattr(tx, path, name, value, flags)
}

// Getxattr gets extended attributes.
func (fs *FS) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO, nil //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doGetxattr(tx, path, name)
}

// Removexattr removes extended attributes.
func (fs *FS) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doRemovexattr(tx, path, name)
}

// Listxattr lists extended attributes.
func (fs *FS) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	tx, err := fs.db.Begin(true)
	if err != nil {
		return -fuse.EIO //unable to acquire lock
	}

	defer endTx(tx, &errc)
	return fs.doListxattr(tx, path, fill)
}

//NewFS sets up the filesystem
func NewFS(db *bolt.DB) *FS {
	fs := FS{db: db}
	fs.store = newNodeStore()
	return &fs
}
