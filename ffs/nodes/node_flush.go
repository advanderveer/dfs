package nodes

import (
	"context"

	"bazil.org/bazil/cas/chunks"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (node *Node) Flush(tx fdb.Transaction, cstore chunks.Store) (errc int) {
	blob := node.blob(tx, cstore)
	m, err := blob.Save(context.Background())
	if err != nil {
		return -fuse.EIO
	}

	node.setManifest(tx, m)
	delete(dirtyBlobs, node.StatGetIno(tx))
	return
}
