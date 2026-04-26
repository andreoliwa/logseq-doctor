package backlog_test

import (
	"testing"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindFirstSectionDivider_FindsSection(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	page, err := graph.OpenPage("focus-with-sections")
	require.NoError(t, err)

	divider := backlog.FindFirstSectionDivider(page)

	require.NotNil(t, divider, "should find a section divider")
	// The divider should contain one of the known section text substrings
	text := logseqext.BlockContentText(divider)
	assert.Contains(t, text, "New tasks")
}

func TestFindFirstSectionDivider_NoDivider(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	// bk.md has only block refs, no section headers
	page, err := graph.OpenPage("bk")
	require.NoError(t, err)

	divider := backlog.FindFirstSectionDivider(page)

	assert.Nil(t, divider, "page with no section headers should return nil")
}

func TestAddBlockRefToFocusPage_NoSectionDivider(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	initialPage, err := graph.OpenPage("bk")
	require.NoError(t, err)

	initialCount := len(initialPage.Blocks())

	err = backlog.AddBlockRefToFocusPage(transaction, "bk", "new-uuid")
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("bk")
	require.NoError(t, err)

	assert.Len(t, page.Blocks(), initialCount+1, "one block should be appended")
}

func TestAddBlockRefToFocusPage_InsertsBeforeSectionDivider(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	err := backlog.AddBlockRefToFocusPage(transaction, "focus-with-sections", "inserted-uuid")
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("focus-with-sections")
	require.NoError(t, err)

	divider := backlog.FindFirstSectionDivider(page)
	require.NotNil(t, divider)

	blocks := page.Blocks()
	dividerIdx := -1

	for idx, block := range blocks {
		if block == divider {
			dividerIdx = idx

			break
		}
	}

	require.NotEqual(t, -1, dividerIdx)

	// The block just before the divider should exist (our inserted block)
	assert.Positive(t, dividerIdx, "divider should not be the first block")
}

func TestMoveBlockRefToTriagedSection_ExistingSection(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	// backlog-with-triaged-child.md has a Triaged section divider with one child
	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-with-triaged-child", "new-triaged-uuid",
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-with-triaged-child")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock, "Triaged section should exist")
	assert.Len(t, triagedBlock.Blocks(), 2) // existing child + newly added
}

func TestMoveBlockRefToTriagedSection_CreatesSectionBeforeScheduled(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-no-someday", "new-triaged-uuid",
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-no-someday")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock, "Triaged section should be created")
	assert.Len(t, triagedBlock.Blocks(), 1)

	scheduledDivider := logseqext.FindBlockContainingText(page, backlog.HeaderScheduled.Label)

	if triagedBlock != nil && scheduledDivider != nil {
		blocks := page.Blocks()
		triagedIdx, scheduledIdx := -1, -1

		for idx, block := range blocks {
			if block == triagedBlock {
				triagedIdx = idx
			}

			if block == scheduledDivider {
				scheduledIdx = idx
			}
		}

		assert.Less(t, triagedIdx, scheduledIdx, "Triaged should appear before Scheduled")
	}
}

// uuidRegularTask is the UUID of the block ref in the regular area of backlog-with-regular-task.md.
const uuidRegularTask = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"

// uuidAlreadyTriaged is the UUID already under Triaged in backlog-triaged-only.md.
const uuidAlreadyTriaged = "a5a5a5a5-a5a5-a5a5-a5a5-a5a5a5a5a5a5"

// uuidDupTask is the UUID that appears in both regular area and Triaged in backlog-duplicate-task.md.
const uuidDupTask = "a6a6a6a6-a6a6-a6a6-a6a6-a6a6a6a6a6a6"

// uuidTriagedNested is a block ref nested 2 levels under Triaged in backlog-triaged-nested.md.
// It is not a direct child of Triaged, so naive direct-children logic would miss it.
const uuidTriagedNested = "c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2"

// uuidRegularAfterNewTasks is the UUID of the top-level block ref that appears after the New tasks
// divider in backlog-regular-after-newtasks.md. It should still be considered part of the regular area.
const uuidRegularAfterNewTasks = "b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b3"

// uuidUnderNewTasks is the UUID of the block ref that is a child of the New tasks divider.
// It should NOT be removed from the regular area.
const uuidUnderNewTasks = "b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2"

// uuidNestedInSection is a block ref nested 2 levels under a non-divider named section
// ("Features > Outline") in backlog-nested-in-section.md.
// It should be removed from that nested position when moved to Triaged.
const uuidNestedInSection = "e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1"

func TestMoveBlockRefToTriagedSection_MovesFromRegularArea(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	// uuidRegularTask exists in regular area of backlog-with-regular-task
	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-with-regular-task", uuidRegularTask,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-with-regular-task")
	require.NoError(t, err)

	// Should be in Triaged
	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock)
	uuidsInTriaged := collectBlockRefUUIDs(triagedBlock)
	assert.Contains(t, uuidsInTriaged, uuidRegularTask, "task should be moved to Triaged")

	// Should NOT be in regular area anymore
	uuidsInRegular := collectRegularAreaBlockRefUUIDs(t, page)
	assert.NotContains(t, uuidsInRegular, uuidRegularTask, "task should be removed from regular area")
}

