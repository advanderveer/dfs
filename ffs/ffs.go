package ffs

import (
	"fmt"
	"math"
	"strings"

	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/examples/shared"
	"github.com/billziss-gh/cgofuse/fuse"
)

func split(path string) []string {
	return strings.Split(path, "/")
}

func trace(vals ...interface{}) func(vals ...interface{}) {
	uid, gid, _ := fuse.Getcontext()
	return shared.Trace(1, fmt.Sprintf("[uid=%v,gid=%v]", uid, gid), vals...)
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

type Memfs struct {
	fuse.FileSystemBase
	store *nodes.Store

	//@TODO find out if we need to store this map on the remote, maybe to act as
	//a locking mechanis or to report recent interactions
	openmap map[uint64]*nodes.NodeT
}

func (self *Memfs) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	stat.Bavail = math.MaxUint64
	return
}

func (self *Memfs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	// defer self.store.Transact()()
	// return self.makeNode(path, mode, dev, nil)
	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.makeNode(path, mode, dev, nil)
	})
}

func (self *Memfs) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	// defer self.store.Transact()()
	// return self.makeNode(path, fuse.S_IFDIR|(mode&07777), 0, nil)
	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.makeNode(path, fuse.S_IFDIR|(mode&07777), 0, nil)
	})
}

func (self *Memfs) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	// defer self.store.Transact()()
	// return self.removeNode(path, false)
	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.removeNode(path, false)
	})
}

func (self *Memfs) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	// defer self.store.Transact()()
	// return self.removeNode(path, true)
	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.removeNode(path, true)
	})
}

func (self *Memfs) Link(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	// defer self.store.Transact()()
	// _, _, oldnode := self.lookupNode(oldpath, nil)
	// if nil == oldnode {
	// 	return -fuse.ENOENT
	// }
	// newprnt, newname, newnode := self.lookupNode(newpath, nil)
	// if nil == newprnt {
	// 	return -fuse.ENOENT
	// }
	// if nil != newnode {
	// 	return -fuse.EEXIST
	// }
	//
	// oldnode.StatIncNlink()
	// newprnt.SetChld(newname, oldnode)
	//
	// tmsp := fuse.Now()
	// oldnode.StatSetCTim(tmsp)
	// newprnt.StatSetCTim(tmsp)
	// newprnt.StatSetMTim(tmsp)
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, oldnode := self.lookupNode(oldpath, nil)
		if nil == oldnode {
			return -fuse.ENOENT
		}
		newprnt, newname, newnode := self.lookupNode(newpath, nil)
		if nil == newprnt {
			return -fuse.ENOENT
		}
		if nil != newnode {
			return -fuse.EEXIST
		}

		oldnode.StatIncNlink()
		newprnt.SetChld(newname, oldnode)

		tmsp := fuse.Now()
		oldnode.StatSetCTim(tmsp)
		newprnt.StatSetCTim(tmsp)
		newprnt.StatSetMTim(tmsp)
		return 0
	})
}

func (self *Memfs) Symlink(target string, newpath string) (errc int) {
	defer trace(target, newpath)(&errc)
	// defer self.store.Transact()()
	// return self.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
	})
}

func (self *Memfs) Readlink(path string) (errc int, target string) {
	defer trace(path)(&errc, &target)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT, ""
	// }
	// if fuse.S_IFLNK != node.Stat().Mode&fuse.S_IFMT {
	// 	return -fuse.EINVAL, ""
	// }
	// return 0, string(node.Data())

	return self.store.TxWithErrcStr(func(tx fdb.Transaction) (int, string) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT, ""
		}
		if fuse.S_IFLNK != node.Stat().Mode&fuse.S_IFMT {
			return -fuse.EINVAL, ""
		}
		return 0, string(node.Data())
	})
}

