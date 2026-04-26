package dashboard

import (
	"fmt"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

func allSectionHeaders() []backlog.Header {
	return []backlog.Header{
		backlog.HeaderFocus, backlog.HeaderOverdue, backlog.HeaderNewTasks,
		backlog.HeaderTriaged, backlog.HeaderScheduled, backlog.HeaderUnranked,
	}
}

// MoveToUnranked moves the given task UUIDs from the regular area of backlogPage
// to under the "🔢 Unranked tasks" section divider, creating the divider if absent.
//
// graphPath is the path to the Logseq graph root directory.
// backlogPageName is the page name (e.g. "my-backlog", without .md extension).
// uuids is the list of task UUIDs to move.
func MoveToUnranked(graphPath, backlogPageName string, uuids []string) error {
	if len(uuids) == 0 {
		return nil
	}

	graph := logseqapi.OpenGraphFromPath(graphPath)
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(backlogPageName)
	if err != nil {
		return fmt.Errorf("open page %q: %w", backlogPageName, err)
	}

	uuidSet := make(map[string]bool, len(uuids))
	for _, u := range uuids {
		uuidSet[u] = true
	}

	unrankedDivider := logseqext.FindBlockContainingText(page, backlog.HeaderUnranked.Label)
	toMove := collectBlocksToMove(page, uuidSet)

	if len(toMove) == 0 {
		return nil // nothing to do
	}

	unrankedDivider = ensureUnrankedDivider(page, unrankedDivider)

	for _, block := range toMove {
		block.RemoveSelf()
		unrankedDivider.AddChild(block)
	}

	err = transaction.Save()
	if err != nil {
		return fmt.Errorf("save transaction: %w", err)
	}

	return nil
}

// collectBlocksToMove returns all blocks (top-level or children of section headers)
// whose block-ref UUID is in uuidSet. Section header blocks themselves are skipped.
func collectBlocksToMove(page logseq.Page, uuidSet map[string]bool) []*content.Block {
	var toMove []*content.Block

	for _, block := range page.Blocks() {
		if isSectionHeaderBlock(block) {
			// Tasks under a section header (e.g. 🆕 New tasks) are descendant blocks.
			block.Children().FindDeep(func(n content.Node) bool {
				childBlock, ok := n.(*content.Block)
				if !ok {
					return false
				}

				uuid := logseqext.ExtractBlockRefUUID(childBlock)
				if uuidSet[uuid] {
					toMove = append(toMove, childBlock)
				}

				return false
			})

			continue
		}

		uuid := logseqext.ExtractBlockRefUUID(block)
		if uuidSet[uuid] {
			toMove = append(toMove, block)
		}
	}

	return toMove
}

// isSectionHeaderBlock reports whether block's text matches any known section header.
func isSectionHeaderBlock(block *content.Block) bool {
	blockText := logseqext.BlockContentText(block)

	for _, header := range allSectionHeaders() {
		if header.Matches(blockText) {
			return true
		}
	}

	return false
}

// ensureUnrankedDivider returns the existing divider block, or creates and inserts one.
func ensureUnrankedDivider(page logseq.Page, existing *content.Block) *content.Block {
	if existing != nil {
		return existing
	}

	dividerBlock := content.NewBlock(content.NewParagraph(content.NewText(backlog.HeaderUnranked.String())))
	scheduledDivider := logseqext.FindBlockContainingText(page, backlog.HeaderScheduled.Label)

	if scheduledDivider != nil {
		page.InsertBlockBefore(dividerBlock, scheduledDivider)
	} else {
		page.AddBlock(dividerBlock)
	}

	return dividerBlock
}
