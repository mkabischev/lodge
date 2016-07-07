package server

import (
	"strconv"
	"testing"
	"time"

	"github.com/mkabischev/lodge/server/lru"
)

func benchMemorySet(s Storage) func(*testing.PB) {
	return func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			s.Set(strconv.Itoa(i), "foo", 0)
		}
	}
}

func benchMemoryGet(s Storage) func(*testing.PB) {
	return func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			s.Get(strconv.Itoa(i))
		}
	}
}

func benchMemoryCombine(n int, s Storage) func(*testing.PB) {
	return func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if i%n == 0 {
				s.Set(strconv.Itoa(i), "foo", 0)
			} else {
				s.Get(strconv.Itoa(i))
			}
		}
	}
}

func BenchmarkStorageMemorySet(b *testing.B) {
	storage := NewMemory(1 * time.Second)

	b.RunParallel(benchMemorySet(storage))
}

func BenchmarkStorageBucketSet(b *testing.B) {
	storage := NewBucketStorage(100, func() Storage {
		return NewLRUStorage(lru.New(1000000))
	})

	b.RunParallel(benchMemorySet(storage))
}

func BenchmarkLStorageRUSet(b *testing.B) {
	storage := NewLRUStorage(lru.New(1000000))

	b.RunParallel(benchMemorySet(storage))
}

func BenchmarkStorageMemoryGet(b *testing.B) {
	storage := NewMemory(1 * time.Second)

	b.RunParallel(benchMemoryGet(storage))
}

func BenchmarkStorageBucketGet(b *testing.B) {
	storage := NewBucketStorage(100, func() Storage {
		return NewLRUStorage(lru.New(1000000))
	})

	b.RunParallel(benchMemoryGet(storage))
}

func BenchmarkLStorageRUGet(b *testing.B) {
	storage := NewLRUStorage(lru.New(1000000))

	b.RunParallel(benchMemoryGet(storage))
}

func BenchmarkStorageMemoryCombine(b *testing.B) {
	storage := NewMemory(1 * time.Second)

	b.RunParallel(benchMemoryCombine(5, storage))
}

func BenchmarkStorageBucketCombine(b *testing.B) {
	storage := NewBucketStorage(100, func() Storage {
		return NewLRUStorage(lru.New(1000000))
	})

	b.RunParallel(benchMemoryCombine(5, storage))
}

func BenchmarkStorageLRUCombine(b *testing.B) {
	storage := NewLRUStorage(lru.New(1000000))

	b.RunParallel(benchMemoryCombine(5, storage))
}