func (self *Memfs) Rename(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	// defer self.store.Transact()()
	// oldprnt, oldname, oldnode := self.lookupNode(oldpath, nil)
	// if nil == oldnode {
	// 	return -fuse.ENOENT
	// }
	// newprnt, newname, newnode := self.lookupNode(newpath, oldnode)
	// if nil == newprnt {
	// 	return -fuse.ENOENT
	// }
	// if "" == newname {
	// 	return -fuse.EINVAL // guard against directory loop creation
	// }
	// if oldprnt == newprnt && oldname == newname {
	// 	return 0
	// }
	// if nil != newnode {
	// 	errc = self.removeNode(newpath, fuse.S_IFDIR == oldnode.Stat().Mode&fuse.S_IFMT)
	// 	if 0 != errc {
	// 		return errc
	// 	}
	// }
	//
	// oldprnt.DelChld(oldname)
	// newprnt.SetChld(newname, oldnode)
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		oldprnt, oldname, oldnode := self.lookupNode(oldpath, nil)
		if nil == oldnode {
			return -fuse.ENOENT
		}
		newprnt, newname, newnode := self.lookupNode(newpath, oldnode)
		if nil == newprnt {
			return -fuse.ENOENT
		}
		if "" == newname {
			return -fuse.EINVAL // guard against directory loop creation
		}
		if oldprnt == newprnt && oldname == newname {
			return 0
		}
		if nil != newnode {
			errc = self.removeNode(newpath, fuse.S_IFDIR == oldnode.Stat().Mode&fuse.S_IFMT)
			if 0 != errc {
				return errc
			}
		}

		oldprnt.DelChld(oldname)
		newprnt.SetChld(newname, oldnode)
		return 0
	})
}

func (self *Memfs) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.StatSetMode((node.Stat().Mode & fuse.S_IFMT) | mode&07777)
	// node.StatSetCTim(fuse.Now())
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetMode((node.Stat().Mode & fuse.S_IFMT) | mode&07777)
		node.StatSetCTim(fuse.Now())
		return 0
	})
}

func (self *Memfs) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// if ^uint32(0) != uid {
	// 	node.StatSetUid(uid)
	// }
	// if ^uint32(0) != gid {
	// 	node.StatSetGid(gid)
	// }
	//
	// node.StatSetCTim(fuse.Now())
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if ^uint32(0) != uid {
			node.StatSetUid(uid)
		}
		if ^uint32(0) != gid {
			node.StatSetGid(gid)
		}

		node.StatSetCTim(fuse.Now())
		return 0
	})
}

func (self *Memfs) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.StatSetCTim(fuse.Now())
	// if nil == tmsp {
	// 	tmsp0 := node.Stat().Ctim
	// 	tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
	// 	tmsp = tmsa[:]
	// }
	//
	// node.StatSetATim(tmsp[0])
	// node.StatSetMTim(tmsp[1])
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetCTim(fuse.Now())
		if nil == tmsp {
			tmsp0 := node.Stat().Ctim
			tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
			tmsp = tmsa[:]
		}

		node.StatSetATim(tmsp[0])
		node.StatSetMTim(tmsp[1])
		return 0
	})
}

func (self *Memfs) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	// defer self.store.Transact()()
	// return self.openNode(path, false)

	return self.store.TxWithErrcUint64(func(tx fdb.Transaction) (int, uint64) {
		return self.openNode(path, false)
	})
}

func (self *Memfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	// defer self.store.Transact()()
	// node := self.getNode(path, fh)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// *stat = node.Stat()
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		node := self.getNode(path, fh)
		if nil == node {
			return -fuse.ENOENT
		}
		*stat = node.Stat()
		return 0
	})
}

func (self *Memfs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	// defer self.store.Transact()()
	// node := self.getNode(path, fh)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.SetData(resize(node.Data(), size, true))
	// node.StatSetSize(size)
	//
	// tmsp := fuse.Now()
	// node.StatSetCTim(tmsp)
	// node.StatSetMTim(tmsp)
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		node := self.getNode(path, fh)
		if nil == node {
			return -fuse.ENOENT
		}

		node.SetData(resize(node.Data(), size, true))
		node.StatSetSize(size)

		tmsp := fuse.Now()
		node.StatSetCTim(tmsp)
		node.StatSetMTim(tmsp)
		return 0
	})
}

func (self *Memfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	// defer self.store.Transact()()
	// node := self.getNode(path, fh)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// endofst := ofst + int64(len(buff))
	// if endofst > node.Stat().Size {
	// 	endofst = node.Stat().Size
	// }
	// if endofst < ofst {
	// 	return 0
	// }
	// n = copy(buff, node.Data()[ofst:endofst])
	// node.StatSetATim(fuse.Now())
	// return

	return self.store.TxWithInt(func(tx fdb.Transaction) (n int) {
		node := self.getNode(path, fh)
		if nil == node {
			return -fuse.ENOENT
		}
		endofst := ofst + int64(len(buff))
		if endofst > node.Stat().Size {
			endofst = node.Stat().Size
		}
		if endofst < ofst {
			return 0
		}
		n = copy(buff, node.Data()[ofst:endofst])
		node.StatSetATim(fuse.Now())
		return
	})
}

