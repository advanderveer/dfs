package simple

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/billziss-gh/cgofuse/fuse"
)

//FileHandle is our in-memory handle for open files
type FileHandle struct {
	openCount int
	*os.File
}

//FS is a basic file system
type FS struct {
	fuse.FileSystemBase
	ReadyCh chan struct{}

	dbdir   string
	handles map[uint64]*FileHandle
}

//NewFS creates the filesystem
func NewFS(dir string) *FS {
	return &FS{
		ReadyCh: make(chan struct{}),

		dbdir:   dir,
		handles: map[uint64]*FileHandle{},
	}
}

var (
	filename = "hello"
)

//Init is called when the file system is created.
func (fs *FS) Init() {
	fs.ReadyCh <- struct{}{}
}

//Open a file, returns a file discriptor that can be used for further interactions on the file.
func (fs *FS) Open(path string, flags int) (errc int, fd uint64) {
	switch path {
	case "/" + filename:
		var err error
		fh := &FileHandle{}
		fh.File, err = os.OpenFile(filepath.Join(fs.dbdir, path), flags, 0777)
		if err != nil {
			return -fuse.EIO, ^uint64(0)
		}

		fd = 100
		fh.openCount = 1
		fs.handles[fd] = fh
		return 0, fd //return a handle that is used by further calls
	default:
		return -fuse.ENOENT, ^uint64(0)
	}
}

//Getattr returns file attributes
func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fd uint64) (errc int) {
	switch path {
	case "/":
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	case "/" + filename:

		fi, err := os.Stat(filepath.Join(fs.dbdir, path))
		if err != nil {
			fmt.Println("stat", err)
			return -fuse.EIO
		}

		stat.Mode = fuse.S_IFREG | 0777 //@TODO set this correct perms
		// stat.Size = int64(len(contents))
		stat.Size = fi.Size()
		return 0
	default:
		return -fuse.ENOENT
	}
}

//Truncate changes the size of a file
func (fs *FS) Truncate(path string, size int64, fd uint64) int {
	err := fs.handles[fd].Truncate(size)
	if err != nil {
		fmt.Println("failed to truncate")
		return -fuse.EIO
	}

	// contents = ""
	return 0
}

//Release closes an open file.
func (fs *FS) Release(path string, fd uint64) int {
	err := fs.handles[fd].Close()
	if err != nil {
		fmt.Println("close", err)
		return -fuse.EIO
	}
	return 0
}

//Write writes data to a file.
func (fs *FS) Write(path string, buff []byte, ofst int64, fd uint64) (n int) {
	var err error
	n, err = fs.handles[fd].WriteAt(buff, ofst)
	if err != nil {
		fmt.Println("write at ", err)
		return -fuse.EIO
	}

	return n
}

//Read file contents
func (fs *FS) Read(path string, buff []byte, ofst int64, fd uint64) (n int) {
	var err error
	fh := fs.handles[fd]
	n, err = fh.ReadAt(buff, ofst)
	if err != nil {
		fmt.Println("readat", err)
		return -fuse.EIO
	}

	return n
}

//Readdir read a path as a directory
func (fs *FS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fd uint64) (errc int) {

	fill(".", nil, 0)
	fill("..", nil, 0)
	fill(filename, nil, 0)
	return 0
}
