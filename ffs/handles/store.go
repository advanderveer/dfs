package handles

import (
	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

type Store struct {
	tr  fdb.Transactor
	ss  subspace.Subspace
	sss subspace.Subspace

	// openmap map[uint64]*nodes.Node
}

func NewStore(tr fdb.Transactor, ss subspace.Subspace, sss subspace.Subspace) *Store {
	return &Store{
		tr:  tr,
		ss:  ss,
		sss: sss,
		// openmap: make(map[uint64]*nodes.Node),
	}
}

func (s *Store) Get(tx fdb.Transaction, fh uint64) (n *nodes.Node) {
	d := tx.Get(s.ss.Pack(tuple.Tuple{int64(fh)})).MustGet()
	if len(d) > 0 {
		return nodes.NewNode(s.sss, fh)
	}

	return
	// return s.openmap[fh]
}

func (s *Store) Set(tx fdb.Transaction, fh uint64, n *nodes.Node) {
	tx.Set(s.ss.Pack(tuple.Tuple{int64(fh)}), []byte{0x01})
	// s.openmap[fh] = n
}

func (s *Store) Del(tx fdb.Transaction, fh uint64) {
	tx.Clear(s.ss.Pack(tuple.Tuple{int64(fh)}))
	// delete(s.openmap, fh)
}
