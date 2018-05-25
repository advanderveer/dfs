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

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/blocks"
	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/advanderveer/dfs/ffs/server"
	"github.com/billziss-gh/cgofuse/fuse"
)

// Your (Storage) product is only as good as its test suite:
// 1/ https://blogs.oracle.com/bill/entry/zfs_and_the_all_singing
// 2/ tools: https://github.com/billziss-gh/secfs.test

func TestQuickIO(t *testing.T) {
	bdir, err := ioutil.TempDir("", "dfs_")
	ok(t, err)

	db, dir, clean := db()
	defer clean()

	if runtime.GOOS == "windows" {
		t.Skip("no windows testing yet")
	} else {
		t.Run("linux/osx fuzzing", func(t *testing.T) {
			fmt.Println("blocks dir:", bdir)

			bstore, err := blocks.NewStore(bdir, "")
			if err != nil {
				t.Fatal("failed to create block store", err)
			}

			defer bstore.Close()
			dfs, err := ffs.NewFS(nodes.NewStore(db, dir), bstore)
			ok(t, err)

			var fsiface fuse.FileSystemInterface = dfs
			fsiface, err = server.NewSimpleRPCFS(dfs)
			if err != nil {
				t.Fatalf("failed to create rpc filesyste,")
			}

			host := fuse.NewFileSystemHost(fsiface)
			host.SetCapReaddirPlus(true)

			mntdir := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), t.Name()))

			go func() {

				for {
					fi, err := os.Stat(mntdir)
					if err == nil && fi.IsDir() {
						break
					}
				}

				t.Run("basic ops", func(t *testing.T) {
					errc := dfs.Mkdir("/foobar", 0777)
					equals(t, 0, errc)
				})

				t.Run("test list xattr", func(t *testing.T) {
					errc := dfs.Setxattr("/", "hello", []byte("bar"), 0)
					equals(t, 0, errc)

					errc, attr := dfs.Getxattr("/", "hello")
					equals(t, 0, errc)
					equals(t, []byte("bar"), attr)

					var n int
					equals(t, 0, dfs.Listxattr("/", func(name string) bool {
						n++
						equals(t, "hello", name)
						return true
					}))

					equals(t, 1, n)
				})

				dira := filepath.Join(mntdir, "a")
				err = os.Mkdir(dira, 0777)
				ok(t, err)

				t.Run("run fsx", func(t *testing.T) {
					cmd := exec.Command("fsx", "-N", "5000", "test", "xxxxxx")
					cmd.Dir = mntdir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					ok(t, err)
					equals(t, true, cmd.ProcessState.Success())

					cmd = exec.Command("fsx", "-e", "-N", "100", "test", "xxxxxx")
					cmd.Dir = mntdir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err = cmd.Run()
					ok(t, err)
					equals(t, true, cmd.ProcessState.Success())
				})

				t.Run("run fstorture", func(t *testing.T) {

					//fstorture
					dirb := filepath.Join(mntdir, "b")
					err = os.Mkdir(dirb, 0777)
					ok(t, err)

					cmd := exec.Command("fstorture", dira, dirb, "6", "-c", "30")
					cmd.Dir = mntdir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err = cmd.Run()
					ok(t, err)
					equals(t, true, cmd.ProcessState.Success())
				})

				t.Run("create and read link", func(t *testing.T) {
					err := os.Symlink(dira, filepath.Join(mntdir, "c"))
					ok(t, err)

					lnk, err := os.Readlink(filepath.Join(mntdir, "c"))
					ok(t, err)
					equals(t, dira, lnk)

					fi, err := os.Stat(filepath.Join(mntdir, "c")) //@TODO sometimes fails?
					ok(t, err)
					equals(t, fi.IsDir(), true)
				})

				t.Run("create and read hard link", func(t *testing.T) {
					d := filepath.Join(mntdir, "d")
					err := ioutil.WriteFile(d, []byte{0x01}, 0777)
					ok(t, err)

					data, err := ioutil.ReadFile(filepath.Join(mntdir, "d"))
					ok(t, err)
					equals(t, []byte{0x01}, data)

					t.Run("through link", func(t *testing.T) {
						err = os.Link(d, filepath.Join(mntdir, "e"))
						ok(t, err)

						data, err := ioutil.ReadFile(filepath.Join(mntdir, "e"))
						ok(t, err)
						equals(t, []byte{0x01}, data)
					})

				})

				t.Run("read dir", func(t *testing.T) {
					fis, err := ioutil.ReadDir(mntdir)
					ok(t, err)
					assert(t, len(fis) > 0, "expected at least some listings, got: %d", len(fis))
				})

				time.Sleep(time.Second * 5)

				//done, unmount
				ok := host.Unmount()
				equals(t, true, ok)
			}()

			ok := host.Mount(mntdir, []string{})
			equals(t, true, ok)
		})
	}

	//think of a turn based locking mechanism, that is passed around based on general "activity" on a sub-tree: Allow lower resolution of locking and releasing (e.g every few seconds). Allow uncontented (high performance) locking of a certain subtree.

	//on the brokeness of linux locking: http://0pointer.de/blog/projects/locking.html
	//also: http://0pointer.de/blog/projects/locking2
	//samba file locking: https://www.samba.org/samba/news/articles/low_point/tale_two_stds_os2.html
}
