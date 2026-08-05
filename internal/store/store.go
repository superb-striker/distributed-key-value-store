package store

import (
	"errors"
	"fmt"

	"go.etcd.io/bbolt"
)

var ErrKeyNotFound = errors.New("store: key not found")

var bucketName = []byte("kv")

type Store struct {
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		 _, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("store: init buckets: %w", err)
	}
	return &Store{db: db, path: path}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Get returns the value for key, or ErrKeyNotFound if it doesn't exist.
func (s *Store) Get(key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		v := b.Get([]byte(key))
		if v == nil {
			return ErrKeyNotFound
		}
		// bbolt's returned slice is only valid for the transaction's
		// lifetime, so it must be copied out before View returns.
		val = append([]byte(nil), v...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Put writes key -> value, overwriting any existing value.
func (s *Store) Put(key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(key), value)
	})
}

// Delete removes key. It is not an error to delete a key that doesn't exist
// (idempotent, matching typical KV-store DELETE semantics).
func (s *Store) Delete(key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(key))
	})
}
