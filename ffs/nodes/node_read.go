package nodes

import (
	"context"
	"io"

	"bazil.org/bazil/cas/blobs"
	"bazil.org/bazil/cas/chunks"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func (node *Node) ReadAt(tx fdb.Transaction, cstore chunks.Store, buff []byte, ofst int64) (n int) {
	blob, err := blobs.Open(cstore, node.manifest(tx))
	if err != nil {
		return -fuse.EIO //@TODO report
	}

	n, err = blob.IO(context.Background()).ReadAt(buff, ofst)
	if err != nil {
		if err == io.EOF {
			return n
		}

		return -fuse.EIO //@TODO report
	}

	return
}
