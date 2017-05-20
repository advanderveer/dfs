package simple

import (
	"fmt"

	"github.com/billziss-gh/cgofuse/fuse"
)

//FS is a basic file system
type FS struct {
	fuse.FileSystemBase
	ReadyCh chan struct{}
}

//NewFS creates the filesystem
func NewFS() *FS {
	return &FS{
		ReadyCh: make(chan struct{}),
	}
}

const (
	filename = "hello"
	contents = "hello, world\n"
)

//Init is called when the file system is created.
func (fs *FS) Init() {
	fs.ReadyCh <- struct{}{}
}

//Open a file
func (fs *FS) Open(path string, flags int) (errc int, fh uint64) {
	switch path {
	case "/" + filename:
		return 0, 0
	default:
		return -fuse.ENOENT, ^uint64(0)
	}
}

//Getattr returns file attributes
func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	fmt.Println("path", path, "fh(0):", fh == ^uint64(0))
	switch path {
	case "/":
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	case "/" + filename:
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = int64(len(contents))
		return 0
	default:
		return -fuse.ENOENT
	}
}

//Read file contents
func (fs *FS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	endofst := ofst + int64(len(buff))
	if endofst > int64(len(contents)) {
		endofst = int64(len(contents))
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, contents[ofst:endofst])
	return
}

//Readdir read a path as a directory
func (fs *FS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)
	fill(filename, nil, 0)
	return 0
}
