package webdav

import (
	"encoding/xml"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	fdbdir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"golang.org/x/net/context"
)

//FDBFSConf configures the filesystem
type FDBFSConf struct {
	MaxChunkSize int64
}

//DefaultFBDFSConf returns a sensible default filesystem configuration
func DefaultFBDFSConf() *FDBFSConf {
	return &FDBFSConf{MaxChunkSize: 100 * 1000} //10Kb
}

// NewFDBFS returns a new fdb FileSystem implementation.
func NewFDBFS(tr fdb.Transactor, ns fdbdir.Directory, cfg *FDBFSConf) FileSystem {
	nodes, err := ns.CreateOrOpen(tr, []string{"nodes"}, nil)
	if err != nil {
		panic("webdav: failed to create 'nodes' subdirectory in namespace: " + err.Error())
	}

	if cfg == nil {
		cfg = DefaultFBDFSConf()
	}

	fs := &FDBFS{
		tr:           tr,
		nodes:        nodes,
		maxChunkSize: cfg.MaxChunkSize,
	}

	return fs
}

// A FDBFS implements FileSystem, storing all metadata and actual file data
// in-fdbory. No limits on filesystem size are used, so it is not recommended
// this be used where the clients are untrusted.
//
// Concurrent access is permitted. The tree structure is protected by a mutex,
// and each node's contents and metadata are protected by a per-node mutex.
//
// TODO: Enforce file permissions.
type FDBFS struct {
	tr           fdb.Transactor
	nodes        fdbdir.DirectorySubspace
	maxChunkSize int64
}

func (fs *FDBFS) RemoveAll(ctx context.Context, name string) (err error) {
	_, err = fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		dir, frag, err := fs.Find(tx, "remove", name) //find parent dir node that should contain the name
		if err != nil {
			return nil, err
		}

		if dir == nil {
			return nil, os.ErrInvalid // We can't remove the root.
		}

		err = fs.Del(tx, dir, frag)
		if err != nil && err != ErrNodeNotExist {
			return nil, err
		}

		return
	})

	return
}

func (fs *FDBFS) Rename(ctx context.Context, oldName, newName string) (err error) {
	_, err = fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		oldName = slashClean(oldName)
		newName = slashClean(newName)
		if oldName == newName {
			return nil, nil
		}
		if strings.HasPrefix(newName, oldName+"/") {
			return nil, os.ErrInvalid // We can't rename oldName to be a sub-directory of itself.
		}

		oDir, oFrag, err := fs.Find(tx, "rename", oldName)
		if err != nil {
			return nil, err
		}
		if oDir == nil {
			return nil, os.ErrInvalid // We can't rename from the root.
		}

		nDir, nFrag, err := fs.Find(tx, "rename", newName)
		if err != nil {
			return nil, err
		}
		if nDir == nil {
			return nil, os.ErrInvalid // We can't rename to the root.
		}

		oNode, err := fs.Get(tx, oDir, oFrag)
		if err != nil && err != ErrNodeNotExist {
			return nil, err
		}

		if oNode == nil {
			return nil, os.ErrNotExist
		}

		if oNode.Mode.IsDir() { //old node is a dir
			nNode, _ := fs.Get(tx, nDir, nFrag)
			if nNode != nil {
				if !nNode.Mode.IsDir() { //new node must also new a directory
					return nil, errNotADirectory
				}

				empty := true
				if err = fs.Children(tx, nNode, func(name string, n *Node) bool {
					empty = false
					return true
				}); err != nil {
					return nil, err
				}

				if !empty { //new node (as directory) must be empty
					return nil, errDirectoryNotEmpty
				}
			}
		}

		//delete old node from old dir
		if err = fs.Del(tx, oDir, oFrag); err != nil {
			return nil, err
		}

		//place old on new parent
		if err = fs.Put(tx, nDir, nFrag, oNode); err != nil {
			return nil, err
		}

		return
	})

	return
}

func (fs *FDBFS) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	_, err = fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		dir, frag, err := fs.Find(tx, "mkdir", name) //find dir node that should contain the name
		if err != nil {
			return nil, err
		}
		if dir == nil { // check if we're creating the root
			return nil, os.ErrInvalid
		}

		if c, _ := fs.Get(tx, dir, frag); c != nil { //check if dir already exists
			return nil, os.ErrExist
		}

		if err = fs.Put(tx, dir, frag, &Node{
			ID:      genID(),
			Mode:    perm.Perm() | os.ModeDir,
			ModTime: time.Now(),
		}); err != nil {
			return nil, err
		}

		return
	})

	return
}

