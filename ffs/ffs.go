package ffs

import (
	"fmt"
	"math"
	"strings"

	"github.com/advanderveer/dfs/ffs/blocks"
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
	nstore *nodes.Store
	bstore *blocks.Store

	//@TODO find out if we need to store this map on the remote, maybe to act as
	//a locking mechanis or to report recent interactions
	openmap map[uint64]*nodes.Node
}

func (self *Memfs) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	stat.Bavail = math.MaxUint64
	return
}

func (self *Memfs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.makeNode(tx, path, mode, dev, nil)
	})
}

func (self *Memfs) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.makeNode(tx, path, fuse.S_IFDIR|(mode&07777), 0, nil)
	})
}

func (self *Memfs) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.removeNode(tx, path, false)
	})
}

func (self *Memfs) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.removeNode(tx, path, true)
	})
}

func (self *Memfs) Link(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, oldnode := self.lookupNode(tx, oldpath, nil)
		if nil == oldnode {
			return -fuse.ENOENT
		}
		newprnt, newname, newnode := self.lookupNode(tx, newpath, nil)
		if nil == newprnt {
			return -fuse.ENOENT
		}
		if nil != newnode {
			return -fuse.EEXIST
		}

		oldnode.StatIncNlink(tx)
		newprnt.SetChld(tx, newname, oldnode)

		tmsp := fuse.Now()
		oldnode.StatSetCTim(tx, tmsp)
		newprnt.StatSetCTim(tx, tmsp)
		newprnt.StatSetMTim(tx, tmsp)
		return 0
	})
}

func (self *Memfs) Symlink(target string, newpath string) (errc int) {
	defer trace(target, newpath)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.makeNode(tx, newpath, fuse.S_IFLNK|00777, 0, []byte(target))
	})
}

func (self *Memfs) Readlink(path string) (errc int, target string) {
	defer trace(path)(&errc, &target)
	return self.nstore.TxWithErrcStr(func(tx fdb.Transaction) (int, string) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT, ""
		}
		if fuse.S_IFLNK != node.Stat(tx).Mode&fuse.S_IFMT {
			return -fuse.EINVAL, ""
		}
		return 0, string(self.bstore.ReadData(node, tx))
	})
}

func (self *Memfs) Rename(oldpath string, newpath string) (errc int) {
	defer trace(oldpath, newpath)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		oldprnt, oldname, oldnode := self.lookupNode(tx, oldpath, nil)
		if nil == oldnode {
			return -fuse.ENOENT
		}
		newprnt, newname, newnode := self.lookupNode(tx, newpath, oldnode)
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
			errc = self.removeNode(tx, newpath, fuse.S_IFDIR == oldnode.Stat(tx).Mode&fuse.S_IFMT)
			if 0 != errc {
				return errc
			}
		}

		oldprnt.DelChld(tx, oldname)
		newprnt.SetChld(tx, newname, oldnode)
		return 0
	})
}

func (self *Memfs) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetMode(tx, (node.Stat(tx).Mode&fuse.S_IFMT)|mode&07777)
		node.StatSetCTim(tx, fuse.Now())
		return 0
	})
}

func (self *Memfs) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if ^uint32(0) != uid {
			node.StatSetUid(tx, uid)
		}
		if ^uint32(0) != gid {
			node.StatSetGid(tx, gid)
		}

		node.StatSetCTim(tx, fuse.Now())
		return 0
	})
}

func (self *Memfs) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetCTim(tx, fuse.Now())
		if nil == tmsp {
			tmsp0 := node.Stat(tx).Ctim
			tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
			tmsp = tmsa[:]
		}

		node.StatSetATim(tx, tmsp[0])
		node.StatSetMTim(tx, tmsp[1])
		return 0
	})
}

func (self *Memfs) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	return self.nstore.TxWithErrcUint64(func(tx fdb.Transaction) (int, uint64) {
		return self.openNode(tx, path, false)
	})
}

func (self *Memfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		node := self.getNode(tx, path, fh)
		if nil == node {
			return -fuse.ENOENT
		}
		*stat = node.Stat(tx)
		return 0
	})
}

