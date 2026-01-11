package internal_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestInsertMarkdownToJournal(t *testing.T) {
	tests := []struct {
		name           string
		date           time.Time
		content        string
		parentText     string
		expectedGolden string
	}{
		{
			name:           "empty content should be a no-op",
			date:           time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			content:        "",
			parentText:     "",
			expectedGolden: "", // No golden file check for empty content
		},
		{
			name:           "insert content with parent text found in simple text",
			date:           time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			content:        "Line 1\nLine 2\nLine 3",
			parentText:     "block",
			expectedGolden: "2025_01_01",
		},
		{
			name:           "insert content with parent text found in a page link",
			date:           time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			content:        "Line 1\nLine 2\nLine 3",
			parentText:     "page",
			expectedGolden: "2025_01_01",
		},
		{
			name:           "insert content with parent text found in a tag",
			date:           time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			content:        "Line 1\nLine 2\nLine 3",
			parentText:     "tag",
			expectedGolden: "2025_01_01",
		},
		{
			name:           "insert content when parent text not found",
			date:           time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			content:        "Line 1\nLine 2\nLine 3",
			parentText:     "header", // This text doesn't exist in the journal
			expectedGolden: "2025_01_02",
		},
		{
			name:           "insert multiline task with property and logbook",
			date:           time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			content:        "DOING 21:12 Some task\ncollapsed:: true\n:LOGBOOK:\nCLOCK: [2025-08-27 Wed 21:12:50]\n:END:",
			parentText:     "work",
			expectedGolden: "2025_01_03",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.StubGraph(t, "")

			opts := &internal.InsertMarkdownOptions{
				Graph:      graph,
				Date:       test.date,
				Content:    test.content,
				ParentText: test.parentText,
			}
			err := internal.InsertMarkdown(opts)
			require.NoError(t, err)

			if test.expectedGolden != "" {
				testutils.AssertGoldenJournals(t, graph, "", []string{test.expectedGolden})
			}
		})
	}
}

func TestInsertMarkdownToPage(t *testing.T) {
	tests := []struct {
		name           string
		page           string
		content        string
		parentText     string
		key            string
		expectedGolden string
	}{
		{
			name:           "key provided, and it doesn't exist",
			page:           "md-key-not-found",
			content:        "New markdown content",
			parentText:     "",
			key:            "unique-key",
			expectedGolden: "md-key-not-found",
		},
		{
			name:           "key provided, and block exists with children, properties and logbook",
			page:           "md-key-with-children",
			content:        "Updated content",
			parentText:     "",
			key:            "groceries",
			expectedGolden: "md-key-with-children",
		},
		{
			name:           "key provided but parent is not provided",
			page:           "md-key-entire-page",
			content:        "Updated content from anywhere",
			parentText:     "",
			key:            "groceries",
			expectedGolden: "md-key-entire-page",
		},
		{
			name:           "key and parent provided: search for key within block and its children",
			page:           "md-key-within-parent",
			content:        "Updated nested content",
			parentText:     "Parent block",
			key:            "groceries",
			expectedGolden: "md-key-within-parent",
		},
		{
			name:           "key and parent provided, block is deeply nested",
			page:           "md-key-deeply-nested",
			content:        "Updated deeply nested content",
			parentText:     "Parent block",
			key:            "groceries",
			expectedGolden: "md-key-deeply-nested",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// TODO: test: move all Markdown test files to separate subdirs
			//   - rename "stub-graph" to "graph-template", keep only "logseq" dir
			//   - create an "testdata/md" dir with "journals" and "pages" dirs
			//   - move all .md files to these dirs
			//   graph := testutils.StubGraph(t, "md")
			graph := testutils.StubGraph(t, "")

			opts := &internal.InsertMarkdownOptions{
				Graph:      graph,
				Page:       test.page,
				Content:    test.content,
				ParentText: test.parentText,
				Key:        test.key,
			}

			err := internal.InsertMarkdown(opts)
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, graph, "", []string{test.expectedGolden})
		})
	}
}
