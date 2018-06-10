package ffs

import (
	"fmt"
	"os"
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
)

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (f *fileInfo) Name() string       { return f.name }
func (f *fileInfo) Size() int64        { return f.size }
func (f *fileInfo) Mode() os.FileMode  { return f.mode }
func (f *fileInfo) ModTime() time.Time { return f.modTime }
func (f *fileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f *fileInfo) Sys() interface{}   { return nil }

type Browser struct {
	fs *Memfs
}

func NewBrowser(fs *Memfs) (b *Browser) {
	b = &Browser{fs: fs}
	return
}

func (b *Browser) Readdir(path string, f func(name string, fi os.FileInfo)) (err error) {
	errc, fh := b.fs.Opendir(path)
	if errc != 0 {
		return fmt.Errorf("failed to open dir: %d", errc)
	}

	defer b.fs.Releasedir("", fh)
	if errc = b.fs.Readdir(path, func(name string, stat *fuse.Stat_t, ofst int64) bool {
		fi := &fileInfo{name: name}

		if stat != nil {
			fi.size = stat.Size
			fi.mode = os.FileMode(stat.Mode & fuse.S_IFMT)
			fi.modTime = stat.Mtim.Time()
		}

		f(name, fi)
		return true
	}, 0, fh); errc != 0 {
		return fmt.Errorf("failed to read dir: %d", errc)
	}

	return
}
