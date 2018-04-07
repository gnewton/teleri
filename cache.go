package main

import (
	"log"
	"time"

	b "github.com/dgraph-io/badger"
)

// cache of FQ filename:timestamp
type FileTimeCache interface {
	HasChanged(filename string, t *time.Time) bool
	Close() error
}

type fileTimeCache struct {
	db *b.DB
}

func NewFileTimeCache(dir string) (FileTimeCache, error) {
	ftc := new(fileTimeCache)
	opts := b.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir

	var err error
	ftc.db, err = b.Open(opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ftc, nil
}

func (f *fileTimeCache) HasChanged(filename string, t *time.Time) bool {
	if true {
		return true
	}
	var tcache time.Time

	err := f.db.View(func(txn *b.Txn) error {
		item, err := txn.Get([]byte(filename))
		if err != nil {
			//log.Println(err)
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		err = tcache.GobDecode(val)
		if err != nil {
			return err
		}
		return nil
	})

	// key found
	if err == nil {
		log.Println(t.After(tcache), t, tcache)
		return t.After(tcache)
	} else {
		// key not found
		if err == b.ErrKeyNotFound {
			//write key to db
			// this should be done in bulk!
			err = f.db.Update(func(txn *b.Txn) error {
				tbytes, errt := t.GobEncode()
				if errt != nil {
					log.Println(errt)
					return errt
				}
				err := txn.Set([]byte(filename), tbytes)
				return err
			})
			if err != nil {
				log.Println(err)
			}
		}
		return true
	}

	// See if in cache
	// if no, add to cache and return true
	// if yes, retieve & compare times
	return true
}

func (f *fileTimeCache) Close() error {
	return nil
}
