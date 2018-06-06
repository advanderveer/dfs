package nodes

import (
	"bazil.org/bazil/cas"
	"bazil.org/bazil/cas/blobs"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

func (n *Node) setManifest(tx fdb.Transaction, m *blobs.Manifest) {
	n.putUint64At(tx, "msize", m.Size)
	n.putUint32At(tx, "csize", m.ChunkSize)
	n.putUint32At(tx, "fanout", m.Fanout)
	tx.Set(n.ss.Pack(tuple.Tuple{"mroot"}), m.Root.Bytes())
}

func (n *Node) manifest(tx fdb.Transaction) (m *blobs.Manifest) {
	m = &blobs.Manifest{Type: "blob"}

	keyd := tx.Get(n.ss.Pack(tuple.Tuple{"mroot"})).MustGet()
	if len(keyd) > 0 {
		m.Root = cas.NewKey(keyd)
	}

	m.Size = n.getUint64At(tx, "msize")
	m.ChunkSize = n.getUint32At(tx, "csize")
	m.Fanout = n.getUint32At(tx, "fanout")
	return
}
