package server

import (
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/rpc"
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

// func TestRPC(t *testing.T) {
// 	svr := rpc.NewServer()
// 	svr.RegisterName("A", new(Arith))
// 	l, err := net.Listen("tcp6", ":")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	go func() {
// 		t.Logf("accepting connections on: %v", l.Addr())
// 		for {
// 			var conn net.Conn
// 			conn, err = l.Accept()
// 			if err != nil {
// 				t.Fatal(err)
// 			}
//
// 			go svr.ServeConn(conn)
// 		}
// 	}()
//
// 	conn, err := net.DialTimeout("tcp6", l.Addr().String(), time.Second)
// 	if err != nil {
// 		t.Fatal("failed to dial", err)
// 	}
//
// 	c := rpc.NewClient(conn)
// 	reply := &Reply{}
// 	err = c.Call("A.Multiply", Args{7, 8}, reply)
// 	if err != nil {
// 		t.Fatal("failed to call", err)
// 	}
//
// 	fmt.Println(reply)
// }

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

	svr := rpc.NewServer()
	rcvr := &Receiver{fs: fs}
	svr.RegisterName("FS", rcvr)
	l, err := net.Listen("tcp6", ":")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		t.Logf("accepting connections on: %v", l.Addr())
		for {
			var conn net.Conn
			conn, err = l.Accept()
			if err != nil {
				t.Fatal(err)
			}

			go svr.ServeConn(conn)
		}
	}()

	conn, err := net.DialTimeout("tcp6", l.Addr().String(), time.Second)
	if err != nil {
		t.Fatal("failed to dial", err)
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
}
