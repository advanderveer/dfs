package webdav

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"golang.org/x/net/context"
)

func testCfg() *FDBFSConf {
	def := DefaultFBDFSConf()
	def.MaxChunkSize = 2
	return def
}

func open(tb testing.TB) (fdb.Database, directory.Directory, func()) {
	fdb.MustAPIVersion(510)
	db, err := fdb.OpenDefault()
	if err != nil {
		tb.Fatal("failed to open database:", err)
	}

	//app dir
	dir, err := directory.CreateOrOpen(db, []string{"fdb-tests", tb.Name()}, nil)
	if err != nil {
		tb.Fatal("failed to create or open app dir:", err)
	}

	return db, dir, func() {
		_, err := dir.Remove(db, nil)
		if err != nil {
			tb.Fatal("failed to remove testing dir:", err)
		}
	}
}

func TestFDBWalk(t *testing.T) {
	type walkStep struct {
		name, frag string
		final      bool
	}

	testCases := []struct {
		dir  string
		want []walkStep
	}{
		{"", []walkStep{
			{"", "", true},
		}},
		{"/", []walkStep{
			{"", "", true},
		}},
		{"/a", []walkStep{
			{"", "a", true},
		}},
		{"/a/", []walkStep{
			{"", "a", true},
		}},
		{"/a/b", []walkStep{
			{"", "a", false},
			{"a", "b", true},
		}},
		{"/a/b/", []walkStep{
			{"", "a", false},
			{"a", "b", true},
		}},
		{"/a/b/c", []walkStep{
			{"", "a", false},
			{"a", "b", false},
			{"b", "c", true},
		}},
		// The following test case is the one mentioned explicitly
		// in the method description.
		{"/foo/bar/x", []walkStep{
			{"", "foo", false},
			{"foo", "bar", false},
			{"bar", "x", true},
		}},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.dir, func(t *testing.T) {
			db, ns, clean := open(t)
			defer clean()
			fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)

			parts := strings.Split(tc.dir, "/")
			for p := 2; p < len(parts); p++ {
				d := strings.Join(parts[:p], "/")
				if err := fs.Mkdir(ctx, d, 0666); err != nil {
					t.Errorf("tc.dir=%q: mkdir: %q: %v", tc.dir, d, err)
				}
			}

			i, prevFrag := 0, ""
			err := fs.walk(fs.tr, "test", tc.dir, func(dir *Node, frag string, final bool) error {
				got := walkStep{
					name:  prevFrag,
					frag:  frag,
					final: final,
				}
				want := tc.want[i]

				if got != want {
					return fmt.Errorf("got %+v, want %+v", got, want)
				}
				i, prevFrag = i+1, frag
				return nil
			})
			if err != nil {
				t.Errorf("tc.dir=%q: %v", tc.dir, err)
			}
		})
	}
}

func TestFDBFS(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()

	testFS(t, NewFDBFS(db, ns, testCfg()))
}

func TestFDBFSRoot(t *testing.T) {
	ctx := context.Background()
	db, ns, clean := open(t)
	defer clean()

	fs := NewFDBFS(db, ns, testCfg())
	for i := 0; i < 5; i++ {
		stat, err := fs.Stat(ctx, "/")
		if err != nil {
			t.Fatalf("i=%d: Stat: %v", i, err)
		}
		if !stat.IsDir() {
			t.Fatalf("i=%d: Stat.IsDir is false, want true", i)
		}

		f, err := fs.OpenFile(ctx, "/", os.O_RDONLY, 0)
		if err != nil {
			t.Fatalf("i=%d: OpenFile: %v", i, err)
		}
		defer f.Close()
		children, err := f.Readdir(-1)
		if err != nil {
			t.Fatalf("i=%d: Readdir: %v", i, err)
		}
		if len(children) != i {
			t.Fatalf("i=%d: got %d children, want %d", i, len(children), i)
		}

		if _, err := f.Write(make([]byte, 1)); err == nil {
			t.Fatalf("i=%d: Write: got nil error, want non-nil", i)
		}

		if err := fs.Mkdir(ctx, fmt.Sprintf("/dir%d", i), 0777); err != nil {
			t.Fatalf("i=%d: Mkdir: %v", i, err)
		}
	}
}

