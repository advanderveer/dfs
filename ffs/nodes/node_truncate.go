package nodes

import (
	"context"

	"bazil.org/bazil/cas/blobs"
	"bazil.org/bazil/cas/chunks"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (node *Node) Truncate(tx fdb.Transaction, cstore chunks.Store, size int64) (errc int) {
	blob, err := blobs.Open(cstore, node.manifest(tx))
	if err != nil {
		return -fuse.EIO //@TODO report
	}

	err = blob.Truncate(context.Background(), uint64(size))
	if err != nil {
		return -fuse.EIO //@TODO report
	}

	m, err := blob.Save(context.Background())
	if err != nil {
		return -fuse.EIO //@TODO report
	}

	node.setManifest(tx, m)
	return
}
