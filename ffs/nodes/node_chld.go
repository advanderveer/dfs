package nodes

import "github.com/apple/foundationdb/bindings/go/src/fdb"

func (n *NodeT) SetChld(tx fdb.Transaction, name string, nn *NodeT) { n.no.chld[name] = nn }
func (n *NodeT) GetChld(tx fdb.Transaction, name string) *NodeT     { return n.no.chld[name] }
func (n *NodeT) DelChld(tx fdb.Transaction, name string)            { delete(n.no.chld, name) }
func (n *NodeT) CountChld(tx fdb.Transaction) int64                 { return int64(len(n.no.chld)) }
func (n *NodeT) ChldEach(tx fdb.Transaction, f func(name string, n *NodeT) bool) {
	for name, n := range n.no.chld {
		stop := f(name, n)
		if stop {
			return
		}
	}
}
