package internal_test

import (
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindFirstQuery(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	finder := internal.NewLogseqFinder(graph)

	tests := []struct {
		name      string
		pageTitle string
		expected  string
	}{
		{
			name:      "no query",
			pageTitle: "query-none",
			expected:  "",
		},
		{
			name:      "one query",
			pageTitle: "query-one",
			expected: "{:title \"Who is using this account?\"\n  :query (property :payment-method [[query-one]])\n" +
				"  :collapsed? false}",
		},
		{
			name:      "multiple queries",
			pageTitle: "query-multiple",
			expected:  "(and (or [[home]] [[phone]]) (task TODO DOING WAITING))",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := finder.FindFirstQuery(test.pageTitle)

			assert.Equal(t, test.expected, query)
		})
	}
}

func TestFindBlockContainingText(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	tests := []struct {
		name         string
		searchText   string
		expectedText string // Expected text in the found block, empty if nil expected
	}{
		{
			name:         "search for parent block",
			searchText:   "parent block",
			expectedText: "Parent block with tasks at first level",
		},
		{
			name:         "non existent text",
			searchText:   "non existent text",
			expectedText: "",
		},
		{
			name:         "empty search text",
			searchText:   "",
			expectedText: "",
		},
		{
			name:         "search for page",
			searchText:   "page",
			expectedText: "Parent block with nested tasks",
		},
		{
			name:         "search for tag",
			searchText:   "tag",
			expectedText: "Parent block with tasks at first level",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := internal.FindBlockContainingText(page, test.searchText)

			if test.expectedText == "" {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				// Verify the block contains the expected text
				blockText := extractTextFromBlock(result)
				assert.Contains(t, blockText, test.expectedText)
			}
		})
	}
}

func TestFindBlockContainingText_EmptyPage(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	page, err := graph.OpenPage("empty-bullets")
	require.NoError(t, err)

	result := internal.FindBlockContainingText(page, "anything")
	assert.Nil(t, result)
}

func TestFindTaskMarkerByKey(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	tests := getFindTaskByKeyTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var parentBlock *content.Block
			if test.parentText != "" {
				parentBlock = internal.FindBlockContainingText(page, test.parentText)
				require.NotNil(t, parentBlock, "parent block should be found")
			}

			result := internal.FindTaskMarkerByKey(page, parentBlock, test.key)

			if test.expectedText == "" {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				// Verify the block contains the expected text
				blockText := extractTextFromBlock(result.ParentBlock())
				assert.Contains(t, blockText, test.expectedText)
			}
		})
	}
}

func getFindTaskByKeyTestCases() []struct {
	name         string
	key          string
	parentText   string
	expectedText string
} {
	return []struct {
		name         string
		key          string
		parentText   string // Text to find parent block, empty if searching entire page
		expectedText string // Expected text in the found task, empty if nil expected
	}{
		{
			name:         "search for root",
			key:          "root",
			parentText:   "",
			expectedText: "TODO Single task at root level",
		},
		{
			name:         "search for second",
			key:          "second",
			parentText:   "",
			expectedText: "DOING Second task",
		},
		{
			name:         "search for third nested",
			key:          "third nested",
			parentText:   "",
			expectedText: "DONE Third nested task",
		},
		{
			name:         "non existent",
			key:          "non existent",
			parentText:   "",
			expectedText: "",
		},
		{
			name:         "empty key",
			key:          "",
			parentText:   "",
			expectedText: "",
		},
		{
			name:         "parent block is Parent block with nested tasks and key is second",
			key:          "second",
			parentText:   "Parent block with nested tasks",
			expectedText: "DOING Second nested task",
		},
		{
			name:         "search for page",
			key:          "page",
			parentText:   "",
			expectedText: "WAITING Fourth nested task with [[page link]]",
		},
		{
			name:         "search for tag",
			key:          "tag",
			parentText:   "",
			expectedText: "WAITING Fourth task with a #tag",
		},
	}
}

func TestFindTaskByKey_EmptyPage(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	page, err := graph.OpenPage("empty-bullets")
	require.NoError(t, err)

	result := internal.FindTaskMarkerByKey(page, nil, "anything")
	assert.Nil(t, result)
}

// extractTextFromBlock extracts all text content from a block for testing purposes.
// It only extracts from the block's content (not nested blocks).
func extractTextFromBlock(block *content.Block) string {
	// TODO: use the Markdown function once it's implemented in https://github.com/aholstenson/logseq-go/issues/1
	var builder strings.Builder
	extractTextFromNodes(block.Content(), &builder)

	return builder.String()
}

func extractTextFromNodes(nodes content.NodeList, builder *strings.Builder) {
	for _, node := range nodes {
		switch nodeTyped := node.(type) {
		case *content.TaskMarker:
			builder.WriteString(taskStatusToString(nodeTyped.Status))
			builder.WriteString(" ")
		case *content.Text:
			builder.WriteString(nodeTyped.Value)
		case *content.PageLink:
			builder.WriteString("[[")
			builder.WriteString(nodeTyped.To)
			builder.WriteString("]]")
		case *content.Hashtag:
			builder.WriteString("#")
			builder.WriteString(nodeTyped.To)
		case content.HasChildren:
			extractTextFromNodes(nodeTyped.Children(), builder)
		}
	}
}

//nolint:cyclop // This is a simple mapping function with many cases
func taskStatusToString(status content.TaskStatus) string {
	// TODO: move this to logseq-go, use stringer; maybe related to https://github.com/aholstenson/logseq-go/issues/1
	switch status {
	case content.TaskStatusTodo:
		return "TODO"
	case content.TaskStatusDoing:
		return "DOING"
	case content.TaskStatusDone:
		return "DONE"
	case content.TaskStatusLater:
		return "LATER"
	case content.TaskStatusNow:
		return "NOW"
	case content.TaskStatusCancelled:
		return "CANCELLED"
	case content.TaskStatusCanceled:
		return "CANCELED"
	case content.TaskStatusInProgress:
		return "IN-PROGRESS"
	case content.TaskStatusWait:
		return "WAIT"
	case content.TaskStatusWaiting:
		return "WAITING"
	case content.TaskStatusNone:
		return ""
	default:
		return ""
	}
}
