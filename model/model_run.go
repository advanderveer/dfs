package model

import (
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	uuid "github.com/nu7hatch/gouuid"
)

func (m *Model) ViewRun(id string) (r *Run, err error) {
	_, err = m.tr.Transact(func(tx fdb.Transaction) (res interface{}, e error) {
		r = &Run{}
		e = m.decode(tx.Get(m.runs.Pack(tuple.Tuple{id})).MustGet(), r)
		if e != nil {
			return nil, e
		}

		return
	})

	return
}

func (m *Model) EachRun(f func(run *Run) bool) error {
	if _, err := m.tr.Transact(func(tx fdb.Transaction) (res interface{}, err error) {
		iter := tx.GetRange(m.runs, fdb.RangeOptions{}).Iterator()
		for iter.Advance() {
			kv := iter.MustGet()

			run := &Run{}
			err = m.decode(kv.Value, run)
			if err != nil {
				fmt.Printf("failed to decode: %v\n", err)
				continue
			}

			if !f(run) {
				break
			}
		}

		return nil, nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *Model) CreateRun(job *Job) (r *Run, err error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	r = &Run{ID: uid.String(), Job: *job}
	if _, err = m.tr.Transact(func(tx fdb.Transaction) (res interface{}, err error) {
		d, err := m.encode(r)
		if err != nil {
			return nil, err
		}

		tx.Set(m.runs.Pack(tuple.Tuple{r.ID}), d)
		return
	}); err != nil {
		return nil, err
	}

	return
}
