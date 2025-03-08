package internal_test

import (
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
)

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
			result := internal.IsValidMarkdownFile(test.filePath)
			if result != test.expected {
				t.Errorf("For %q, expected %v, got %v", test.filePath, test.expected, result)
			}
		})
	}
}

func TestAppendRawMarkdownToJournal(t *testing.T) {
	graph := testutils.OpenTestGraph(t)

	now := time.Now()

	size, err := internal.AppendRawMarkdownToJournal(graph, now, "")
	require.NoError(t, err)
	assert.Equal(t, 0, size)

	contentToAppend, err := os.ReadFile(filepath.Join("testdata", "append-raw-journal.md"))
	require.NoError(t, err)

	testCase := func(day int, expectedFilename string) func(*testing.T) {
		return func(*testing.T) {
			date := time.Date(2024, 12, day, 0, 0, 0, 0, time.UTC)

			_, err = internal.AppendRawMarkdownToJournal(graph, date, string(contentToAppend))
			require.NoError(t, err)

			modifiedContents, err := os.ReadFile(filepath.Join(graph.Directory(), "journals",
				expectedFilename+".md"))
			require.NoError(t, err)
			golden.Assert(t, string(modifiedContents), filepath.Join(graph.Directory(), "journals",
				expectedFilename+".md.golden"))
		}
	}

	t.Run("Journal exists and has content", testCase(24, "2024_12_24"))
	t.Run("Journal doesn't exist", testCase(25, "2024_12_25"))
	t.Run("Journal exists but it's an empty file", testCase(26, "2024_12_25"))
}
