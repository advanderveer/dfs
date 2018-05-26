package fsrpc

import (
	"fmt"
	"net"
	"net/rpc"
)

type Svr struct {
	addr string
	fs   FS
	l    net.Listener
}

func NewServer(fs FS, addr string) (svr *Svr, err error) {
	svr = &Svr{fs: fs, addr: addr}
	svr.l, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	return svr, nil
}

func (svr *Svr) Addr() net.Addr {
	return svr.l.Addr()
}

func (svr *Svr) ListenAndServe() (err error) {
	s := rpc.NewServer()

	rcvr := &Receiver{fs: svr.fs}
	s.RegisterName("FS", rcvr)

	fmt.Println("Accepting connections on:", svr.l.Addr())
	for {
		var conn net.Conn
		conn, err = svr.l.Accept()
		if err != nil {
			fmt.Println("Err accepting:", err)
			continue
		}

		fmt.Printf("Accepted conn from: %v\n", conn.RemoteAddr())
		go s.ServeConn(conn)
	}
}