func TestMoveBlockRefToTriagedSection_AlreadyInTriaged_Idempotent(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-triaged-only", uuidAlreadyTriaged,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-triaged-only")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock)

	// Should appear exactly once
	uuidsInTriaged := collectBlockRefUUIDs(triagedBlock)
	count := 0

	for _, u := range uuidsInTriaged {
		if u == uuidAlreadyTriaged {
			count++
		}
	}

	assert.Equal(t, 1, count, "task should appear exactly once in Triaged")
}

func TestMoveBlockRefToTriagedSection_InBothAreas_RemovesFromRegular(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-duplicate-task", uuidDupTask,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-duplicate-task")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock)

	uuidsInTriaged := collectBlockRefUUIDs(triagedBlock)
	count := 0

	for _, u := range uuidsInTriaged {
		if u == uuidDupTask {
			count++
		}
	}

	assert.Equal(t, 1, count, "should appear exactly once in Triaged")

	uuidsInRegular := collectRegularAreaBlockRefUUIDs(t, page)
	assert.NotContains(t, uuidsInRegular, uuidDupTask, "should be removed from regular area")
}

// collectBlockRefUUIDs returns all block-ref UUIDs that are direct children of block.
func collectBlockRefUUIDs(block *content.Block) []string {
	var uuids []string

	for _, child := range block.Blocks() {
		uuid := logseqext.ExtractBlockRefUUID(child)
		if uuid != "" {
			uuids = append(uuids, uuid)
		}
	}

	return uuids
}

func TestMoveBlockRefToTriagedSection_NestedInTriaged_Idempotent(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	// uuidTriagedNested is nested 2 levels under Triaged (not a direct child).
	// It should be recognized as already in Triaged and not added again.
	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-triaged-nested", uuidTriagedNested,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-triaged-nested")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock)

	// Count occurrences of the nested UUID across all descendants of Triaged.
	count := 0

	triagedBlock.Children().FindDeep(func(node content.Node) bool {
		if blockRef, ok := node.(*content.BlockRef); ok {
			if blockRef.ID == uuidTriagedNested {
				count++
			}
		}

		return false
	})

	assert.Equal(t, 1, count, "nested UUID should appear exactly once in Triaged (not duplicated)")
}

func TestMoveBlockRefToTriagedSection_RegularAreaAfterNewTasks(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	// uuidRegularAfterNewTasks is a top-level block ref that appears AFTER the New tasks divider.
	// It should still be considered part of the regular area and be moved to Triaged.
	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-regular-after-newtasks", uuidRegularAfterNewTasks,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-regular-after-newtasks")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock)

	uuidsInTriaged := collectBlockRefUUIDs(triagedBlock)
	assert.Contains(t, uuidsInTriaged, uuidRegularAfterNewTasks, "task after New tasks divider should be moved to Triaged")

	// The block should no longer exist in the regular area
	uuidsInRegular := collectRegularAreaBlockRefUUIDs(t, page)
	assert.NotContains(t, uuidsInRegular, uuidRegularAfterNewTasks, "task should be removed from regular area")
}

func TestMoveBlockRefToTriagedSection_DoesNotRemoveFromNewTasksChildren(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	// uuidUnderNewTasks is a child of the New tasks divider, not in the regular area.
	// MoveBlockRefToTriagedSection should not remove it from there.
	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-regular-after-newtasks", uuidUnderNewTasks,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-regular-after-newtasks")
	require.NoError(t, err)

	newTasksBlock := logseqext.FindBlockContainingText(page, backlog.HeaderNewTasks.Label)
	require.NotNil(t, newTasksBlock, "New tasks section should still exist")

	uuidsInNewTasks := collectBlockRefUUIDs(newTasksBlock)
	assert.Contains(t, uuidsInNewTasks, uuidUnderNewTasks, "task under New tasks should NOT be removed")
}

