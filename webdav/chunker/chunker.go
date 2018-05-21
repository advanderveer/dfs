package chunker

import (
	"bytes"
	"io"
)

//Chunker can chunk
type Chunker interface {
	Reset(r io.Reader)
	Next() (*Chunk, error)
}

//Chunk is returned by a chunker
type Chunk struct {
	Offset int64
	Data   []byte
}

type fixedSizeChunker struct {
	b []byte
	n int64
	r io.Reader
	p int64
}

//FixedSize chunker chunks data read from 'r' into sizes of length 'n'
func FixedSize(n int64) Chunker {
	return &fixedSizeChunker{n: n, r: bytes.NewReader(nil), b: make([]byte, n)}
}

func (c *fixedSizeChunker) Reset(r io.Reader) {
	c.r = r
	c.b = make([]byte, c.n)
}

func (c *fixedSizeChunker) Next() (*Chunk, error) {
	chunk := make([]byte, c.n)
	n, err := c.r.Read(chunk)
	if err != nil {
		return nil, err
	}

	res := &Chunk{
		Data:   chunk[:n],
		Offset: c.p,
	}

	c.p += int64(n)
	return res, nil
}

var _ Chunker = &fixedSizeChunker{}
