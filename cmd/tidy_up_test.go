package cmd //nolint:testpackage

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
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
		{
			"unsorted with duplicates",
			[]string{"orange", "banana", "apple", "apple", "orange"},
			[]string{"apple", "banana", "orange"},
		},
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
	if err := os.WriteFile(validFilePath, []byte("# Test"), 0o600); err != nil {
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

func setupPage(t *testing.T, name string) logseq.Page {
	t.Helper()

	ctx := context.Background()

	graph, err := logseq.Open(ctx, filepath.Join("testdata", "graph"))
	if err != nil {
		t.Fatal(err)
	}

	page, err := graph.OpenPage(name)
	if err != nil {
		t.Fatal(err)
	}

	return page
}

func TestCheckForbiddenReferences(t *testing.T) {
	invalid := setupPage(t, "forbidden")
	assert.Equal(t, tidyInfo{"remove 4 forbidden references to pages/tags: Inbox, quick capture", false},
		checkForbiddenReferences(invalid))

	valid := setupPage(t, "valid")
	assert.Equal(t, tidyInfo{"", false}, checkForbiddenReferences(valid))
}

func TestCheckRunningTasks(t *testing.T) {
	invalid := setupPage(t, "running")
	assert.Equal(t, tidyInfo{"stop 2 running task(s): DOING, IN-PROGRESS", false}, checkRunningTasks(invalid))

	valid := setupPage(t, "valid")
	assert.Equal(t, tidyInfo{"", false}, checkRunningTasks(valid))
}

func TestRemoveDoubleSpaces(t *testing.T) {
	invalid := setupPage(t, "spaces")
	assert.Equal(t,
		tidyInfo{"fixed 4 double spaces: 'Link   With  Spaces  ', 'Regular   text with  spaces'," +
			" 'some  page   title  with  spaces', 'some  tag with   spaces'", true},
		removeDoubleSpaces(invalid))

	// TODO: compare the saved Markdown file with spaces.md.golden.
	//  I tested manually and it works but I need to do something like:
	// actual := os.ReadFile(invalid.Path())
	// expected := os.ReadFile(invalid.Path() + ".golden")
	// assert.Equal(t, expected, actual)

	valid := setupPage(t, "valid")
	assert.Equal(t, tidyInfo{"", false}, removeDoubleSpaces(valid))
}