func (self *Memfs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		node := self.getNode(tx, path, fh)
		if nil == node {
			return -fuse.ENOENT
		}

		self.bstore.WriteData(node, tx, resize(self.bstore.ReadData(node, tx), size, true))
		node.StatSetSize(tx, size)

		tmsp := fuse.Now()
		node.StatSetCTim(tx, tmsp)
		node.StatSetMTim(tx, tmsp)
		return 0
	})
}

func (self *Memfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	return self.nstore.TxWithInt(func(tx fdb.Transaction) (n int) {
		node := self.getNode(tx, path, fh)
		if nil == node {
			return -fuse.ENOENT
		}
		endofst := ofst + int64(len(buff))
		if endofst > node.Stat(tx).Size {
			endofst = node.Stat(tx).Size
		}
		if endofst < ofst {
			return 0
		}
		n = copy(buff, self.bstore.ReadData(node, tx)[ofst:endofst])
		node.StatSetATim(tx, fuse.Now())
		return
	})
}

func (self *Memfs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	return self.nstore.TxWithInt(func(tx fdb.Transaction) (n int) {
		node := self.getNode(tx, path, fh)
		if nil == node {
			return -fuse.ENOENT
		}

		endofst := ofst + int64(len(buff))
		if endofst > node.Stat(tx).Size {
			// node.SetData(tx, resize(self.bstore.ReadData(node, tx), endofst, true))
			self.bstore.WriteData(node, tx, resize(self.bstore.ReadData(node, tx), endofst, true))
			node.StatSetSize(tx, endofst)
		}

		//@TODO this copy is not persisted as is
		n = copy(self.bstore.ReadData(node, tx)[ofst:endofst], buff)

		tmsp := fuse.Now()
		node.StatSetCTim(tx, tmsp)
		node.StatSetMTim(tx, tmsp)
		return
	})
}

func (self *Memfs) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.closeNode(tx, fh)
	})
}

func (self *Memfs) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	return self.nstore.TxWithErrcUint64(func(tx fdb.Transaction) (int, uint64) {
		return self.openNode(tx, path, true)
	})

}

func (self *Memfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer trace(path, fill, ofst, fh)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		node := self.openmap[fh]
		sta := node.Stat(tx)

		fill(".", &sta, 0)
		fill("..", nil, 0)
		node.ChldEach(tx, func(name string, chld *nodes.Node) (stop bool) {
			csta := chld.Stat(tx)
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
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		return self.closeNode(tx, fh)
	})
}

func (self *Memfs) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if "com.apple.ResourceFork" == name {
			return -fuse.ENOTSUP
		}
		if fuse.XATTR_CREATE == flags {
			if _, ok := node.XAtrGet(tx, name); ok {
				return -fuse.EEXIST
			}
		} else if fuse.XATTR_REPLACE == flags {
			if _, ok := node.XAtrGet(tx, name); !ok {
				return -fuse.ENOATTR
			}
		}

		xatr := make([]byte, len(value))
		copy(xatr, value)
		node.XAtrSet(tx, name, xatr)
		return 0
	})
}

func (self *Memfs) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	return self.nstore.TxWithErrcBytes(func(tx fdb.Transaction) (int, []byte) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT, nil
		}
		if "com.apple.ResourceFork" == name {
			return -fuse.ENOTSUP, nil
		}

		xatr, ok := node.XAtrGet(tx, name)
		if !ok {
			return -fuse.ENOATTR, nil
		}
		return 0, xatr
	})
}

func (self *Memfs) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}
		if "com.apple.ResourceFork" == name {
			return -fuse.ENOTSUP
		}

		if _, ok := node.XAtrGet(tx, name); !ok {
			return -fuse.ENOATTR
		}

		node.XAtrDel(tx, name)
		return 0
	})
}

func (self *Memfs) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		return node.XAtrEach(tx, func(name string) int {
			if !fill(name) {
				return -fuse.ERANGE
			}
			return 0
		})
	})
}

func (self *Memfs) Chflags(path string, flags uint32) (errc int) {
	defer trace(path, flags)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetFlags(tx, flags)
		node.StatSetCTim(tx, fuse.Now())
		return 0
	})
}

