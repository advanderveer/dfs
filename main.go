package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/memfs"
	"github.com/advanderveer/dfs/msg"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/cenkalti/backoff"
)

func main() {
	logs := log.New(os.Stderr, "ffs/", log.Lshortfile)
	if len(os.Args) < 3 {
		logs.Fatalf("ffs [addr|'local'|'memfs'] [mountpoint]")
	}

	logs.Printf("mounting filesystem from '%s' at '%s'", os.Args[1], os.Args[2])
	defer logs.Printf("unmounted, done!")

	var (
		fs  fuse.FileSystemInterface
		err error
	)

	switch os.Args[1] {
	case "local":
		logs.Println("using a own-mounted fs")
		var err error

		fdb.MustAPIVersion(510)
		db, err := fdb.OpenDefault()
		if err != nil {
			logs.Fatal(err)
		}

		var clean func() error
		fs, clean, err = ffs.NewTempFS("", db)
		if err != nil {
			logs.Fatal(err)
		}

		defer clean()

	case "memfs":
		logs.Println("using a memory fs")
		fs = memfs.NewMemfs()
	default:
		logs.Println("using a remote fs")
		fs, err = fsrpc.DialHTTP(os.Args[1], "/fs")
		if err != nil {
			log.Fatalf("failed to dial: %v", err)
		}

		//exploring the ability to run docker on top of the fs
		//@TODO move this to a package
		go func() {
			dexe, err := exec.LookPath("docker")
			if err != nil {
				logs.Printf("failed to find Docker executable in PATH: %v, do not register as worker", err)
				return
			}

			logs.Printf("found docker executable '%s', register this PC as worker", dexe)
			if err = backoff.Retry(func() (err error) {
				for {
					resp, err := http.Get("http://" + os.Args[1] + "/events?timeout=10&category=runs")
					if err != nil {
						logs.Printf("failed get runs: %v", err)
						return err
					}

					if resp.StatusCode != 200 {
						logs.Printf("unexpected status code: %v", resp.StatusCode)
						return errors.New("unexpected status code")
					}

					buf := bytes.NewBuffer(nil)

					v := msg.EventReponse{}
					dec := json.NewDecoder(io.TeeReader(resp.Body, buf))
					err = dec.Decode(&v)
					if err != nil {
						logs.Printf("failed to decode: %v", err)
						return err
					}

					for _, ev := range v.Events {
						run := ev.Data
						if run == nil {
							continue
						}

						job := run.Job
						log.Printf("received job workspace: %s, tasks: %d, job: %#v", job.Workspace, len(job.Tasks), job)
						for name, t := range job.Tasks {
							args := []string{"run",
								"--name", fmt.Sprintf("%s_%s", run.ID, name),
								"--log-driver", "json-file",
							}

							//@TODO run as detached container and ship logs async
							log.Printf("task %s, data: %d", name, len(t.Data))
							for src, data := range t.Data {
								log.Printf("adding mount for data %s", src)
								args = append(args, fmt.Sprintf(`--mount=type=bind,src=%s,dst=%s`,
									filepath.Join(os.Args[2], job.Workspace, src),
									data.Dest,
								))
							}

							args = append(args, t.Image)
							args = append(args, t.Command...)

							stdout, err := os.Create(filepath.Join(os.Args[2], job.Workspace, fmt.Sprintf("%s_%s.stdout.log", name, run.ID)))
							if err != nil {
								logs.Printf("failed to create stdout file, discarding")
								stdout = os.Stdout
							}

							stderr, err := os.Create(filepath.Join(os.Args[2], job.Workspace, fmt.Sprintf("%s_%s.stderr.log", name, run.ID)))
							if err != nil {
								logs.Printf("failed to create stdout file, discarding")
								stderr = os.Stderr
							}

							cmd := exec.Command("docker", args...)
							cmd.Stdout = stdout
							cmd.Stderr = stderr

							//@TODO run as a logshipping package that uses Docker API
							go func() {
								log.Printf("running: docker %v", cmd.Args)
								err := cmd.Run()
								if err != nil {
									log.Printf("failed to run task container: %v", err)
									return
								}

								log.Printf("ran: docker %v", cmd.Args)
							}()
						}
					}

					//@TODO sleep for a minimum amount (increased exponentially?)
				}
			}, backoff.NewExponentialBackOff()); err != nil {
				logs.Printf("polling failed: %v", err)
				return
			}
		}()

	}

	host := fuse.NewFileSystemHost(fs)
	if !host.Mount("", os.Args[2:]) {
		os.Exit(1) //mount failed
	}
}
