/*
 * memfs.go
 *
 * Copyright 2017-2018 Bill Zissimopoulos
 */
/*
 * This file is part of Cgofuse.
 *
 * It is licensed under the MIT license. The full license text can be found
 * in the License.txt file at the root of this project.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/advanderveer/fdb-tests/pkg/webdav"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/billziss-gh/cgofuse/fuse"
)

const (
	filename = "hello"
	contents = "hello, world\n"
)

type fdbFS struct {
	fuse.FileSystemBase
	tr    fdb.Transactor
	davfs *webdav.FDBFS
}

func (fs *fdbFS) rerr(err error) (errc int) {
	if err != nil {
		return -fuse.EBUSY
	}

	return 0
}

func (fs *fdbFS) tx(op, path string, f func(tx fdb.Transaction, dir *webdav.Node, frag string) int) (errc int) {
	_, err := fs.tr.Transact(func(tx fdb.Transaction) (interface{}, error) {
		fmt.Printf("[%s] path: %s", op, path)
		dir, frag, err := fs.davfs.Find(fs.tr, op, path)
		fmt.Printf("\t\tdir: %p, frag: %s, finderr: %v", dir, frag, err)
		if err != nil {
			return nil, fmt.Errorf("fdbfuse: failed to find node: %v", err)
		}

		errc = f(tx, dir, frag)
		fmt.Printf("\t errc: %d\n", errc)
		return nil, nil
	})

	if err != nil {
		return -fuse.EIO
	}

	return
}

func (fs *fdbFS) Open(path string, flags int) (errc int, fh uint64) {
	return -fuse.ENOSYS, ^uint64(0)
}

func (fs *fdbFS) Mknod(path string, mode uint32, dev uint64) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Opendir(path string) (errc int, fh uint64) {
	return -fuse.ENOSYS, ^uint64(0)
}

func (fs *fdbFS) Unlink(path string) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Rmdir(path string) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Link(oldpath string, newpath string) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Symlink(target string, newpath string) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Readlink(path string) (errc int, target string) {
	return -fuse.ENOSYS, ""
}

func (fs *fdbFS) Chmod(path string, mode uint32) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Chown(path string, uid uint32, gid uint32) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Truncate(path string, size int64, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Release(path string, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Releasedir(path string, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Getxattr(path string, name string) (errc int, xatr []byte) {
	return -fuse.ENOSYS, nil
}

func (fs *fdbFS) Removexattr(path string, name string) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Listxattr(path string, fill func(name string) bool) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Chflags(path string, flags uint32) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Setcrtime(path string, tmsp fuse.Timespec) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Setchgtime(path string, tmsp fuse.Timespec) (errc int) {
	return -fuse.ENOSYS
}

func (fs *fdbFS) Rename(oldpath string, newpath string) (errc int) {
	errc = fs.tx("Rename", oldpath, func(tx fdb.Transaction, dir *webdav.Node, frag string) (errc int) {
		//@TODO do not wrap in extra tx that also attempts to get node
		err := fs.davfs.Rename(context.TODO(), oldpath, newpath)
		if err != nil {
			switch err {
			case os.ErrInvalid:
				return -fuse.EINVAL
			case os.ErrNotExist:
				return -fuse.ENOENT
			default:
				//@TODO determine what to do with the webdav errors
				return fs.rerr(err)
			}
		}

		return
	})

	return
}

//Mkdir will attempt to create a directory at the provided 'path', it returns
//an error exists if the directory already exists
func (fs *fdbFS) Mkdir(path string, mode uint32) (errc int) {
	errc = fs.tx("Mkdir", path, func(tx fdb.Transaction, dir *webdav.Node, frag string) (errc int) {
		//@TODO set context and pass on mode/perm
		//@TODO do not wrap in extra tx that also attempts to get node
		err := fs.davfs.Mkdir(context.TODO(), path, 0777)
		if err != nil && err != os.ErrExist {
			return fs.rerr(err)
		}

		if err == os.ErrExist {
			return -fuse.EEXIST
		}

		return
	})

	return
}

//Getattr returns status information of a node, either by the path name or the
//provide fh. @TODO implement case in which fh is provided
func (fs *fdbFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	errc = fs.tx("Getattr", path, func(tx fdb.Transaction, dir *webdav.Node, frag string) (errc int) {
		var (
			err  error
			node *webdav.Node
		)

		if dir == nil { //opening root
			node, err = fs.davfs.Root(tx)
			if err != nil {
				return fs.rerr(err)
			}

			node.Mode = 0777 | os.ModeDir //@TODO what is a sane permission for the root for FUSE and add it upon
			//creating the root instead of when reading it

		} else { //opening dir child
			node, err = fs.davfs.Get(tx, dir, frag)
			if err != nil && err != webdav.ErrNodeNotExist {
				return fs.rerr(err)
			}
		}

		if node == nil { //no entry found
			return -fuse.ENOENT
		}

		if node.Mode.IsDir() { //size is not relevant for dirs, just set the mode
			stat.Mode = fuse.S_IFDIR | uint32(node.Mode.Perm())
			return 0
		}

		//get size for regular file and set mode
		stat.Mode = fuse.S_IFREG | uint32(node.Mode.Perm())
		stat.Size, err = fs.davfs.SizeOf(tx, node.ID)
		return fs.rerr(err)
	})

	return
}

//Readdir will be called to fill in directory entries using the provided fill
//function. It should do so using the provided 'path' or directly using 'fh'.
//@TODO handle the case in which fh is provided
func (fs *fdbFS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	errc = fs.tx("Readdir", path, func(tx fdb.Transaction, dir *webdav.Node, frag string) (errc int) {
		var err error
		if dir == nil { //opening root
			dir, err = fs.davfs.Root(tx)
			if err != nil {
				return fs.rerr(err)
			}
		}

		//fill directory
		fill(".", nil, 0)
		fill("..", nil, 0)
		if err := fs.davfs.Children(tx, dir, func(name string, n *webdav.Node) (stop bool) {
			//@TODO deal with offset
			//@TODO fill stat struct
			return !fill(name, &fuse.Stat_t{}, 0)
		}); err != nil {
			log.Print(err)
		}

		return 0
	})

	return
}

// func (fs *fdbFS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
// 	if errc := fs.tx("Read", path, func(tx fdb.Transaction, dir *webdav.Node, frag string) (errc int) {
//
// 		endofst := ofst + int64(len(buff))
// 		if endofst > int64(len(contents)) {
// 			endofst = int64(len(contents))
// 		}
// 		if endofst < ofst {
// 			return 0
// 		}
// 		n = copy(buff, contents[ofst:endofst])
// 		return
//
// 	}); errc > 0 {
// 		return 0 //@TODO how to fail reads?
// 	}
//
// 	return
// }

var _ fuse.FileSystemChflags = (*fdbFS)(nil)
var _ fuse.FileSystemSetcrtime = (*fdbFS)(nil)
var _ fuse.FileSystemSetchgtime = (*fdbFS)(nil)

func main() {
	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	dir, err := directory.CreateOrOpen(db, []string{"fdb-tests", "litmus"}, nil)
	if err != nil {
		log.Fatal("failed to create or open app dir:", err)
	}

	davfs := webdav.NewFDBFS(db, dir, nil).(*webdav.FDBFS)
	defer func() {
		log.Println("cleaning up database...")
		_, err := dir.Remove(db, nil)
		if err != nil {
			log.Fatal("failed to remove testing dir:", err)
		}
		log.Println("done, exiting!")
	}()

	davfs.Mkdir(context.TODO(), "/my-file", 0777)

	hellofs := &fdbFS{tr: db, davfs: davfs}
	host := fuse.NewFileSystemHost(hellofs)
	host.SetCapReaddirPlus(true)
	host.Mount("", os.Args[1:])
}
