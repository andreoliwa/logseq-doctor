package logseqext

import (
	"regexp"
	"strings"
	"unicode"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Regex patterns for extracting tags from task content.
var (
	markdownLinkPattern  = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	pageRefPattern       = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	bracketHashtagPatter = regexp.MustCompile(`#\[\[([^\]]+)\]\]`)
	simpleHashtagPattern = regexp.MustCompile(`#([\w\-/]+)`)
	timePrefixPattern    = regexp.MustCompile(`^\*?\*?(\d{1,2}:\d{2})\*?\*?\s+`)
)

// CleanTaskName extracts the task name from content: first line, marker stripped, time prefix stripped.
func CleanTaskName(taskContent, marker string) string {
	firstLine, _, _ := strings.Cut(taskContent, "\n")

	firstLine = strings.TrimPrefix(firstLine, marker+" ")
	firstLine = timePrefixPattern.ReplaceAllString(firstLine, "")

	return strings.TrimSpace(firstLine)
}

// ExtractDirectTags extracts #hashtags and [[page refs]] from content text.
func ExtractDirectTags(contentText string) []string {
	if contentText == "" {
		return nil
	}

	// Strip markdown links to avoid parsing tags within URLs.
	contentClean := markdownLinkPattern.ReplaceAllString(contentText, "$1")

	var tags []string

	// [[page references]]
	for _, match := range pageRefPattern.FindAllStringSubmatch(contentClean, -1) {
		tags = append(tags, match[1])
	}

	// #[[tag with spaces]]
	for _, match := range bracketHashtagPatter.FindAllStringSubmatch(contentClean, -1) {
		tags = append(tags, match[1])
	}

	// #hashtags (word chars, hyphens, slashes)
	for _, match := range simpleHashtagPattern.FindAllStringSubmatch(contentClean, -1) {
		tags = append(tags, match[1])
	}

	result := UniqueStrings(tags)

	if len(result) == 0 {
		return []string{}
	}

	return result
}

// NormalizeTagPrefixes ensures all tags start with "#", are lowercase, and slugified
// (accents removed, non-alphanumeric chars stripped) to match Python's slugify(t, separator=").
func NormalizeTagPrefixes(tags []string) {
	for idx, tag := range tags {
		body := strings.TrimPrefix(tag, "#")
		tags[idx] = "#" + slugifyTag(body)
	}
}

// slugifyTag removes accents and non-alphanumeric characters, then lowercases —
// matching Python's slugify(tag, separator=") behaviour.
func slugifyTag(s string) string {
	// Decompose to NFD (separate base letters from diacritics), strip marks, recompose.
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	normalized, _, _ := transform.String(t, s)

	// Keep only alphanumeric chars (strip spaces, hyphens, slashes, etc.).
	var builder strings.Builder

	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}

	return builder.String()
}

// UniqueStrings deduplicates a string slice preserving order.
func UniqueStrings(items []string) []string {
	seen := make(map[string]bool, len(items))

	var result []string

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ExtractBlockRefUUIDs extracts all block ref UUIDs from a page (ordered).
func ExtractBlockRefUUIDs(page logseq.Page) []string {
	var uuids []string

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if ref, ok := n.(*content.BlockRef); ok {
				uuids = append(uuids, ref.ID)
			}

			return false
		})
	}

	return uuids
}

// SectionedUUID pairs a block-ref UUID with whether it lives under a section
// header (ranked=false means it is under Overdue, New tasks, Triaged, Scheduled,
// or Unranked — any divider that is not the implicit "ranked" area at the top).
type SectionedUUID struct {
	UUID   string
	Ranked bool
}

// ExtractSectionedBlockRefUUIDs scans a backlog page and returns every block-ref
// UUID together with whether it is in the ranked area (above all section dividers)
// or the unranked area (under any divider whose text matches one of
// unrankedSectionTexts, case-insensitive substring).
//
// The ranked area is defined as top-level blocks that are NOT section-header
// blocks and NOT descendants of any section-header block.
func ExtractSectionedBlockRefUUIDs(page logseq.Page, unrankedSectionTexts []string) []SectionedUUID {
	// First pass: find all section-header blocks.
	sectionHeaders := make(map[*content.Block]bool)

	for _, block := range page.Blocks() {
		text := BlockContentText(block)
		for _, headerText := range unrankedSectionTexts {
			if strings.Contains(strings.ToLower(text), strings.ToLower(headerText)) {
				sectionHeaders[block] = true

				break
			}
		}
	}

	// Second pass: classify each ref.
	var result []SectionedUUID

	for _, block := range page.Blocks() {
		if sectionHeaders[block] {
			// This is a divider block itself — collect refs under it as unranked.
			block.Children().FindDeep(func(n content.Node) bool {
				if ref, ok := n.(*content.BlockRef); ok {
					result = append(result, SectionedUUID{UUID: ref.ID, Ranked: false})
				}

				return false
			})

			continue
		}

		// Top-level block: ranked only if it is not a descendant of any header.
		underHeader := false

		for header := range sectionHeaders {
			if isDescendantBlock(block, header) {
				underHeader = true

				break
			}
		}

		block.Children().FindDeep(func(n content.Node) bool {
			if ref, ok := n.(*content.BlockRef); ok {
				result = append(result, SectionedUUID{UUID: ref.ID, Ranked: !underHeader})
			}

			return false
		})
	}

	return result
}

// isDescendantBlock returns true if block is a descendant of ancestor
// by traversing the parent chain upward.
func isDescendantBlock(block, ancestor *content.Block) bool {
	for block != nil {
		parent := block.Parent()
		if parent == nil {
			break
		}

		parentBlock, ok := parent.(*content.Block)
		if !ok {
			break
		}

		if parentBlock == ancestor {
			return true
		}

		block = parentBlock
	}

	return false
}
