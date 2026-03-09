package store

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/UneBaguette/shorten-go/internal/model"

	"github.com/dgraph-io/badger/v4"
)

const MaxLinksPerIP = 5
const MaxCreationsPerDay = 15
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

func (s *Store) Set(url *model.URL, ttl time.Duration) error {
	count, err := s.CountByIP(url.IP)

	if err != nil {
		return err
	}

	if count >= MaxLinksPerIP {
		return fmt.Errorf("limit_reached")
	}

	creations, err := s.GetCreations(url.IP)

	if err != nil {
		return err
	}

	if creations > MaxCreationsPerDay {
		return fmt.Errorf("daily_limit_reached")
	}

	data, err := json.Marshal(url)

	if err != nil {
		return err
	}

	err = s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(url.Code), data).WithTTL(ttl)

		if err := txn.SetEntry(entry); err != nil {
			return err
		}

		ipKey := []byte("ip:" + url.IP + ":" + url.Code)
		ipEntry := badger.NewEntry(ipKey, []byte{}).WithTTL(ttl)

		return txn.SetEntry(ipEntry)
	})

	if err != nil {
		return err
	}

	_, err = s.IncrementCreations(url.IP, ttl)
	return err
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

func (s *Store) Delete(code string, ip string) error {
	return s.db.Update(func(txn *badger.Txn) error {

		if err := txn.Delete([]byte(code)); err != nil {
			return err
		}

		ipKey := []byte("ip:" + ip + ":" + code)
		return txn.Delete(ipKey)
	})
}

func (s *Store) GenerateCode() string {
	b := make([]byte, 6)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func (s *Store) CountByIP(ip string) (int, error) {
	count := 0

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := []byte("ip:" + ip + ":")

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}

		return nil
	})

	return count, err
}

func (s *Store) IncrementCreations(ip string, ttl time.Duration) (int, error) {
	key := []byte("creations:" + ip)
	var count int

	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		if err == nil {
			err = item.Value(func(val []byte) error {
				count = int(val[0])
				return nil
			})

			if err != nil {
				return err
			}
		}

		count++
		entry := badger.NewEntry(key, []byte{byte(count)}).WithTTL(24 * time.Hour)

		return txn.SetEntry(entry)
	})

	return count, err
}

func (s *Store) GetCreations(ip string) (int, error) {
	key := []byte("creations:" + ip)
	var count int

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		if err == badger.ErrKeyNotFound {
			return nil
		}

		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			count = int(val[0])
			return nil
		})
	})

	return count, err
}

func GenerateToken() string {
	b := make([]byte, 32)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
