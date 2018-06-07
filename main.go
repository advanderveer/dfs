package main

import (
	"log"
	"os"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/memfs"
	"github.com/billziss-gh/cgofuse/fuse"
)

func main() {
	logs := log.New(os.Stderr, "ffs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("ffs [addr|'local'|'memfs'] [mountpoint]")
	}

	logs.Printf("mounting filesystem from '%s' at '%s'", os.Args[1], os.Args[2])
	defer logs.Printf("unmounted, done!")

	var (
		fs  fuse.FileSystemInterface
		err error
	)

	switch os.Args[1] {
	case "local":
		logs.Println("using a own-mounted fs")
		var clean func() error
		fs, clean, err = ffs.NewTempFS("")
		if err != nil {
			logs.Fatal(err)
		}

		defer clean()

	case "memfs":
		logs.Println("using a memory fs")
		fs = memfs.NewMemfs()
	default:
		logs.Println("using a remote fs")
		fs, err = fsrpc.Dial(os.Args[1])
		if err != nil {
			log.Fatalf("failed to dial: %v", err)
		}
	}

	host := fuse.NewFileSystemHost(fs)
	if !host.Mount("", os.Args[2:]) {
		os.Exit(1) //mount failed
	}
}
