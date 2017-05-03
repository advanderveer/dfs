package dfs

import "github.com/billziss-gh/cgofuse/fuse"

type nodeT struct {
	stat    fuse.Stat_t
	xatr    map[string][]byte
	chld    map[string]*nodeT
	data    []byte
	opencnt int
}

func newNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *nodeT {
	tmsp := fuse.Now()
	fs := nodeT{
		fuse.Stat_t{
			Dev:      dev,
			Ino:      ino,
			Mode:     mode,
			Nlink:    1,
			Uid:      uid,
			Gid:      gid,
			Atim:     tmsp,
			Mtim:     tmsp,
			Ctim:     tmsp,
			Birthtim: tmsp,
		},
		nil,
		nil,
		nil,
		0}
	if fuse.S_IFDIR == fs.stat.Mode&fuse.S_IFMT {
		fs.chld = map[string]*nodeT{}
	}
	return &fs
}
