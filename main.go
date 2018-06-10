package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/advanderveer/dfs/ffs"
	"github.com/advanderveer/dfs/ffs/fsrpc"
	"github.com/advanderveer/dfs/memfs"
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
			type lpEvent struct {
				// Timestamp is milliseconds since epoch to match javascrits Date.getTime()
				Timestamp int64  `json:"timestamp"`
				Category  string `json:"category"`
				// NOTE: Data can be anything that is able to passed to json.Marshal()
				Data interface{} `json:"data"`
			}

			type eventResponse struct {
				Events []lpEvent `json:"events"`
			}

			dexe, err := exec.LookPath("docker")
			if err != nil {
				logs.Printf("failed to find Docker executable in PATH: %v, do not register as worker", err)
				return
			}

			logs.Printf("found docker executable '%s', register this PC as worker", dexe)
			if err = backoff.Retry(func() (err error) {
				for {
					resp, err := http.Get("http://" + os.Args[1] + "/runs?timeout=10&category=runs")
					if err != nil {
						logs.Printf("failed get runs: %v", err)
						return err
					}

					if resp.StatusCode != 200 {
						logs.Printf("unexpected status code: %v", resp.StatusCode)
						return errors.New("unexpected status code")
					}

					v := eventResponse{}
					dec := json.NewDecoder(resp.Body)
					err = dec.Decode(&v)
					if err != nil {
						logs.Printf("failed to decode: %v", err)
						return err
					}

					for _, ev := range v.Events {
						logs.Println("event", ev.Data)
					}

					//@TODO sleep for a mimum amount (increased exponentially?)
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
