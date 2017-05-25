package nodes

import (
	"os"

	"github.com/billziss-gh/cgofuse/fuse"
)

type nodeData struct {
	Stat fuse.Stat_t
	Xatr map[string][]byte
	Link []byte
	Chld map[string]uint64
}

//Node represents a filesystem node
type Node struct {
	nodeData
	opencnt int
	handle  *os.File
	chlds   map[string]*Node
}

func newNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *Node {
	tmsp := fuse.Now()
	fs := Node{
		nodeData{
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
		},
		0,
		nil,
		nil}
	if fuse.S_IFDIR == fs.Stat.Mode&fuse.S_IFMT {
		fs.Chld = map[string]uint64{}
		fs.chlds = map[string]*Node{}
	}
	return &fs
}

//Ino returns the nodes identity
func (node *Node) Ino() uint64 {
	return node.Stat.Ino
}

//EachChild calls next for each child, if it returns false it will stop
func (node *Node) EachChild(next func(name string, n *Node) bool) {
	for name, child := range node.chlds {
		ok := next(name, child)
		if !ok {
			break
		}
	}
}

//PutChild sets a child node by name
func (node *Node) PutChild(name string, n *Node) {
	node.Chld[name] = n.Stat.Ino
	node.chlds[name] = n
}

//DelChild deletes a child node by name
func (node *Node) DelChild(name string) {
	// ino, ok := node.Chld[name]
	// if ok {
	delete(node.Chld, name)
	delete(node.chlds, name)
	// }
}

//ReadAt implements: https://godoc.org/os#File.ReadAt
func (node *Node) ReadAt(b []byte, off int64) (n int, err error) {
	return node.handle.ReadAt(b, off)
}

//WriteAt implements: https://godoc.org/os#File.WriteAt
func (node *Node) WriteAt(b []byte, off int64) (n int, err error) {
	return node.handle.WriteAt(b, off)
}

//Truncate implements: https://godoc.org/os#File.Truncate
func (node *Node) Truncate(size int64) error {
	return node.handle.Truncate(size)
}
