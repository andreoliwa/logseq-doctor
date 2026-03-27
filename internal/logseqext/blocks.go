package logseqext

import (
	"context"
	"fmt"
	"log"
	"strings"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

// JournalDayDivisorYear is used to extract the year from a journalDay integer (YYYYMMDD).
const JournalDayDivisorYear = 10000

// JournalDayDivisorMonth is used to extract the month from a journalDay integer (YYYYMMDD).
const JournalDayDivisorMonth = 100

// BlockProperties returns the Properties node for a task block.
// logseq-go's block.Properties() only checks the first child — task blocks have a
// Paragraph first, so it would prepend a new empty Properties node before the TODO line.
// The parser creates an empty Properties placeholder at position 0; the real properties
// (id::, collapsed::, etc.) appear after the first Paragraph.
// This helper finds the first Properties that comes after a Paragraph in the content.
func BlockProperties(block *content.Block) *content.Properties {
	seenParagraph := false

	for _, node := range block.Content() {
		if _, ok := node.(*content.Paragraph); ok {
			seenParagraph = true

			continue
		}

		if p, ok := node.(*content.Properties); ok && seenParagraph {
			return p
		}
	}

	// No properties after a paragraph — fall back to block.Properties() which creates one.
	// This path is taken for blocks that truly have no properties yet.
	return block.Properties()
}

// FindBlockByIDProperty finds a block in a page by its id:: property value.
// It searches Properties nodes within block content because logseq-go's block.Properties()
// only checks the first child — task blocks have a paragraph first, then properties.
func FindBlockByIDProperty(page logseq.Page, uuid string) *content.Block {
	return page.Blocks().FindDeep(func(block *content.Block) bool {
		return blockHasIDProperty(block, uuid)
	})
}

// blockHasIDProperty checks whether a block's id:: property matches the given UUID.
// Searches inside content nodes to handle task blocks where properties follow the paragraph.
func blockHasIDProperty(block *content.Block, uuid string) bool {
	found := false

	block.Content().FindDeep(func(node content.Node) bool {
		props, ok := node.(*content.Properties)
		if !ok {
			return false
		}

		for _, v := range props.Get("id") {
			if t, ok := v.(*content.Text); ok && strings.Contains(t.Value, uuid) {
				found = true

				return true
			}
		}

		return false
	})

	return found
}

// BlockContentText extracts the text content from a block's content nodes.
func BlockContentText(block *content.Block) string {
	var text string

	block.Content().FindDeep(func(node content.Node) bool {
		if textNode, ok := node.(*content.Text); ok {
			text = textNode.Value

			return true
		}

		return false
	})

	return text
}

// SetTaskCanceled changes the task marker to CANCELED using logseq-go's WithStatus API.
func SetTaskCanceled(block *content.Block) error {
	var taskMarker *content.TaskMarker

	block.Content().FindDeep(func(node content.Node) bool {
		if marker, ok := node.(*content.TaskMarker); ok {
			taskMarker = marker

			return true
		}

		return false
	})

	if taskMarker == nil {
		return nil // No task marker found, nothing to change
	}

	_, err := taskMarker.WithStatus(content.TaskStatusCanceled)
	if err != nil {
		return fmt.Errorf("failed to change task status to canceled: %w", err)
	}

	return nil
}

// AddSibling inserts newBlock into page relative to the given anchor blocks.
// It inserts after the last non-nil block in after, or before the before block,
// or appends to the page if neither is provided.
func AddSibling(page logseq.Page, newBlock, before *content.Block, after ...*content.Block) {
	for _, a := range after {
		if a != nil {
			page.InsertBlockAfter(newBlock, a)

			return
		}
	}

	if before != nil {
		page.InsertBlockBefore(newBlock, before)

		return
	}

	page.AddBlock(newBlock)
}

// RemoveEmptyBlocks removes blocks that have no child blocks and returns true if any were removed.
func RemoveEmptyBlocks(save bool, blocks ...*content.Block) bool {
	for _, block := range blocks {
		if block != nil && len(block.Blocks()) == 0 {
			block.RemoveSelf()

			save = true
		}
	}

	return save
}

// OpenGraphFromPath opens a Logseq graph from the given directory path.
// Aborts the program if path is empty or the graph cannot be opened,
// to avoid error-handling boilerplate at every call site.
func OpenGraphFromPath(path string) *logseq.Graph {
	if path == "" {
		log.Fatalln("path is empty, maybe the LOGSEQ_GRAPH_PATH environment variable is not set")
	}

	ctx := context.Background()

	graph, err := logseq.Open(ctx, path)
	if err != nil {
		log.Fatalf("error opening graph: %v", err)
	}

	return graph
}

// OpenPage opens a page in the Logseq graph.
// Aborts the program in case of error to avoid boilerplate at every call site.
func OpenPage(graph *logseq.Graph, pageTitle string) logseq.Page {
	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		log.Fatalf("error opening page: %v", err)
	}

	return page
}
