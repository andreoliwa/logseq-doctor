package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

// ErrPageIsNil is returned when a page is nil.
var ErrPageIsNil = errors.New("page is nil")

// InsertMarkdownOptions contains options for inserting Markdown content.
type InsertMarkdownOptions struct {
	Graph      *logseq.Graph
	Date       time.Time
	Page       string // Page name to add content to (empty = journal)
	Content    string
	ParentText string // Partial text to search for in parent blocks
	Key        string // Unique key to search for existing block (case-insensitive)
}

// InsertMarkdown inserts Markdown content to a page or journal.
// If Page is provided, adds to that page. Otherwise, adds to journal for Date.
// If Key is provided, it searches for an existing block containing that key (case-insensitive)
// and updates it. Otherwise, creates a new block.
// If ParentText is provided, it searches for the first block containing that text
// and inserts the content among its direct children. Timed content uses sibling chronology.
func InsertMarkdown(opts *InsertMarkdownOptions) error {
	if strings.TrimSpace(opts.Content) == "" {
		return nil
	}

	transaction := opts.Graph.NewTransaction()

	var targetPage logseq.Page

	var err error

	if opts.Page != "" {
		targetPage, err = transaction.OpenPage(opts.Page)
	} else {
		targetPage, err = transaction.OpenJournal(opts.Date)
	}

	if err != nil {
		return fmt.Errorf("error opening target page for transaction: %w", err)
	}

	var parentBlock *content.Block
	if opts.ParentText != "" {
		parentBlock = logseqext.FindBlockContainingText(targetPage, opts.ParentText)
		// If parent not found, parentBlock will be nil and content will be added to top level
	}

	var existingBlock *content.Block
	if opts.Key != "" {
		existingBlock = logseqext.FindBlockByKey(targetPage, parentBlock, opts.Key)
	}

	if existingBlock != nil {
		err = updateExistingBlock(existingBlock, opts.Content)
		if err != nil {
			return fmt.Errorf("error updating block: %w", err)
		}

		logseqext.RepositionBlock(existingBlock)
	} else {
		err = addContent(targetPage, parentBlock, opts.Content)
		if err != nil {
			return fmt.Errorf("error adding content: %w", err)
		}
	}

	err = transaction.Save()
	if err != nil {
		return fmt.Errorf("error saving transaction: %w", err)
	}

	return nil
}

// updateExistingBlock updates an existing block's content while preserving children, properties, and logbook.
func updateExistingBlock(block *content.Block, newContent string) error {
	if block == nil {
		return ErrPageIsNil
	}

	trimmedContent := strings.TrimSpace(newContent)
	if trimmedContent == "" {
		return nil
	}

	// Parse the new content as Markdown
	parsedBlock, err := logseq.ParseBlock(trimmedContent)
	if err != nil {
		return fmt.Errorf("error parsing content as markdown: %w", err)
	}

	removeOldContentNodes(block)
	insertNewContentNodes(block, parsedBlock)

	return nil
}

// removeOldContentNodes removes content nodes from a block while preserving Properties, Logbook, and child Blocks.
func removeOldContentNodes(block *content.Block) {
	var nodesToRemove []content.Node

	for node := block.FirstChild(); node != nil; node = node.NextSibling() {
		if shouldPreserveNode(node) {
			continue
		}
		// This is a content node (Paragraph, etc.), mark for removal
		nodesToRemove = append(nodesToRemove, node)
	}

	// Remove the old content nodes
	for _, node := range nodesToRemove {
		node.RemoveSelf()
	}
}

// shouldPreserveNode returns true if the node should be preserved (Properties, Logbook, or child Blocks).
func shouldPreserveNode(node content.Node) bool {
	if _, ok := node.(*content.Properties); ok {
		return true
	}

	if _, ok := node.(*content.Logbook); ok {
		return true
	}

	if _, ok := node.(*content.Block); ok {
		return true
	}

	return false
}

// insertNewContentNodes inserts new content nodes from parsedBlock into block.
func insertNewContentNodes(block *content.Block, parsedBlock *content.Block) {
	if len(parsedBlock.Children()) == 0 {
		return
	}

	// Place visible content before metadata as well as children so the first
	// paragraph remains the block's first line after a key update.
	var insertBefore content.Node

	for node := block.FirstChild(); node != nil; node = node.NextSibling() {
		if shouldPreserveNode(node) {
			insertBefore = node

			break
		}
	}

	children := parsedBlock.Children()
	// Always iterate forward to maintain order, whether inserting before a node or appending
	for i := range len(children) {
		child := children[i]
		// Skip child blocks from the parsed content, only add content nodes
		if _, ok := child.(*content.Block); ok {
			continue
		}

		if insertBefore != nil {
			block.InsertChildBefore(child, insertBefore)
		} else {
			block.AddChild(child)
		}
	}
}

// addContent adds content either as a child block to the specified parent or as a top-level block to the page.
func addContent(page logseq.Page, parentBlock *content.Block, contentText string) error {
	if page == nil {
		return ErrPageIsNil
	}

	trimmedContent := strings.TrimSpace(contentText)
	if trimmedContent == "" {
		return nil
	}

	// The parser distinguishes explicit root outlines from a single block with
	// supporting paragraphs and child blocks. Never flatten the latter.
	parsedBlock, _, outlined, err := logseq.ParseSource(trimmedContent)
	// Source boundaries are unnecessary for a single block. A parsed result
	// remains usable when only range mapping failed on unoutlined input.
	if err != nil && (outlined || parsedBlock == nil) {
		return fmt.Errorf("error parsing content as markdown: %w", err)
	}

	if outlined {
		for _, block := range parsedBlock.Blocks() {
			logseqext.PlaceBlock(page, parentBlock, block)
		}
	} else if len(parsedBlock.Children()) > 0 {
		logseqext.PlaceBlock(page, parentBlock, parsedBlock)
	}

	return nil
}
