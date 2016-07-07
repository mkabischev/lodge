package server

import (
	"hash/crc32"
	"math"
)

func NewBucketStorage(n int, factory func() Storage) Storage {
	buckets := make([]Storage, n)

	for i := 0; i < n; i++ {
		buckets[i] = factory()
	}

	return &bucketStorage{
		buckets: buckets,
	}
}

type bucketStorage struct {
	buckets []Storage
}

func (s *bucketStorage) Set(key, value string, ttl int64) error {
	return s.bucket(key).Set(key, value, ttl)
}

func (s *bucketStorage) Get(key string) (string, error) {
	return s.bucket(key).Get(key)
}

func (s *bucketStorage) HSet(key, field, value string) error {
	return s.bucket(key).HSet(key, field, value)
}

func (s *bucketStorage) HGet(key, field string) (string, error) {
	return s.bucket(key).HGet(key, field)
}

func (s *bucketStorage) HGetAll(key string) (map[string]string, error) {
	return s.bucket(key).HGetAll(key)
}

func (s *bucketStorage) Delete(key string) error {
	return s.bucket(key).Delete(key)
}

func (s *bucketStorage) Keys() ([]string, error) {
	result := make([]string, 0)

	for _, bucket := range s.buckets {
		bucketKeys, err := bucket.Keys()
		if err != nil {
			return nil, err
		}

		result = append(result, bucketKeys...)
	}

	return result, nil
}

func (s *bucketStorage) Expire(key string, ttl int64) error {
	return s.bucket(key).Expire(key, ttl)
}

func (s *bucketStorage) bucket(key string) Storage {
	sum := crc32.ChecksumIEEE([]byte(key))
	i := int(math.Mod(float64(sum), float64(len(s.buckets))))

	return s.buckets[i]
}
