package cmd

import (
	"reflect"
	"testing"
)

func TestSortAndRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty slice", []string{}, []string{}},
		{"one element", []string{"apple"}, []string{"apple"}},
		{"duplicates", []string{"orange", "apple", "banana", "apple"}, []string{"apple", "banana", "orange"}},
		{"sorted unique", []string{"orange", "banana", "apple"}, []string{"apple", "banana", "orange"}},
		{"unsorted with duplicates", []string{"orange", "banana", "apple", "apple", "orange"}, []string{"apple", "banana", "orange"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := sortAndRemoveDuplicates(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}
