package testutils

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// slugRefPattern matches (( any slug with spaces )) in committed .md files.
var slugRefPattern = regexp.MustCompile(`\(\(\s*([^)]+?)\s*\)\)`)

// slugToUUID derives a deterministic UUID from a slug using FNV-32a.
// Format: {h8}-0000-0000-0000-{h8}0000 where h8 = zero-padded 8-char hex of fnv32a(slug).
func slugToUUID(slug string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(slug))
	n := h.Sum32()

	return fmt.Sprintf("%08x-0000-0000-0000-%08x0000", n, n)
}

// buildSlugMap builds slug→UUID and UUID→slug maps from a slice of blocks.
// Panics if two different slugs produce the same UUID (collision detection).
func buildSlugMap(blocks []Block) (map[string]string, map[string]string) {
	slugToUUIDMap := make(map[string]string, len(blocks))
	uuidToSlugMap := make(map[string]string, len(blocks))

	for _, block := range blocks {
		uuid := slugToUUID(block.Slug)
		if existing, ok := uuidToSlugMap[uuid]; ok {
			panic(fmt.Sprintf("slug UUID collision: %q and %q both hash to %s", block.Slug, existing, uuid))
		}

		slugToUUIDMap[block.Slug] = uuid
		uuidToSlugMap[uuid] = block.Slug
	}

	return slugToUUIDMap, uuidToSlugMap
}

// expandSlugs replaces (( slug )) with ((uuid)) in content.
// Slugs not present in slugToUUIDMap are left unchanged.
func expandSlugs(content string, slugToUUIDMap map[string]string) string {
	const minSubLen = 2

	return slugRefPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := slugRefPattern.FindStringSubmatch(match)
		if len(sub) < minSubLen {
			return match
		}

		slug := strings.TrimSpace(sub[1])

		uuid, ok := slugToUUIDMap[slug]
		if !ok {
			return match
		}

		return "((" + uuid + "))"
	})
}

// collapseSlugs replaces ((uuid)) with (( slug )) in content.
// UUIDs not present in uuidToSlugMap are left unchanged.
func collapseSlugs(content string, uuidToSlugMap map[string]string) string {
	const minSubLen = 2

	uuidRefPattern := regexp.MustCompile(`\(\(([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\)\)`)

	return uuidRefPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := uuidRefPattern.FindStringSubmatch(match)
		if len(sub) < minSubLen {
			return match
		}

		slug, ok := uuidToSlugMap[sub[1]]
		if !ok {
			return match
		}

		return "(( " + slug + " ))"
	})
}

// ExportBuildSlugMap exposes buildSlugMap for white-box testing.
var ExportBuildSlugMap = buildSlugMap //nolint:gochecknoglobals

// ExportExpandSlugs exposes expandSlugs for white-box testing.
var ExportExpandSlugs = expandSlugs //nolint:gochecknoglobals

// ExportCollapseSlugs exposes collapseSlugs for white-box testing.
var ExportCollapseSlugs = collapseSlugs //nolint:gochecknoglobals
