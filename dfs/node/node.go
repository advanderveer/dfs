package node

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

func split(path string) []string {
	return strings.Split(path, "/")
}

//MustDelNode will remove a node or panic if anything fails
func MustDelNode(tx *bolt.Tx, node *N) {
	b := tx.Bucket(BucketNameNodes)

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, node.Stat.Ino)

	err := b.Delete(key)
	if err != nil {
		fmt.Println("Error 1 - MustDelNode", err)
		//@TODO panic(?)
	}
}

//MustPutNode will write a node or panic if anything fails
func MustPutNode(tx *bolt.Tx, node *N) {
	b := tx.Bucket(BucketNameNodes)

	data, err := json.Marshal(node)
	if err != nil {
		fmt.Println("Error 1 - MustPutNode", err)
		panic(err)
	}

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, node.Stat.Ino)

	err = b.Put(key, data)
	if err != nil {
		fmt.Println("Error 2 - MustPutNode", err)
		panic(err)
	}

}

//MustGetNode will read a node or panic if anything fails
func MustGetNode(tx *bolt.Tx, ino uint64) (node *N) {
	b := tx.Bucket(BucketNameNodes)

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, ino)

	data := b.Get(key)
	if data == nil {
		return nil
	}

	node = newNode(0, 0, 0, 0, 0)
	err := json.Unmarshal(data, node)
	if err != nil {
		fmt.Println("Error - MustGetNode", err)
		panic(err)
	}

	return node
}

//N is a filesystem node
type N struct {
	Stat fuse.Stat_t       `json:"s"`
	Data []byte            `json:"d"`
	Xatr map[string][]byte `json:"x"`
	Chld map[string]uint64 `json:"c"`

	chlds   map[string]*N
	opencnt int
}

//Ino uniquely identifes the node
func (n *N) Ino() uint64 {
	return n.Stat.Ino
}

//ListChld lists children of a Node
func (n *N) ListChld(tx *bolt.Tx) (chlds map[string]*N) {

	chlds = map[string]*N{}
	for name, c := range n.chlds {
		//@TODO load children lazily?
		chlds[name] = c
	}

	return chlds
}

//DelChld deletes a child
func (n *N) DelChld(tx *bolt.Tx, name string) {
	delete(n.Chld, name)
	delete(n.chlds, name)
}

//PutChld (over)writes a (new) child
func (n *N) PutChld(tx *bolt.Tx, name string, node *N) {
	n.Chld[name] = node.Stat.Ino
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
		nil,
		0}
	if fuse.S_IFDIR == fs.Stat.Mode&fuse.S_IFMT {
		fs.chlds = map[string]*N{}
		fs.Chld = map[string]uint64{}
	}
	return &fs
}
