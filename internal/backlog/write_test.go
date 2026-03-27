package backlog_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
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

func TestAddBlockRefToSomedaySection_ExistingSection(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	err := backlog.AddBlockRefToSomedaySection(
		transaction, "backlog-someday-test", "new-someday-uuid",
		backlog.SectionSomeday, backlog.SectionScheduled,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-someday-test")
	require.NoError(t, err)

	somedayBlock := logseqext.FindBlockContainingText(page, backlog.SectionSomeday)
	require.NotNil(t, somedayBlock)
	assert.Len(t, somedayBlock.Blocks(), 2, "should have original + new child")
}

func TestAddBlockRefToSomedaySection_CreatesSectionBeforeScheduled(t *testing.T) {
	//nolint:staticcheck
	graph := testutils.StubGraph(t, "")
	transaction := graph.NewTransaction()

	err := backlog.AddBlockRefToSomedaySection(
		transaction, "backlog-no-someday", "new-someday-uuid",
		backlog.SectionSomeday, backlog.SectionScheduled,
	)
	require.NoError(t, err)

	err = transaction.Save()
	require.NoError(t, err)

	page, err := graph.OpenPage("backlog-no-someday")
	require.NoError(t, err)

	somedayBlock := logseqext.FindBlockContainingText(page, backlog.SectionSomeday)
	require.NotNil(t, somedayBlock, "Someday section should be created")
	assert.Len(t, somedayBlock.Blocks(), 1)

	// Verify Someday appears before Scheduled in the page
	somedayDivider := logseqext.FindBlockContainingText(page, backlog.SectionSomeday)
	scheduledDivider := logseqext.FindBlockContainingText(page, backlog.SectionScheduled)

	if somedayDivider != nil && scheduledDivider != nil {
		blocks := page.Blocks()
		somedayIdx, scheduledIdx := -1, -1

		for idx, block := range blocks {
			if block == somedayDivider {
				somedayIdx = idx
			}

			if block == scheduledDivider {
				scheduledIdx = idx
			}
		}

		assert.Less(t, somedayIdx, scheduledIdx, "Someday should appear before Scheduled")
	}
}
