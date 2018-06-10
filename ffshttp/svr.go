package ffshttp

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/gorilla/mux"
	"github.com/jcuga/golongpoll"
)

type Server struct {
	d fdb.Database
	m *golongpoll.LongpollManager
	l net.Listener
	r *mux.Router
	s *http.Server
}

func NewServer(fsrcp *rpc.Server, d fdb.Database, addr string) (s *Server, err error) {
	s = &Server{}
	s.l, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.m, err = golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		return nil, err
	}

	s.r = mux.NewRouter()
	s.r.Handle("/fs", fsrcp)
	s.r.HandleFunc("/runs", s.m.SubscriptionHandler)

	s.s = &http.Server{
		Handler: s.r,
		//@TODO add sensible timeouts for a webserver that handles a fs
	}

	return s, nil
}

func (s *Server) Addr() net.Addr {
	return s.l.Addr()
}

func (s *Server) Serve() error {
	return s.s.Serve(s.l)
}
