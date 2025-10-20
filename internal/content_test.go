package internal_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
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
	t.Run("empty content should be a no-op", func(t *testing.T) {
		graph := testutils.StubGraph(t, "")
		now := time.Now()

		size, err := internal.AppendRawMarkdownToJournal(graph, now, "")
		require.NoError(t, err)
		assert.Equal(t, 0, size)
	})

	contentToAppend, err := os.ReadFile(filepath.Join("testdata", "append-raw-journal.md"))
	require.NoError(t, err)

	tests := []struct {
		name             string
		day              int
		expectedFilename string
	}{
		{
			name:             "Journal exists and has content",
			day:              24,
			expectedFilename: "2024_12_24",
		},
		{
			name:             "Journal doesn't exist",
			day:              25,
			expectedFilename: "2024_12_25",
		},
		{
			name:             "Journal exists but it's an empty file",
			day:              26,
			expectedFilename: "2024_12_26",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.StubGraph(t, "")
			date := time.Date(2024, 12, test.day, 0, 0, 0, 0, time.UTC)

			_, err := internal.AppendRawMarkdownToJournal(graph, date, string(contentToAppend))
			require.NoError(t, err)

			testutils.AssertGoldenJournals(t, graph, "", []string{test.expectedFilename})
		})
	}
}
