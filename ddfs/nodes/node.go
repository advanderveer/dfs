package nodes

import (
	"os"

	"github.com/billziss-gh/cgofuse/fuse"
)

//Data is the persisting part of a node
type Data struct {
	Stat fuse.Stat_t
	Xatr map[string][]byte
	Link []byte
	Chld map[string]uint64
}

//Node represents a filesystem node
type Node struct {
	Data
	opencnt int
	handle  *os.File
	chlds   map[string]*Node
}

func newNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *Node {
	tmsp := fuse.Now()
	node := Node{
		Data{
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
	node.initMaps()
	return &node
}

func (node *Node) initMaps() {
	if node.IsDir() {
		node.Data.Chld = map[string]uint64{}
		node.chlds = map[string]*Node{}
	}
}

//IsDir return if the node is a dir, this doesn't chnage over the lifecycle
func (node *Node) IsDir() bool {
	return fuse.S_IFDIR == node.Data.Stat.Mode&fuse.S_IFMT
}

//Ino returns the nodes identity, this doesn't change over the lifecycle
func (node *Node) Ino() uint64 {
	return node.Data.Stat.Ino
}

//PutChild sets a child node by name
func (node *Node) PutChild(name string, n *Node) {
	node.Data.Chld[name] = n.Data.Stat.Ino
	node.chlds[name] = n
}

//DelChild deletes a child node by name
func (node *Node) DelChild(name string) {
	delete(node.Data.Chld, name)
	delete(node.chlds, name)
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
