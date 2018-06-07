package nodes

import (
	"context"

	"bazil.org/bazil/cas/chunks"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (node *Node) WriteAt(tx fdb.Transaction, cstore chunks.Store, buff []byte, ofst int64) (n int) {
	endofst := ofst + int64(len(buff))
	if endofst > node.Stat(tx).Size {
		n = node.Truncate(tx, cstore, endofst) //@TODO truncate without trying to open another manifest
		if n != 0 {
			return n
		}

		node.StatSetSize(tx, endofst)
	}

	blob := node.blob(tx, cstore)
	n, err := blob.IO(context.Background()).WriteAt(buff, ofst)
	if err != nil {
		return -fuse.EIO
	}

	return n
}
