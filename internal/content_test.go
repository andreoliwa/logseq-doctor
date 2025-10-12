package internal_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
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

	err := os.WriteFile(validFilePath, []byte("# Test"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Update the file path for the valid markdown file test case
	tests[4].filePath = validFilePath

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.IsValidMarkdownFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("For %q, expected %v, got %v", tt.filePath, tt.expected, result)
			}
		})
	}
}

func TestAppendRawMarkdownToJournal(t *testing.T) {
	graph := testutils.StubGraph(t, "")

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

			testutils.AssertGoldenJournals(t, graph, "", []string{expectedFilename})
		}
	}

	t.Run("Journal exists and has content", testCase(24, "2024_12_24"))
	t.Run("Journal doesn't exist", testCase(25, "2024_12_25"))
	t.Run("Journal exists but it's an empty file", testCase(26, "2024_12_25"))
}