func (self *Memfs) Setcrtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetBirthTim(tx, tmsp)
		node.StatSetCTim(tx, fuse.Now())
		return 0
	})
}

func (self *Memfs) Setchgtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	return self.nstore.TxWithErrc(func(tx fdb.Transaction) (errc int) {
		_, _, node := self.lookupNode(tx, path, nil)
		if nil == node {
			return -fuse.ENOENT
		}

		node.StatSetBirthTim(tx, tmsp)
		node.StatSetCTim(tx, fuse.Now())
		return 0
	})
}

func (self *Memfs) makeNode(tx fdb.Transaction, path string, mode uint32, dev uint64, data []byte) int {
	prnt, name, node := self.lookupNode(tx, path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}

	self.nstore.IncIno(tx)
	uid, gid, _ := fuse.Getcontext()
	node = self.nstore.NewNode(tx, dev, self.nstore.Ino(tx), mode, uid, gid)
	if nil != data {
		// node.SetData(tx, make([]byte, len(data)))
		self.bstore.WriteData(node, tx, make([]byte, len(data)))
		node.StatSetSize(tx, int64(len(data)))
		// node.CopyData(tx, data)
		self.bstore.CopyData(node, tx, data)
	}

	prnt.SetChld(tx, name, node)
	prnt.StatSetCTim(tx, node.Stat(tx).Ctim)
	prnt.StatSetMTim(tx, node.Stat(tx).Ctim)
	return 0
}

func (self *Memfs) removeNode(tx fdb.Transaction, path string, dir bool) int {
	prnt, name, node := self.lookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.Stat(tx).Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.Stat(tx).Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR
	}

	if 0 < node.CountChld(tx) {
		return -fuse.ENOTEMPTY
	}

	node.StatDecNlink(tx)
	prnt.DelChld(tx, name)

	tmsp := fuse.Now()
	node.StatSetCTim(tx, tmsp)
	prnt.StatSetCTim(tx, tmsp)
	prnt.StatSetMTim(tx, tmsp)
	return 0
}

func (self *Memfs) openNode(tx fdb.Transaction, path string, dir bool) (int, uint64) {
	_, _, node := self.lookupNode(tx, path, nil)
	if nil == node {
		return -fuse.ENOENT, ^uint64(0)
	}
	if !dir && fuse.S_IFDIR == node.Stat(tx).Mode&fuse.S_IFMT {
		return -fuse.EISDIR, ^uint64(0)
	}
	if dir && fuse.S_IFDIR != node.Stat(tx).Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR, ^uint64(0)
	}

	node.IncOpencnt(tx)
	if 1 == node.Opencnt(tx) {
		self.openmap[node.Stat(tx).Ino] = node
	}
	return 0, node.Stat(tx).Ino
}

func (self *Memfs) closeNode(tx fdb.Transaction, fh uint64) int {
	node := self.openmap[fh]
	node.DecOpencnt(tx)
	if 0 == node.Opencnt(tx) {
		delete(self.openmap, node.Stat(tx).Ino)
	}

	return 0
}

func (self *Memfs) getNode(tx fdb.Transaction, path string, fh uint64) *nodes.Node {
	if ^uint64(0) == fh {
		_, _, node := self.lookupNode(tx, path, nil)
		return node
	} else {
		return self.openmap[fh]
	}
}

func (self *Memfs) lookupNode(tx fdb.Transaction, path string, ancestor *nodes.Node) (prnt *nodes.Node, name string, node *nodes.Node) {
	prnt = self.nstore.Root(tx)
	name = ""
	node = self.nstore.Root(tx)
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c
			node = node.GetChld(tx, name)
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

func NewFS(nstore *nodes.Store, bstore *blocks.Store) (*Memfs, error) {
	self := Memfs{}
	self.nstore = nstore
	self.bstore = bstore
	self.openmap = map[uint64]*nodes.Node{}
	return &self, nil
}

var _ fuse.FileSystemChflags = (*Memfs)(nil)
var _ fuse.FileSystemSetcrtime = (*Memfs)(nil)
var _ fuse.FileSystemSetchgtime = (*Memfs)(nil)
