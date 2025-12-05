package util

import "sync"

type SafeMap[K comparable, V any] struct {
	m sync.Map
}

func (m *SafeMap[K, V]) Get(key K) (value V, ok bool) {
	ivalue, ok := m.m.Load(key)
	if !ok {
		return
	}
	return ivalue.(V), true
}

func (m *SafeMap[K, V]) Set(key K, value V) {
	m.m.Store(key, value)
}

func (m *SafeMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *SafeMap[K, V]) Clear() {
	m.m.Range(func(key, value any) bool {
		m.m.Delete(key)
		return true
	})
}
