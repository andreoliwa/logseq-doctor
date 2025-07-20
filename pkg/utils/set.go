package utils

import (
	"cmp"
	"sort"
)

// Set is a simple implementation of a set using a map.
type Set[T cmp.Ordered] struct {
	data map[T]struct{}
}

// NewSet creates and returns a new set.
func NewSet[T cmp.Ordered]() *Set[T] {
	return &Set[T]{data: make(map[T]struct{})}
}

// Add inserts an element into the set.
func (s *Set[T]) Add(value T) {
	s.data[value] = struct{}{}
}

// Remove deletes an element from the set.
func (s *Set[T]) Remove(value T) {
	delete(s.data, value)
}

// Contains checks if an element exists in the set.
func (s *Set[T]) Contains(value T) bool {
	_, exists := s.data[value]

	return exists
}

// Size returns the number of elements in the set.
func (s *Set[T]) Size() int {
	return len(s.data)
}

// Values returns all elements in the set as a slice.
func (s *Set[T]) Values() []T {
	keys := make([]T, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}

	return keys
}

// ValuesSorted returns all elements in the set as a sorted slice.
func (s *Set[T]) ValuesSorted() []T {
	keys := s.Values()
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

// Diff returns a new set containing elements that are in the current set but not in the provided sets.
func (s *Set[T]) Diff(sets ...*Set[T]) *Set[T] {
	result := NewSet[T]()

	// Copy all elements of the current set into the result
	for key := range s.data {
		result.Add(key)
	}

	// Remove elements that are present in any of the given sets
	for _, otherSet := range sets {
		for key := range otherSet.data {
			result.Remove(key)
		}
	}

	return result
}

// Update adds all elements from the given sets into the current set.
func (s *Set[T]) Update(sets ...*Set[T]) {
	for _, otherSet := range sets {
		for key := range otherSet.data {
			s.Add(key)
		}
	}
}

// Clear removes all elements from the set.
func (s *Set[T]) Clear() {
	s.data = make(map[T]struct{})
}
