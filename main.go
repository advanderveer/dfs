package main

import (
	"log"
	"os"

	"github.com/advanderveer/dfs/simple"
	"github.com/billziss-gh/cgofuse/fuse"
)

func main() {
	logs := log.New(os.Stderr, "dfs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("please provide the mount path and db dir")
	}

	logs.Printf("mounting filesystem from '%s'", os.Args[1])
	defer logs.Printf("unmounted, done!")
	// fs, err := dfs.NewFS(os.Args[1])
	// if err != nil {
	// 	logs.Fatalf("failed to create filesystem: %v", err)
	// }

	fs := &simple.Hellofs{}

	host := fuse.NewFileSystemHost(fs)
	if !host.Mount("", os.Args[2:]) {
		os.Exit(1) //mount failed
	}
}
