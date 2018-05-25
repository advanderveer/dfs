package fsrpc

import (
	"bytes"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/rpc"
	"reflect"
	"testing"
	"time"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/blocks"
	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/billziss-gh/cgofuse/fuse"
)

func db() (tr fdb.Transactor, ss directory.DirectorySubspace, f func()) {
	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	dir, err := directory.CreateOrOpen(db, []string{"fdb-tests", "litmus"}, nil)
	if err != nil {
		log.Fatal("failed to create or open app dir:", err)
	}

	return db, dir, func() {
		_, err := dir.Remove(db, nil)
		if err != nil {
			log.Fatal("failed to remove testing dir:", err)
		}
	}
}

//@TODO test byte copy off read procedure
func TestFSRPC(t *testing.T) {
	bdir, err := ioutil.TempDir("", "dfs_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, dir, clean := db()
	defer clean()

	bstore, err := blocks.NewStore(bdir, "blocks")
	if err != nil {
		t.Fatalf("failed to create block store: %v", err)
	}

	defer bstore.Close()
	fs, err := ffs.NewFS(nodes.NewStore(db, dir), bstore)
	if err != nil {
		t.Fatalf("failed to create filesystem: %v", err)
	}

	svr, err := NewServer(fs, ":")
	if err != nil {
		t.Fatal(err)
	}

	go svr.ListenAndServe()
	time.Sleep(time.Second)
	conn, err := net.DialTimeout("tcp", svr.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}

	c := rpc.NewClient(conn)
	args := &StatfsArgs{Path: "/", Stat: &fuse.Statfs_t{}}
	reply := &StatfsReply{}
	err = c.Call("FS.Statfs", args, reply)
	if err != nil {
		t.Fatal("failed to call", err)
	}

	if reply.Args.Path != "/" {
		t.Fatal("expected reply to contain completed args")
	}

	if reply.Args.Stat.Bavail != math.MaxUint64 {
		t.Fatal("failed to return available blocks")
	}

	sndr := &Sender{c, nil}
	var _ FS = sndr //check if the rpc client statisfies the filesystem interface

	stfs := &fuse.Statfs_t{}
	errc := sndr.Statfs("/", stfs)
	if errc != 0 {
		t.Fatal("expected errc to be zero")
	}

	if sndr.LastErr != nil {
		t.Fatal("expected last error to be nil")
	}

	if stfs.Bavail != math.MaxUint64 {
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
