package server

import (
	"fmt"
	"sync"
	"time"
)

var errNotFound = fmt.Errorf("Element not found")
var errWrongType = fmt.Errorf("Wrong type")

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
}

func NewMemory() *Memory {
	return &Memory{
		values: make(map[string]Item),
	}
}

type Item struct {
	expiresAt int64
	value     interface{}
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
			if str, ok := item.value.(string); ok {
				return str, nil
			} else {
				return "", errWrongType
			}
		}
	}

	return "", errNotFound
}

func (m *Memory) HSet(key, field, value string, ttl int64) error {
	m.valuesLock.Lock()
	defer m.valuesLock.Unlock()

	item := Item{
		value:     value,
		expiresAt: expiresAt(ttl),
	}

	if hashItem, ok := m.values[key]; ok {
		if hash, ok := hashItem.value.(map[string]Item); ok {
			hash[field] = item
		} else {
			return errWrongType
		}
	} else {
		m.values[key] = Item{
			value: map[string]Item{
				field: Item{expiresAt: expiresAt(ttl), value: value},
			},
		}
	}

	return nil
}

func (m *Memory) HGet(key, field string) (string, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()

	if hashItem, ok := m.values[key]; ok {
		if hash, ok := hashItem.value.(map[string]Item); ok {
			if item, ok := hash[field]; ok {
				if !item.expired() {
					return item.value.(string), nil
				}
			}
		} else {
			return "", errWrongType
		}
	}

	return "", errNotFound
}

func (m *Memory) HGetAll(key string) (map[string]string, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()

	if hashItem, ok := m.values[key]; ok {
		if hash, ok := hashItem.value.(map[string]Item); ok {
			result := make(map[string]string, len(hash))
			for field, item := range hash {
				if !item.expired() {
					result[field] = item.value.(string)
				}
			}

			return result, nil
		}
	}

	return nil, errNotFound
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
