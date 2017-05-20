package simple_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/advanderveer/dfs/simple"
	"github.com/billziss-gh/cgofuse/fuse"
)

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func TestReadingFIles(t *testing.T) {
	dbdir, err := ioutil.TempDir("", "dfs_")
	ok(t, err)
	err = ioutil.WriteFile(filepath.Join(dbdir, "hello"), []byte("hello, world\n"), 0777)
	ok(t, err)

	defer func() {
		data, err := ioutil.ReadFile(filepath.Join(dbdir, "hello"))
		ok(t, err)
		equals(t, "hello, fuse\n", string(data))
	}()

	t.Run("fs_1", func(t *testing.T) {
		fs := simple.NewFS(dbdir)
		host := fuse.NewFileSystemHost(fs)
		fsdir := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), t.Name()))

		go func() {
			ok := host.Mount(fsdir, []string{})
			equals(t, true, ok)
		}()

		<-fs.ReadyCh

		t.Run("read file", func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join(fsdir, "hello"))
			ok(t, err)
			equals(t, "hello, world\n", string(data))
		})

		t.Run("write file", func(t *testing.T) {
			err := WriteFile(filepath.Join(fsdir, "hello"), []byte("hello, fuse\n"), 0777)
			ok(t, err)
		})

		t.Run("read file again", func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join(fsdir, "hello"))
			ok(t, err)
			equals(t, "hello, fuse\n", string(data))
		})

		t.Run("unmounting", func(t *testing.T) {
			ok := host.Unmount()
			equals(t, true, ok)
		})
	})
}
