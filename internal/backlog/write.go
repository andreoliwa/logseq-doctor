package backlog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/fatih/color"

	"github.com/andreoliwa/logseq-doctor/internal"
	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
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
	dividerTriaged   *content.Block
	dividerUnranked  *content.Block

	deletedCount        int
	movedCount          int
	movedScheduledCount int
	unpinnedCount       int

	result           *Result
	pinnedBlockRefs  *set.Set[string]
	triagedBlockRefs *set.Set[string] // UUIDs already in the Triaged section
}

func newPageState() *pageState {
	return &pageState{ //nolint:exhaustruct // zero values for all pointer/int fields are correct defaults
		result:           &Result{FocusRefsFromPage: set.NewSet[string](), ShowQuickCapture: false},
		pinnedBlockRefs:  set.NewSet[string](),
		triagedBlockRefs: set.NewSet[string](),
	}
}

func insertAndRemoveRefs(
	graph *logseq.Graph, pageTitle string, newBlockRefs, obsoleteBlockRefs,
	overdueBlockRefs, futureScheduledBlockRefs *set.Set[string],
	taskLookup map[logseqapi.TaskUUID]logseqapi.TaskJSON,
) (*Result, error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	state := newPageState()

	normalised := NormaliseHeaderText(page)
	scanPageBlocks(page, state, obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs)
	insertOverdueTasks(page, state, overdueBlockRefs)

	save := insertNewTasks(page, state, newBlockRefs, overdueBlockRefs, futureScheduledBlockRefs)
	save = save || normalised

	insertScheduledTasks(page, state, futureScheduledBlockRefs)

	sortTriagedSection(state, taskLookup)
	save = logseqext.RemoveEmptyBlocks(save,
		state.dividerNewTasks, state.dividerOverdue, state.dividerScheduled,
		state.dividerTriaged, state.dividerUnranked)
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

// NormaliseHeaderText scans all top-level blocks on the page and normalises any
// block whose text node (trimmed) is exactly a known header text (with or without
// the canonical emoji, case-insensitively) to the canonical "🎯 Focus" / "🆕 New
// tasks" form. Only the text node is updated; sibling nodes (e.g. [[quick capture]]
// page links) are left intact. Returns true if any block was changed.
//
// A block whose text is "Focus" or "🎯 FOCUS" gets normalised to "🎯 Focus ".
// A block whose text is "⏰ Scheduled tasks section already exists" is left alone
// because it is not purely a header — it has extra words after the header text.
func NormaliseHeaderText(page logseq.Page) bool {
	changed := false

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(node content.Node) bool {
			text, ok := node.(*content.Text)
			if !ok {
				return false
			}

			trimmed := strings.TrimSpace(text.Value)

			for _, hdr := range allHeaders {
				// Only match if the text node is exactly the header text (ignoring
				// emoji prefix and case), not if it merely contains the header text.
				if !strings.EqualFold(trimmed, hdr.Text) && !strings.EqualFold(trimmed, hdr.String()) {
					continue
				}

				if trimmed == hdr.String() {
					// Already canonical — nothing to do.
					break
				}

				// Preserve any trailing whitespace (separates text from page-link sibling).
				suffix := text.Value[len(trimmed):]
				text.Value = hdr.String() + suffix
				changed = true

				break
			}

			return false
		})
	}

	return changed
}

// scanPageBlocks is a two-pass coordinator: first it normalises headers, then
// collects UUIDs already in the Triaged section, then processes all blocks
// (which uses those UUIDs for deduplication in the regular area).
func scanPageBlocks(
	page logseq.Page, state *pageState,
	obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs *set.Set[string],
) {
	collectTriagedRefs(page, state)
	processAllBlocks(page, state, obsoleteBlockRefs, overdueBlockRefs, futureScheduledBlockRefs)
}

// collectTriagedRefs performs a first pass over the page to find the Triaged
// section divider and record all block-ref UUIDs that are descendants of it.
func collectTriagedRefs(page logseq.Page, state *pageState) {
	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(node content.Node) bool {
			if text, ok := node.(*content.Text); ok {
				if HeaderTriaged.Matches(text.Value) {
					state.dividerTriaged = block
				}
			}

			return false
		})

		if state.dividerTriaged != nil {
			break
		}
	}

	if state.dividerTriaged == nil {
		return
	}

	state.dividerTriaged.Children().FindDeep(func(node content.Node) bool {
		if blockRef, ok := node.(*content.BlockRef); ok {
			state.triagedBlockRefs.Add(blockRef.ID)
		}

		return false
	})
}

