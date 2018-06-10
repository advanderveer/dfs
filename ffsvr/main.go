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
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	logs := log.New(os.Stderr, "ffs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("ffsvr [dbdir] [addr]")
	}

	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		logs.Fatal(err)
	}

	fs, clean, err := ffs.NewTempFS(os.Args[1], db)
	if err != nil {
		logs.Fatalf("failed to setup fs: %v", err)
	}

	_ = clean
	// defer clean()
	svr, err := ffshttp.NewServer(fsrpc.New(fs), ffs.NewBrowser(fs), db, os.Args[2])
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
