package webdav

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/advanderveer/fdb-tests/pkg/webdav/chunker"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

// A fdbFile2 is a File implementation for a fdbFSNode. It is a per-file (not
// per-node) read/write position, and a snapshot of the fdbFS' tree structure
// (a node's name and children) for that node.
type fdbFile struct {
	n                *FDBFSNode
	nameSnapshot     string
	childrenSnapshot []os.FileInfo
	pos              int
}

// A *fdbFile implements the optional DeadPropsHolder interface.
var _ DeadPropsHolder = (*fdbFile)(nil)

func (f *fdbFile) DeadProps() (map[xml.Name]Property, error)     { return f.n.deadProps() }
func (f *fdbFile) Patch(patches []Proppatch) ([]Propstat, error) { return f.n.patch(patches) }

func (f *fdbFile) Close() (err error) { return }

func (f *fdbFile) Readdir(count int) (fis []os.FileInfo, err error) {
	_, err = f.n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		if !f.n.mode.IsDir() {
			return nil, os.ErrInvalid
		}
		old := f.pos
		if old >= len(f.childrenSnapshot) {
			// The os.File Readdir docs say that at the end of a directory,
			// the error is io.EOF if count > 0 and nil if count <= 0.
			if count > 0 {
				return nil, io.EOF
			}
			return nil, nil
		}
		if count > 0 {
			f.pos += count
			if f.pos > len(f.childrenSnapshot) {
				f.pos = len(f.childrenSnapshot)
			}
		} else {
			f.pos = len(f.childrenSnapshot)
			old = 0
		}

		fis = f.childrenSnapshot[old:f.pos]
		return
	})

	return
}

func (f *fdbFile) Seek(offset int64, whence int) (pos int64, err error) {
	_, err = f.n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		size, err := f.n.fs.SizeOf(tx, f.n.id)
		if err != nil {
			return nil, err
		}

		npos := f.pos
		// TODO: How to handle offsets greater than the size of system int?
		switch whence {
		case os.SEEK_SET:
			npos = int(offset)
		case os.SEEK_CUR:
			npos += int(offset)
		case os.SEEK_END:
			npos = int(size) + int(offset)
		default:
			npos = -1
		}
		if npos < 0 {
			return int64(0), os.ErrInvalid
		}
		f.pos = npos
		pos = int64(f.pos)

		return
	})

	return
}

func (f *fdbFile) Stat() (os.FileInfo, error) {
	return f.n.stat(f.nameSnapshot)
}

func (f *fdbFile) Read(p []byte) (n int, err error) {
	_, err = f.n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		if f.n.mode.IsDir() {
			return 0, os.ErrInvalid
		}

		n, err = f.n.fs.rAt(tx, f.n.id, int64(f.pos), p)
		if err != nil {
			if err != io.EOF {
				err = fmt.Errorf("failed to read %d bytes at %d of node %x: %v", len(p), f.pos, f.n.id, err)
			}

			return n, err
		}

		f.pos += n
		return
	})

	return
}

func (f *fdbFile) Write(p []byte) (n int, err error) {
	_, err = f.n.fs.tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		lenp := len(p)
		if f.n.mode.IsDir() {
			return 0, os.ErrInvalid
		}

		n, err = f.n.fs.wAt(tx, chunker.FixedSize(f.n.fs.maxChunkSize), f.n.id, int64(f.pos), p)
		if err != nil {
			return 0, fmt.Errorf("failed to write %d bytes at %d of node %x: %v", len(p), f.pos, f.n.id, err)
		}

		f.pos += n

		f.n.modTime = time.Now() //@TODO write to disk, test changed mod time after write
		return lenp, nil
	})

	return
}
