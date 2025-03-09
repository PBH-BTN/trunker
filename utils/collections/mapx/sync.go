//go:build go1.24

package mapx

import (
	"sync"
)

// Map a wrapper for sync.Map to support generic
type Map[V any] struct {
	m sync.Map
}

var _ SyncStringMap[int] = &Map[int]{}

func (m *Map[V]) Delete(key string) bool {
	_, ok := m.LoadAndDelete(key)
	return ok
}
func (m *Map[V]) Load(key string) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}
func (m *Map[V]) LoadAndDelete(key string) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return value, loaded
	}
	return v.(V), loaded

}
func (m *Map[V]) LoadOrStore(key string, value V) (actual V, loaded bool) {
	a, loaded := m.m.LoadOrStore(key, value)
	return a.(V), loaded
}
func (m *Map[V]) Range(f func(key string, value V) bool) {
	m.m.Range(func(key, value any) bool { return f(key.(string), value.(V)) })
}
func (m *Map[V]) Store(key string, value V) {
	m.m.Store(key, value)
}

func (m *Map[V]) LoadOrStoreLazy(key string, lazy func() V) (actual V, loaded bool) {
	v, loaded := m.m.Load(key)
	if !loaded {
		v = lazy()
		a, loaded := m.m.LoadOrStore(key, v)
		return a.(V), loaded
	}
	return v.(V), loaded
}

func (m *Map[V]) Len() int {
	count := 0
	m.m.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func New[V any]() SyncStringMap[V] {
	return &Map[V]{}
}
