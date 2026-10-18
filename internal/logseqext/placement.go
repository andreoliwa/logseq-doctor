package logseqext

import (
	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"
)

// PlaceBlock inserts before the first later timed sibling, or appends.
// Existing siblings retain their relative order and time spelling.
func PlaceBlock(page logseq.Page, parent *content.Block, block *content.Block) {
	siblings := page.Blocks()
	if parent != nil {
		siblings = parent.Blocks()
	}

	anchor := laterTimeAnchor(block, siblings)
	switch {
	case parent != nil:
		if anchor == nil {
			parent.AddChild(block)
		} else {
			parent.InsertChildBefore(block, anchor)
		}
	case anchor == nil:
		page.AddBlock(block)
	default:
		page.InsertBlockBefore(block, anchor)
	}
}

// RepositionBlock moves a timed update within its actual parent only.
func RepositionBlock(block *content.Block) {
	if _, timed := content.LeadingTime(block); !timed {
		return
	}

	if parent, ok := block.Parent().(*content.Block); ok {
		if anchor := laterTimeAnchor(block, parent.Blocks()); anchor != nil {
			parent.InsertChildBefore(block, anchor)
		} else {
			parent.AddChild(block)
		}
	}
}

// LaterTimeAnchor exposes the shared first-later rule for source-slice placement.
func LaterTimeAnchor(block *content.Block, siblings content.BlockList) *content.Block {
	return laterTimeAnchor(block, siblings)
}
