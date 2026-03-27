package backlog

import (
	"fmt"
	"strings"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/fatih/color"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/pkg/set"
)

// pageState holds mutable state accumulated while scanning a backlog page.
type pageState struct {
	firstBlock       *content.Block
	dividerNewTasks  *content.Block
	dividerOverdue   *content.Block
	dividerFocus     *content.Block
	dividerScheduled *content.Block
	dividerSomeday   *content.Block

	deletedCount        int
	movedCount          int
	movedScheduledCount int
	unpinnedCount       int

	result          *Result
	pinnedBlockRefs *set.Set[string]
}

func newPageState() *pageState {
	return &pageState{ //nolint:exhaustruct // zero values for all pointer/int fields are correct defaults
		result:          &Result{FocusRefsFromPage: set.NewSet[string](), ShowQuickCapture: false},
		pinnedBlockRefs: set.NewSet[string](),
	}
}

func insertAndRemoveRefs(
	graph *logseq.Graph, pageTitle string, newBlockRefs, obsoleteBlockRefs,
	overdueBlockRefs, futureScheduledBlockRefs *set.Set[string],
) (*Result, error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	state := newPageState()

	scanPageBlocks(page, state, obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs)
	insertOverdueTasks(page, state, overdueBlockRefs)

	save := insertNewTasks(page, state, newBlockRefs, overdueBlockRefs, futureScheduledBlockRefs)
	insertScheduledTasks(page, state, futureScheduledBlockRefs)

	save = logseqext.RemoveEmptyBlocks(save, state.dividerNewTasks, state.dividerOverdue, state.dividerScheduled)
	save = reportCounts(state, save)

	if save {
		err = transaction.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save transaction: %w", err)
		}
	} else {
		color.Yellow(" no changes")
	}

	if state.dividerFocus == nil {
		state.result.FocusRefsFromPage.Clear()
	}

	return state.result, nil
}

// scanPageBlocks iterates all blocks on the page, records section dividers,
// and removes or unpins block refs that are obsolete, overdue, or scheduled.
func scanPageBlocks(
	page logseq.Page, state *pageState,
	obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs *set.Set[string],
) {
	for i, block := range page.Blocks() {
		if i == 0 {
			state.firstBlock = block
		}

		block.Children().FindDeep(func(node content.Node) bool {
			if text, ok := node.(*content.Text); ok {
				recordSectionDivider(block, text.Value, state)
			}

			if blockRef, ok := node.(*content.BlockRef); ok {
				processBlockRef(node, blockRef, block, state, obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs)
			}

			return false
		})
	}
}

// recordSectionDivider updates state with a block if its text matches a known section header.
func recordSectionDivider(block *content.Block, textValue string, state *pageState) {
	switch {
	case strings.Contains(textValue, sectionNewTasksText):
		state.dividerNewTasks = block
	case strings.Contains(textValue, SectionOverdue):
		state.dividerOverdue = block
	case textValue == SectionFocus:
		state.dividerFocus = block
	case strings.Contains(textValue, SectionScheduled):
		state.dividerScheduled = block
	case strings.Contains(textValue, SectionSomeday):
		state.dividerSomeday = block
	}
}

