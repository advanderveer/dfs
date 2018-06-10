package ffshttp

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"

	"github.com/advanderveer/dfs/ffs"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/gorilla/mux"
	"github.com/hashicorp/hcl"
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
	s.r.HandleFunc("/run", s.runHandle).Name("run")
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

func (s *Server) runHandle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "failed to parse form: %v", err)
		return
	}

	fname := r.FormValue("file")
	if fname == "" {
		fmt.Fprintf(w, "no file in form")
		return
	}

	buf := bytes.NewBuffer(nil)
	err = s.b.Readfile(fname, buf)
	if err != nil {
		fmt.Fprintf(w, "failed to parse form: %v", err)
		return
	}

	v := map[string]interface{}{}
	err = hcl.Unmarshal(buf.Bytes(), &v)
	if err != nil {
		fmt.Fprintf(w, "failed to parse hcl: %v", err)
		return
	}

	fmt.Fprintf(w, "run! %#v", v)
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

		runurl, _ := s.r.Get("run").URL()
		if strings.HasSuffix(name, ".hcl") {
			fmt.Fprintf(w,
				`<td><form action="%s" method="post">
					<input type="hidden" name="file" value="%s"/>
					<button type="submit">run!</button>
				</form></td>`,
				runurl.String(), filepath.Join(path, name),
			)
		} else {
			fmt.Fprintf(w, `<td></td>`)
		}

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
