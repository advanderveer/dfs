package fsrpc

import (
	"bytes"
	"net/rpc"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffshttp"
	"github.com/advanderveer/dfs/model"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
)

func TestHTTPRPC(t *testing.T) {
	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		t.Fatal(err)
	}

	fs, clean, err := ffs.NewTempFS("e2e", db)
	if err != nil {
		t.Fatal(err)
	}

	defer clean()
	m, clean2, err := model.New(db)
	if err != nil {
		t.Fatalf("failed to setup mode: %v", err)
	}

	defer clean2()
	fsr := New(fs)
	svr, err := ffshttp.NewServer(fsr, ffs.NewBrowser(fs), m, "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	go svr.Serve()
	time.Sleep(time.Second * 2)

	c, err := rpc.DialHTTPPath("tcp", svr.Addr().String(), "/fs")
	if err != nil {
		t.Fatal(err)
	}

	args := &StatfsArgs{Path: "/", Stat: &fuse.Statfs_t{}}
	reply := &StatfsReply{}
	err = c.Call("FS.Statfs", args, reply)
	if err != nil {
		t.Fatal("failed to call", err)
	}

	if reply.Args.Path != "/" {
		t.Fatal("expected reply to contain completed args")
	}

	if reply.Args.Stat.Bavail == 0 {
		t.Fatal("failed to return available blocks")
	}
}

//@TODO test byte copy off read procedure
func TestFSRPC(t *testing.T) {
	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		t.Fatal(err)
	}

	fs, clean, err := ffs.NewTempFS("e2e", db)
	if err != nil {
		t.Fatal(err)
	}

	defer clean()

	svr, err := NewServer(fs, "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	go svr.ListenAndServe()
	time.Sleep(time.Second)

	c, err := rpc.Dial("tcp", svr.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	args := &StatfsArgs{Path: "/", Stat: &fuse.Statfs_t{}}
	reply := &StatfsReply{}
	err = c.Call("FS.Statfs", args, reply)
	if err != nil {
		t.Fatal("failed to call", err)
	}

	if reply.Args.Path != "/" {
		t.Fatal("expected reply to contain completed args")
	}

	if reply.Args.Stat.Bavail == 0 {
		t.Fatal("failed to return available blocks")
	}

	sndr := &Sender{rpc: c, LastErr: nil, uid: uint32(os.Getuid()), gid: uint32(os.Getgid())}
	var _ FS = sndr //check if the rpc client statisfies the filesystem interface

	stfs := &fuse.Statfs_t{}
	errc := sndr.Statfs("/", stfs)
	if errc != 0 {
		t.Fatal("expected errc to be zero")
	}

	if sndr.LastErr != nil {
		t.Fatal("expected last error to be nil")
	}

	if stfs.Bavail == 0 {
		t.Fatalf("expected statf to return correct values, got: %#v\n", stfs)
	}

	t.Run("make dir remote, then read dirs", func(t *testing.T) {
		errc := sndr.Mkdir("/foo", 0777)
		if errc != 0 || sndr.LastErr != nil {
			t.Fatalf("failed to create dir (%d): %v\n", errc, sndr.LastErr)
		}

		errc, fh := sndr.Opendir("/")
		if errc != 0 || sndr.LastErr != nil {
			t.Fatal(err)
		}

		//@TODO test readdir and readxargs
		dirnames := []string{}
		if errc = sndr.Readdir("/", func(name string, stat *fuse.Stat_t, ofst int64) bool {
			dirnames = append(dirnames, name)
			return true
		}, 0, fh); errc != 0 || sndr.LastErr != nil {
			t.Fatalf("failed to create dir (%d): %v\n", errc, sndr.LastErr)
		}

		if reflect.DeepEqual(dirnames, []string{"..", ".", "foo"}) {
			t.Fatal("readdir should work")
		}
	})

	t.Run("xattr list", func(t *testing.T) {

		errc := sndr.Setxattr("/", "hello", []byte("bar"), 0)
		if errc != 0 || sndr.LastErr != nil {
			t.Fatal(sndr.LastErr)
		}

		errc, attr := sndr.Getxattr("/", "hello")
		if errc != 0 || sndr.LastErr != nil {
			t.Fatal(sndr.LastErr)
		}

		if !bytes.Equal(attr, []byte("bar")) {
			t.Fatal("expected the attr to equal")
		}

		attrs := []string{}
		if errc = sndr.Listxattr("/", func(name string) bool {
			attrs = append(attrs, name)
			return true
		}); errc != 0 || sndr.LastErr != nil {
			t.Fatal(sndr.LastErr)
		}
	})
}
