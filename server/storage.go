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
	HSet(key, field, value string) error
	HGet(key, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
	Delete(key string) error
	Keys() ([]string, error)
	Expire(key string, ttl int64) error
}

type Memory struct {
	values        map[string]Item
	valuesLock    sync.RWMutex
	cleanupPeriod time.Duration
}

func NewMemory(cleanupPeriod time.Duration) *Memory {
	storage := &Memory{
		values:        make(map[string]Item),
		cleanupPeriod: cleanupPeriod,
	}

	go storage.cleanup()

	return storage
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

func (m *Memory) cleanup() {
	for _ = range time.Tick(m.cleanupPeriod) {
		m.doCleanup()
	}
}

func (m *Memory) doCleanup() {
	m.valuesLock.Lock()
	defer m.valuesLock.Unlock()

	for key, item := range m.values {
		if item.expired() {
			delete(m.values, key)
		}
	}
}

func (m *Memory) Set(key, value string, ttl int64) error {
	m.valuesLock.Lock()
	defer m.valuesLock.Unlock()

	if item, ok := m.values[key]; ok {
		if _, ok := item.value.(string); !ok {
			return errWrongType
		}
	}

	m.values[key] = Item{
		value:     value,
		expiresAt: expiresAfter(ttl),
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

func (m *Memory) HSet(key, field, value string) error {
	m.valuesLock.Lock()
	defer m.valuesLock.Unlock()

	if hashItem, ok := m.values[key]; ok {
		if hash, ok := hashItem.value.(map[string]string); ok {
			hash[field] = value
		} else {
			return errWrongType
		}
	} else {
		m.values[key] = Item{
			value: map[string]string{
				field: value,
			},
		}
	}

	return nil
}

func (m *Memory) HGet(key, field string) (string, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()

	if hashItem, ok := m.values[key]; ok {
		if hash, ok := hashItem.value.(map[string]string); ok {
			if value, ok := hash[field]; ok {
				return value, nil
			}
		}
	} else {
		return "", errWrongType
	}

	return "", errNotFound
}

func (m *Memory) HGetAll(key string) (map[string]string, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()

	if hashItem, ok := m.values[key]; ok {
		if hash, ok := hashItem.value.(map[string]string); ok {
			result := make(map[string]string, len(hash))
			for field, value := range hash {
				result[field] = value
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

func (m *Memory) Expire(key string, ttl int64) error {
	m.valuesLock.Lock()
	m.valuesLock.Unlock()

	if item, ok := m.values[key]; ok {
		if !item.expired() {
			item.expiresAt = expiresAfter(ttl)
			return nil
		}
	}

	return errNotFound
}

func expiresAfter(ttl int64) int64 {
	var expiresAt int64

	if ttl != 0 {
		expiresAt = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	}

	return expiresAt
}
