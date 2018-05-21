package webdav

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestFDBFileRW(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)

	ctx := context.Background()

	f1, err := fs.OpenFile(ctx, "hello.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f1.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	err = f1.Close()
	if err != nil {
		t.Fatal(err)
	}

	f2, err := fs.OpenFile(ctx, "hello.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}

	fi, _ := f2.Stat()
	if fi.Size() != 5 {
		t.Fatal("expected this size")
	}

	data, err := ioutil.ReadAll(f2)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, []byte("hello")) {
		t.Fatal("expected to read back data")
	}
}

func TestFDBNodeRoot(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)

	root, err := fs.Root(fs.tr)
	if err != nil {
		t.Fatal(err)
	}

	if root == nil {
		t.Fatal("root node should get something")
	}

	if root.ModTime.Sub(time.Now()) > time.Millisecond {
		t.Fatal("root node time should be recent (new)")
	}

	if !root.Mode.IsDir() {
		t.Fatal("root node must be directory")
	}

	root2, err := fs.Root(fs.tr)
	if err != nil {
		t.Fatal(err)
	}

	if !root.ModTime.Equal(root2.ModTime) {
		t.Fatal("second root call should get root that equals first")
	}
}

func TestProps(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)
	root, _ := fs.Root(fs.tr)

	if err := fs.PutProp(fs.tr, root.ID, xml.Name{Space: "x:", Local: "ring"}, Property{
		XMLName:  xml.Name{Space: "x:", Local: "ring"},
		InnerXML: []byte("1 shilling"),
	}); err != nil {
		t.Fatal(err)
	}

	if err := fs.PutProp(fs.tr, root.ID, xml.Name{Space: "x:", Local: "ring2"}, Property{
		XMLName:  xml.Name{Space: "Y:", Local: "ring2"},
		InnerXML: []byte("2 shilling"),
	}); err != nil {
		t.Fatal(err)
	}

	n := 0
	vals := []byte{}
	if err := fs.Props(fs.tr, root.ID, func(p Property) (r bool) {
		vals = append(vals, p.InnerXML...)
		n++
		return
	}); err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatal("expected 2 props")
	}

	if string(vals) != "1 shilling2 shilling" {
		t.Fatal("expected property values to be intact")
	}

	if err := fs.DelProp(fs.tr, root.ID, xml.Name{Space: "x:", Local: "ring"}); err != nil {
		t.Fatal(err)
	}

	n = 0
	vals = []byte{}
	if err := fs.Props(fs.tr, root.ID, func(p Property) (r bool) {
		vals = append(vals, p.InnerXML...)
		n++
		return
	}); err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatal("expected 1 props")
	}

	if string(vals) != "2 shilling" {
		t.Fatal("expected property values to be intact")
	}
}

func TestFDBNodePutGetDelGet(t *testing.T) {
	db, ns, clean := open(t)
	defer clean()
	fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)
	root, _ := fs.Root(fs.tr)

	n1 := &Node{ID: genID(), ModTime: time.Now(), Mode: 0666 | os.ModeDir}
	err := fs.Put(fs.tr, root, "hello.txt", n1)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Put(fs.tr, root, "hello2.txt", n1)
	if err != nil {
		t.Fatal(err)
	}

	n2, err := fs.Get(fs.tr, root, "hello.txt")
	if err != nil {
		t.Fatal(err)
	}

	if n2 == nil {
		t.Fatal("should find the node by filename")
	}

	if n1.ID != n2.ID || !n1.ModTime.Equal(n2.ModTime) || n1.Mode != n2.Mode {
		t.Fatal("expected nodes to be equal")
	}

	i := 0
	if err = fs.Children(fs.tr, root, func(frag string, n *Node) (r bool) {
		i++
		return
	}); err != nil {
		t.Fatal("iteration should succeed")
	}

	if i != 2 {
		t.Fatal("expected to iterate over two children")
	}

	i = 0
	if err = fs.Children(fs.tr, root, func(frag string, n *Node) (r bool) {
		i++
		return true
	}); err != nil {
		t.Fatal("iteration should succeed")
	}

	if i != 1 {
		t.Fatal("expected to iterate over one child")
	}

	err = fs.Del(fs.tr, root, "hello.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Get(fs.tr, root, "hello.txt")
	if err != ErrNodeNotExist {
		t.Fatal("expected node not exist error, got: ", err.Error())
	}
}

func TestFDBWalk2(t *testing.T) {
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

func TestFDBFSRoot2(t *testing.T) {
	ctx := context.Background()
	db, ns, clean := open(t)
	defer clean()

	fs := NewFDBFS(db, ns, testCfg()).(*FDBFS)
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
