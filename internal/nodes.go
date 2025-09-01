package internal

import "github.com/andreoliwa/logseq-go/content"

// IsAncestor checks if ancestor is an ancestor of block by traversing up the parent chain.
// Returns true if ancestor is found in the parent hierarchy of block, false otherwise.
// Returns true if block and ancestor are the same block.
func IsAncestor(block, ancestor *content.Block) bool {
	for block != nil {
		if block == ancestor {
			return true
		}

		parent := block.Parent()
		if parent == nil {
			break
		}

		var ok bool

		block, ok = parent.(*content.Block)
		if !ok {
			break
		}
	}

	return false
}

// TODO: move other generic functions here too: AddSibling(), nextChildHasPin(), removeEmptyDividers()
