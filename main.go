package main

import (
	"log"
	"os"

	"github.com/advanderveer/dfs/ffs"
	"github.com/billziss-gh/cgofuse/fuse"
)

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

	fs, err := ffs.NewFS(nil)
	if err != nil {
		logs.Fatalf("failed to create filesystem: %v", err)
	}

	host := fuse.NewFileSystemHost(fs)
	if !host.Mount("", os.Args[2:]) {
		os.Exit(1) //mount failed
	}
}
