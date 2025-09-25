package utils

import (
	"slices"
	"sync"
)

type ComparableType interface {
	Equals(other ComparableType) bool
}

const (
	initialMapCapacity int = 16
)

// ComparableMap is a thread-safe map-like data structure that stores keys and values.
// Keys must implement the ComparableType interface, allowing for comparison operations.
// Values are stored as pointers to allow for efficient updates and nil checks.
// The map maintains a slice of keys and a corresponding slice of value pointers.
// The capacity field tracks the maximum number of elements the map can hold.
// Access to the map is synchronized using a read-write mutex to ensure safe concurrent usage.
type ComparableMap[K ComparableType, V any] struct {
	keys     []K
	values   []*V
	capacity int
	mu       sync.RWMutex
}

// NewComparableMap creates and returns a new instance of ComparableMap with the default initial capacity.
// The map uses keys of type K, which must satisfy the ComparableType constraint, and values of any type V.
// This constructor is useful for initializing a ComparableMap without specifying a custom capacity.
func NewComparableMap[K ComparableType, V any]() *ComparableMap[K, V] {
	return NewComparableMapWithCapacity[K, V](initialMapCapacity)
}

// NewComparableMapWithCapacity creates and returns a new ComparableMap with the specified initial capacity.
// The map uses keys of type K, which must satisfy the ComparableType constraint, and values of type V.
// The initial capacity determines the underlying slice capacities for keys and values, optimizing memory allocation
// for scenarios where the expected number of elements is known in advance.
func NewComparableMapWithCapacity[K ComparableType, V any](initialCapacity int) *ComparableMap[K, V] {
	return &ComparableMap[K, V]{
		keys:     make([]K, 0, initialCapacity),
		values:   make([]*V, 0, initialCapacity),
		capacity: initialCapacity,
		mu:       sync.RWMutex{},
	}
}

// Put inserts or updates the value associated with the specified key.
// If the key does not exist, it is added with the provided value pointer.
// If the key exists, its value pointer is updated.
// Automatically grows internal storage if capacity is reached.
// Thread-safe.
func (m *ComparableMap[K, V]) Put(key K, value *V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := m.Index(key)
	if idx < 0 {
		// Grow internal key/value slices if capacity is reached.
		if len(m.keys) == m.capacity-1 {
			m.capacity += initialMapCapacity
			m.keys = slices.Grow(m.keys, m.capacity)
			m.values = slices.Grow(m.values, m.capacity)
		}
		m.keys = append(m.keys, key)
		m.values = append(m.values, value)
	} else {
		m.values[idx] = value
	}
}

// PutValue inserts or updates the value associated with the specified key.
// Accepts a value (not pointer) and stores its address.
// Thread-safe.
func (m *ComparableMap[K, V]) PutValue(key K, value V) {
	m.Put(key, &value)
}

// Get retrieves the value pointer associated with the specified key.
// Returns the pointer and true if found, or nil and false otherwise.
// Thread-safe.
func (m *ComparableMap[K, V]) Get(key K) (*V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for idx, key := range m.keys {
		if key.Equals(key) {
			return m.values[idx], true
		}
	}
	return nil, false
}

// Index returns the index of the specified key in the map.
// Returns -1 if the key is not found.
// Thread-safe.
func (m *ComparableMap[K, V]) Index(key K) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for idx, key := range m.keys {
		if key.Equals(key) {
			return idx
		}
	}
	return -1
}

// Size returns the number of key-value pairs currently stored in the map.
// Thread-safe.
func (m *ComparableMap[K, V]) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.keys)
}

// ContainsKey checks whether the specified key exists in the map.
// Returns true if the key is present, false otherwise.
// Thread-safe.
func (m *ComparableMap[K, V]) ContainsKey(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, k := range m.keys {
		if k.Equals(key) {
			return true
		}
	}
	return false
}

// Delete removes the key-value pair associated with the specified key.
// Returns the value pointer if the key was present, or nil otherwise.
// Thread-safe.
func (m *ComparableMap[K, V]) Delete(key K) *V {
	m.mu.Lock()
	defer m.mu.Unlock()

	for idx, k := range m.keys {
		if k.Equals(key) {
			value := m.values[idx]

			m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
			m.values = append(m.values[:idx], m.values[idx+1:]...)

			return value
		}
	}
	return nil
}

// Clear removes all key-value pairs from the map, resetting its size.
// Retains the current capacity for future use.
// Thread-safe.
func (m *ComparableMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys = make([]K, 0, m.capacity)
	m.values = make([]*V, 0, m.capacity)
}
