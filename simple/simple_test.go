package simple_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/advanderveer/dfs/simple"
	"github.com/billziss-gh/cgofuse/fuse"
)

func TestReadingFIles(t *testing.T) {
	dir, err := ioutil.TempDir("", "dfs_")
	ok(t, err)

	_ = dir

	t.Run("fs_1", func(t *testing.T) {
		fs := simple.NewFS()
		host := fuse.NewFileSystemHost(fs)
		dir := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), t.Name()))

		go func() {
			ok := host.Mount(dir, []string{})
			equals(t, true, ok)
		}()

		<-fs.ReadyCh

		t.Run("read file", func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join(dir, "hello"))
			ok(t, err)
			equals(t, "hello, world\n", string(data))
		})

		t.Run("unmounting", func(t *testing.T) {
			ok := host.Unmount()
			equals(t, true, ok)
		})
	})
}
