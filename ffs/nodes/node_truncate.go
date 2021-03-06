package nodes

import (
	"context"

	"bazil.org/bazil/cas/chunks"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (node *Node) Truncate(tx fdb.Transaction, cstore chunks.Store, size int64) (errc int) {
	blob := node.blob(tx, cstore)
	err := blob.Truncate(context.Background(), uint64(size))
	if err != nil {
		return -fuse.EIO //@TODO report
	}

	return
}
