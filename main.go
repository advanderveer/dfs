package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/advanderveer/dfs/dfs"
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/boltdb/bolt"
)

func main() {
	logs := log.New(os.Stderr, "dfs/", log.Lshortfile)
	if len(os.Args) < 2 {
		fmt.Println("please provide the mount path")
		os.Exit(1)
	}

	dir, err := ioutil.TempDir("", "dfs_")
	if err != nil {
		logs.Fatalf("failed to create file system store directory: %v", err)
	}

	db, err := bolt.Open(filepath.Join(dir, "buf.db"), 0600, nil)
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