func (self *Memfs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	// defer self.store.Transact()()
	// node := self.getNode(path, fh)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// endofst := ofst + int64(len(buff))
	// if endofst > node.Stat().Size {
	// 	node.SetData(resize(node.Data(), endofst, true))
	// 	node.StatSetSize(endofst)
	// }
	//
	// n = copy(node.Data()[ofst:endofst], buff)
	//
	// tmsp := fuse.Now()
	// node.StatSetCTim(tmsp)
	// node.StatSetMTim(tmsp)
	// return

	return self.store.TxWithInt(func(tx fdb.Transaction) (n int) {
		node := self.getNode(path, fh)
		if nil == node {
			return -fuse.ENOENT
		}
		endofst := ofst + int64(len(buff))
		if endofst > node.Stat().Size {
			node.SetData(resize(node.Data(), endofst, true))
			node.StatSetSize(endofst)
		}

		n = copy(node.Data()[ofst:endofst], buff)

		tmsp := fuse.Now()
		node.StatSetCTim(tmsp)
		node.StatSetMTim(tmsp)
		return
	})
}

func (self *Memfs) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	// defer self.store.Transact()()
	// return self.closeNode(fh)

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.closeNode(fh)
	})
}

func (self *Memfs) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	// defer self.store.Transact()()
	// return self.openNode(path, true)

	return self.store.TxWithErrcUint64(func(tx fdb.Transaction) (int, uint64) {
		return self.openNode(path, true)
	})

}

func (self *Memfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer trace(path, fill, ofst, fh)(&errc)
	// defer self.store.Transact()()
	// node := self.openmap[fh]
	// sta := node.Stat()
	//
	// fill(".", &sta, 0)
	// fill("..", nil, 0)
	// node.ChldEach(func(name string, chld *nodes.NodeT) (stop bool) {
	// 	csta := chld.Stat()
	// 	if !fill(name, &csta, 0) {
	// 		return true
	// 	}
	//
	// 	return
	// })
	//
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		node := self.openmap[fh]
		sta := node.Stat()

		fill(".", &sta, 0)
		fill("..", nil, 0)
		node.ChldEach(func(name string, chld *nodes.NodeT) (stop bool) {
			csta := chld.Stat()
			if !fill(name, &csta, 0) {
				return true
			}

			return
		})

		return 0
	})
}

func (self *Memfs) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	// defer self.store.Transact()()
	// return self.closeNode(fh)

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.closeNode(fh)
	})
}

func (self *Memfs) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// if "com.apple.ResourceFork" == name {
	// 	return -fuse.ENOTSUP
	// }
	// if fuse.XATTR_CREATE == flags {
	// 	if _, ok := node.XAtrGet(name); ok {
	// 		return -fuse.EEXIST
	// 	}
	// } else if fuse.XATTR_REPLACE == flags {
	// 	if _, ok := node.XAtrGet(name); !ok {
	// 		return -fuse.ENOATTR
	// 	}
	// }
	//
	// xatr := make([]byte, len(value))
	// copy(xatr, value)
	// node.XAtrSet(name, xatr)
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if "com.apple.ResourceFork" == name {
			return -fuse.ENOTSUP
		}
		if fuse.XATTR_CREATE == flags {
			if _, ok := node.XAtrGet(name); ok {
				return -fuse.EEXIST
			}
		} else if fuse.XATTR_REPLACE == flags {
			if _, ok := node.XAtrGet(name); !ok {
				return -fuse.ENOATTR
			}
		}

		xatr := make([]byte, len(value))
		copy(xatr, value)
		node.XAtrSet(name, xatr)
		return 0
	})
}

func (self *Memfs) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT, nil
	// }
	// if "com.apple.ResourceFork" == name {
	// 	return -fuse.ENOTSUP, nil
	// }
	//
	// xatr, ok := node.XAtrGet(name)
	// if !ok {
	// 	return -fuse.ENOATTR, nil
	// }
	// return 0, xatr

	return self.store.TxWithErrcBytes(func(tx fdb.Transaction) (int, []byte) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT, nil
		}
		if "com.apple.ResourceFork" == name {
			return -fuse.ENOTSUP, nil
		}

		xatr, ok := node.XAtrGet(name)
		if !ok {
			return -fuse.ENOATTR, nil
		}
		return 0, xatr
	})
}

