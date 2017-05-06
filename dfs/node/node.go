package node

import (
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
)

func split(path string) []string {
	return strings.Split(path, "/")
}

//N is a filesystem node
type N struct {
	Stat fuse.Stat_t
	Data []byte
	Xatr map[string][]byte

	chlds   map[string]*N
	opencnt int
}

//ListChld lists children of a Node
func (n *N) ListChld() map[string]*N {
	return n.chlds
}

//DelChld deletes a child
func (n *N) DelChld(name string) {
	delete(n.chlds, name)
}

//PutChld (over)writes a (new) child
func (n *N) PutChld(name string, node *N) {
	n.chlds[name] = node
}

func newNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *N {
	tmsp := fuse.Now()
	fs := N{
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
	if fuse.S_IFDIR == fs.Stat.Mode&fuse.S_IFMT {
		fs.chlds = map[string]*N{}
	}
	return &fs
}
