package webdav

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

//@TODO fix: The Finder can’t complete the operation because some data in “6g copy.bin” can’t be read or written. (Error code -36)
// related: https://github.com/cryptomator/cryptomator/issues/579
// related: https://community.cryptomator.org/t/mac-app-crashing-error-36/683/2
//@TODO change fs methods on nodes to GetChild/DelChild/PutChild
//@TODO use other encoding scheme than JSON for headers (that allows setting modtime directly)
//@TODO come up with garbage collection mechanism (to clean up unlinked nodes)
//@TODO implement readdir without using a snapshot
//@TODO clean up copied tests

var (
	nilNodeID  = nodeID{}
	rootNodeID = nodeID{0x01}
)

type nodeID [sha256.Size]byte

func genID() (id nodeID) {
	_, err := rand.Read(id[:])
	if err != nil {
		panic("failed to generate id: " + err.Error())
	}

	return id
}

func encodeNode(n *Node) ([]byte, error) {
	return json.Marshal(n)
}

func decodeNode(d []byte) (*Node, error) {
	n := &Node{}
	return n, json.Unmarshal(d, n)
}

var (
	ErrNodeNotExist  = errors.New("node does not exist")
	errInvalidNodeID = errors.New("node id is invalid")
)

func (fs *FDBFS) getHdr(tx fdb.Transaction, nid nodeID) (node *Node, err error) {
	if nid == nilNodeID {
		return nil, errInvalidNodeID
	}

	d := tx.Get(fs.nodes.Pack(tuple.Tuple{nid[:]})).MustGet()
	if len(d) > 0 {
		node, err = decodeNode(d)
		if err != nil {
			return nil, err
		}

		node.ID = nid
		return
	}

	return nil, ErrNodeNotExist
}

func (fs *FDBFS) putHdr(tx fdb.Transaction, n *Node) error {
	d, err := encodeNode(n)
	if err != nil {
		return err
	}

	if n.ID == nilNodeID {
		return errInvalidNodeID
	}

	tx.Set(fs.nodes.Pack(tuple.Tuple{n.ID[:]}), d)
	return nil
}

//Node holds information about a node and refers to other nodes as children
type Node struct {
	ID      nodeID
	ModTime time.Time
	Mode    os.FileMode
}

// func (n *Node) stat(tr name string) *fdbFileInfo {
// 	return &fdbFileInfo{
// 		name:    name,
// 		size:    int64(len(n.data)),
// 		mode:    n.Mode,
// 		modTime: n.ModTime,
// 	}
// }

func decodeProp(d []byte) (Property, error) {
	p := Property{}
	return p, xml.Unmarshal(d, &p)
}

func encodeProp(p Property) ([]byte, error) {
	return xml.Marshal(p)
}

func (fs *FDBFS) PutProp(tr fdb.Transactor, nid nodeID, name xml.Name, p Property) (err error) {
	if nid == nilNodeID {
		return errInvalidNodeID
	}

	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		d, e := encodeProp(p)
		if e != nil {
			return nil, e
		}

		tx.Set(fs.nodes.Pack(tuple.Tuple{nid[:], "p", name.Space, name.Local}), d)
		return
	})

	return
}

func (fs *FDBFS) DelProp(tr fdb.Transactor, nid nodeID, name xml.Name) (err error) {
	if nid == nilNodeID {
		return errInvalidNodeID
	}

	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		tx.Clear(fs.nodes.Pack(tuple.Tuple{nid[:], "p", name.Space, name.Local}))
		return
	})

	return
}

//Props allows iteration over the properties of node 'p'
func (fs *FDBFS) Props(tr fdb.Transactor, nid nodeID, f func(p Property) bool) (err error) {
	if nid == nilNodeID {
		return errInvalidNodeID
	}

	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		iter := tx.GetRange(fs.nodes.Sub(nid[:], "p"), fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv := iter.MustGet()
			t, e := fs.nodes.Unpack(kv.Key)
			if e != nil {
				return nil, e
			}

			if len(t) < 4 {
				return nil, errors.New("unexpected tuple format")
			}

			p, e := decodeProp(kv.Value)
			if e != nil {
				return nil, e
			}

			if f(p) {
				return nil, nil
			}
		}

		return
	})

	return
}

//Root will read the root node, if there is none it will be created and written immediately
func (fs *FDBFS) Root(tr fdb.Transactor) (node *Node, err error) {
	if _, err = fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		node, err = fs.getHdr(tx, rootNodeID)
		if err != ErrNodeNotExist {
			return nil, err
		}

		node = &Node{ID: rootNodeID, ModTime: time.Now(), Mode: 0660 | os.ModeDir}
		err = fs.putHdr(tx, node)
		if err != nil {
			return nil, err
		}

		return
	}); err != nil {
		return nil, err
	}

	return
}