// processAllBlocks iterates all blocks on the page, records section dividers,
// and removes or unpins block refs that are obsolete, overdue, or scheduled.
func processAllBlocks(
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
	case HeaderNewTasks.Matches(textValue):
		state.dividerNewTasks = block
	case HeaderOverdue.Matches(textValue):
		state.dividerOverdue = block
	case HeaderFocus.Matches(textValue):
		state.dividerFocus = block
	case HeaderScheduled.Matches(textValue):
		state.dividerScheduled = block
	case HeaderTriaged.Matches(textValue):
		state.dividerTriaged = block
	case HeaderUnranked.Matches(textValue):
		state.dividerUnranked = block
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
	underTriaged := state.dividerTriaged != nil && internal.IsAncestor(block, state.dividerTriaged)
	underUnranked := state.dividerUnranked != nil && internal.IsAncestor(block, state.dividerUnranked)

	// Preserve tasks under the Unranked divider — they are intentionally unranked.
	if underUnranked {
		return
	}

	// If already in Triaged and this ref is in the regular area, remove it (deduplication).
	if state.triagedBlockRefs.Contains(blockRef.ID) && !underTriaged {
		blockRef.Parent().Parent().RemoveSelf()

		state.deletedCount++

		return
	}

	switch {
	case obsoleteBlockRefs.Contains(blockRef.ID) && !underTriaged:
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
			state.dividerOverdue = content.NewBlock(HeaderOverdue.NewParagraph())
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
			state.dividerNewTasks = content.NewBlock(HeaderNewTasks.NewParagraph())
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
			state.dividerScheduled = content.NewBlock(HeaderScheduled.NewParagraph())
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

// taskSortKey holds the fields used to sort tasks in the Triaged section.
type taskSortKey struct {
	priority    content.PriorityValue // PriorityNone=0 sorts FIRST (unprioritized at top)
	createdDate time.Time             // oldest first
	firstLine   string                // alphabetical tiebreaker
	id          logseqapi.TaskUUID    // UUID, guaranteed unique final tiebreaker
	block       *content.Block        // reference to the block for reordering
}

// sortTriagedSection sorts children of the Triaged divider by priority, date, name, and ID.
func sortTriagedSection(state *pageState, taskLookup map[logseqapi.TaskUUID]logseqapi.TaskJSON) {
	// TODO: the same triaged task can belong to multiple backlogs...
	//  It will be moved in one backlog but not in the others. How to solve this?
	if state.dividerTriaged == nil {
		return
	}

	children := state.dividerTriaged.Blocks()
	if len(children) == 0 {
		return
	}

	keys, unprioritizedCount := newSortKeys(children, taskLookup)

	// Sort: priority asc (none=0 first), date asc, name asc, id asc
	sort.SliceStable(keys, func(left, right int) bool {
		return taskSortKeyLess(keys[left], keys[right])
	})

	// Reorder child blocks without touching the paragraph (block text).
	// SetChildren would strip the paragraph, so we remove child blocks and re-add in order.
	childNodes := make([]content.Node, len(children))
	for i, c := range children {
		childNodes[i] = c
	}

	state.dividerTriaged.RemoveChildren(childNodes...)

	for _, k := range keys {
		state.dividerTriaged.AddChild(k.block)
	}

	if unprioritizedCount > 0 {
		color.Yellow(" %s in Triaged without priority",
			FormatCount(unprioritizedCount, "task", "tasks"))
	}
}

// newSortKeys creates sort keys for all children of the Triaged section.
func newSortKeys(
	children content.BlockList, taskLookup map[logseqapi.TaskUUID]logseqapi.TaskJSON,
) ([]taskSortKey, int) {
	keys := make([]taskSortKey, 0, len(children))
	unprioritizedCount := 0

	for _, child := range children {
		uuid := logseqext.ExtractBlockRefUUID(child)
		key := taskSortKey{block: child, id: uuid} //nolint:exhaustruct // zero values are correct defaults

		if task, ok := taskLookup[uuid]; ok {
			key.priority = logseqext.ParsePriorityFromContent(task.Content)
			key.createdDate = logseqext.JournalDayToTime(task.Page.JournalDay)
			key.firstLine = logseqext.ExtractFirstLine(task.Content)
		}

		if key.priority == content.PriorityNone {
			unprioritizedCount++
		}

		keys = append(keys, key)
	}

	return keys, unprioritizedCount
}

// taskSortKeyLess compares two sort keys for ordering in the Triaged section.
func taskSortKeyLess(left, right taskSortKey) bool {
	if left.priority != right.priority {
		return left.priority < right.priority
	}

	if !left.createdDate.Equal(right.createdDate) {
		return left.createdDate.Before(right.createdDate)
	}

	if left.firstLine != right.firstLine {
		return left.firstLine < right.firstLine
	}

	return left.id < right.id
}

// focusSectionHeaders are the section dividers on the Focus page.
// Used to find the insertion point for new block refs.
//
//nolint:gochecknoglobals // constant lookup table for section detection
var focusSectionHeaders = []Header{HeaderOverdue, HeaderNewTasks, HeaderTriaged, HeaderScheduled}

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

// FindFirstSectionDivider finds the first block whose text matches a known section header.
func FindFirstSectionDivider(page logseq.Page) *content.Block {
	for _, block := range page.Blocks() {
		blockText := logseqext.BlockContentText(block)

		for _, header := range focusSectionHeaders {
			if header.Matches(blockText) {
				return block
			}
		}
	}

	return nil
}

// regularAreaSectionHeaders are section headers that mark the boundary of the regular area.
// A top-level block matching any of these is a section divider, not part of the regular area.
//
//nolint:gochecknoglobals // constant lookup table for regular area detection
var regularAreaSectionHeaders = []Header{
	HeaderFocus, HeaderOverdue, HeaderNewTasks, HeaderTriaged, HeaderScheduled, HeaderUnranked,
}

// BlockRefExistsUnder returns true if a block ref with the given UUID exists
// anywhere in the descendant tree of parent.
func BlockRefExistsUnder(parent *content.Block, uuid logseqapi.TaskUUID) bool {
	found := false

	parent.Children().FindDeep(func(node content.Node) bool {
		if blockRef, ok := node.(*content.BlockRef); ok && blockRef.ID == uuid {
			found = true

			return true
		}

		return false
	})

	return found
}

// RemoveBlockRefFromRegularArea removes the block ref with the given UUID from the regular area
// of the page. The regular area is any top-level block ref that is not itself a section divider.
// Section dividers (Focus, Overdue, New tasks, Triaged, Scheduled, Unranked) and their children
// are not part of the regular area. Since we walk only top-level blocks, child refs are never seen.
func RemoveBlockRefFromRegularArea(page logseq.Page, uuid logseqapi.TaskUUID) {
	for _, block := range page.Blocks() {
		blockText := logseqext.BlockContentText(block)

		isDivider := false

		for _, header := range regularAreaSectionHeaders {
			if header.Matches(blockText) {
				isDivider = true

				break
			}
		}

		if isDivider {
			continue
		}

		if logseqext.ExtractBlockRefUUID(block) == uuid {
			block.RemoveSelf()

			return
		}

		// Also search descendants of non-divider blocks (e.g. block refs nested under
		// named sub-sections like "Features", "Bugs", "Convert from Python > Outline").
		var found *content.Block

		block.Children().FindDeep(func(node content.Node) bool {
			if b, ok := node.(*content.Block); ok {
				if logseqext.ExtractBlockRefUUID(b) == uuid {
					found = b

					return true
				}
			}

			return false
		})

		if found != nil {
			found.RemoveSelf()

			return
		}
	}
}

// MoveBlockRefToTriagedSection moves a block ref to the Triaged section of a backlog page.
// If the ref exists in the regular area (not under Focus, New tasks, Overdue, or Scheduled),
// it is removed from there. If it's already in Triaged, no duplicate is added.
// Creates the Triaged section if it doesn't exist.
func MoveBlockRefToTriagedSection(
	transaction *logseq.Transaction, backlogPage string, uuid logseqapi.TaskUUID, triagedText, scheduledText string,
) error {
	page, err := transaction.OpenPage(backlogPage)
	if err != nil {
		return fmt.Errorf("failed to open backlog page %s: %w", backlogPage, err)
	}

	triagedBlock := logseqext.FindBlockContainingText(page, triagedText)
	alreadyInTriaged := triagedBlock != nil && BlockRefExistsUnder(triagedBlock, uuid)

	// Remove from regular area if present (regardless of whether it's in Triaged).
	RemoveBlockRefFromRegularArea(page, uuid)

	if alreadyInTriaged {
		// Already in Triaged; removal from regular area is sufficient.
		return nil
	}

	if triagedBlock == nil {
		return createTriagedSectionWithRef(page, uuid, triagedText, scheduledText)
	}

	ref := content.NewBlock(content.NewParagraph(content.NewBlockRef(uuid)))
	triagedBlock.AddChild(ref)

	return nil
}

// createTriagedSectionWithRef creates a new Triaged section with a block reference.
// It inserts the section before the Scheduled section if found, or appends to the end.
func createTriagedSectionWithRef(
	page logseq.Page, uuid logseqapi.TaskUUID, triagedText, scheduledText string,
) error {
	scheduledBlock := logseqext.FindBlockContainingText(page, scheduledText)

	triagedDivider := content.NewBlock(content.NewParagraph(content.NewText(triagedText)))
	ref := content.NewBlock(content.NewParagraph(content.NewBlockRef(uuid)))
	triagedDivider.AddChild(ref)

	if scheduledBlock != nil {
		page.InsertBlockBefore(triagedDivider, scheduledBlock)
	} else {
		page.AddBlock(triagedDivider)
	}

	return nil
}
