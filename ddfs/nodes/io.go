package nodes

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

var errNodeNotExist = errors.New("node doesn't exist")

func loadChild(b *bolt.Bucket, node *Node, name string) (child *Node, err error) {
	ino, ok := node.Chld[name]
	if !ok {
		return nil, nil
	}

	child, ok = node.chlds[name]
	if !ok {
		child, err = loadNode(b, ino)
		if err != nil {
			fmt.Println("failed to load")
			//@TODO handle errors
			return nil, err
		}

		node.chlds[name] = child
	} else {
		err = loadNodeData(b, child.Ino(), &child.Data)
		if err != nil {
			fmt.Println("failed to load data")
			//@TODO handle errors
			return nil, err
		}
	}

	return child, nil
}

func loadNodeData(b *bolt.Bucket, ino uint64, ndata *Data) (err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, ino)

	data := b.Get(key)
	if data == nil {
		return errNodeNotExist
	}

	buf := bytes.NewBuffer(data)
	dec := json.NewDecoder(buf)
	err = dec.Decode(ndata)
	if err != nil {
		return err
	}

	return nil
}

func loadNode(b *bolt.Bucket, ino uint64) (node *Node, err error) {
	node = &Node{}
	err = loadNodeData(b, ino, &node.Data)
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
