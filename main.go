package main

import (
	"log"
	"os"

	"github.com/advanderveer/dfs/ffs"
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

func main() {
	logs := log.New(os.Stderr, "dfs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("please provide the mount path and db dir")
	}

	logs.Printf("mounting filesystem from '%s'", os.Args[1])
	defer logs.Printf("unmounted, done!")

	err := os.MkdirAll(os.Args[1], 0777)
	if err != nil {
		logs.Fatalf("failed to create storage dir: %v", err)
	}

	db, dir, _ := db()
	fs, err := ffs.NewFS(nodes.NewStore(db, dir))
	if err != nil {
		logs.Fatalf("failed to create filesystem: %v", err)
	}

	host := fuse.NewFileSystemHost(fs)
	if !host.Mount("", os.Args[2:]) {
		os.Exit(1) //mount failed
	}
}
