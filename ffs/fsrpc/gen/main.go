package main

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"text/template"
)

type ParamDecl struct {
	IsPointer bool
	FieldName string
	Name      string
	Type      string
}

type ResultDecl struct {
	Type string
}

type ProcedureDecl struct {
	Name    string
	Params  []ParamDecl
	Results []ResultDecl
}

type ServerDecl struct {
	Package string
	Procs   []ProcedureDecl
}

var t = template.Must(template.New("svr").Parse(`// Code automtically generated. DO NOT EDIT.
package {{.Package}}

import(
	"fmt"
	"os"
	"github.com/billziss-gh/cgofuse/fuse"
)

type ReaddirCall struct {
	Name string
	Stat *fuse.Stat_t
	Ofst int64
}

type ListxattrCall struct {
	Name string
}

{{range $i, $proc := .Procs}}
type {{$proc.Name}}Args struct {
	{{range $j, $param := $proc.Params}}{{$param.FieldName}} {{$param.Type}}
	{{end}}
}

type {{$proc.Name}}Reply struct {
	Args *{{$proc.Name}}Args
	{{range $j, $res := $proc.Results}}R{{$j}} {{$res.Type}}
	{{end}}
	{{if eq $proc.Name "Readdir"}}Fills []ReaddirCall{{end}}
	{{if eq $proc.Name "Listxattr"}}Fills []ListxattrCall{{end}}
}
func (rcvr *Receiver) {{$proc.Name}}(a *{{$proc.Name}}Args, r *{{$proc.Name}}Reply) (err error) {
	{{if eq $proc.Name "Readdir"}}a.Fill = func(name string, stat *fuse.Stat_t, ofst int64) bool{
		r.Fills = append(r.Fills, ReaddirCall{Name: name, Stat: stat, Ofst: ofst})
		return true
	}{{else if eq $proc.Name "Listxattr"}}
	a.Fill = func(name string) bool{
		r.Fills = append(r.Fills, ListxattrCall{Name: name})
		return true
	}{{end}}
	{{if $proc.Results}}{{range $j, $res := $proc.Results}}{{if ne $j 0}},{{end}}r.R{{$j}} {{end}} = {{end}}rcvr.fs.{{$proc.Name}}({{range $j, $param := $proc.Params}}{{if ne $j 0}}, {{end}}a.{{$param.FieldName}} {{end}})
	r.Args = a
	return
}

func (sndr *Sender) {{$proc.Name}}({{range $j, $param := $proc.Params}}{{if ne $j 0}}, {{end}}{{$param.Name}}  {{$param.Type}}{{end}}) {{if $proc.Results}}({{range $j, $res := $proc.Results}}{{if ne $j 0}},{{end}}{{$res.Type}}{{end}}){{end}} {
	{{if $proc.Results}}r := &{{$proc.Name}}Reply{}{{else}}r := &struct{}{}{{end}}
	a := &{{$proc.Name}}Args{
		{{range $j, $param := $proc.Params}}{{$param.FieldName}}: {{$param.Name}},
		{{end}}
	}

	sndr.LastErr = sndr.rpc.Call("FS.{{$proc.Name}}", a, r)
	if sndr.LastErr != nil {
		fmt.Println("Transport Error:", sndr.LastErr.Error())
	}

	{{range $j, $param := $proc.Params}}{{if $param.IsPointer}}*{{$param.Name}} = *r.Args.{{$param.FieldName}}{{end}}
	{{end}}
	{{if eq $proc.Name "Read"}}copy(buff, r.Args.Buff){{end}}
	{{if eq $proc.Name "Readdir"}}
	for _, c := range r.Fills {
		if c.Stat != nil {
			c.Stat.Uid = uint32(os.Getuid())
			c.Stat.Gid = uint32(os.Getgid())
		}
		if !fill(c.Name, c.Stat, c.Ofst) {
			break
		}
	}
	{{else if eq $proc.Name "Listxattr"}}
	for _, c := range r.Fills {
		if !fill(c.Name) {
			break
		}
	}
	{{else if eq $proc.Name "Getattr"}}
	stat.Uid = uint32(os.Getuid())
	stat.Gid = uint32(os.Getgid())
	{{end}}

	return {{range $j, $res := $proc.Results}}{{if ne $j 0}},{{end}}{{if eq $res.Type "int"}}errc(r.R{{$j}}){{else}}r.R{{$j}}{{end}}{{end}}
}

{{end}}

`))

func write(logs *log.Logger, path string, svc ServerDecl) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open output file: %v", err)
	}

	defer f.Close()

	err = t.Option("missingkey=error").Execute(f, svc)
	if err != nil {
		return fmt.Errorf("failed to execute generator template: %v", err)
	}

	return nil
}

func parse(logs *log.Logger, fset *token.FileSet, name string) (err error) {
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %v", name, err)
	}

	defer f.Close()
	astf, err := parser.ParseFile(fset, "", f, 0)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	imp := importer.Default()
	conf := types.Config{Importer: imp}
	pkg, err := conf.Check("", fset, []*ast.File{astf}, nil)
	if err != nil {
		log.Fatal(err)
	}

	fsobj := pkg.Scope().Lookup("FS") //@TODO can be generator argument
	named, ok := fsobj.Type().(*types.Named)
	if !ok {
		return fmt.Errorf("object is not a named type")
	}

	iface, ok := named.Underlying().(*types.Interface)
	if !ok {
		return fmt.Errorf("object is not a interface type")
	}

	svrDecl := ServerDecl{
		Procs:   make([]ProcedureDecl, iface.NumMethods()),
		Package: pkg.Name(),
	}

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		sig, ok := m.Type().(*types.Signature)
		if !ok {
			return fmt.Errorf("iterface method type is not a signature")
		}

		procDecl := ProcedureDecl{
			Name:    m.Name(),
			Params:  make([]ParamDecl, sig.Params().Len()),
			Results: make([]ResultDecl, sig.Results().Len()),
		}

		for j := 0; j < sig.Params().Len(); j++ {
			p := sig.Params().At(j)
			procDecl.Params[j] = ParamDecl{
				FieldName: strings.Title(p.Name()),
				Name:      p.Name(),
				Type:      strings.Replace(p.Type().String(), "github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse", "fuse", 1),
			}

			_, procDecl.Params[j].IsPointer = p.Type().Underlying().(*types.Pointer)
		}

		for j := 0; j < sig.Results().Len(); j++ {
			r := sig.Results().At(j)
			procDecl.Results[j] = ResultDecl{
				Type: r.Type().String(),
			}
		}

		svrDecl.Procs[i] = procDecl
	}

	err = write(logs, strings.Replace(name, ".go", "_rpc.go", 1), svrDecl)
	if err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}

	cmd := exec.CommandContext(context.TODO(), "go", "fmt")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run go fmt")
	}

	return
}

func run(logs *log.Logger, args []string) (err error) {
	gofile := os.Getenv("GOFILE")
	if gofile == "" {
		return errors.New("GOFILE environment variable is empty, this program must be run as a generator")
	}

	wdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %v", err)
	}

	fset := token.NewFileSet()
	name := filepath.Join(wdir, gofile)
	err = parse(logs, fset, name)
	if err != nil {
		return fmt.Errorf("failed to parse file '%s': %v", name, err)
	}

	return
}

func main() {
	err := run(log.New(os.Stderr, "gwgen/", log.Lshortfile), os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