// processBlockRef decides whether to delete, pin, unpin, or keep a block ref.
//
//nolint:cyclop // complexity comes from the inherent number of cases, not poor structure
func processBlockRef(
	node content.Node, blockRef *content.BlockRef, block *content.Block,
	state *pageState,
	obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs *set.Set[string],
) {
	shouldDelete := false
	underSomeday := state.dividerSomeday != nil && internal.IsAncestor(block, state.dividerSomeday)

	switch {
	case obsoleteBlockRefs.Contains(blockRef.ID) && !underSomeday:
		shouldDelete = true
		state.deletedCount++

	case overdueBlockRefs.Contains(blockRef.ID):
		if nextChildHasPin(node) {
			state.pinnedBlockRefs.Add(blockRef.ID)
		} else {
			shouldDelete = true
			state.movedCount++
		}

	case futureScheduledBlockRefs.Contains(blockRef.ID):
		shouldDelete = true

		if state.dividerScheduled == nil || !internal.IsAncestor(block, state.dividerScheduled) {
			state.movedScheduledCount++
		}

	default:
		// Existing non-overdue task: remove the pin marker if present.
		if nextChildHasPin(node) {
			nextChild := node.NextSibling()

			if nextChild != nil {
				nextChild.RemoveSelf()

				state.unpinnedCount++
			}
		}
	}

	if shouldDelete {
		// Block ref's parents are: paragraph and block
		// TODO: handle cases when the block ref is nested under another block ref.
		//  This will remove the obsolete block and its children.
		//  Should I show a warning message to the user and prevent the block from being deleted?
		blockRef.Parent().Parent().RemoveSelf()
	} else if state.dividerFocus == nil {
		// Keep adding tasks to the focus section until the divider is found.
		state.result.FocusRefsFromPage.Add(blockRef.ID)
	}
}

// insertOverdueTasks inserts overdue task refs under the overdue divider (creating it if needed).
//
//	Sections order: Focus / Overdue / New tasks / all other tasks / Scheduled tasks
//
// Overdue tasks go after the focus section and before new ones so the user
// can manually decide which overdue tasks deserve focus.
func insertOverdueTasks(page logseq.Page, state *pageState, overdueBlockRefs *set.Set[string]) {
	for _, blockRef := range overdueBlockRefs.ValuesSorted() {
		if state.pinnedBlockRefs.Contains(blockRef) {
			continue
		}

		if state.dividerOverdue == nil {
			state.dividerOverdue = content.NewBlock(content.NewParagraph(
				content.NewText(SectionOverdue+" "),
				content.NewPageLink(PageQuickCapture),
			))
			logseqext.AddSibling(page, state.dividerOverdue, state.firstBlock, state.dividerFocus)
		}

		overdueTask := content.NewBlock(content.NewParagraph(
			content.NewBlockRef(blockRef),
			content.NewText("📅📌"),
		))
		state.dividerOverdue.AddChild(overdueTask)
	}
}

// insertNewTasks inserts new task refs under the new-tasks divider (creating it if needed).
// Returns updated save flag.
func insertNewTasks(
	page logseq.Page, state *pageState,
	newBlockRefs, overdueBlockRefs, futureScheduledBlockRefs *set.Set[string],
) bool {
	if newBlockRefs.Size() == 0 {
		return false
	}

	for _, blockRef := range newBlockRefs.ValuesSorted() {
		if overdueBlockRefs.Contains(blockRef) {
			// Don't add overdue tasks again as new tasks.
			continue
		}

		if futureScheduledBlockRefs.Contains(blockRef) {
			// Don't add future scheduled tasks as new tasks but count them as moved.
			state.movedScheduledCount++

			continue
		}

		if state.dividerNewTasks == nil {
			state.dividerNewTasks = content.NewBlock(content.NewParagraph(
				content.NewText(SectionNewTasks+" "),
				content.NewPageLink(PageQuickCapture),
			))
			logseqext.AddSibling(page, state.dividerNewTasks, state.firstBlock, state.dividerOverdue, state.dividerFocus)
		}

		state.dividerNewTasks.AddChild(content.NewBlock(content.NewBlockRef(blockRef)))
	}

	color.Green(" %s", FormatCount(newBlockRefs.Size(), "new task", "new tasks"))

	state.result.ShowQuickCapture = true

	return true
}

// insertScheduledTasks moves future-scheduled tasks to the bottom of the page.
func insertScheduledTasks(page logseq.Page, state *pageState, futureScheduledBlockRefs *set.Set[string]) {
	if futureScheduledBlockRefs.Size() == 0 {
		return
	}

	for _, blockRef := range futureScheduledBlockRefs.ValuesSorted() {
		if state.dividerScheduled == nil {
			state.dividerScheduled = content.NewBlock(content.NewParagraph(
				content.NewText(SectionScheduled+" "),
				content.NewPageLink(PageQuickCapture),
			))
			page.AddBlock(state.dividerScheduled)
		}

		state.dividerScheduled.AddChild(content.NewBlock(content.NewBlockRef(blockRef)))
	}
}

