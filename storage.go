package logde

import (
	"fmt"
	"sync"
	"time"
)

var ErrNotFound = fmt.Errorf("Element not found")

type Storage interface {
	Set(key, value string, ttl int64) error
	Get(key string) (string, error)
	HSet(key, field, value string, ttl int64) error
	HGet(key, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
	Delete(key string) error
	Keys() ([]string, error)
}

type Memory struct {
	values     map[string]Item
	valuesLock sync.RWMutex

	hashes     map[string]map[string]Item
	hashesLock sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		values: make(map[string]Item),
		hashes: make(map[string]map[string]Item),
	}
}

type Item struct {
	expiresAt int64
	value     string
}

func (i Item) expired() bool {
	if i.expiresAt == 0 {
		return false
	}

	return i.expiresAt < time.Now().Unix()
}

func (m *Memory) Set(key, value string, ttl int64) error {
	m.valuesLock.Lock()
	defer m.valuesLock.Unlock()

	m.values[key] = Item{
		value:     value,
		expiresAt: expiresAt(ttl),
	}

	return nil
}

func (m *Memory) Get(key string) (string, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()
	if item, ok := m.values[key]; ok {
		if !item.expired() {
			return item.value, nil
		}
	}

	return "", ErrNotFound
}

func (m *Memory) HSet(key, field, value string, ttl int64) error {
	m.hashesLock.Lock()
	defer m.hashesLock.Unlock()

	item := Item{
		value:     value,
		expiresAt: expiresAt(ttl),
	}

	if hash, ok := m.hashes[key]; ok {
		hash[field] = item
	} else {
		m.hashes[key] = map[string]Item{
			field: item,
		}
	}

	return nil
}

func (m *Memory) HGet(key, field string) (string, error) {
	m.hashesLock.RLock()
	defer m.hashesLock.RUnlock()

	if hash, ok := m.hashes[key]; ok {
		if item, ok := hash[field]; ok {
			if !item.expired() {
				return item.value, nil
			}
		}
	}

	return "", ErrNotFound
}

func (m *Memory) HGetAll(key string) (map[string]string, error) {
	m.hashesLock.RLock()
	defer m.hashesLock.RUnlock()

	if hash, ok := m.hashes[key]; ok {
		result := make(map[string]string, len(hash))
		for field, item := range hash {
			if !item.expired() {
				result[field] = item.value
			}
		}

		return result, nil
	}

	return nil, ErrNotFound
}

func (m *Memory) Delete(key string) error {
	m.valuesLock.Lock()
	defer m.valuesLock.Unlock()

	delete(m.values, key)

	return nil
}

func (m *Memory) Keys() ([]string, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()

	result := make([]string, len(m.values))
	i := 0
	for key, _ := range m.values {
		result[i] = key
		i++
	}

	return result, nil
}

func expiresAt(ttl int64) int64 {
	var expiresAt int64

	if ttl != 0 {
		expiresAt = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	}

	return expiresAt
}
