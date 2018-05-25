package main

import (
	"log"
	"os"

	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/billziss-gh/cgofuse/fuse"
)

func db(ns string) (tr fdb.Transactor, ss directory.DirectorySubspace, f func()) {
	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	dir, err := directory.CreateOrOpen(db, []string{"fdb-tests", ns}, nil)
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

func main() {
	logs := log.New(os.Stderr, "ffs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("ffs [addr] [mountpoint]")
	}

	uid := os.Getuid()
	gid := os.Getgid()
	logs.Printf("mounting filesystem from '%s' at '%s' (uid: %d, gid: %d)", os.Args[1], os.Args[2], uid, gid)
	defer logs.Printf("unmounted, done!")

	// conn, err := net.DialTimeout("tcp", os.Args[1], time.Second*2)
	// if err != nil {
	// 	logs.Fatalf("failed to dial: %v", err)
	// }

	// err := os.MkdirAll(os.Args[1], 0777)
	// if err != nil {
	// 	logs.Fatalf("failed to create block storage dir: %v", err)
	// }
	//
	// // db, dir, _ := db(os.Args[1])
	// // defer clean()
	//
	// // bstore, err := blocks.NewStore(os.Args[1], "blocks")
	// // if err != nil {
	// // 	logs.Fatalf("failed to create block store: %v", err)
	// // }
	// //
	// // defer bstore.Close()
	// // fs, err := ffs.NewFS(nodes.NewStore(db, dir), bstore)
	// // if err != nil {
	// // 	logs.Fatalf("failed to create filesystem: %v", err)
	// // }
	//
	// // var fsiface fuse.FileSystemInterface = fs
	// // fsiface, err = server.NewSimpleRPCFS(fs)
	// // if err != nil {
	// // 	logs.Fatalf("failed to create rpc filesyste,")
	// // }

	fs, err := fsrpc.Dial(os.Args[1])
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	host := fuse.NewFileSystemHost(fs)
	if !host.Mount("", os.Args[2:]) {
		os.Exit(1) //mount failed
	}
}