func (fs *FDBFS) Stat(ctx context.Context, name string) (fi os.FileInfo, err error) {
	_, err = fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		dir, frag, err := fs.Find(tx, "stat", name) //find parent dir by name
		if err != nil {
			return nil, err
		}
		if dir == nil {
			var root *Node
			root, err = fs.Root(tx)
			if err != nil {
				return nil, err
			}

			fi, _ = fs.stat(fs.tr, root.ID, "/")
			if err != nil {
				return nil, err
			}

			return fi, nil //read root's status information
		}

		var n *Node
		n, err = fs.Get(tx, dir, frag)
		if err != nil && err != ErrNodeNotExist {
			return nil, err
		}

		if n != nil {
			fi, err = fs.stat(fs.tr, n.ID, path.Base(name))
			if err != nil {
				return nil, err
			}

			return
		}

		return nil, os.ErrNotExist
	})

	return
}

func (fs *FDBFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (file File, err error) {
	_, err = fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		dir, frag, err := fs.Find(tx, "open", name) //find dir that should contain node
		if err != nil {
			return nil, err
		}
		var n *Node //work towards getting a files node
		if dir == nil {
			// We're opening the root.
			if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
				return nil, os.ErrPermission
			}

			n, err = fs.Root(tx)
			if err != nil {
				return nil, err
			}

			frag = "/"

		} else {
			n, err = fs.Get(tx, dir, frag)
			if err != nil && err != ErrNodeNotExist {
				return nil, err
			}

			if flag&(os.O_SYNC|os.O_APPEND) != 0 {
				return nil, os.ErrInvalid
			}
			if flag&os.O_CREATE != 0 {
				if flag&os.O_EXCL != 0 && n != nil {
					return nil, os.ErrExist
				}
				if n == nil {
					n = &Node{
						ID:   genID(),
						Mode: perm.Perm(),
					}

					err = fs.Put(tx, dir, frag, n)
					if err != nil {
						return nil, err
					}
				}
			}
			if n == nil {
				return nil, os.ErrNotExist
			}

			if flag&(os.O_WRONLY|os.O_RDWR) != 0 && flag&os.O_TRUNC != 0 {
				err = fs.trunc(fs.tr, n.ID)
				if err != nil {
					return nil, err
				}
			}
		}

		children := []os.FileInfo{}
		if err = fs.Children(tx, n, func(name string, nn *Node) (b bool) {
			fi, _ := fs.stat(fs.tr, nn.ID, name)
			children = append(children, fi)
			return
		}); err != nil {
			return nil, err
		}

		file = &fdbFile{
			n: &FDBFSNode{
				id:      n.ID,
				fs:      fs,
				mode:    n.Mode,
				modTime: n.ModTime,
			},
			nameSnapshot:     frag,
			childrenSnapshot: children,
		}

		return
	})

	return
}

// A FDBFSNode represents a single entry in the fbd filesystem
type FDBFSNode struct {
	id      nodeID
	fs      *FDBFS
	mode    os.FileMode
	modTime time.Time
}

func (n *FDBFSNode) stat(name string) (fi *fdbFileInfo, err error) {
	_, err = n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, err error) {
		var size int64
		size, err = n.fs.SizeOf(tx, n.id)
		if err != nil {
			return nil, err
		}

		fi = &fdbFileInfo{
			name:    name,
			size:    size,
			mode:    n.mode,
			modTime: n.modTime,
		}

		return
	})

	return
}

func (n *FDBFSNode) deadProps() (props map[xml.Name]Property, err error) {
	_, err = n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, err error) {
		props = map[xml.Name]Property{}
		if err = n.fs.Props(n.fs.tr, n.id, func(p Property) (r bool) {
			props[p.XMLName] = p
			return
		}); err != nil {
			return nil, err
		}

		return
	})

	return
}

func (n *FDBFSNode) patch(patches []Proppatch) (pstats []Propstat, err error) {
	_, err = n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, err error) {
		pstat := Propstat{Status: http.StatusOK}
		for _, patch := range patches {
			for _, p := range patch.Props {
				pstat.Props = append(pstat.Props, Property{XMLName: p.XMLName})
				if patch.Remove {

					_ = n.fs.DelProp(tx, n.id, p.XMLName)
					continue
				}

				_ = n.fs.PutProp(tx, n.id, p.XMLName, p)
			}
		}

		pstats = []Propstat{pstat}
		return
	})

	return
}

// Implements os.FileInfo
type fdbFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (f *fdbFileInfo) Name() string       { return f.name }
func (f *fdbFileInfo) Size() int64        { return f.size }
func (f *fdbFileInfo) Mode() os.FileMode  { return f.mode }
func (f *fdbFileInfo) ModTime() time.Time { return f.modTime }
func (f *fdbFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f *fdbFileInfo) Sys() interface{}   { return nil }
