package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/advanderveer/dfs/dfs"
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

func main() {
	logs := log.New(os.Stderr, "dfs/", log.Lshortfile)
	if len(os.Args) < 3 {
		fmt.Println("please provide the mount path and db dir")
		os.Exit(1)
	}

	db, err := bolt.Open(filepath.Join(os.Args[2]), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	logs.Printf("mounting filesystem on '%s'", os.Args[1])
	defer logs.Printf("unmounted, done!")
	dfs := dfs.NewFS(db)
	host := fuse.NewFileSystemHost(dfs)
	if !host.Mount(os.Args[1], []string{}) {
		os.Exit(1) //mount failed
	}
}
