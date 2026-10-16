package internal_test

import (
	"fmt"
	"os"
	"strings"
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
			graph := testutils.NewStubGraph(t, "md")

			opts := &internal.InsertMarkdownOptions{
				Graph:      graph,
				Date:       test.date,
				Content:    test.content,
				ParentText: test.parentText,
			}
			err := internal.InsertMarkdown(opts)
			require.NoError(t, err)

			if test.expectedGolden != "" {
				testutils.AssertGoldenJournals(t, graph, "md", []string{test.expectedGolden})
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
			page:           "key-not-found",
			content:        "New markdown content",
			parentText:     "",
			key:            "unique-key",
			expectedGolden: "key-not-found",
		},
		{
			name:           "key provided, and block exists with children, properties and logbook",
			page:           "key-with-children",
			content:        "Updated content",
			parentText:     "",
			key:            "groceries",
			expectedGolden: "key-with-children",
		},
		{
			name:           "key provided but parent is not provided",
			page:           "key-entire-page",
			content:        "Updated content from anywhere",
			parentText:     "",
			key:            "groceries",
			expectedGolden: "key-entire-page",
		},
		{
			name:           "key and parent provided: search for key within block and its children",
			page:           "key-within-parent",
			content:        "Updated nested content",
			parentText:     "Parent block",
			key:            "groceries",
			expectedGolden: "key-within-parent",
		},
		{
			name:           "key and parent provided, block is deeply nested",
			page:           "key-deeply-nested",
			content:        "Updated deeply nested content",
			parentText:     "Parent block",
			key:            "groceries",
			expectedGolden: "key-deeply-nested",
		},
		{
			name:           "update block with multi-paragraph content preserves order",
			page:           "multi-paragraph-update",
			content:        "First paragraph\n\nSecond paragraph\n\nThird paragraph",
			parentText:     "",
			key:            "test-key",
			expectedGolden: "multi-paragraph-update",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.NewStubGraph(t, "md")

			opts := &internal.InsertMarkdownOptions{
				Graph:      graph,
				Page:       test.page,
				Content:    test.content,
				ParentText: test.parentText,
				Key:        test.key,
			}

			err := internal.InsertMarkdown(opts)
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, graph, "md", []string{test.expectedGolden})
		})
	}
}

// These fixtures catch wrapping explicit roots, flattening support content, and
// applying chronology outside the selected sibling list.
func TestMarkdownChronologicalCreation(t *testing.T) {
	for _, test := range []struct {
		name, before, input, parent, want string
	}{
		{
			name: "batch evolves with equal times and untimed roots",
			before: "- 10:00 Existing\n" +
				"- 11:00 Later\n",
			input: "- **9h00** First\n" +
				"  - 12:00 Child\n" +
				"- 10h00 Equal one\n" +
				"- *10.00* Equal two\n" +
				"- Notes\n" +
				"  - 08:00 Group child\n",
			parent: "",
			want: "- **9h00** First\n" +
				"\t- 12:00 Child\n" +
				"- 10:00 Existing\n" +
				"- 10h00 Equal one\n" +
				"- *10.00* Equal two\n" +
				"- 11:00 Later\n" +
				"- Notes\n" +
				"\t- 08:00 Group child\n",
		},
		{
			name:   "single block keeps paragraphs and child",
			before: "- 10:00 Existing\n",
			input: "9h00 Call\n" +
				"\n" +
				"Details\n" +
				"\n" +
				"- 12:00 Child",
			parent: "",
			want: "- 9h00 Call\n" +
				"\n" +
				"  Details\n" +
				"\n" +
				"\t- 12:00 Child\n" +
				"- 10:00 Existing\n",
		},
		{
			name:   "unbulleted thematic keeps single-block semantics",
			before: "- 10:00 Existing\n",
			input:  "9h00 Call\n\n- One\n\n---\n\n- Two",
			want:   "- 9h00 Call\n\n\t- One\n\t- ---\n\t- Two\n- 10:00 Existing\n",
		},
		{
			name: "parent scopes batch",
			before: "- 08:00 Outside\n" +
				"- Project\n" +
				"  - 10:00 Existing\n" +
				"- 12:00 Outside\n",
			input: "- 9h00 First\n" +
				"- 11:00 Last",
			parent: "Project",
			want: "- 08:00 Outside\n" +
				"- Project\n" +
				"\t- 9h00 First\n" +
				"\t- 10:00 Existing\n" +
				"\t- 11:00 Last\n" +
				"- 12:00 Outside\n",
		},
		{
			name:   "missing parent falls back to roots",
			before: "- 10:00 Existing\n",
			input:  "9h00 First",
			parent: "Missing",
			want: "- 9h00 First\n" +
				"- 10:00 Existing\n",
		},
		{
			name: "ref unit uses first child",
			before: "- ((12345678-1234-1234-1234-123456789abc))\n" +
				"  collapsed:: true\n" +
				"  - 12:00 First\n" +
				"  - 12:20 Second\n" +
				"- 12:27 Later\n",
			input:  "12h10 Incoming",
			parent: "",
			want: "- ((12345678-1234-1234-1234-123456789abc))\n" +
				"  collapsed:: true\n" +
				"\t- 12:00 First\n" +
				"\t- 12:20 Second\n" +
				"- 12h10 Incoming\n" +
				"- 12:27 Later\n",
		},
		{
			name:   "supplied status properties and logbook survive",
			before: "- 10:00 Existing\n",
			input: "- DOING **9h00** Work\n" +
				"  collapsed:: true\n" +
				"  :LOGBOOK:\n" +
				"  CLOCK: [2026-10-16 Fri 09:00:00]\n" +
				"  :END:\n",
			parent: "",
			want: "- DOING **9h00** Work\n" +
				"  collapsed:: true\n" +
				"  :LOGBOOK:\n" +
				"  CLOCK: [2026-10-16 Fri 09:00:00]\n" +
				"  :END:\n" +
				"- 10:00 Existing\n",
		},
		{
			name: "disorder remains and untimed appends",
			before: "- 11:00 First\n" +
				"- 09:00 Second\n",
			input: "- 10:00 Incoming\n" +
				"- Untimed",
			parent: "",
			want: "- 10:00 Incoming\n" +
				"- 11:00 First\n" +
				"- 09:00 Second\n" +
				"- Untimed\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph, path := testutils.NewMarkdownGraph(t, test.before)
			require.NoError(t, internal.InsertMarkdown(&internal.InsertMarkdownOptions{
				Graph: graph, Page: "timeline", Content: test.input, ParentText: test.parent,
			}))

			actual, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, strings.TrimSuffix(test.want, "\n"), string(actual))
		})
	}
}

