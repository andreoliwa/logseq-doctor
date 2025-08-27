package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

// ErrPageIsNil is returned when a page is nil.
var ErrPageIsNil = errors.New("page is nil")

// InsertMarkdownOptions contains options for inserting markdown content.
type InsertMarkdownOptions struct {
	Graph      *logseq.Graph
	Date       time.Time
	Content    string
	ParentText string // Partial text to search for in parent blocks
}

// InsertMarkdownToJournal inserts markdown content to a journal page
// If ParentText is provided, it searches for the first block containing that text
// and inserts the content as a child block. Otherwise, appends to the end.
func InsertMarkdownToJournal(opts *InsertMarkdownOptions) error {
	if opts.Content == "" {
		return nil
	}

	transaction := opts.Graph.NewTransaction()

	journalTx, err := transaction.OpenJournal(opts.Date)
	if err != nil {
		return fmt.Errorf("error opening journal page for transaction: %w", err)
	}

	var parentBlock *content.Block
	if opts.ParentText != "" {
		parentBlock = findBlockContainingText(journalTx, opts.ParentText)
		// If parent not found, parentBlock will be nil and content will be added to top level
	}

	err = addContent(journalTx, parentBlock, opts.Content)
	if err != nil {
		return fmt.Errorf("error adding content: %w", err)
	}

	err = transaction.Save()
	if err != nil {
		return fmt.Errorf("error saving transaction: %w", err)
	}

	return nil
}

// findBlockContainingText searches for the first block containing the specified text using FindDeep.
func findBlockContainingText(page logseq.Page, searchText string) *content.Block {
	if page == nil || searchText == "" {
		return nil
	}

	searchTextLower := strings.ToLower(searchText)

	return page.Blocks().FindDeep(func(block *content.Block) bool {
		textNode := block.Children().FindDeep(func(node content.Node) bool {
			if text, ok := node.(*content.Text); ok {
				return strings.Contains(strings.ToLower(text.Value), searchTextLower)
			}

			if pageLink, ok := node.(*content.PageLink); ok {
				return strings.Contains(strings.ToLower(pageLink.To), searchTextLower)
			}

			if hashtag, ok := node.(*content.Hashtag); ok {
				return strings.Contains(strings.ToLower(hashtag.To), searchTextLower)
			}

			return false
		})

		return textNode != nil
	})
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

	// Parse the content as markdown to handle multi-line text properly
	parsedBlock, err := logseq.ParseBlock(trimmedContent)
	if err != nil {
		return fmt.Errorf("error parsing content as markdown: %w", err)
	}

	if len(parsedBlock.Children()) > 0 {
		newBlock := content.NewBlock()

		children := parsedBlock.Children()
		for _, child := range children {
			newBlock.AddChild(child)
		}

		if parentBlock != nil {
			parentBlock.AddChild(newBlock)
		} else {
			page.AddBlock(newBlock)
		}
	}

	return nil
}
