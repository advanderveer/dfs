package webdav

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"testing"

	"github.com/advanderveer/fdb-tests/pkg/webdav/chunker"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"golang.org/x/net/context"
)

func TestChunkedWrite(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)

	ctx := context.Background()

	for ci, c := range []struct {
		with    []byte
		write   []byte
		at      int64
		exp     []byte
		expat   int64
		expn    int
		expeof  bool
		expsize int64
	}{
		{ //write initial data
			with:    []byte{},
			write:   []byte{0x01, 0x02},
			at:      0,
			exp:     []byte{0x01, 0x02},
			expn:    2,
			expsize: 2,
		},
		{ //write nothing, read part
			with:    []byte{0x01, 0x02},
			write:   []byte{},
			at:      0,
			exp:     []byte{0x02},
			expn:    1,
			expat:   1,
			expsize: 2,
		},
		{ //overwrite part of existing data
			with:    []byte{0x01, 0x02},
			write:   []byte{0x02, 0x03},
			at:      1,
			exp:     []byte{0x01, 0x02, 0x03},
			expn:    3,
			expat:   0,
			expsize: 3,
		},

		{ //write new part exactly at the end
			with:    []byte{0x01, 0x02},
			write:   []byte{0x03, 0x04},
			at:      2,
			exp:     []byte{0x01, 0x02, 0x03, 0x04},
			expn:    4,
			expat:   0,
			expsize: 4,
		},

		{ //write new part with uneven chunk before
			with:  []byte{0x01},
			write: []byte{0x02, 0x03}, at: 1,
			exp:     []byte{0x01, 0x02, 0x03},
			expn:    3,
			expat:   0,
			expsize: 3,
		},

		{ //overwrite with multiple chunks
			with:    []byte{0x01, 0x02, 0x03},
			write:   []byte{0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			at:      1,
			exp:     []byte{0x02, 0x03, 0x04, 0x05},
			expn:    4,
			expat:   1,
			expsize: 7,
		},
		{ //overwrite one chunk exactly in the middle
			with:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			write:   []byte{0x03, 0x04},
			at:      2,
			exp:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			expn:    6,
			expat:   0,
			expsize: 6,
		},

		{ //write nothing read nothing
			with:    []byte{},
			write:   []byte{},
			at:      0,
			exp:     []byte{},
			expn:    0,
			expat:   0,
			expsize: 0,
		},
		{ //read until end
			//ReadAt reads len(b) bytes from the File starting at byte offset off. It returns the number of
			//bytes read and the error, if any. ReadAt always returns a non-nil error when n < len(b). At end of
			//file, that error is io.EOF.
			with:    []byte{},
			write:   []byte{},
			at:      0,
			exp:     make([]byte, 1),
			expat:   0,
			expn:    0, //n < len(b)
			expeof:  true,
			expsize: 0,
		},
		{ //read midway without content
			with:    []byte{},
			write:   []byte{},
			at:      0,
			exp:     []byte{},
			expat:   1,
			expn:    0,
			expeof:  true,
			expsize: 0,
		},

		{ //write hole without any content
			with:    []byte{},
			write:   []byte{0x05},
			at:      3,
			exp:     []byte{0x00, 0x00, 0x00, 0x05},
			expat:   0,
			expn:    4,
			expsize: 4,
		},
	} {
		name := fmt.Sprintf("with %v write %v at %d, read %d at %d", c.with, c.write, c.at, len(c.exp), c.expat)
		t.Run(name, func(t *testing.T) {
			f, err := fs.OpenFile(ctx, fmt.Sprintf("%d.txt", ci), os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				t.Fatal("failed to open file:", err)
			}

			defer f.Close()
			fv2, ok := f.(*fdbFile)
			if !ok {
				t.Fatal("must be file version 2")
			}

			n, err := fv2.n.fs.wAt(fs.tr, chunker.FixedSize(2), fv2.n.id, 0, c.with)
			if err != nil || n != len(c.with) {
				t.Fatalf("failed to write base data, expected %d, got %d, err: %v", len(c.with), n, err)
			}

			n, err = fv2.n.fs.wAt(fs.tr, chunker.FixedSize(2), fv2.n.id, c.at, c.write)
			if err != nil || n != len(c.write) {
				t.Fatalf("failed to write data, expected %d, got %d, err: %v", len(c.write), n, err)
			}

			buf := make([]byte, len(c.exp))
			n, err = fv2.n.fs.rAt(fs.tr, fv2.n.id, c.expat, buf)
			if c.expeof && err != io.EOF {
				t.Fatal("should get EOF error", err) //either nil or somet other error
			} else if !c.expeof && err != nil {
				t.Fatal("unexpected error", err)
			}

			if n != c.expn || !bytes.Equal(buf, c.exp) {
				t.Fatalf("unexpected data at %d, expected %v, got %v (n=%d), err: %v", c.expat, c.exp, buf, n, err)
			}

			size, err := fv2.n.fs.SizeOf(fs.tr, fv2.n.id)
			if err != nil {
				t.Fatal("failed to get size", err)
			}

			if size != c.expsize {
				t.Fatalf("expected size to be %d, got %d", c.expsize, size)
			}

			prevSize := -1
			fv2.n.fs.tr.Transact(func(tx fdb.Transaction) (interface{}, error) {
				if err = walkChunks(tx, fv2.n.fs.nodes, fv2.n.id, -1, math.MaxInt64, func(i int, o int64, data []byte) bool {
					if prevSize > -1 && prevSize != 2 {
						t.Fatal("invalid block structure")
					}

					prevSize = len(data)
					return false
				}); err != nil {
					t.Fatal("failed to walk for chunk structure")
				}

				return nil, nil
			})

			if err = fs.trunc(fs.tr, fv2.n.id); err != nil {
				t.Fatal("failed to truncate", err)
			}

			size, err = fs.SizeOf(fs.tr, fv2.n.id)
			if err != nil {
				t.Fatal("failed to get size", err)
			}

			if size != 0 {
				t.Fatalf("expected size to be %d, got %d", 0, size)
			}
		})

	}
}
