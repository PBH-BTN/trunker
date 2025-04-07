//go:build go1.24

package mapx

import (
	"sync"
	"sync/atomic"
)

// Map a wrapper for sync.Map to support generic
type Map[K, V any] struct {
	m      sync.Map
	length int64
}

var _ SyncStringMap[int] = &Map[string, int]{}

func (m *Map[K, V]) Delete(key K) bool {
	_, ok := m.m.LoadAndDelete(key)
	if ok {
		atomic.AddInt64(&m.length, -1)
	}
	return ok
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if loaded {
		atomic.AddInt64(&m.length, -1)
		return v.(V), loaded
	}
	var empty V
	return empty, loaded
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := m.m.LoadOrStore(key, value)
	if !loaded {
		atomic.AddInt64(&m.length, 1)
	}
	return a.(V), loaded
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool { return f(key.(K), value.(V)) })
}

func (m *Map[K, V]) Store(key K, value V) {
	_, loaded := m.m.LoadOrStore(key, value)
	if !loaded {
		atomic.AddInt64(&m.length, 1)
	} else {
		m.m.Store(key, value)
	}
}

func (m *Map[K, V]) LoadOrStoreLazy(key K, lazy func() V) (actual V, loaded bool) {
	v, loaded := m.m.Load(key)
	if !loaded {
		v = lazy()
		a, loaded := m.m.LoadOrStore(key, v)
		if !loaded {
			atomic.AddInt64(&m.length, 1)
		}
		return a.(V), loaded
	}
	return v.(V), loaded
}

func (m *Map[K, V]) Len() int {
	return int(atomic.LoadInt64(&m.length))
}

func NewSyncMap[K, V any]() Map[K, V] {
	return Map[K, V]{}
}
