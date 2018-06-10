package model

import (
	"bytes"
	"encoding/gob"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
)

type Run struct {
	ID  string
	Job Job
}

type Data struct {
	Dest string
}

type Task struct {
	Image   string
	Command []string
	Data    map[string]Data `hcl:"data"`
}

type Job struct {
	Workspace string
	Tasks     map[string]Task `hcl:"task"`
}

type Model struct {
	tr   fdb.Transactor
	ss   directory.DirectorySubspace
	runs directory.DirectorySubspace
}

func New(tr fdb.Transactor) (m *Model, clean func() error, err error) {
	m = &Model{tr: tr}
	m.ss, err = directory.CreateOrOpen(tr, []string{"ffs_model"}, nil)
	if err != nil {
		return nil, nil, err
	}

	gob.Register(&Run{})
	m.runs, err = m.ss.CreateOrOpen(tr, []string{"runs"}, nil)
	if err != nil {
		return nil, nil, err
	}

	return m, func() error {
		_, rerr := m.ss.Remove(tr, nil)
		if rerr != nil {
			return rerr
		}

		return nil
	}, nil
}

func (m *Model) encode(v interface{}) (d []byte, err error) {
	buf := bytes.NewBuffer(d)
	err = gob.NewEncoder(buf).Encode(v)
	return buf.Bytes(), err
}

func (m *Model) decode(buf []byte, v interface{}) (err error) {
	return gob.NewDecoder(bytes.NewBuffer(buf)).Decode(v)
}
