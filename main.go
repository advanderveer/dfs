package main

import (
	"fmt"
	"log"
	"os"

	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/memfs"
	"github.com/billziss-gh/cgofuse/fuse"
)

// func db(ns string) (tr fdb.Transactor, ss directory.DirectorySubspace, f func()) {
// 	fdb.MustAPIVersion(510)
// 	db, err := fdb.OpenDefault()
// 	if err != nil {
// 		log.Fatal("failed to open database:", err)
// 	}
//
// 	dir, err := directory.CreateOrOpen(db, []string{"fdb-tests", ns}, nil)
// 	if err != nil {
// 		log.Fatal("failed to create or open app dir:", err)
// 	}
//
// 	return db, dir, func() {
// 		_, err := dir.Remove(db, nil)
// 		if err != nil {
// 			log.Fatal("failed to remove testing dir:", err)
// 		}
// 	}
// }

func main() {
	logs := log.New(os.Stderr, "ffs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("ffs [addr|'local'|'memfs'] [mountpoint]")
	}

	uid := os.Getuid()
	gid := os.Getgid()
	fmt.Println("ENODATA", fuse.ENODATA, "ENOSYS", fuse.ENOSYS, "ENOTSUP", fuse.ENOTSUP, "ENOATTR", fuse.ENOATTR)
	fmt.Println("fuse.S_IFDIR", fuse.S_IFDIR, "fuse.S_IFMT", fuse.S_IFMT)
	logs.Printf("mounting filesystem from '%s' at '%s' (uid: %d, gid: %d)", os.Args[1], os.Args[2], uid, gid)
	defer logs.Printf("unmounted, done!")

	// conn, err := net.DialTimeout("tcp", os.Args[1], time.Second*2)
	// if err != nil {
	// 	logs.Fatalf("failed to dial: %v", err)
	// }

	var (
		fs  fuse.FileSystemInterface
		err error
	)

	switch os.Args[1] {
	// case "local":
	// 	logs.Println("using a own-mounted fs")
	// 	tpdir, err := ioutil.TempDir("", "ffs_")
	// 	if err != nil {
	// 		logs.Fatalf("failed to creat temp dir: %v", err)
	// 	}
	//
	// 	db, dir, clean := db(tpdir)
	// 	defer clean()
	//
	// 	bstore, err := blocks.NewStore(tpdir, "blocks")
	// 	if err != nil {
	// 		logs.Fatalf("failed to create block store: %v", err)
	// 	}
	//
	// 	defer bstore.Close()
	// 	fs, err = ffs.NewFS(nodes.NewStore(db, dir), bstore)
	// 	if err != nil {
	// 		logs.Fatalf("failed to create filesystem: %v", err)
	// 	}
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
