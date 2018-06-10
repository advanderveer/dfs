package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/ffshttp"
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

	fs, clean, err := ffs.NewTempFS(os.Args[1])
	if err != nil {
		logs.Fatalf("failed to setup fs: %v", err)
	}

	defer clean()
	svr, err := ffshttp.NewServer(fsrpc.New(fs), os.Args[2])
	if err != nil {
		logs.Fatalf("failed to create server: %v", err)
	}

	defer fmt.Println("exited")
	go func() {
		logs.Printf("starting http on: %v", os.Args[2])
		logs.Println(svr.Serve())
	}()
	<-c
}
