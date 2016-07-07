package lru

import (
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	lru := New(1)

	lru.Set("foo", "bar", 0)
	assertValue(t, lru, "foo", "bar")
}

func TestSetGetWithEviction(t *testing.T) {
	lru := New(2)

	lru.Set("key1", "value1", 0)
	lru.Set("key2", "value2", 0)
	lru.Set("key3", "value3", 0)

	assertNotFound(t, lru, "key1")
	assertFound(t, lru, "key2")
	assertFound(t, lru, "key3")
}

func TestSetGetExpire(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	lru := New(3)

	lru.Set("key1", "value1", 0)
	lru.Set("key2", "value2", 1)
	lru.Set("key3", "value3", 1)

	time.Sleep(2 * time.Second)

	// key1 was insert first, so it has to be last
	assertLastKey(t, lru, "key1")
	assertValue(t, lru, "key1", "value1")

	// retrieving key2, so it moves to back
	assertNotFound(t, lru, "key2")
	assertLastKey(t, lru, "key2")

	assertNotFound(t, lru, "key3")

	// retrieving key3, so it moves to back
	assertLastKey(t, lru, "key3")
}

func assertValue(t *testing.T, l *LRU, key, value string) {
	val, ok := l.Get(key)

	if !ok {
		t.Fatalf("Key '%v' not found, but expected", key)
	}

	if val.(string) != value {
		t.Fatalf("Expected: %v. Got: %v", value, val)
	}
}

func assertFound(t *testing.T, l *LRU, key string) {
	if _, ok := l.Get(key); !ok {
		t.Fatalf("Key '%v' not found, but expected", key)
	}
}

func assertNotFound(t *testing.T, l *LRU, key string) {
	if _, ok := l.Get(key); ok {
		t.Fatalf("Key '%v' found, but not expected", key)
	}
}

func assertLastKey(t *testing.T, l *LRU, key string) {
	lastItem := l.list.Back().Value.(*item)
	if lastItem.key != key {
		t.Fatalf("Expected last key: %v. Got: %v", key, lastItem.key)
	}
}
