package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/advanderveer/dfs/dfs"
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

// Your (Storage) product is only as good as its test suite:
// 1/ https://blogs.oracle.com/bill/entry/zfs_and_the_all_singing
// 2/ tools: https://github.com/billziss-gh/secfs.test

func TestQuickIO(t *testing.T) {
	dir, err := ioutil.TempDir("", "dfs_")
	ok(t, err)

	db, err := bolt.Open(filepath.Join(dir, "buf.db"), 0600, nil)
	ok(t, err)

	if runtime.GOOS == "windows" {
		t.Skip("no windows testing yet")
	} else {
		t.Run("linux/osx fuzzing", func(t *testing.T) {
			dfs := dfs.NewFS(db)
			host := fuse.NewFileSystemHost(dfs)
			dir := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), t.Name()))

			go func() {
				for {
					fi, err := os.Stat(dir)
					if err == nil && fi.IsDir() {
						break
					}
				}

				//fsx
				cmd := exec.Command("fsx", "-N", "5000", "test", "xxxxxx")
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				ok(t, err)
				equals(t, true, cmd.ProcessState.Success())

				//fsx (attr)
				cmd = exec.Command("fsx", "-e", "-N", "100", "test", "xxxxxx")
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				ok(t, err)
				equals(t, true, cmd.ProcessState.Success())

				//fstorture
				dira := filepath.Join(dir, "a")
				err = os.Mkdir(dira, 0777)
				ok(t, err)

				dirb := filepath.Join(dir, "b")
				err = os.Mkdir(dirb, 0777)
				ok(t, err)

				cmd = exec.Command("fstorture", dira, dirb, "6", "-c", "30")
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				ok(t, err)
				equals(t, true, cmd.ProcessState.Success())

				//done, unmount
				ok := host.Unmount()
				equals(t, true, ok)
			}()

			ok := host.Mount(dir, []string{})
			equals(t, true, ok)
		})
	}

	//think of a turn based locking mechanism, that is passed around based on general "activity" on a sub-tree: Allow lower resolution of locking and releasing (e.g every few seconds). Allow uncontented (high performance) locking of a certain subtree.

	//on the brokeness of linux locking: http://0pointer.de/blog/projects/locking.html
	//also: http://0pointer.de/blog/projects/locking2
	//samba file locking: https://www.samba.org/samba/news/articles/low_point/tale_two_stds_os2.html
}