func TestFDBFileReaddir(t *testing.T) {
	ctx := context.Background()
	db, ns, clean := open(t)
	defer clean()

	fs := NewFDBFS(db, ns, testCfg())
	if err := fs.Mkdir(ctx, "/foo", 0777); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	readdir := func(count int) ([]os.FileInfo, error) {
		f, err := fs.OpenFile(ctx, "/foo", os.O_RDONLY, 0)
		if err != nil {
			t.Fatalf("OpenFile: %v", err)
		}
		defer f.Close()
		return f.Readdir(count)
	}
	if got, err := readdir(-1); len(got) != 0 || err != nil {
		t.Fatalf("readdir(-1): got %d fileInfos with err=%v, want 0, <nil>", len(got), err)
	}
	if got, err := readdir(+1); len(got) != 0 || err != io.EOF {
		t.Fatalf("readdir(+1): got %d fileInfos with err=%v, want 0, EOF", len(got), err)
	}
}

func TestFDBFile(t *testing.T) {
	testCases := []string{
		"wantData ",
		"wantSize 0",
		"write abc",
		"wantData abc",
		"write de",
		"wantData abcde",
		"wantSize 5",
		"write 5*x",
		"write 4*y+2*z",
		"write 3*st",
		"wantData abcdexxxxxyyyyzzststst",
		"wantSize 22",
		"seek set 4 want 4",
		"write EFG",
		"wantData abcdEFGxxxyyyyzzststst",
		"wantSize 22",
		"seek set 2 want 2",
		"read cdEF",
		"read Gx",
		"seek cur 0 want 8",
		"seek cur 2 want 10",
		"seek cur -1 want 9",
		"write J",
		"wantData abcdEFGxxJyyyyzzststst",
		"wantSize 22",
		"seek cur -4 want 6",
		"write ghijk",
		"wantData abcdEFghijkyyyzzststst",
		"wantSize 22",
		"read yyyz",
		"seek cur 0 want 15",
		"write ",
		"seek cur 0 want 15",
		"read ",
		"seek cur 0 want 15",
		"seek end -3 want 19",
		"write ZZ",
		"wantData abcdEFghijkyyyzzstsZZt",
		"wantSize 22",
		"write 4*A",
		"wantData abcdEFghijkyyyzzstsZZAAAA",
		"wantSize 25",
		"seek end 0 want 25",
		"seek end -5 want 20",
		"read Z+4*A",
		"write 5*B",
		"wantData abcdEFghijkyyyzzstsZZAAAABBBBB",
		"wantSize 30",
		"seek end 10 want 40",
		"write C",
		"wantData abcdEFghijkyyyzzstsZZAAAABBBBB..........C",
		"wantSize 41",
		"write D",
		"wantData abcdEFghijkyyyzzstsZZAAAABBBBB..........CD",
		"wantSize 42",
		"seek set 43 want 43",
		"write E",
		"wantData abcdEFghijkyyyzzstsZZAAAABBBBB..........CD.E",
		"wantSize 44",
		"seek set 0 want 0",
		"write 5*123456789_",
		"wantData 123456789_123456789_123456789_123456789_123456789_",
		"wantSize 50",
		"seek cur 0 want 50",
		"seek cur -99 want err",
	}

	ctx := context.Background()

	const filename = "/foo"

	db, ns, clean := open(t)
	defer clean()

	fs := NewFDBFS(db, ns, testCfg())
	f, err := fs.OpenFile(ctx, filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	for i, tc := range testCases {
		j := strings.IndexByte(tc, ' ')
		if j < 0 {
			t.Fatalf("test case #%d %q: invalid command", i, tc)
		}
		op, arg := tc[:j], tc[j+1:]

		// Expand an arg like "3*a+2*b" to "aaabb".
		parts := strings.Split(arg, "+")
		for j, part := range parts {
			if k := strings.IndexByte(part, '*'); k >= 0 {
				repeatCount, repeatStr := part[:k], part[k+1:]
				n, err := strconv.Atoi(repeatCount)
				if err != nil {
					t.Fatalf("test case #%d %q: invalid repeat count %q", i, tc, repeatCount)
				}
				parts[j] = strings.Repeat(repeatStr, n)
			}
		}
		arg = strings.Join(parts, "")

		switch op {
		default:
			t.Fatalf("test case #%d %q: invalid operation %q", i, tc, op)

		case "read":
			buf := make([]byte, len(arg))
			if _, err := io.ReadFull(f, buf); err != nil {
				t.Fatalf("test case #%d %q: ReadFull: %v", i, tc, err)
			}
			if got := string(buf); got != arg {
				t.Fatalf("test case #%d %q:\ngot  %q\nwant %q", i, tc, got, arg)
			}

		case "seek":
			parts := strings.Split(arg, " ")
			if len(parts) != 4 {
				t.Fatalf("test case #%d %q: invalid seek", i, tc)
			}

			whence := 0
			switch parts[0] {
			default:
				t.Fatalf("test case #%d %q: invalid seek whence", i, tc)
			case "set":
				whence = os.SEEK_SET
			case "cur":
				whence = os.SEEK_CUR
			case "end":
				whence = os.SEEK_END
			}
			offset, err := strconv.Atoi(parts[1])
			if err != nil {
				t.Fatalf("test case #%d %q: invalid offset %q", i, tc, parts[1])
			}

			if parts[2] != "want" {
				t.Fatalf("test case #%d %q: invalid seek", i, tc)
			}
			if parts[3] == "err" {
				_, err := f.Seek(int64(offset), whence)
				if err == nil {
					t.Fatalf("test case #%d %q: Seek returned nil error, want non-nil", i, tc)
				}
			} else {
				got, err := f.Seek(int64(offset), whence)
				if err != nil {
					t.Fatalf("test case #%d %q: Seek: %v", i, tc, err)
				}
				want, err := strconv.Atoi(parts[3])
				if err != nil {
					t.Fatalf("test case #%d %q: invalid want %q", i, tc, parts[3])
				}
				if got != int64(want) {
					t.Fatalf("test case #%d %q: got %d, want %d", i, tc, got, want)
				}
			}

		case "write":
			n, err := f.Write([]byte(arg))
			if err != nil {
				t.Fatalf("test case #%d %q: write: %v", i, tc, err)
			}
			if n != len(arg) {
				t.Fatalf("test case #%d %q: write returned %d bytes, want %d", i, tc, n, len(arg))
			}

		case "wantData":
			g, err := fs.OpenFile(ctx, filename, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatalf("test case #%d %q: OpenFile: %v", i, tc, err)
			}
			gotBytes, err := ioutil.ReadAll(g)
			if err != nil {
				t.Fatalf("test case #%d %q: ReadAll: %v", i, tc, err)
			}
			for i, c := range gotBytes {
				if c == '\x00' {
					gotBytes[i] = '.'
				}
			}
			got := string(gotBytes)
			if got != arg {
				t.Fatalf("test case #%d %q:\ngot  %q\nwant %q", i, tc, got, arg)
			}
			if err := g.Close(); err != nil {
				t.Fatalf("test case #%d %q: Close: %v", i, tc, err)
			}

		case "wantSize":
			n, err := strconv.Atoi(arg)
			if err != nil {
				t.Fatalf("test case #%d %q: invalid size %q", i, tc, arg)
			}
			fi, err := fs.Stat(ctx, filename)
			if err != nil {
				t.Fatalf("test case #%d %q: Stat: %v", i, tc, err)
			}
			if got, want := fi.Size(), int64(n); got != want {
				t.Fatalf("test case #%d %q: got %d, want %d", i, tc, got, want)
			}
		}
	}
}

// TestFDBFileWriteAllocs tests that writing N consecutive 1KiB chunks to a
// memFile doesn't allocate a new buffer for each of those N times. Otherwise,
// calling io.Copy(aFDBFile, src) is likely to have quadratic complexity.
func TestFDBFileWriteAllocs(t *testing.T) {
	if runtime.Compiler == "gccgo" {
		t.Skip("gccgo allocates here")
	}
	ctx := context.Background()
	db, ns, clean := open(t)
	defer clean()

	fs := NewFDBFS(db, ns, testCfg())
	f, err := fs.OpenFile(ctx, "/xxx", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	xxx := make([]byte, 1024)
	for i := range xxx {
		xxx[i] = 'x'
	}

	a := testing.AllocsPerRun(100, func() {
		f.Write(xxx)
	})
	// AllocsPerRun returns an integral value, so we compare the rounded-down
	// number to zero.
	if a > 0 {
		//@TODO we might want to fatal here, due to potential complexity
		t.Logf("%v allocs per run, want 0", a)
	}
}

func BenchmarkFDBFileWrite(b *testing.B) {
	ctx := context.Background()
	db, ns, clean := open(b)
	defer clean()
	fs := NewFDBFS(db, ns, nil)
	xxx := make([]byte, 1024)
	for i := range xxx {
		xxx[i] = 'x'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, err := fs.OpenFile(ctx, "/xxx", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			b.Fatalf("OpenFile: %v", err)
		}
		for j := 0; j < 100; j++ {
			f.Write(xxx)
		}
		if err := f.Close(); err != nil {
			b.Fatalf("Close: %v", err)
		}
		if err := fs.RemoveAll(ctx, "/xxx"); err != nil {
			b.Fatalf("RemoveAll: %v", err)
		}
	}
}

func TestCopyMoveFDBProps(t *testing.T) {
	ctx := context.Background()
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg())
	create := func(name string) error {
		f, err := fs.OpenFile(ctx, name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		_, wErr := f.Write([]byte("contents"))
		cErr := f.Close()
		if wErr != nil {
			return wErr
		}
		return cErr
	}
	patch := func(name string, patches ...Proppatch) error {
		f, err := fs.OpenFile(ctx, name, os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		_, pErr := f.(DeadPropsHolder).Patch(patches)
		cErr := f.Close()
		if pErr != nil {
			return pErr
		}
		return cErr
	}
	props := func(name string) (map[xml.Name]Property, error) {
		f, err := fs.OpenFile(ctx, name, os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		m, pErr := f.(DeadPropsHolder).DeadProps()
		cErr := f.Close()
		if pErr != nil {
			return nil, pErr
		}
		if cErr != nil {
			return nil, cErr
		}
		return m, nil
	}

	p0 := Property{
		XMLName:  xml.Name{Space: "x:", Local: "boat"},
		InnerXML: []byte("pea-green"),
	}
	p1 := Property{
		XMLName:  xml.Name{Space: "x:", Local: "ring"},
		InnerXML: []byte("1 shilling"),
	}
	p2 := Property{
		XMLName:  xml.Name{Space: "x:", Local: "spoon"},
		InnerXML: []byte("runcible"),
	}
	p3 := Property{
		XMLName:  xml.Name{Space: "x:", Local: "moon"},
		InnerXML: []byte("light"),
	}

	if err := create("/src"); err != nil {
		t.Fatalf("create /src: %v", err)
	}
	if err := patch("/src", Proppatch{Props: []Property{p0, p1}}); err != nil {
		t.Fatalf("patch /src +p0 +p1: %v", err)
	}
	if _, err := copyFiles(ctx, fs, "/src", "/tmp", true, infiniteDepth, 0); err != nil {
		t.Fatalf("copyFiles /src /tmp: %v", err)
	}
	if _, err := moveFiles(ctx, fs, "/tmp", "/dst", true); err != nil {
		t.Fatalf("moveFiles /tmp /dst: %v", err)
	}
	if err := patch("/src", Proppatch{Props: []Property{p0}, Remove: true}); err != nil {
		t.Fatalf("patch /src -p0: %v", err)
	}
	if err := patch("/src", Proppatch{Props: []Property{p2}}); err != nil {
		t.Fatalf("patch /src +p2: %v", err)
	}
	if err := patch("/dst", Proppatch{Props: []Property{p1}, Remove: true}); err != nil {
		t.Fatalf("patch /dst -p1: %v", err)
	}
	if err := patch("/dst", Proppatch{Props: []Property{p3}}); err != nil {
		t.Fatalf("patch /dst +p3: %v", err)
	}

	gotSrc, err := props("/src")
	if err != nil {
		t.Fatalf("props /src: %v", err)
	}

	wantSrc := map[xml.Name]Property{
		p1.XMLName: p1,
		p2.XMLName: p2,
	}
	if !reflect.DeepEqual(gotSrc, wantSrc) {
		t.Fatalf("props /src:\ngot  %v\nwant %v", gotSrc, wantSrc)
	}

	gotDst, err := props("/dst")
	if err != nil {
		t.Fatalf("props /dst: %v", err)
	}
	wantDst := map[xml.Name]Property{
		p0.XMLName: p0,
		p3.XMLName: p3,
	}
	if !reflect.DeepEqual(gotDst, wantDst) {
		t.Fatalf("props /dst:\ngot  %v\nwant %v", gotDst, wantDst)
	}
}

func TestFDBWalkFS(t *testing.T) {
	testCases := []struct {
		desc    string
		buildfs []string
		startAt string
		depth   int
		walkFn  filepath.WalkFunc
		want    []string
	}{{
		"just root",
		[]string{},
		"/",
		infiniteDepth,
		nil,
		[]string{
			"/",
		},
	}, {
		"infinite walk from root",
		[]string{
			"mkdir /a",
			"mkdir /a/b",
			"touch /a/b/c",
			"mkdir /a/d",
			"mkdir /e",
			"touch /f",
		},
		"/",
		infiniteDepth,
		nil,
		[]string{
			"/",
			"/a",
			"/a/b",
			"/a/b/c",
			"/a/d",
			"/e",
			"/f",
		},
	}, {
		"infinite walk from subdir",
		[]string{
			"mkdir /a",
			"mkdir /a/b",
			"touch /a/b/c",
			"mkdir /a/d",
			"mkdir /e",
			"touch /f",
		},
		"/a",
		infiniteDepth,
		nil,
		[]string{
			"/a",
			"/a/b",
			"/a/b/c",
			"/a/d",
		},
	}, {
		"depth 1 walk from root",
		[]string{
			"mkdir /a",
			"mkdir /a/b",
			"touch /a/b/c",
			"mkdir /a/d",
			"mkdir /e",
			"touch /f",
		},
		"/",
		1,
		nil,
		[]string{
			"/",
			"/a",
			"/e",
			"/f",
		},
	}, {
		"depth 1 walk from subdir",
		[]string{
			"mkdir /a",
			"mkdir /a/b",
			"touch /a/b/c",
			"mkdir /a/b/g",
			"mkdir /a/b/g/h",
			"touch /a/b/g/i",
			"touch /a/b/g/h/j",
		},
		"/a/b",
		1,
		nil,
		[]string{
			"/a/b",
			"/a/b/c",
			"/a/b/g",
		},
	}, {
		"depth 0 walk from subdir",
		[]string{
			"mkdir /a",
			"mkdir /a/b",
			"touch /a/b/c",
			"mkdir /a/b/g",
			"mkdir /a/b/g/h",
			"touch /a/b/g/i",
			"touch /a/b/g/h/j",
		},
		"/a/b",
		0,
		nil,
		[]string{
			"/a/b",
		},
	}, {
		"infinite walk from file",
		[]string{
			"mkdir /a",
			"touch /a/b",
			"touch /a/c",
		},
		"/a/b",
		0,
		nil,
		[]string{
			"/a/b",
		},
	}, {
		"infinite walk with skipped subdir",
		[]string{
			"mkdir /a",
			"mkdir /a/b",
			"touch /a/b/c",
			"mkdir /a/b/g",
			"mkdir /a/b/g/h",
			"touch /a/b/g/i",
			"touch /a/b/g/h/j",
			"touch /a/b/z",
		},
		"/",
		infiniteDepth,
		func(path string, info os.FileInfo, err error) error {
			if path == "/a/b/g" {
				return filepath.SkipDir
			}
			return nil
		},
		[]string{
			"/",
			"/a",
			"/a/b",
			"/a/b/c",
			"/a/b/z",
		},
	}}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fs, clean, err := buildFDBTestFS(t, tc.buildfs)
			if err != nil {
				t.Fatalf("%s: cannot create test filesystem: %v", tc.desc, err)
			}

			defer clean()

			var got []string
			traceFn := func(path string, info os.FileInfo, err error) error {
				if tc.walkFn != nil {
					err = tc.walkFn(path, info, err)
					if err != nil {
						return err
					}
				}
				got = append(got, path)
				return nil
			}
			fi, err := fs.Stat(ctx, tc.startAt)
			if err != nil {
				t.Fatalf("%s: cannot stat: %v", tc.desc, err)
			}
			err = walkFS(ctx, fs, tc.depth, tc.startAt, fi, traceFn)
			if err != nil {
				t.Errorf("%s:\ngot error %v, want nil", tc.desc, err)
				return
			}
			sort.Strings(got)
			sort.Strings(tc.want)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%s:\ngot  %q\nwant %q", tc.desc, got, tc.want)
				return
			}
		})

	}
}

