package server

import (
	"fmt"
	"net"
	"net/rpc"
	"time"
)

func NewSimpleRPCFS(fs FS) (FS, error) {
	conn, err := SimpleRPC(fs)
	if err != nil {
		return nil, err
	}

	return &Sender{rpc.NewClient(conn), nil}, nil
}

func SimpleRPC(fs FS) (net.Conn, error) {
	svr := rpc.NewServer()

	rcvr := &Receiver{fs: fs}
	svr.RegisterName("FS", rcvr)
	l, err := net.Listen("tcp6", ":")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	go func() {
		fmt.Println("Accepting connections on:", l.Addr())
		for {
			var conn net.Conn
			conn, err = l.Accept()
			if err != nil {
				fmt.Println("Err accepting:", err)
				continue
			}

			svr.ServeConn(conn)
		}
	}()

	conn, err := net.DialTimeout("tcp6", l.Addr().String(), time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return conn, nil
}
