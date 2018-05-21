package chunker

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestChunker(t *testing.T) {
	in := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
	}
	c1 := FixedSize(2)
	c1.Reset(bytes.NewBuffer(in))

	out := []byte{}
	offs := []int64{}
	for {
		c, err := c1.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			t.Fatal(err)
		}

		out = append(out, c.Data...)
		offs = append(offs, c.Offset)
	}

	if !reflect.DeepEqual(offs, []int64{0, 2, 4, 6, 8}) {
		t.Fatal("unexpected offsets", offs)
	}

	if !bytes.Equal(out, in) {
		t.Fatal("expected chunker input to equal output")
	}
}
