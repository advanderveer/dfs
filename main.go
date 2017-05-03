package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/advanderveer/dfs/dfs"
	"github.com/billziss-gh/cgofuse/fuse"
)

func main() {
	memfs := dfs.NewMemfs()
	host := fuse.NewFileSystemHost(memfs)
	if runtime.GOOS != "windows" {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			if !host.Unmount() {
				os.Exit(2) //unmount failed
			}
		}()
	}

	if !host.Mount(os.Args) {
		os.Exit(1) //mount failed
	}
}
