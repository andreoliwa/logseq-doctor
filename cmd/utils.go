package cmd

import (
	"fmt"
	"github.com/fatih/color"
)

// Set is a simple implementation of a set using a map.
type Set[T comparable] struct {
	data map[T]struct{}
}

// NewSet creates and returns a new set.
func NewSet[T comparable]() *Set[T] {
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

// FormatCount returns a string with the count and the singular or plural form of a word.
func FormatCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}

	return fmt.Sprintf("%d %s", count, plural)
}

// PageColor is a color function for page names.
var PageColor = color.New(color.FgHiWhite).SprintfFunc() //nolint:gochecknoglobals
