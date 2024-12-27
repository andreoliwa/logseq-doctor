package cmd

import (
	"os"
	"path/filepath"
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

func TestIsValidMarkdownFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"empty path", "", false},
		{"invalid extension", "file.txt", false},
		{"non-existent file", "non_existent_file.md", false},
		{"directory path", "./", false},
		{"valid markdown file", "valid_markdown_file.md", true},
	}

	// Create a temporary directory for test files
	dir := t.TempDir()

	// Create a valid markdown file
	validFilePath := filepath.Join(dir, "valid_markdown_file.md")
	err := os.WriteFile(validFilePath, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Update the file path for the valid markdown file test case
	tests[4].filePath = validFilePath

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isValidMarkdownFile(test.filePath)
			if result != test.expected {
				t.Errorf("For %q, expected %v, got %v", test.filePath, test.expected, result)
			}
		})
	}
}
