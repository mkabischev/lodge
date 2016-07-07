package server

import (
	"sync"

	"github.com/mkabischev/lodge/server/lru"
)

type lruStorage struct {
	sync.Mutex

	data *lru.LRU
}

func NewLRUStorage(l *lru.LRU) Storage {
	return &lruStorage{
		data: l,
	}
}

func (s *lruStorage) Set(key, value string, ttl int64) error {
	s.Lock()
	defer s.Unlock()

	s.data.Set(key, value, ttl)

	return nil
}

func (s *lruStorage) Get(key string) (string, error) {
	s.Lock()
	defer s.Unlock()

	if val, ok := s.data.Get(key); ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}

	return "", errNotFound
}

func (s *lruStorage) HSet(key, field, value string) error {
	s.Lock()
	defer s.Unlock()

	if val, ok := s.data.Get(key); ok {
		if hash, ok := val.(map[string]string); ok {
			hash[field] = value
			s.data.Set(key, hash, 0)

			return nil
		}

		return errWrongType
	}

	hash := map[string]string{
		field: value,
	}

	s.data.Set(key, hash, 0)

	return nil
}

func (s *lruStorage) HGet(key, field string) (string, error) {
	if val, ok := s.data.Get(key); ok {
		if hash, ok := val.(map[string]string); ok {
			if value, ok := hash[field]; ok {
				return value, nil
			}
		}
	}

	return "", errNotFound
}

func (s *lruStorage) HGetAll(key string) (map[string]string, error) {
	if val, ok := s.data.Get(key); ok {
		if hash, ok := val.(map[string]string); ok {
			result := make(map[string]string, len(hash))
			for k, v := range hash {
				result[k] = v
			}

			return result, nil
		} else {
			return nil, errWrongType
		}
	}

	return nil, errNotFound
}

func (s *lruStorage) Delete(key string) error {
	s.Lock()
	defer s.Unlock()

	s.data.Delete(key)

	return nil
}

func (s *lruStorage) Keys() ([]string, error) {
	s.Lock()
	defer s.Unlock()

	return s.data.Keys(), nil
}

func (s *lruStorage) Expire(key string, ttl int64) error {
	if ok := s.data.Expire(key, ttl); ok {
		return nil
	}

	return errNotFound
}
