package webdav

import (
	"bytes"
	"errors"
	"io"
	"math"
	"os"

	"github.com/advanderveer/fdb-tests/pkg/webdav/chunker"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	fdbdir "github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

func walkChunks(tx fdb.Transaction, nodes fdbdir.DirectorySubspace, nid nodeID, from, to int64, f func(i int, o int64, data []byte) bool) error {
	r := fdb.KeyRange{
		Begin: nodes.Pack(tuple.Tuple{nid[:], "c", from}),
		End:   nodes.Pack(tuple.Tuple{nid[:], "c", to}),
	}

	var rev bool
	if from > to {
		rev = true
		r = fdb.KeyRange{
			Begin: nodes.Pack(tuple.Tuple{nid[:], "c", to}),
			End:   nodes.Pack(tuple.Tuple{nid[:], "c", from}),
		}
	}

	iter := tx.GetRange(r, fdb.RangeOptions{Reverse: rev}).Iterator()

	var i int
	for iter.Advance() {
		kv := iter.MustGet()
		t, e := nodes.Unpack(kv.Key)
		if e != nil {
			return e //unexpected key
		}

		if len(t) < 3 {
			return errors.New("unexpected tuple format") //unexpected key
		}

		offset, ok := t[2].(int64)
		if !ok {
			return errors.New("no offset in tuple")
		}

		stop := f(i, offset, kv.Value)
		if stop {
			break
		}

		i++
	}

	return nil
}

func (fs *FDBFS) stat(tr fdb.Transactor, nid nodeID, name string) (fi os.FileInfo, err error) {
	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		node, err := fs.getHdr(tx, nid)
		if err != nil {
			return nil, err
		}

		size, err := fs.SizeOf(tx, nid)
		if err != nil {
			return nil, err
		}

		fi = &fdbFileInfo{
			name:    name,
			size:    size,
			mode:    node.Mode,
			modTime: node.ModTime,
		}

		return
	})

	return
}

func (fs *FDBFS) trunc(tr fdb.Transactor, nid nodeID) (err error) {
	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		tx.ClearRange(fdb.KeyRange{
			Begin: fs.nodes.Pack(tuple.Tuple{nid[:], "c", -1}),
			End:   fs.nodes.Pack(tuple.Tuple{nid[:], "c", math.MaxInt64}),
		})

		return
	})

	return
}

func (fs *FDBFS) rAt(tr fdb.Transactor, nid nodeID, pos int64, p []byte) (n int, err error) {
	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		var (
			selection []byte //our selection relevant chunkcs concatenated
			selOffset int64  //offset at which our selection start
		)

		//step 1: get selection by reading all chunks
		if err = walkChunks(tx, fs.nodes, nid, pos+int64(len(p)), -1, func(i int, o int64, d []byte) (ok bool) {
			e := o + int64(len(d))
			if e <= pos {
				selOffset = e
				return true
			}

			selection = append(d, selection...)
			return
		}); err != nil {
			return nil, err
		}

		//step 2: read correct bytes from selection
		relPos := pos - selOffset
		if relPos < 0 {
			return nil, errors.New("wedav: encountered negative replos for selection writing")
		}

		if int(relPos) > len(selection) {
			return nil, io.EOF //we selected pas what the selection can give use: EOF
		}

		n = copy(p, selection[relPos:])
		if n < len(p) {
			return nil, io.EOF //nothing more to copy: EOF
		}

		return
	})

	return
}

func (fs *FDBFS) SizeOf(tr fdb.Transactor, nid nodeID) (size int64, err error) {
	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		if err = walkChunks(tx, fs.nodes, nid, math.MaxInt64, -1, func(i int, off int64, cdata []byte) bool {
			size = off + int64(len(cdata))
			return true
		}); err != nil {
			return nil, err
		}
		return
	})

	return
}

func (fs *FDBFS) wAt(tr fdb.Transactor, c chunker.Chunker, nid nodeID, pos int64, p []byte) (n int, err error) {
	_, err = tr.Transact(func(tx fdb.Transaction) (r interface{}, e error) {
		var (
			selReader io.Reader //our selection is the relevant data that is to change
			selOffset int64     //offset at which our selection start
		)

		//@TODO add case that if the pos is larger then the current size, we create
		//a reader that can be read to provide zero bytes to the chunker with the
		//new data at the end, such that we don't have to load it all into memory
		//the chunker can then decide to chunk empty blocks differently
		if true {
			var selection []byte //case for selections that fit into memory

			//[1,2][3,4][4]   _    --> [1,2][3,4][4,0][0,7][8]
			//[1,2][3,4][4,5]_      --> [1,2][3,4][4,5][6,7][8]
			//[1,2][3,_][4,5]       --> [1,2][3,4][4,5]

			//step 1: read all existing data that may be affected by our write by walking back
			if err = walkChunks(tx, fs.nodes, nid, pos+int64(len(p)), -1, func(i int, o int64, data []byte) (stop bool) {
				e := o + int64(len(data))
				if e < pos {
					selOffset = e
					return true
				}

				selection = append(data, selection...)
				return
			}); err != nil {
				return nil, err
			}

			//step 2: write actual byte changes to our selection
			if len(selection) == 0 {
				if pos > selOffset {
					//ther are no bytes between position and the last chunk (or the beginning)
					//since our selection is empty. Yet, our position is larger then the last
					//seen chunk byte so we prepend the selection with zero to create a hole
					selection = append(make([]byte, (pos-selOffset)), p...)
				} else {
					selection = p
				}

				n = len(p) //we will always write p as amount now
			} else {
				relPos := pos - selOffset
				if relPos < 0 {
					//this occurs if the first chunk with offset larger then our pos-maxChunkSize
					//is larger then the position, this means we're writing to an empty part of the file
					//this should probably never occur as we'll be creating empty chunks
					return nil, errors.New("not implemented")
				}

				//if necessary grow the selection slice to accomodate writing new data to the end
				minSelSize := int(relPos) + len(p)
				if minSelSize > len(selection) {
					selection = append(selection, make([]byte, minSelSize-(len(selection)))...)
				}

				n = copy(selection[relPos:], p) //we write
				if n < len(p) {
					//somehow the selected bytes from existing chunks are not large enough to copy over p
					return nil, errors.New("selection slice was not big enough to accomodate new data")
				}
			}

			selReader = bytes.NewBuffer(selection)
		}

		//step 3: (re)chunk our selection and persist to database
		c.Reset(selReader)
		for {
			c, err := c.Next()
			if err != nil {
				if err == io.EOF {
					break
				}

				return nil, err
			}

			absOffset := selOffset + c.Offset
			tx.Set(fs.nodes.Pack(tuple.Tuple{nid[:], "c", absOffset}), c.Data)
		}

		return
	})

	return
}
