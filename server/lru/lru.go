package lru

import (
	"container/list"
	"time"
)

type expire interface {
	expired() bool
}

type neverExpires struct{}

func (e neverExpires) expired() bool {
	return false
}

type expireAt struct {
	at int64
}

func (e expireAt) expired() bool {
	return time.Now().Unix() > e.at
}

type item struct {
	e     expire
	key   string
	value interface{}
}

func (i *item) expired() bool {
	return i.e.expired()
}

func (i *item) setTTL(ttl int64) {
	if ttl == 0 {
		i.e = neverExpires{}
		return
	}

	i.e = expireAt{time.Now().Unix() + ttl}
}

type LRU struct {
	size  int
	list  *list.List
	items map[string]*list.Element
}

func New(size int) *LRU {
	return &LRU{
		size:  size,
		list:  list.New(),
		items: make(map[string]*list.Element),
	}
}

func (l *LRU) Set(key string, value interface{}, ttl int64) {
	if it, ok := l.items[key]; ok {
		l.list.MoveToFront(it)
		item := it.Value.(*item)

		item.value = value
		item.setTTL(ttl)
	} else {
		it := &item{
			key:   key,
			value: value,
		}

		it.setTTL(ttl)

		l.items[key] = l.list.PushFront(it)

		if l.list.Len() > l.size {
			l.evict()
		}
	}
}

func (l *LRU) Get(key string) (interface{}, bool) {
	if it, ok := l.items[key]; ok {
		item := it.Value.(*item)
		if !item.expired() {
			l.list.MoveToFront(it)
			return item.value, true
		} else {
			l.list.MoveToBack(it)
			return nil, false
		}
	}

	return nil, false
}

func (l *LRU) Delete(key string) {
	if it, ok := l.items[key]; ok {
		l.list.Remove(it)
		delete(l.items, key)
	}
}

func (l *LRU) evict() {
	it := l.list.Back()
	if it != nil {
		l.list.Remove(it)
		delete(l.items, it.Value.(*item).key)
	}
}

func (l *LRU) Keys() []string {
	keys := make([]string, len(l.items))

	i := 0
	for k, v := range l.items {
		if !v.Value.(*item).expired() {
			keys[i] = k
			i++
		}
	}

	return keys[:i]
}

func (l *LRU) Expire(key string, ttl int64) bool {
	if it, ok := l.items[key]; ok {
		l.list.MoveToFront(it)
		it.Value.(*item).setTTL(ttl)

		return true
	}

	return false
}
