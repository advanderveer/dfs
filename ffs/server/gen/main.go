package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"text/template"
)

// func parseProcedure(logs *log.Logger, ftype *ast.FuncType) (p ProcedureDecl, err error) {
// 	for _, param := range ftype.Params.List {
// 		if len(param.Names) < 0 {
// 			continue
// 		}
//
// 		if p.InputDecl == nil {
// 			p.InputDecl = make(map[string]ValueDecl)
// 		}
//
// 		p.InputDecl[param.Names[0].Name] = ValueDecl{param.Type}
// 		fmt.Println("\t->", param.Type)
// 	}
//
// 	if ftype.Results != nil {
// 		for _, res := range ftype.Results.List {
// 			p.OutputDecl = append(p.OutputDecl, ValueDecl{res.Type})
// 			fmt.Println("\t<-", res.Type)
// 		}
// 	}
//
// 	// fmt.Println(ftype.Params.List)
// 	// if len(ftype.Params.List) != 2 {
// 	// 	return p, fmt.Errorf("service procedure should have two parameters, got: %v", ftype.Results)
// 	// }
// 	//
// 	// //first param is expected to be context, grab the second:
// 	// if inputPtr, ok := ftype.Params.List[1].Type.(*ast.StarExpr); ok {
// 	// 	if inputIdent, ok := inputPtr.X.(*ast.Ident); ok {
// 	// 		p.InputDecl = inputIdent
// 	// 	}
// 	// }
// 	//
// 	// if p.InputDecl == nil {
// 	// 	return p, fmt.Errorf("failed to parse procedure input decl from token '%v'", ftype)
// 	// }
// 	//
// 	// if len(ftype.Results.List) != 2 {
// 	// 	return p, fmt.Errorf("service procedure should have two return values, got: %d", len(ftype.Results.List))
// 	// }
// 	//
// 	// //second result is always error, grab the first
// 	// if outputPtr, ok := ftype.Results.List[0].Type.(*ast.StarExpr); ok {
// 	// 	if outputIdent, ok := outputPtr.X.(*ast.Ident); ok {
// 	// 		p.OutputDecl = outputIdent
// 	// 	}
// 	// }
// 	//
// 	// if p.OutputDecl == nil {
// 	// 	return p, fmt.Errorf("failed to parse procedure output decl from token '%v'", ftype)
// 	// }
//
// 	return
// }
//
// func parseInterface(logs *log.Logger, itype *ast.InterfaceType) (err error) {
// 	for _, m := range itype.Methods.List {
// 		if len(m.Names) < 1 {
// 			continue
// 		}
//
// 		if ftype, ok := m.Type.(*ast.FuncType); ok {
// 			fmt.Println(m.Names)
// 			decl, err := parseProcedure(logs, ftype)
// 			if err != nil {
// 				return fmt.Errorf("failed to parce procedure: %v", err)
// 			}
//
// 			_ = decl
// 		}
// 	}
//
// 	return nil
// }

type ParamDecl struct {
	Name string
	Type string
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
	"github.com/billziss-gh/cgofuse/fuse"
)

//Receiver receives RPC requests and returns results
type Receiver struct {
	fs FS
}
{{range $i, $proc := .Procs}}
type {{$proc.Name}}Args struct {
	{{range $j, $param := $proc.Params}}{{$param.Name}} {{$param.Type}}
	{{end}}
}

type {{$proc.Name}}Reply struct {
	Args *{{$proc.Name}}Args
	{{range $j, $res := $proc.Results}}R{{$j}} {{$res.Type}}
	{{end}}
}
func (rcvr *Receiver) {{$proc.Name}}(a *{{$proc.Name}}Args, r *{{$proc.Name}}Reply) (err error) {
	{{if $proc.Results}}{{range $j, $res := $proc.Results}}{{if ne $j 0}},{{end}}r.R{{$j}} {{end}} = {{end}}rcvr.fs.{{$proc.Name}}({{range $j, $param := $proc.Params}}{{if ne $j 0}}, {{end}}a.{{$param.Name}} {{end}})
	r.Args = a
	return
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

		fmt.Println(m.Name())
		procDecl := ProcedureDecl{
			Name:    m.Name(),
			Params:  make([]ParamDecl, sig.Params().Len()),
			Results: make([]ResultDecl, sig.Results().Len()),
		}

		for j := 0; j < sig.Params().Len(); j++ {
			p := sig.Params().At(j)
			procDecl.Params[j] = ParamDecl{
				Name: strings.Title(p.Name()),
				Type: strings.Replace(p.Type().String(), "github.com/advanderveer/dfs/vendor/github.com/billziss-gh/cgofuse/fuse", "fuse", 1),
			}

			fmt.Printf("\t->%s %#v\n", p.Name(), p.Type().String())
		}

		for j := 0; j < sig.Results().Len(); j++ {
			r := sig.Results().At(j)
			procDecl.Results[j] = ResultDecl{
				Type: r.Type().String(),
			}

			fmt.Println("\t<-", r)
		}

		svrDecl.Procs[i] = procDecl
	}

	err = write(logs, strings.Replace(name, ".go", "_svr.go", 1), svrDecl)
	if err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}

	// ast.Inspect(astf, func(node ast.Node) bool {
	// 	decl, ok := node.(*ast.GenDecl)
	// 	if !ok || decl.Tok != token.TYPE {
	// 		return true //we care only about types
	// 	}
	//
	// 	for _, spec := range decl.Specs {
	// 		if tspec, ok := spec.(*ast.TypeSpec); ok {
	// 			if itype, ok := tspec.Type.(*ast.InterfaceType); ok {
	//
	// 				err = parseInterface(logs, itype)
	// 				if err != nil {
	// 					err = fmt.Errorf("failed to parse service interface: %v", err)
	// 					return false
	// 				}
	//
	// 				_ = itype
	// 			}
	// 		}
	// 	}
	//
	// 	return true
	// })

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
