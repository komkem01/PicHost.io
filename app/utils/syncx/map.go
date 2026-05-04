// Package syncx provides a generic wrapper around sync.Map, offering type safety for keys and values.
package syncx

import "sync"

// Map is a generic wrapper around sync.Map, providing type safety for keys and values.
type Map[K comparable, V any] struct {
	m sync.Map
}

// Delete removes the value for a key.
func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Load retrieves the value for a key. The ok result indicates whether the key was found.
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if ok {
		value = v.(V)
	}
	return
}

// LoadAndDelete retrieves and removes the value for a key. The loaded result indicates whether the key was found.
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if loaded {
		value = v.(V)
	}
	return
}

// LoadOrStore retrieves the existing value for a key if present. Otherwise, it stores and returns the given value.
// The loaded result indicates whether the value was loaded (true) or stored (false).
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := m.m.LoadOrStore(key, value)
	actual = a.(V)
	return
}

// Range calls the given function for each key and value pair in the map.
// If the function returns false, Range stops the iteration.
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}
