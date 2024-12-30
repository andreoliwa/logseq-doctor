package cmd //nolint:testpackage

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
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

type resultSetupFileContents struct {
	oldContents string
	goldenPath  string
}

func setupFileContents(t *testing.T, name string) resultSetupFileContents {
	t.Helper()

	subdir := filepath.Join("graph", "pages", name+".md")
	path := filepath.Join("testdata", subdir)

	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return resultSetupFileContents{string(bytes), subdir + ".golden"}
}

func TestCheckForbiddenReferences(t *testing.T) {
	invalid := setupPage(t, "forbidden")
	assert.Equal(t, changedPage{"remove 4 forbidden references to pages/tags: Inbox, quick capture", false},
		checkForbiddenReferences(invalid))

	valid := setupPage(t, "valid")
	assert.Equal(t, changedPage{"", false}, checkForbiddenReferences(valid))
}

func TestCheckRunningTasks(t *testing.T) {
	invalid := setupPage(t, "running")
	assert.Equal(t, changedPage{"stop 2 running task(s): DOING, IN-PROGRESS", false}, checkRunningTasks(invalid))

	valid := setupPage(t, "valid")
	assert.Equal(t, changedPage{"", false}, checkRunningTasks(valid))
}

func TestRemoveDoubleSpaces(t *testing.T) {
	invalid := setupPage(t, "spaces")
	assert.Equal(t,
		changedPage{"fixed 4 double spaces: 'Link   With  Spaces  ', 'Regular   text with  spaces'," +
			" 'some  page   title  with  spaces', 'some  tag with   spaces'", true},
		removeDoubleSpaces(invalid))

	// TODO: compare the saved Markdown file with spaces.md.golden.
	//  I tested manually and it works but I need to do something like:
	// actual := os.ReadFile(invalid.Path())
	// expected := os.ReadFile(invalid.Path() + ".golden")
	// assert.Equal(t, expected, actual)

	valid := setupPage(t, "valid")
	assert.Equal(t, changedPage{"", false}, removeDoubleSpaces(valid))
}

func TestRemoveUnnecessaryBracketsFromTags(t *testing.T) {
	invalid := setupFileContents(t, "tag-brackets")
	changed := removeUnnecessaryBracketsFromTags(invalid.oldContents)
	assert.Equal(t, "removed unnecessary brackets from tags", changed.msg)
	golden.Assert(t, changed.newContents, invalid.goldenPath)

	valid := setupFileContents(t, "valid")
	assert.Equal(t, changedContents{"", ""}, removeUnnecessaryBracketsFromTags(valid.oldContents))
}