// reportCounts prints colored summaries and returns updated save flag.
func reportCounts(state *pageState, save bool) bool {
	if state.deletedCount > 0 {
		color.Red(" %s removed", FormatCount(state.deletedCount, "task was", "tasks were"))

		save = true
	}

	if state.movedCount > 0 {
		color.Magenta(" %s moved around", FormatCount(state.movedCount, "task was", "tasks were"))

		save = true
		state.result.ShowQuickCapture = true
	}

	if state.movedScheduledCount > 0 {
		color.Blue(" %s moved to scheduled tasks", FormatCount(state.movedScheduledCount, "task was", "tasks were"))

		save = true
	}

	if state.unpinnedCount > 0 {
		color.Cyan(" %s unpinned", FormatCount(state.unpinnedCount, "task was", "tasks were"))

		save = true
	}

	return save
}

func nextChildHasPin(node content.Node) bool {
	nextChild := node.NextSibling()
	if nextChild != nil {
		if text, ok := nextChild.(*content.Text); ok {
			return strings.Contains(text.Value, "📌")
		}
	}

	return false
}

// focusSectionTexts are text substrings that identify section dividers on the Focus page.
//
//nolint:gochecknoglobals // constant lookup table for section detection
var focusSectionTexts = []string{SectionOverdue, SectionNewTasks, SectionSomeday, SectionScheduled}

// AddBlockRefToFocusPage adds a block ref ((uuid)) to the Focus page.
func AddBlockRefToFocusPage(transaction *logseq.Transaction, focusPageTitle, uuid string) error {
	page, err := transaction.OpenPage(focusPageTitle)
	if err != nil {
		return fmt.Errorf("failed to open Focus page: %w", err)
	}

	ref := content.NewBlock(content.NewParagraph(content.NewBlockRef(uuid)))
	insertBefore := FindFirstSectionDivider(page)

	if insertBefore != nil {
		page.InsertBlockBefore(ref, insertBefore)
	} else {
		page.AddBlock(ref)
	}

	return nil
}

// FindFirstSectionDivider finds the first block whose text contains a known section header.
func FindFirstSectionDivider(page logseq.Page) *content.Block {
	for _, block := range page.Blocks() {
		blockText := logseqext.BlockContentText(block)

		for _, sectionText := range focusSectionTexts {
			if strings.Contains(blockText, sectionText) {
				return block
			}
		}
	}

	return nil
}

// AddBlockRefToSomedaySection adds a block ref to the Someday section of a backlog page.
// Creates the section if it doesn't exist.
func AddBlockRefToSomedaySection(
	transaction *logseq.Transaction, backlogPage, uuid, somedayText, scheduledText string,
) error {
	page, err := transaction.OpenPage(backlogPage)
	if err != nil {
		return fmt.Errorf("failed to open backlog page %s: %w", backlogPage, err)
	}

	somedayBlock := logseqext.FindBlockContainingText(page, somedayText)

	if somedayBlock == nil {
		return createSomedaySectionWithRef(page, uuid, somedayText, scheduledText)
	}

	ref := content.NewBlock(content.NewParagraph(content.NewBlockRef(uuid)))
	somedayBlock.AddChild(ref)

	return nil
}

// createSomedaySectionWithRef creates a new Someday section with a block reference.
// It inserts the section before the Scheduled section if found, or appends to the end.
func createSomedaySectionWithRef(
	page logseq.Page, uuid, somedayText, scheduledText string,
) error {
	scheduledBlock := logseqext.FindBlockContainingText(page, scheduledText)

	somedayDivider := content.NewBlock(content.NewParagraph(content.NewText(somedayText)))
	ref := content.NewBlock(content.NewParagraph(content.NewBlockRef(uuid)))
	somedayDivider.AddChild(ref)

	if scheduledBlock != nil {
		page.InsertBlockBefore(somedayDivider, scheduledBlock)
	} else {
		page.AddBlock(somedayDivider)
	}

	return nil
}