func buildFDBTestFS(tb testing.TB, buildfs []string) (FileSystem, func(), error) {
	// TODO: Could this be merged with the build logic in TestFS?

	ctx := context.Background()
	db, ns, clean := open(tb)

	fs := NewFDBFS(db, ns, testCfg())
	for _, b := range buildfs {
		op := strings.Split(b, " ")
		switch op[0] {
		case "mkdir":
			err := fs.Mkdir(ctx, op[1], os.ModeDir|0777)
			if err != nil {
				return nil, clean, err
			}
		case "touch":
			f, err := fs.OpenFile(ctx, op[1], os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return nil, clean, err
			}
			f.Close()
		case "write":
			f, err := fs.OpenFile(ctx, op[1], os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return nil, clean, err
			}
			_, err = f.Write([]byte(op[2]))
			f.Close()
			if err != nil {
				return nil, clean, err
			}
		default:
			return nil, clean, fmt.Errorf("unknown file operation %q", op[0])
		}
	}
	return fs, clean, nil
}

//these test fail with the -race flag
func TestFBDRaceTests(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg())
	ctx := context.Background()
	t.Run("FileSystem race conditions", func(t *testing.T) {
		t.Run("FS.Stat after FS.Mkdir", func(t *testing.T) {
			go fs.Mkdir(ctx, "/hello", 0777)
			go fs.Stat(ctx, "/hello")
		})

		t.Run("FS.Stat after FS.OpenFile", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				go fs.OpenFile(ctx, fmt.Sprintf("%d.txt", i), os.O_CREATE, 0777)
				go fs.Stat(ctx, fmt.Sprintf("%d.txt", i))
			}
		})

		t.Run("FS.Stat after FS.RemoveAll", func(t *testing.T) {
			fs.Mkdir(ctx, "pew", 0777)
			go fs.RemoveAll(ctx, "pew")
			go fs.Stat(ctx, "/pew")
		})

		t.Run("FS.Stat after FS.Rename", func(t *testing.T) {
			fs.Mkdir(ctx, "pew", 0777)
			go fs.Rename(ctx, "pew", "bar")
			go fs.Stat(ctx, "/bar")
		})
	})

	t.Run("File race conditions", func(t *testing.T) {
		t.Run("File.DeadProps after Patch", func(t *testing.T) {
			root, _ := fs.OpenFile(ctx, "/", os.O_RDONLY, 0777)
			proot := root.(DeadPropsHolder)

			go proot.Patch([]Proppatch{{
				Props: []Property{{
					XMLName: xml.Name{Space: "foo", Local: "bar"},
				}},
			}})

			proot.DeadProps()
		})

		t.Run("File.Stat after Write", func(t *testing.T) {
			f, _ := fs.OpenFile(ctx, "/rw.txt", os.O_RDWR|os.O_CREATE, 0777)
			go f.Write([]byte{0x01})
			f.Stat()
		})

		t.Run("File.Read after Write", func(t *testing.T) {
			f, _ := fs.OpenFile(ctx, "/rw.txt", os.O_RDWR|os.O_CREATE, 0777)
			go f.Write([]byte{0x01})
			f.Read(make([]byte, 100))
		})

		t.Run("File.Seek after Write", func(t *testing.T) {
			f, _ := fs.OpenFile(ctx, "/rw.txt", os.O_RDWR|os.O_CREATE, 0777)
			go f.Write([]byte{0x01})
			f.Seek(1, 1)
		})

		t.Run("File.Readdir manual trigger", func(t *testing.T) {
			fs.Mkdir(ctx, "/raceme", 0777)
			dir, _ := fs.OpenFile(ctx, "/", os.O_RDONLY, 0777)
			for i := 0; i < 100; i++ {
				dir1 := dir.(*fdbFile)
				dir1.childrenSnapshot[0] = nil //trigger race detector

				go dir1.Readdir(100)
			}
		})
	})
}
