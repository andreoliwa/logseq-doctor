package internal_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestInsertMarkdownToJournal(t *testing.T) { //nolint:funlen
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
			err := internal.InsertMarkdownToJournal(opts)
			require.NoError(t, err)

			if test.expectedGolden != "" {
				testutils.AssertGoldenJournals(t, graph, "", []string{test.expectedGolden})
			}
		})
	}
}
