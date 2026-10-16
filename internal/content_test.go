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
		graph := testutils.NewStubGraph(t, "stub-graph")
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
			graph := testutils.NewStubGraph(t, "stub-graph")
			date := time.Date(2024, 12, test.day, 0, 0, 0, 0, time.UTC)

			_, err := internal.AppendRawMarkdownToJournal(graph, date, string(contentToAppend))
			require.NoError(t, err)

			testutils.AssertGoldenJournals(t, graph, "stub-graph", []string{test.expectedFilename})
		})
	}
}

// These fixtures catch append-only placement, byte rewriting and subtree flattening.
func TestRawChronologicalBatches(t *testing.T) {
	const ref = "((12345678-1234-1234-1234-123456789abc))"
	for _, testCase := range []struct{ name, before, incoming, want string }{
		{
			"interleave",
			"- 10:00 Existing\n",
			"- **9h00** Early\n- 11:00 Late\n",
			"- **9h00** Early\n- 10:00 Existing\n- 11:00 Late\n",
		},
		{
			"equal arrival",
			"- 10:00 Existing\n- 11:00 Later\n",
			"- 10:00 First\n- 10:00 Second\n",
			"- 10:00 Existing\n- 10:00 First\n- 10:00 Second\n- 11:00 Later\n",
		},
		{"disorder", "- 11:00 Later\n- 09:00 Earlier\n", "- 10:00 New\n", "- 10:00 New\n- 11:00 Later\n- 09:00 Earlier\n"},
		{
			"evolving",
			"- 12:00 Existing\n",
			"- 11:00 First\n- 09:00 Second\n- 11:00 Third\n",
			"- 09:00 Second\n- 11:00 First\n- 11:00 Third\n- 12:00 Existing\n",
		},
		{
			"subtree",
			"- 10:00 Existing [literal] [link](https://example.com)\r\n",
			"- *9h00* Early\r\n  id:: abc\r\n  supporting [text]\r\n\t- 12:00 Child\r\n",
			"- *9h00* Early\r\n  id:: abc\r\n  supporting [text]\r\n\t- 12:00 Child\r\n" +
				"- 10:00 Existing [literal] [link](https://example.com)\r\n",
		},
		{"untimed group", "- 10:00 Existing\n", "- Group\n  - 09:00 Child\n", "- 10:00 Existing\n- Group\n  - 09:00 Child\n"},
		{
			"ref unit",
			"- " + ref + "\n  - 12:00 First\n  - 12:20 Second\n- 12:27 Later\n",
			"- 12:10 New\n",
			"- " + ref + "\n  - 12:00 First\n  - 12:20 Second\n- 12:10 New\n- 12:27 Later\n",
		},
		{
			"fenced",
			"- 10:00 Existing\n  ```md\n  - 07:00 Not a root\n  ```\n",
			"- 09:00 New\n",
			"- 09:00 New\n- 10:00 Existing\n  ```md\n  - 07:00 Not a root\n  ```\n",
		},
		{"raw unoutlined", "- 10:00 Existing", "09:00 raw [literal]", "- 10:00 Existing\n09:00 raw [literal]"},
		{"missing final newline", "- 10:00 Existing", "- 09:00 New", "- 09:00 New\n- 10:00 Existing"},
		{"empty journal", "", "- 09:00 New", "- 09:00 New"},
		{"empty input", "- 10:00 Existing\n", "", "- 10:00 Existing\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			graph, _ := testutils.NewMarkdownGraph(t, "")
			date := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
			path, err := graph.JournalPath(date)
			require.NoError(t, err)
			// Group-readable mode verifies replacement preserves existing permissions.
			require.NoError(t, os.WriteFile(path, []byte(testCase.before), 0o640)) //nolint:gosec
			_, err = internal.AppendRawMarkdownToJournal(graph, date, testCase.incoming)
			require.NoError(t, err)
			got, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, testCase.want, string(got))

			info, err := os.Stat(path)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o640), info.Mode().Perm())
		})
	}
}

func TestRawPagePrefixAndUnoutlinedRangeFallback(t *testing.T) {
	for _, test := range []struct{ name, before, incoming, want string }{
		{
			"page properties",
			"title:: Day\n- 08:00 Before\n- 10:00 Later\n",
			"tags:: capture\n- 09:00 New\n",
			"title:: Day\ntags:: capture\n- 08:00 Before\n- 09:00 New\n- 10:00 Later\n",
		},
		{
			"unoutlined range fallback",
			"- 10:00 Existing\n",
			"Introduction\n\n- One\n\n---\n\n- Two",
			"- 10:00 Existing\nIntroduction\n\n- One\n\n---\n\n- Two",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph, _ := testutils.NewMarkdownGraph(t, "")
			date := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
			path, err := graph.JournalPath(date)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, []byte(test.before), 0o600))
			_, err = internal.AppendRawMarkdownToJournal(graph, date, test.incoming)
			require.NoError(t, err)
			got, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, test.want, string(got))
		})
	}
}

func TestRawUnsafeJoinsLeaveJournalUnchanged(t *testing.T) {
	for _, test := range []struct{ name, before, incoming string }{
		{"indented root becomes child", "- 08:00 Before\n- 10:00 Later\n", "  - 09:00 New\n"},
		{"existing root becomes child", "  - 10:00 Later\n", "- 09:00 New\n"},
		{"unsupported source ranges", "- 10:00 Existing\n", "- One\n\n---\n\n- Two"},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph, _ := testutils.NewMarkdownGraph(t, "")
			date := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
			path, err := graph.JournalPath(date)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, []byte(test.before), 0o600))
			_, err = internal.AppendRawMarkdownToJournal(graph, date, test.incoming)
			require.Error(t, err)
			got, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, test.before, string(got))
		})
	}
}
