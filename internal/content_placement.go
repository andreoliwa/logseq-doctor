package internal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

var (
	errInvalidSourceRange     = errors.New("invalid parser source range")
	errIncompleteSourceRanges = errors.New("parser ranges do not cover source suffix")
	errUnsafeSourceJoin       = errors.New("insertion would change root block ownership")
)

type rawBlock struct {
	block  *content.Block
	source string
}

func placeRawContent(original, incoming string) (string, error) {
	parsed, additions, outlined, err := logseq.ParseSource(incoming)
	if !outlined && parsed != nil {
		return appendRaw(original, incoming), nil
	}

	if err != nil {
		return "", fmt.Errorf("parse incoming content: %w", err)
	}

	if !outlined {
		return appendRaw(original, incoming), nil
	}

	err = validateSourceRanges(incoming, additions)
	if err != nil {
		return "", err
	}

	_, existing, _, err := logseq.ParseSource(original)
	if err != nil {
		return "", fmt.Errorf("parse journal: %w", err)
	}

	err = validateSourceRanges(original, existing)
	if err != nil {
		return "", err
	}

	prefix, blocks := existingRawBlocks(original, existing, len(additions))

	// The parser assigns the preamble to the page, not to the first root.
	prefix = joinRaw(prefix, incoming[:additions[0].Start])
	for _, item := range additions {
		blocks = insertRawBlock(blocks, item.Block, incoming[item.Start:item.End])
	}

	result, starts := assembleRawBlocks(prefix, blocks)

	err = validateRawLayout(result, starts)
	if err != nil {
		return "", err
	}

	return result, nil
}

func validateSourceRanges(source string, ranges []logseq.BlockSourceRange) error {
	previous := -1
	for _, item := range ranges {
		if item.Block == nil || item.Start < 0 || item.End <= item.Start || item.End > len(source) ||
			(previous >= 0 && item.Start != previous) {
			return fmt.Errorf("%w [%d:%d]", errInvalidSourceRange, item.Start, item.End)
		}

		previous = item.End
	}

	if previous >= 0 && previous != len(source) {
		return errIncompleteSourceRanges
	}

	return nil
}

func joinRaw(before, after string) string {
	if before == "" || after == "" || strings.HasSuffix(before, "\n") {
		return before + after
	}

	newline := "\n"
	if strings.Contains(before, "\r\n") || strings.Contains(after, "\r\n") {
		newline = "\r\n"
	}

	return before + newline + after
}

func appendRaw(original, incoming string) string {
	// Preserve the old empty-journal placeholder behavior for unoutlined input.
	if strings.TrimLeft(strings.TrimRight(original, "\n"), "- ") == "" {
		return incoming
	}

	return joinRaw(original, incoming)
}

func insertRawBlock(blocks []rawBlock, block *content.Block, source string) []rawBlock {
	siblings := make(content.BlockList, 0, len(blocks))
	for _, item := range blocks {
		siblings = append(siblings, item.block)
	}

	anchor := logseqext.LaterTimeAnchor(block, siblings)

	index := len(blocks)
	for j, item := range blocks {
		if item.block == anchor {
			index = j

			break
		}
	}

	blocks = append(blocks, rawBlock{block: nil, source: ""})
	copy(blocks[index+1:], blocks[index:])
	blocks[index] = rawBlock{block: block, source: source}

	return blocks
}

func existingRawBlocks(original string, existing []logseq.BlockSourceRange, additional int) (string, []rawBlock) {
	prefix := original

	blocks := make([]rawBlock, 0, len(existing)+additional)
	if len(existing) != 0 {
		prefix = original[:existing[0].Start]
		for _, item := range existing {
			blocks = append(blocks, rawBlock{block: item.Block, source: original[item.Start:item.End]})
		}
	}

	return prefix, blocks
}

func assembleRawBlocks(prefix string, blocks []rawBlock) (string, []int) {
	result := prefix
	starts := make([]int, 0, len(blocks))

	for _, block := range blocks {
		joined := joinRaw(result, block.source)
		starts = append(starts, len(joined)-len(block.source))
		result = joined
	}

	return result, starts
}

func validateRawLayout(result string, starts []int) error {
	_, ranges, _, err := logseq.ParseSource(result)
	if err != nil {
		return fmt.Errorf("validate inserted content: %w", err)
	}

	if len(ranges) != len(starts) {
		return errUnsafeSourceJoin
	}

	for i, item := range ranges {
		if item.Start != starts[i] {
			return errUnsafeSourceJoin
		}
	}

	return nil
}
