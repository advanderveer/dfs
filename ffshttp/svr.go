package ffshttp

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"

	"github.com/advanderveer/dfs/ffs"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/gorilla/mux"
	"github.com/jcuga/golongpoll"
)

var (
	browsePrefix = "/browse"
)

type Server struct {
	b *ffs.Browser
	d fdb.Database
	m *golongpoll.LongpollManager
	l net.Listener
	r *mux.Router
	s *http.Server
}

func NewServer(fsrcp *rpc.Server, fsb *ffs.Browser, db fdb.Database, addr string) (s *Server, err error) {
	s = &Server{b: fsb}
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
	s.r.PathPrefix(browsePrefix).Handler(http.HandlerFunc(s.handleBrowse))
	s.r.HandleFunc("/runs", s.m.SubscriptionHandler)
	s.r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<p><a href="browse/">browse</a></p>`)
	})

	s.s = &http.Server{
		Handler: s.r,
		//@TODO add sensible timeouts for a webserver that handles a fs
	}

	return s, nil
}

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimLeft(r.URL.Path, browsePrefix)

	fmt.Fprintf(w, `<table>`)
	if err := s.b.Readdir(path, func(name string, fi os.FileInfo) {
		fmt.Fprintf(w, `<tr>`)
		fmt.Fprintf(w, `
			<td><a href="%s/">%s</p></td>
			<td>size: %db</td>
		`, name, name, fi.Size())

		fmt.Fprintf(w, `<tr>`)
	}); err != nil {
		fmt.Fprintf(w, "error: %v", err)
	}
	fmt.Fprintf(w, `</table>`)

	return
}

func (s *Server) Addr() net.Addr {
	return s.l.Addr()
}

func (s *Server) Serve() error {
	return s.s.Serve(s.l)
}
