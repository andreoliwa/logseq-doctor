package logseqext_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockContentText(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	block := logseqext.FindBlockContainingText(page, "Single task at root level")
	require.NotNil(t, block)

	text := logseqext.BlockContentText(block)
	assert.Contains(t, text, "Single task at root level")
}

func TestBlockContentText_EmptyBlock(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("empty-bullets")
	require.NoError(t, err)

	blocks := page.Blocks()
	if len(blocks) == 0 {
		t.Skip("no blocks on page")
	}

	// The function should return empty string for a block with no text nodes
	// (or a non-empty string if the block has text — either way it shouldn't panic).
	text := logseqext.BlockContentText(blocks[0])
	assert.IsType(t, "", text)
}

func TestSetTaskCanceled(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	// Find a TODO task block
	block := logseqext.FindBlockContainingText(page, "Single task at root level")
	require.NotNil(t, block)

	err = logseqext.SetTaskCanceled(block)
	require.NoError(t, err)

	// Verify status changed: find a TaskMarker node in the block's content
	var marker *content.TaskMarker

	block.Content().FindDeep(func(node content.Node) bool {
		if m, ok := node.(*content.TaskMarker); ok {
			marker = m

			return true
		}

		return false
	})
	require.NotNil(t, marker)
	assert.Equal(t, content.TaskStatusCanceled, marker.Status)
}

func TestSetTaskCanceled_NonTaskBlock(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	// A plain text block (no task marker) — should be a no-op
	block := logseqext.FindBlockContainingText(page, "Line before 1")
	require.NotNil(t, block)

	err = logseqext.SetTaskCanceled(block)
	require.NoError(t, err) // Should not return an error
}

func TestAddSibling_AppendsWhenNoAnchors(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("empty-bullets")
	require.NoError(t, err)

	initialCount := len(page.Blocks())
	newBlock := content.NewBlock(content.NewParagraph(content.NewText("new block")))
	logseqext.AddSibling(page, newBlock, nil)

	assert.Len(t, page.Blocks(), initialCount+1)
}

func TestAddSibling_InsertsBeforeAnchor(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	before := logseqext.FindBlockContainingText(page, "Line after 2")
	require.NotNil(t, before)

	newBlock := content.NewBlock(content.NewParagraph(content.NewText("inserted before")))
	logseqext.AddSibling(page, newBlock, before)

	// Verify the new block appears right before `before` in the page blocks
	blocks := page.Blocks()

	var insertedIdx, beforeIdx int

	for idx, block := range blocks {
		text := logseqext.BlockContentText(block)

		if text == "inserted before" {
			insertedIdx = idx
		}

		if text == "Line after 2" {
			beforeIdx = idx
		}
	}

	assert.Equal(t, beforeIdx-1, insertedIdx)
}

func TestRemoveEmptyBlocks_RemovesEmpty(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	// Find a leaf block (no children)
	leafBlock := logseqext.FindBlockContainingText(page, "Single task at root level")
	require.NotNil(t, leafBlock)
	assert.Empty(t, leafBlock.Blocks(), "test prerequisite: block should have no children")

	save := logseqext.RemoveEmptyBlocks(false, leafBlock)
	assert.True(t, save)
}

func TestRemoveEmptyBlocks_SkipsNonEmpty(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	// Find a block with children
	parent := logseqext.FindBlockContainingText(page, "Parent block with nested tasks")
	require.NotNil(t, parent)
	require.NotEmpty(t, parent.Blocks(), "test prerequisite: block should have children")

	save := logseqext.RemoveEmptyBlocks(false, parent)
	assert.False(t, save) // No change — block has children
}

func TestRemoveEmptyBlocks_NilBlockSkipped(t *testing.T) {
	save := logseqext.RemoveEmptyBlocks(false, nil)
	assert.False(t, save)
}

func TestRemoveEmptyBlocks_PropagatesSaveTrue(t *testing.T) {
	// When save starts true, it should stay true even if nothing was removed.
	save := logseqext.RemoveEmptyBlocks(true, nil)
	assert.True(t, save)
}

func TestSetPriority_InsertAfterTaskMarker(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	page, err := graph.OpenPage("finder")
	require.NoError(t, err)

	// Find a TODO task block without priority
	block := logseqext.FindBlockContainingText(page, "Single task at root level")
	require.NotNil(t, block)

	err = logseqext.SetPriority(block, content.PriorityHigh)
	require.NoError(t, err)

	// Verify priority was inserted
	var priority *content.Priority

	block.Content().FindDeep(func(node content.Node) bool {
		if p, ok := node.(*content.Priority); ok {
			priority = p

			return true
		}

		return false
	})
	require.NotNil(t, priority)
	assert.Equal(t, content.PriorityHigh, priority.Priority)
}

func TestSetPriority_ReplaceExisting(t *testing.T) {
	// Build a block with an existing priority in-memory
	block := content.NewBlock(content.NewParagraph(
		content.NewTaskMarkerFromString("TODO"),
		content.NewPriority(content.PriorityHigh),
		content.NewText("Task with priority"),
	))

	err := logseqext.SetPriority(block, content.PriorityLow)
	require.NoError(t, err)

	var priority *content.Priority

	block.Content().FindDeep(func(node content.Node) bool {
		if p, ok := node.(*content.Priority); ok {
			priority = p

			return true
		}

		return false
	})
	require.NotNil(t, priority)
	assert.Equal(t, content.PriorityLow, priority.Priority)
}

func TestSetPriority_PlainBlock(t *testing.T) {
	// A block with no task marker — priority should be prepended
	block := content.NewBlock(content.NewParagraph(
		content.NewText("Plain text block"),
	))

	err := logseqext.SetPriority(block, content.PriorityMedium)
	require.NoError(t, err)

	var priority *content.Priority

	block.Content().FindDeep(func(node content.Node) bool {
		if p, ok := node.(*content.Priority); ok {
			priority = p

			return true
		}

		return false
	})
	require.NotNil(t, priority)
	assert.Equal(t, content.PriorityMedium, priority.Priority)
}

func TestParsePriorityFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected content.PriorityValue
	}{
		{"high", "TODO [#A] High priority task", content.PriorityHigh},
		{"medium", "TODO [#B] Medium priority task", content.PriorityMedium},
		{"low", "TODO [#C] Low priority task", content.PriorityLow},
		{"none", "TODO Regular task without priority", content.PriorityNone},
		{"with time", "TODO [#A] 17:51 Task with time", content.PriorityHigh},
		{"with properties", "TODO [#B] Task\nid:: abc-123\ngroomed:: [[Monday]]", content.PriorityMedium},
		{"empty", "", content.PriorityNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logseqext.ParsePriorityFromContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
