package fsrpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
)

type Svr struct {
	addr string
	fs   FS
	l    net.Listener
	s    *rpc.Server
}

func NewServer(fs FS, addr string) (svr *Svr, err error) {
	svr = &Svr{fs: fs, addr: addr}
	svr.l, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	svr.s = rpc.NewServer()
	svr.s.RegisterName("FS", &Receiver{fs: svr.fs})

	return svr, nil
}

func (svr *Svr) Addr() net.Addr {
	return svr.l.Addr()
}

func (svr *Svr) ListenAndServeHTTP() (err error) {
	fmt.Println("Accepting HTTP on:", svr.l.Addr())
	// svr.s.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	httpsvr := http.Server{
		Handler: svr.s,
	}

	return httpsvr.Serve(svr.l)
}

func (svr *Svr) ListenAndServe() (err error) {
	fmt.Println("Accepting connections on:", svr.l.Addr())
	for {
		var conn net.Conn
		conn, err = svr.l.Accept()
		if err != nil {
			fmt.Println("Err accepting:", err)
			continue
		}

		fmt.Printf("Accepted conn from: %v\n", conn.RemoteAddr())
		go svr.s.ServeConn(conn)
	}
}
