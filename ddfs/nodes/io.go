package nodes

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
)

var errNodeNotExist = errors.New("node doesn't exist")

func loadChildOrNil(b *bolt.Bucket, node *Node, name string) (child *Node) {
	// fmt.Println("prnt:", node.Ino(), "name:", name)

	ino, ok := node.Chld[name]
	if !ok {
		return nil
	}

	_ = ino

	// ino, ok1 := node.Chld[name]
	child = node.chlds[name]

	// fmt.Println(ino, ok1, ok2)

	// ino, ok := node.Chld[name]
	// if !ok {
	// 	return nil
	// }
	//
	// child = node.chlds[name]
	// _ = ino
	return

	// ino, ok := node.Chld[name]
	// if ok {
	// 	child = node.chlds[name]
	//
	// 	if child == nil {
	// 		fmt.Println("AAAAA")
	// 	}
	//
	// 	// if child == nil {
	// 	// 	child = &Node{}
	// 	// 	err := loadIno(b, ino, &child.Data)
	// 	// 	if err != nil {
	// 	// 		fmt.Printf("failed to load ino '%d': %v\n", ino, err)
	// 	// 		return nil
	// 	// 	}
	// 	// 	child.initMaps()
	// 	// }
	//
	// 	_ = ino
	// 	//@TODO if nil, load from disk using ino
	// }
	//
	// return child
}

// func loadIno(b *bolt.Bucket, ino uint64, ndata *Data) (err error) {
// 	key := make([]byte, 8)
// 	binary.BigEndian.PutUint64(key, ino)
//
// 	data := b.Get(key)
// 	if data == nil {
// 		return errNodeNotExist
// 	}
//
// 	buf := bytes.NewBuffer(data)
// 	dec := json.NewDecoder(buf)
// 	err = dec.Decode(ndata)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }

// func loadNode(b *bolt.Bucket, node *Node) error {
//
// 	loadIno(b, node.Ino(), &node.Data)
//
// 	// key := make([]byte, 8)
// 	// binary.BigEndian.PutUint64(key, node.Ino())
// 	//
// 	// data := b.Get(key)
// 	// if data == nil {
// 	// 	return errNodeNotExist
// 	// }
// 	//
// 	// buf := bytes.NewBuffer(data)
// 	// dec := json.NewDecoder(buf)
// 	// err := dec.Decode(&node.Data)
// 	// if err != nil {
// 	// 	return err
// 	// }
//
// 	return nil
// }

func loadNode(b *bolt.Bucket, ino uint64) (node *Node, err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, ino)

	data := b.Get(key)
	if data == nil {
		return nil, errNodeNotExist
	}

	node = &Node{}
	buf := bytes.NewBuffer(data)
	dec := json.NewDecoder(buf)
	err = dec.Decode(&node.Data)
	if err != nil {
		return nil, err
	}

	node.initMaps()
	return node, nil
}

func saveNode(b *bolt.Bucket, node *Node) error {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(node.Data)
	if err != nil {
		return err
	}

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, node.Ino())
	return b.Put(key, buf.Bytes())
}
