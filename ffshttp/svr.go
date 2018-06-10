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
	"github.com/advanderveer/dfs/model"
	"github.com/gorilla/mux"
	"github.com/hashicorp/hcl"
	"github.com/jcuga/golongpoll"
)

var (
	browsePrefix = "/browse"
)

var (
	routeNameCreateRun = "create_run"
	routeNameListRuns  = "list_runs"
	routeNameViewRun   = "view_run"
	routeNameBrowse    = "browse"
)

type Server struct {
	m  *model.Model
	b  *ffs.Browser
	lp *golongpoll.LongpollManager
	l  net.Listener
	r  *mux.Router
	s  *http.Server
}

func NewServer(fsrcp *rpc.Server, fsb *ffs.Browser, m *model.Model, addr string) (s *Server, err error) {
	s = &Server{b: fsb, m: m}
	s.l, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.lp, err = golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		return nil, err
	}

	s.r = mux.NewRouter()
	s.r.Handle("/fs", fsrcp)
	s.r.PathPrefix(browsePrefix).Handler(http.HandlerFunc(s.handleBrowse)).Name(routeNameBrowse)
	s.r.HandleFunc("/viewRun/{id}", s.viewRun).Name(routeNameViewRun)
	s.r.HandleFunc("/createRun", s.createRun).Name(routeNameCreateRun)
	s.r.HandleFunc("/listRuns", s.listRuns).Name(routeNameListRuns)
	s.r.HandleFunc("/events", s.lp.SubscriptionHandler)
	s.r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		browse, _ := s.r.Get(routeNameBrowse).URL()
		listRuns, _ := s.r.Get(routeNameListRuns).URL()
		fmt.Fprintf(w, `<p><a href="%s/">browse</a></p>`, browse.String())
		fmt.Fprintf(w, `<p><a href="%s">list runs</a></p>`, listRuns.String())
	})

	s.s = &http.Server{
		Handler: s.r,
		//@TODO add sensible timeouts for a webserver that handles a fs
	}

	return s, nil
}

func (s *Server) viewRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := s.m.ViewRun(vars["id"])
	if err != nil {
		fmt.Fprintf(w, "failed to view run: %v", err)
		return
	}

	fmt.Fprintf(w, "run: %v", run)
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	if err := s.m.EachRun(func(r *model.Run) bool {
		viewurl, _ := s.r.Get(routeNameViewRun).URL("id", r.ID)
		fmt.Fprintf(w, `<p><a href="%s">run: %v</a></p>`, viewurl.String(), r.ID)
		return true
	}); err != nil {
		fmt.Fprintf(w, "failed to list runs: %v", err)
		return
	}
}

func (s *Server) createRun(w http.ResponseWriter, r *http.Request) {
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

	job := &model.Job{}
	err = hcl.Unmarshal(buf.Bytes(), &job)
	if err != nil {
		fmt.Fprintf(w, "failed to parse hcl: %v", err)
		return
	}

	job.Workspace = filepath.Dir(fname)
	run, err := s.m.CreateRun(job)
	if err != nil {
		fmt.Fprintf(w, "failed to create run: %v", err)
		return
	}

	fmt.Fprintf(w, "created run %#v", run)
	s.lp.Publish("runs", run)
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

		runurl, _ := s.r.Get(routeNameCreateRun).URL()
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
