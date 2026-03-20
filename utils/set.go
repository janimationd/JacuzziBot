package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Set is a thread-safe collection of unique elements.
type Set[T comparable] struct {
	mu   sync.RWMutex
	data map[T]struct{}
}

// Initializes a new thread-safe set.
func NewSet[T comparable]() *Set[T] {
	return NewSetN[T](0)
}

// Initializes a new thread-safe set with initial capacity.
func NewSetN[T comparable](capacity int) *Set[T] {
	return &Set[T]{
		data: make(map[T]struct{}, capacity),
	}
}

// Add inserts a value. Returns false if the value already existed.
func (s *Set[T]) Add(obj T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[obj]; exists {
		return false
	}
	s.data[obj] = struct{}{}
	return true
}

// Remove deletes a value. Returns false if the value was not found.
func (s *Set[T]) Remove(obj T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[obj]; !exists {
		return false
	}
	delete(s.data, obj)
	return true
}

// Contains returns true if the value is in the set.
func (s *Set[T]) Contains(obj T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.data[obj]
	return exists
}

// All returns a thread-safe iterator.
// The RLock is held for the entire duration of the loop *in calling code*, which is pretty neat and prevents panics.
func (s *Set[T]) All() func(func(T) bool) {
	return func(yield func(T) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		for v := range s.data {
			if !yield(v) {
				return
			}
		}
	}
}

// Returns the size/length of the set.
func (s *Set[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

// Since making the class thread-safe has the tradeoff of hiding the map data in a private variable,
// we need to write custom JSON ser/de logic. At least the serialized output will look like a normal
// JSON array now.
func (s *Set[T]) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert the set to a simple list for JSON
	keys := make([]T, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[T]struct{})
	for _, item := range items {
		s.data[item] = struct{}{}
	}
	return nil
}

func (s *Set[T]) ToString() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parts := make([]string, 0, len(s.data))
	for v := range s.data {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return strings.Join(parts, ", ")
}
