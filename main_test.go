package main

import (
	"bytes"
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

func TestEnd2End(t *testing.T) {
	bdir, err := ioutil.TempDir("", "dfs_")
	ok(t, err)

	db, dir, clean := testdb(bdir)
	defer clean()

	bstore, err := blocks.NewStore(bdir, "")
	if err != nil {
		t.Fatal("failed to create block store", err)
	}

	nstore := nodes.NewStore(db, dir)
	hstore := handles.NewStore(db, dir.Sub(tuple.Tuple{"handles"}), dir)

	defer bstore.Close()
	dfs, err := ffs.NewFS(nstore, bstore, hstore, func() (uint32, uint32, int) { return 1, 1, 1 })
	ok(t, err)

	svr, err := fsrpc.NewServer(dfs, "localhost:")
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

	var mntdir string
	switch runtime.GOOS {
	case "darwin", "linux":
		mntdir = filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), "e2e"))
	case "windows":
		mntdir = "M:"
	default:
		t.Skipf("no e2e testing available for '%s'", runtime.GOOS)
	}

	go func() {
		for {
			fi, err := os.Stat(mntdir)
			if err == nil && fi.IsDir() {
				break
			}
		}

		//common test
		time.Sleep(time.Millisecond * 200)
		CommonEnd2End(mntdir, t)

		//platform specific e2e tests
		if !testing.Short() {
			switch runtime.GOOS {
			case "windows":
				WindowsEnd2End(mntdir, t)
			case "darwin", "linux":
				LinuxDarwinEnd2End(mntdir, remotefs, t)
			}
		}

		//wait a bit and unmount
		time.Sleep(time.Second)
		ok := host.Unmount()
		equals(t, true, ok)
	}()

	//mount until either win or linux testing decides to end the mount
	fmt.Println("Mounting...")
	ok := host.Mount(mntdir, []string{})
	equals(t, true, ok)
}

func CommonEnd2End(mntdir string, t *testing.T) {
	t.Run("basic file writing", func(t *testing.T) {
		f, err := os.Create(filepath.Join(mntdir, "bar.txt"))
		ok(t, err)
		defer f.Close()

		n, err := f.WriteAt([]byte{0x04, 0x05, 0x06}, 3)
		ok(t, err)
		equals(t, 3, n)

		n, err = f.WriteAt([]byte{0x01, 0x02, 0x03}, 0)
		ok(t, err)
		equals(t, 3, n)

		buf := make([]byte, 6)
		n, err = f.ReadAt(buf, 0)
		ok(t, err)
		equals(t, 6, n)
		equals(t, true, bytes.Equal(buf, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}))
	})
}

func WindowsEnd2End(mntdir string, t *testing.T) {
	cmd := exec.Command("winfsp-tests-x64.exe",
		"--external",
		"--resilient",
		// `--share-prefix=\gomemfs\share`,
		"-create_allocation_test",
		"-create_fileattr_test",
		"-getfileinfo_name_test",
		"-setfileinfo_test",
		"-delete_access_test",
		"-setsecurity_test",
		"-querydir_namelen_test",
		"-reparse*",
		"-stream*")

	cmd.Dir = mntdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	ok(t, err)
	equals(t, true, cmd.ProcessState.Success())
}

func LinuxDarwinEnd2End(mntdir string, remotefs fuse.FileSystemInterface, t *testing.T) {
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
	err := os.Mkdir(dira, 0777)
	ok(t, err)

	t.Run("run fsx", func(t *testing.T) {
		cmd := exec.Command("fsx", "-N", "5000", "test", "xxxxxx")
		cmd.Dir = mntdir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
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
}
