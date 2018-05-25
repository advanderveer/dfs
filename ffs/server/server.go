package server

import (
	"errors"

	"github.com/billziss-gh/cgofuse/fuse"
)

//go:generate go run gen/main.go
type FS interface {
	fuse.FileSystemInterface
	fuse.FileSystemChflags
	fuse.FileSystemSetcrtime
	fuse.FileSystemSetchgtime
}

//settling on an rpc setup is largely based on:
// - preformance benchmarks: https://github.com/cockroachdb/rpc-bench
// - discussions about the future of net/rpc: https://github.com/golang/go/issues/16844
// - Pros GRPC: cancelation, wide language support, type checked client-server contracts
// - Cons GRPC: slow (see benchmarks), hard dependencies, format learning curve
// - pro gob: handle native types for Stat and timespec more easilty

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}