func TestMoveBlockRefToTriagedSection_MovesFromNestedNamedSection(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	// uuidNestedInSection is nested 2 levels under a non-divider section ("Features > Outline").
	// It should be removed from that nested position and added to Triaged.
	err := backlog.MoveBlockRefToTriagedSection(
		transaction, "backlog-nested-in-section", uuidNestedInSection,
		backlog.HeaderTriaged.Label, backlog.HeaderScheduled.Label,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-nested-in-section")
	require.NoError(t, err)

	triagedBlock := logseqext.FindBlockContainingText(page, backlog.HeaderTriaged.Label)
	require.NotNil(t, triagedBlock, "Triaged section should be created")

	uuidsInTriaged := collectBlockRefUUIDs(triagedBlock)
	assert.Contains(t, uuidsInTriaged, uuidNestedInSection, "task should be moved to Triaged")

	// The block ref should no longer exist under the nested named section
	page.Blocks()[0].Children().FindDeep(func(node content.Node) bool {
		if blockRef, ok := node.(*content.BlockRef); ok {
			assert.NotEqual(t, uuidNestedInSection, blockRef.ID, "task should be removed from nested section")
		}

		return false
	})
}

// uuidScheduledNoDate is the task in backlog-scheduled-no-date.md — it is listed under the
// Scheduled divider but carries no scheduled date in the task data.
const uuidScheduledNoDate = "aaaa0001-0000-0000-0000-000000000001"

// TestProcessOne_UnscheduledTaskMovedFromScheduledDivider verifies that a task sitting under
// the Scheduled divider that no longer has a future scheduled date is moved to New tasks.
func TestProcessOne_UnscheduledTaskMovedFromScheduledDivider(t *testing.T) {
	back := testutils.StubBacklog(t, "bk", "", &testutils.StubAPIResponses{})

	tasks := logseqapi.NewCategorizedTasks()
	tasks.All.Add(uuidScheduledNoDate)
	tasks.TaskLookup[uuidScheduledNoDate] = logseqapi.TaskJSON{
		UUID:   uuidScheduledNoDate,
		Marker: "TODO",
	}

	_, err := back.ProcessOne("backlog-scheduled-no-date", func() (*logseqapi.CategorizedTasks, error) {
		return &tasks, nil
	})
	require.NoError(t, err)

	page, err := back.Graph().OpenPage("backlog-scheduled-no-date")
	require.NoError(t, err)

	newTasksBlock := logseqext.FindBlockContainingText(page, backlog.HeaderNewTasks.Label)
	require.NotNil(t, newTasksBlock, "New tasks divider should be created")

	uuidsInNew := collectBlockRefUUIDs(newTasksBlock)
	assert.Contains(t, uuidsInNew, uuidScheduledNoDate, "task should be moved to New tasks")

	scheduledBlock := logseqext.FindBlockContainingText(page, backlog.HeaderScheduled.Label)
	if scheduledBlock != nil {
		uuidsInScheduled := collectBlockRefUUIDs(scheduledBlock)
		assert.NotContains(t, uuidsInScheduled, uuidScheduledNoDate, "task should not remain in Scheduled")
	}
}

// collectRegularAreaBlockRefUUIDs returns all block-ref UUIDs in the regular area.
// The regular area is any top-level block ref that is NOT a section divider itself.
// (Section dividers are top-level blocks whose text contains a known section header.)
func collectRegularAreaBlockRefUUIDs(t *testing.T, page logseq.Page) []string {
	t.Helper()

	sectionHeaders := []backlog.Header{
		backlog.HeaderFocus, backlog.HeaderOverdue, backlog.HeaderNewTasks,
		backlog.HeaderTriaged, backlog.HeaderScheduled,
	}

	var uuids []string

	for _, block := range page.Blocks() {
		text := logseqext.BlockContentText(block)

		isDivider := false

		for _, h := range sectionHeaders {
			if h.Matches(text) {
				isDivider = true

				break
			}
		}

		if isDivider {
			continue
		}

		uuid := logseqext.ExtractBlockRefUUID(block)
		if uuid != "" {
			uuids = append(uuids, uuid)
		}
	}

	return uuids
}
