package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/blocks"
	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/ffs/nodes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	logs := log.New(os.Stderr, "ffs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("ffsvr [dbdir] [addr]")
	}

	err := os.MkdirAll(os.Args[1], 0777)
	if err != nil {
		logs.Fatalf("failed to create block storage dir: %v", err)
	}

	db, dir, clean := db(os.Args[1])
	defer clean()

	bstore, err := blocks.NewStore(os.Args[1], "blocks")
	if err != nil {
		logs.Fatalf("failed to create block store: %v", err)
	}

	defer bstore.Close()
	fs, err := ffs.NewFS(nodes.NewStore(db, dir), bstore)
	if err != nil {
		logs.Fatalf("failed to create filesystem: %v", err)
	}

	svr, err := fsrpc.NewServer(fs, os.Args[2])
	if err != nil {
		logs.Fatalf("failed to setup filesystem server")
	}

	defer fmt.Println("exited")
	go func() {
		fmt.Println(svr.ListenAndServe())
	}()
	<-c
}
