package cmd

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
