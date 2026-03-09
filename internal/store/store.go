package store

import (
	"encoding/json"
	"math/rand"

	"github.com/UneBaguette/shorten-go/internal/model"

	"github.com/dgraph-io/badger/v4"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Store struct {
	db *badger.DB
}

func New(path string) (*Store, error) {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Set(url *model.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(url.Code), data)
	})
}

func (s *Store) Get(code string) (*model.URL, error) {
	var url model.URL
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(code))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &url)
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return &url, err
}

func (s *Store) Delete(code string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(code))
	})
}

func (s *Store) GenerateCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
