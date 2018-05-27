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
	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/ffs/handles"
	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/billziss-gh/cgofuse/fuse"
)

// Your (Storage) product is only as good as its test suite:
// 1/ https://blogs.oracle.com/bill/entry/zfs_and_the_all_singing
// 2/ tools: https://github.com/billziss-gh/secfs.test

//@TODO add a test that checks if the ino correct after new mount
//@TODO add a test that checks if remote dial fs works with torture and fsx

func TestQuickIO(t *testing.T) {
	bdir, err := ioutil.TempDir("", "dfs_")
	ok(t, err)

	db, dir, clean := testdb(bdir)
	defer clean()

	fmt.Println("scope", bdir)
	if runtime.GOOS == "windows" {
		t.Skip("no windows testing yet")
	} else {
		t.Run("linux/osx fuzzing", func(t *testing.T) {
			fmt.Println("blocks dir:", bdir)

			bstore, err := blocks.NewStore(bdir, "")
			if err != nil {
				t.Fatal("failed to create block store", err)
			}

			nstore := nodes.NewStore(db, dir)
			hstore := handles.NewStore(db, dir.Sub(tuple.Tuple{"handles"}), dir)

			defer bstore.Close()
			dfs, err := ffs.NewFS(nstore, bstore, hstore, func() (uint32, uint32, int) { return 1, 1, 1 })
			ok(t, err)

			svr, err := fsrpc.NewServer(dfs, ":")
			if err != nil {
				t.Fatal(err)
			}

			go svr.ListenAndServe()
			time.Sleep(time.Second)

			remotefs, err := fsrpc.Dial(svr.Addr().String())
			if err != nil {
				t.Fatal(err)
			}

			host := fuse.NewFileSystemHost(remotefs)
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
					errc := remotefs.Mkdir("/foobar", 0777)
					equals(t, 0, errc)
				})

				t.Run("test list xattr", func(t *testing.T) {
					errc := remotefs.Setxattr("/", "hello", []byte("bar"), 0)
					equals(t, 0, errc)

					errc, attr := remotefs.Getxattr("/", "hello")
					equals(t, 0, errc)
					equals(t, []byte("bar"), attr)

					var n int
					equals(t, 0, remotefs.Listxattr("/", func(name string) bool {
						n++
						equals(t, "hello", name)
						return true
					}))

					equals(t, 1, n)
				})

				dira := filepath.Join(mntdir, "a")
				err = os.Mkdir(dira, 0777)
				ok(t, err)

				// Seed set to 1527358203
				// All operations - 100 - completed A-OK!
				// === RUN   TestQuickIO/linux/osx_fuzzing/run_fstorture
				// root1 = /var/folders/8g/sd7s1zr94f948ds6_q0v3q180000gn/T/1527273901808862767_TestQuickIO/linux/osx_fuzzing/a does not exist
				// main_test.go:126: unexpected error: exit status 1
				//
				// === RUN   TestQuickIO/linux/osx_fuzzing/create_and_read_link
				// main_test.go:139: unexpected error: stat /var/folders/8g/sd7s1zr94f948ds6_q0v3q180000gn/T/1527273901808862767_TestQuickIO/linux/osx_fuzzing/c: no such file or directory
				//
				// === RUN   TestQuickIO/linux/osx_fuzzing/create_and_read_hard_link
				// === RUN   TestQuickIO/linux/osx_fuzzing/create_and_read_hard_link/through_link
				// time.Sleep(time.Second * 10) //@TODO remove me, sometimes the error above is shown

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

					fi, err := os.Stat(filepath.Join(mntdir, "a"))
					ok(t, err)
					_ = fi
					//@TODO test if uid/gid is masked correctly by client
					// st := fi.Sys().(*syscall.Stat_t)
					// equals(t, uint32(os.Getuid()), st.Uid)
					// equals(t, uint32(os.Getgid()), st.Gid)
					//
					// for _, fi := range fis {
					// 	st := fi.Sys().(*syscall.Stat_t)
					// 	equals(t, uint32(os.Getuid()), st.Uid)
					// 	equals(t, uint32(os.Getgid()), st.Gid)
					// }

					errc, fh := remotefs.Opendir("/")
					equals(t, 0, errc)

					//@TODO these will  cause fuse.Context to not work correctly
					equals(t, 0, remotefs.Readdir(mntdir, func(name string, st *fuse.Stat_t, ofst int64) bool {
						if st != nil {
							equals(t, uint32(os.Getuid()), st.Uid)
							equals(t, uint32(os.Getgid()), st.Gid)
						}

						return true
					}, 0, fh))
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
