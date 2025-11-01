package internal

import (
	"strings"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

type LogseqFinder interface {
	FindFirstQuery(pageTitle string) string
}

type logseqFinderImpl struct {
	graph *logseq.Graph
}

func NewLogseqFinder(graph *logseq.Graph) LogseqFinder {
	return &logseqFinderImpl{graph: graph}
}

func (f logseqFinderImpl) FindFirstQuery(pageTitle string) string {
	var query string

	page := OpenPage(f.graph, pageTitle)

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if q, ok := n.(*content.Query); ok {
				query = q.Query
			} else if qc, ok := n.(*content.QueryCommand); ok {
				query = strings.Trim(qc.Query, " \n\t")
			}

			if query != "" {
				// Stop after finding one query
				return true
			}

			return false
		})
	}

	if query == "" {
		return ""
	}

	return replaceCurrentPage(query, pageTitle)
}

// replaceCurrentPage replaces the current page placeholder in the query with the actual page name.
func replaceCurrentPage(query, pageTitle string) string {
	return strings.ReplaceAll(query, "<% current page %>", "[["+pageTitle+"]]")
}

// containsTextCaseInsensitive checks if a node contains the specified text (case-insensitive).
// It checks Text, PageLink, and Hashtag nodes.
func containsTextCaseInsensitive(node content.Node, searchTextLower string) bool {
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
}

// FindBlockContainingText searches for the first block containing the specified text using FindDeep.
func FindBlockContainingText(page logseq.Page, searchText string) *content.Block {
	if page == nil || searchText == "" {
		return nil
	}

	searchTextLower := strings.ToLower(searchText)

	return page.Blocks().FindDeep(func(block *content.Block) bool {
		textNode := block.Children().FindDeep(func(node content.Node) bool {
			return containsTextCaseInsensitive(node, searchTextLower)
		})

		return textNode != nil
	})
}

// FindTaskMarkerByKey searches for a task marker containing the specified key (case-insensitive).
// If parentBlock is provided, searches only among its children.
// Otherwise, searches in the entire page.
// Returns the TaskMarker if found, nil otherwise.
func FindTaskMarkerByKey(page logseq.Page, parentBlock *content.Block, key string) *content.TaskMarker {
	if key == "" {
		return nil
	}

	keyLower := strings.ToLower(key)

	predicate := func(block *content.Block) bool {
		// Check if this block has a task marker in its immediate content (not in nested blocks)
		hasTaskMarker := false

		block.Content().FindDeep(func(node content.Node) bool {
			if _, ok := node.(*content.TaskMarker); ok {
				hasTaskMarker = true

				return true
			}

			return false
		})

		if !hasTaskMarker {
			return false
		}

		// Check if the key is in the immediate content of this block (not in nested blocks)
		textNode := block.Content().FindDeep(func(node content.Node) bool {
			return containsTextCaseInsensitive(node, keyLower)
		})

		return textNode != nil
	}

	var block *content.Block
	if parentBlock != nil {
		block = parentBlock.Blocks().FindDeep(predicate)
	} else {
		block = page.Blocks().FindDeep(predicate)
	}

	if block == nil {
		return nil
	}

	// Find the TaskMarker in the block's content
	taskMarkerNode := block.Content().FindDeep(func(node content.Node) bool {
		_, ok := node.(*content.TaskMarker)

		return ok
	})

	if taskMarkerNode == nil {
		return nil
	}

	//nolint:forcetypeassert // We know it's a TaskMarker from the predicate above
	return taskMarkerNode.(*content.TaskMarker)
}