//Get retrieves a node by its 'name' in a parent node 'p'
func (fs *FDBFS) Get(tr fdb.Transactor, p *Node, name string) (n *Node, err error) {
	if p.ID == nilNodeID {
		return nil, errInvalidNodeID
	}

	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		d := tx.Get(fs.nodes.Pack(tuple.Tuple{p.ID[:], "d", name})).MustGet()
		if len(d) < 1 {
			return nil, ErrNodeNotExist
		}

		var nid nodeID
		copy(nid[:], d)
		n, e = fs.getHdr(tx, nid)
		if e != nil {
			return nil, e
		}

		// n.data, e = fs.ReadData(tx, nid)
		return n, e
	})

	return n, err
}

//Put will (over)write node 'n' in parent node 'p' as name
func (fs *FDBFS) Put(tr fdb.Transactor, p *Node, name string, n *Node) (err error) {
	if n.ID == nilNodeID || p.ID == nilNodeID {
		return errInvalidNodeID
	}

	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		tx.Set(fs.nodes.Pack(tuple.Tuple{p.ID[:], "d", name}), n.ID[:]) //ref
		return nil, fs.putHdr(tx, n)
	})

	return
}

//Del will remove the node that is under the provide name in parent 'p'
func (fs *FDBFS) Del(tr fdb.Transactor, p *Node, name string) (err error) {
	if p.ID == nilNodeID {
		return errInvalidNodeID
	}

	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		d := tx.Get(fs.nodes.Pack(tuple.Tuple{p.ID[:], "d", name})).MustGet()
		if len(d) < 1 {
			return nil, ErrNodeNotExist
		}

		tx.Clear(fs.nodes.Pack(tuple.Tuple{p.ID[:], "d", name})) //remove ref
		return
	})

	return
}

//Children allows iteration over the children of node 'p'
func (fs *FDBFS) Children(tr fdb.Transactor, p *Node, f func(name string, n *Node) bool) (err error) {
	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		iter := tx.GetRange(fs.nodes.Sub(p.ID[:], "d"), fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv := iter.MustGet()
			t, e := fs.nodes.Unpack(kv.Key)
			if e != nil {
				return nil, e
			}

			if len(t) < 3 {
				return nil, errors.New("unexpected tuple format")
			}

			name, ok := t[2].(string)
			if !ok {
				return nil, errors.New("unexpected tuple")
			}

			var nid nodeID
			copy(nid[:], kv.Value)
			n, e := fs.getHdr(tx, nid)
			if e != nil {
				return nil, e
			}

			if f(name, n) {
				return nil, nil
			}
		}

		return
	})

	return
}

// walk walks the directory tree for the fullname, calling f at each step. If f
// returns an error, the walk will be aborted and return that same error.
//
// dir is the directory at that step, frag is the name fragment, and final is
// whether it is the final step. For example, walking "/foo/bar/x" will result
// in 3 calls to f:
//   - "/", "foo", false
//   - "/foo/", "bar", false
//   - "/foo/bar/", "x", true
// The frag argument will be empty only if dir is the root node and the walk
// ends at that root node.
func (fs *FDBFS) walk(tr fdb.Transactor, op, fullname string, f func(dir *Node, frag string, final bool) error) error {
	_, e := tr.Transact(func(tx fdb.Transaction) (interface{}, error) {
		original := fullname
		fullname = slashClean(fullname)

		// Strip any leading "/"s to make fullname a relative path, as the walk
		// starts at fs.root.
		if fullname[0] == '/' {
			fullname = fullname[1:]
		}

		dir, err := fs.Root(tr)
		if err != nil {
			return nil, err
		}

		for {
			frag, remaining := fullname, ""
			i := strings.IndexRune(fullname, '/')
			final := i < 0
			if !final {
				frag, remaining = fullname[:i], fullname[i+1:]
			}
			if frag == "" && dir.ID != rootNodeID {
				panic("webdav: empty path fragment for a clean path")
			}
			if err := f(dir, frag, final); err != nil {
				return nil, &os.PathError{
					Op:   op,
					Path: original,
					Err:  err,
				}
			}
			if final {
				break
			}

			child, err := fs.Get(tr, dir, frag)
			if err != nil && err != ErrNodeNotExist {
				return nil, err
			}

			if child == nil {
				return nil, &os.PathError{
					Op:   op,
					Path: original,
					Err:  os.ErrNotExist,
				}
			}
			if !child.Mode.IsDir() {
				return nil, &os.PathError{
					Op:   op,
					Path: original,
					Err:  os.ErrInvalid,
				}
			}
			dir, fullname = child, remaining
		}

		return nil, nil
	})

	return e
}

// find returns the parent of the named node and the relative name fragment
// from the parent to the child. For example, if finding "/foo/bar/baz" then
// parent will be the node for "/foo/bar" and frag will be "baz".
//
// If the fullname names the root node, then parent, frag and err will be zero.
//
// find returns an error if the parent does not already exist or the parent
// isn't a directory, but it will not return an error per se if the child does
// not already exist. The error returned is either nil or an *os.PathError
// whose Op is op.
func (fs *FDBFS) Find(tr fdb.Transactor, op, fullname string) (parent *Node, frag string, err error) {
	err = fs.walk(tr, op, fullname, func(parent0 *Node, frag0 string, final bool) error {
		if !final {
			return nil
		}
		if frag0 != "" {
			parent, frag = parent0, frag0
		}
		return nil
	})
	return parent, frag, err
}