func (self *Memfs) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// if "com.apple.ResourceFork" == name {
	// 	return -fuse.ENOTSUP
	// }
	//
	// if _, ok := node.XAtrGet(name); !ok {
	// 	return -fuse.ENOATTR
	// }
	//
	// node.XAtrDel(name)
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if "com.apple.ResourceFork" == name {
			return -fuse.ENOTSUP
		}

		if _, ok := node.XAtrGet(name); !ok {
			return -fuse.ENOATTR
		}

		node.XAtrDel(name)
		return 0
	})
}

func (self *Memfs) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// return node.XAtrEach(func(name string) int {
	// 	if !fill(name) {
	// 		return -fuse.ERANGE
	// 	}
	// 	return 0
	// })

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		return node.XAtrEach(func(name string) int {
			if !fill(name) {
				return -fuse.ERANGE
			}
			return 0
		})
	})
}

func (self *Memfs) Chflags(path string, flags uint32) (errc int) {
	defer trace(path, flags)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.StatSetFlags(flags)
	// node.StatSetCTim(fuse.Now())
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetFlags(flags)
		node.StatSetCTim(fuse.Now())
		return 0
	})
}

func (self *Memfs) Setcrtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.StatSetBirthTim(tmsp)
	// node.StatSetCTim(fuse.Now())
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetBirthTim(tmsp)
		node.StatSetCTim(fuse.Now())
		return 0
	})
}

func (self *Memfs) Setchgtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.StatSetCTim(tmsp)
	// return 0
	//
	// defer self.store.Transact()()
	// _, _, node := self.lookupNode(path, nil)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	//
	// node.StatSetBirthTim(tmsp)
	// node.StatSetCTim(fuse.Now())
	// return 0

	return self.store.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetBirthTim(tmsp)
		node.StatSetCTim(fuse.Now())
		return 0
	})
}

func (self *Memfs) makeNode(path string, mode uint32, dev uint64, data []byte) int {
	prnt, name, node := self.lookupNode(path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}

	self.store.IncIno()
	uid, gid, _ := fuse.Getcontext()
	node = nodes.NewNode(dev, self.store.Ino(), mode, uid, gid)
	if nil != data {
		node.SetData(make([]byte, len(data)))
		node.StatSetSize(int64(len(data)))
		node.CopyData(data)
	}

	prnt.SetChld(name, node)
	prnt.StatSetCTim(node.Stat().Ctim)
	prnt.StatSetMTim(node.Stat().Ctim)
	return 0
}

func (self *Memfs) removeNode(path string, dir bool) int {
	prnt, name, node := self.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.Stat().Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.Stat().Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR
	}

	if 0 < node.CountChld() {
		return -fuse.ENOTEMPTY
	}

	node.StatDecNlink()
	prnt.DelChld(name)

	tmsp := fuse.Now()
	node.StatSetCTim(tmsp)
	prnt.StatSetCTim(tmsp)
	prnt.StatSetMTim(tmsp)
	return 0
}

func (self *Memfs) openNode(path string, dir bool) (int, uint64) {
	_, _, node := self.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, ^uint64(0)
	}
	if !dir && fuse.S_IFDIR == node.Stat().Mode&fuse.S_IFMT {
		return -fuse.EISDIR, ^uint64(0)
	}
	if dir && fuse.S_IFDIR != node.Stat().Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR, ^uint64(0)
	}

	node.IncOpencnt()
	if 1 == node.Opencnt() {
		self.openmap[node.Stat().Ino] = node
	}
	return 0, node.Stat().Ino
}

func (self *Memfs) closeNode(fh uint64) int {
	node := self.openmap[fh]
	node.DecOpencnt()
	if 0 == node.Opencnt() {
		delete(self.openmap, node.Stat().Ino)
	}

	return 0
}

func (self *Memfs) getNode(path string, fh uint64) *nodes.NodeT {
	if ^uint64(0) == fh {
		_, _, node := self.lookupNode(path, nil)
		return node
	} else {
		return self.openmap[fh]
	}
}

func (self *Memfs) lookupNode(path string, ancestor *nodes.NodeT) (prnt *nodes.NodeT, name string, node *nodes.NodeT) {
	prnt = self.store.Root()
	name = ""
	node = self.store.Root()
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c
			node = node.GetChld(name)
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

func NewFS(store *nodes.Store) (*Memfs, error) {
	self := Memfs{}
	self.store = store
	self.openmap = map[uint64]*nodes.NodeT{}
	return &self, nil
}

var _ fuse.FileSystemChflags = (*Memfs)(nil)
var _ fuse.FileSystemSetcrtime = (*Memfs)(nil)
var _ fuse.FileSystemSetchgtime = (*Memfs)(nil)
