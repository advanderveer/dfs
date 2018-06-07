package ffs

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/billziss-gh/cgofuse/fuse"
)

func TestRW(t *testing.T) {
	fs, clean, err := NewTempFS("")
	assert(t, fs != nil, "expected fs not to be nil")
	ok(t, err)
	defer clean()

	errc := fs.Mknod("foo.txt", fuse.S_IFREG, 0)
	equals(t, 0, errc)

	errc, fh := fs.Open("foo.txt", fuse.O_CREAT|fuse.O_RDWR)
	equals(t, 0, errc)
	equals(t, uint64(3), fh)

	n := fs.Write("foo.txt", []byte{0x01, 0x02, 0x03}, 0, fh)
	equals(t, 3, n)

	n = fs.Write("foo.txt", []byte{0x03, 0x04, 0x05}, 2, fh)
	equals(t, 3, n)

	buf := make([]byte, 6)
	n = fs.Read("foo.txt", buf, 0, fh)
	equals(t, 5, n)
	equals(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x00}, buf)

	errc = fs.Truncate("foo.txt", 3, fh)
	equals(t, 0, errc)
	buf = make([]byte, 3)
	n = fs.Read("foo.txt", buf, 0, fh)
	equals(t, 3, n)
	equals(t, []byte{0x01, 0x02, 0x03}, buf)

}

func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
