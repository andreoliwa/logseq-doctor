package internal

import (
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"strings"
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
