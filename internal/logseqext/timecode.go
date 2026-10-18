package logseqext

import (
	"strings"

	"github.com/andreoliwa/logseq-go/v2/content"
)

// EffectiveTime returns a block's leading time, or the first timed direct
// child's leading time for a ref-only block. It never resolves references.
func EffectiveTime(block *content.Block) (content.TimeOfDay, bool) {
	if value, ok := content.LeadingTime(block); ok {
		return value, true
	}

	if !isRefOnly(block) {
		return content.TimeOfDay{}, false
	}

	for _, child := range block.Blocks() {
		if value, ok := content.LeadingTime(child); ok {
			return value, true
		}
	}

	return content.TimeOfDay{}, false
}

func isRefOnly(block *content.Block) bool {
	refs := 0

	for _, node := range block.Content() {
		if _, ok := node.(*content.Properties); ok {
			continue
		}

		paragraph, ok := node.(*content.Paragraph)
		if !ok {
			return false
		}

		for _, inline := range paragraph.Children() {
			switch value := inline.(type) {
			case *content.BlockRef:
				refs++
			case *content.Text:
				if strings.TrimSpace(value.Value) != "" {
					return false
				}
			default:
				return false
			}
		}
	}

	return refs == 1
}

// laterTimeAnchor deliberately scans in stored order rather than sorting.
func laterTimeAnchor(block *content.Block, siblings content.BlockList) *content.Block {
	clock, ok := content.LeadingTime(block)
	if !ok {
		return nil
	}

	for _, sibling := range siblings {
		if sibling == block {
			continue
		}

		if other, timed := EffectiveTime(sibling); timed && other.Minutes() > clock.Minutes() {
			return sibling
		}
	}

	return nil
}
