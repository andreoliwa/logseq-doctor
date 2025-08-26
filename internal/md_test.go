package internal_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/testutils"
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
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			graph := testutils.StubGraph(t, "")

			opts := &internal.InsertMarkdownOptions{
				Graph:      graph,
				Date:       testCase.date,
				Content:    testCase.content,
				ParentText: testCase.parentText,
			}
			err := internal.InsertMarkdownToJournal(opts)
			require.NoError(t, err)

			if testCase.expectedGolden != "" {
				testutils.AssertGoldenJournals(t, graph, "", []string{testCase.expectedGolden})
			}
		})
	}
}
