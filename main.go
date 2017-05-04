package main

import (
	"fmt"
	"log"
	"os"

	"github.com/advanderveer/dfs/dfs"
	"github.com/billziss-gh/cgofuse/fuse"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please provide the mount path")
		os.Exit(1)
	}

	logs := log.New(os.Stderr, "dfs/", log.Lshortfile)
	logs.Printf("mounting filesystem on '%s'", os.Args[1])
	defer logs.Printf("unmounted, done!")

	memfs := dfs.NewMemfs()
	host := fuse.NewFileSystemHost(memfs)
	if !host.Mount(os.Args[1], []string{}) {
		os.Exit(1) //mount failed
	}
}