func TestMarkdownTimedKeyUpdate(t *testing.T) {
	for _, test := range []struct{ name, input, want string }{
		{
			name:  "move earlier",
			input: "DOING **8h30** key",
			want: "- DOING **8h30** key\n" +
				"  id:: 12345678-1234-1234-1234-123456789abc\n" +
				"  custom:: keep\n" +
				"  :LOGBOOK:\n" +
				"  CLOCK: [2026-10-16 Fri 10:00:00]\n" +
				"  :END:\n" +
				"\t- Child\n" +
				"- 09:00 First\n" +
				"- 11:00 Last\n",
		},
		{
			name:  "move later",
			input: "DOING *12.00* key",
			want: "- 09:00 First\n" +
				"- 11:00 Last\n" +
				"- DOING *12.00* key\n" +
				"  id:: 12345678-1234-1234-1234-123456789abc\n" +
				"  custom:: keep\n" +
				"  :LOGBOOK:\n" +
				"  CLOCK: [2026-10-16 Fri 10:00:00]\n" +
				"  :END:\n" +
				"\t- Child\n",
		},
		{
			name:  "equal excludes self",
			input: "DOING 11h00 key",
			want: "- 09:00 First\n" +
				"- 11:00 Last\n" +
				"- DOING 11h00 key\n" +
				"  id:: 12345678-1234-1234-1234-123456789abc\n" +
				"  custom:: keep\n" +
				"  :LOGBOOK:\n" +
				"  CLOCK: [2026-10-16 Fri 10:00:00]\n" +
				"  :END:\n" +
				"\t- Child\n",
		},
		{
			name:  "untimed retains position",
			input: "DOING key",
			want: "- 09:00 First\n" +
				"- DOING key\n" +
				"  id:: 12345678-1234-1234-1234-123456789abc\n" +
				"  custom:: keep\n" +
				"  :LOGBOOK:\n" +
				"  CLOCK: [2026-10-16 Fri 10:00:00]\n" +
				"  :END:\n" +
				"\t- Child\n" +
				"- 11:00 Last\n",
		},
	} {
		for _, nested := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/nested=%t", test.name, nested), func(t *testing.T) {
				before := "- 09:00 First\n- DOING 10:00 key\n" +
					"  id:: 12345678-1234-1234-1234-123456789abc\n  custom:: keep\n" +
					"  :LOGBOOK:\n  CLOCK: [2026-10-16 Fri 10:00:00]\n  :END:\n  - Child\n- 11:00 Last\n"
				want := test.want

				if nested {
					before = "- Project\n\t" + strings.ReplaceAll(strings.TrimSuffix(before, "\n"), "\n", "\n\t") +
						"\n- 07:00 Outside\n"
					want = "- Project\n\t" + strings.ReplaceAll(strings.TrimSuffix(want, "\n"), "\n", "\n\t") + "\n- 07:00 Outside\n"
				}

				graph, path := testutils.NewMarkdownGraph(t, before)
				require.NoError(t, internal.InsertMarkdown(&internal.InsertMarkdownOptions{
					Graph: graph, Page: "timeline", Key: "key", Content: test.input,
				}))

				actual, err := os.ReadFile(path)
				require.NoError(t, err)
				require.Equal(t, strings.TrimSuffix(want, "\n"), string(actual))
			})
		}
	}
}

func TestMarkdownInvalidOutlineLeavesPageUnchanged(t *testing.T) {
	const original = "- 10:00 Existing\n"

	graph, path := testutils.NewMarkdownGraph(t, original)
	err := internal.InsertMarkdown(&internal.InsertMarkdownOptions{
		Graph: graph, Page: "timeline", Content: "- One\n\n---\n\n- Two",
	})
	require.Error(t, err)

	actual, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, original, string(actual))
}

func TestMarkdownWhitespaceKeyUpdateDoesNotMoveExistingTime(t *testing.T) {
	graph, path := testutils.NewMarkdownGraph(t, "- 11:00 key\n- 09:00 Earlier\n")
	require.NoError(t, internal.InsertMarkdown(&internal.InsertMarkdownOptions{
		Graph: graph, Page: "timeline", Key: "key", Content: " \n\t ",
	}))

	actual, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "- 11:00 key\n- 09:00 Earlier\n", string(actual))
}
