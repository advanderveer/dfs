package fsrpc

import (
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
)

//go:generate go run gen/main.go
type FS interface {
	fuse.FileSystemInterface
	fuse.FileSystemChflags
	fuse.FileSystemSetcrtime
	fuse.FileSystemSetchgtime
}

//Receiver responds to RPC requests
type Receiver struct {
	fs FS
}

func NewReceiver(fs FS) *Receiver {
	return &Receiver{fs: fs}
}

//Sender dispatches RPC requests
type Sender struct {
	uid uint32 //@TODO protect setting these with a lock
	gid uint32 //@TODO make in inpossible to make nodes when these are not set
	rpc interface {
		Call(serviceMethod string, args interface{}, reply interface{}) error
	}
	LastErr error
}

//Dial the filesystem at the provided address as the provided user and group
func Dial(addr string) (*Sender, error) {
	conn, err := net.DialTimeout("tcp", addr, time.Second*30)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	s := &Sender{rpc: rpc.NewClient(conn), LastErr: nil}
	return s, nil
}

//DialHTTP the filesystem at the provided address as the provided user and group
func DialHTTP(addr, path string) (*Sender, error) {
	//@TODO dial with exponential retry
	c, err := rpc.DialHTTPPath("tcp", addr, path) //@TODO dial with timeout
	if err != nil {
		return nil, fmt.Errorf("failed to dial HTTP: %v", err)
	}

	s := &Sender{rpc: c, LastErr: nil}
	return s, nil
}
