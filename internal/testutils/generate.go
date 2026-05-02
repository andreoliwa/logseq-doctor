package testutils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	tparse "github.com/karrick/tparse/v2"
)

// blockIDBase is the starting numeric ID for fixture blocks.
const blockIDBase = 1000

// pageIDBase is the starting numeric ID for fixture pages.
const pageIDBase = 2000

// resolveRelativeDate parses a relative date string like "+3d", "-1w", "+2m", "+1y"
// and returns the resolved time. Empty string returns the zero time.
// "0" returns now. Supported units: d (days), w (weeks), m (months), y (years).
func resolveRelativeDate(rel string, now time.Time) (time.Time, error) {
	if rel == "" {
		return time.Time{}, nil
	}

	if rel == "0" {
		return now, nil
	}

	result, err := tparse.AddDuration(now, rel)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid relative date %q: %w", rel, err)
	}

	return result, nil
}

// toJournalDayInt converts a time.Time to the Logseq journalDay integer format (YYYYMMDD).
func toJournalDayInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// blockJSON is the JSON structure matching testdata/stub-api/*.jsonl entries.
type blockJSON struct {
	UUID                 string            `json:"uuid"`
	Marker               string            `json:"marker,omitempty"`
	Content              string            `json:"content"`
	Deadline             int               `json:"deadline,omitempty"`
	Scheduled            int               `json:"scheduled,omitempty"`
	Page                 pageJSON          `json:"page"`
	Properties           map[string]string `json:"properties"`
	PropertiesOrder      []string          `json:"propertiesOrder"`
	PropertiesTextValues map[string]string `json:"propertiesTextValues"`
	Refs                 []refJSON         `json:"refs"`
	PathRefs             []refJSON         `json:"pathRefs"`
	Format               string            `json:"format"`
	ID                   int               `json:"id"`
	Parent               refJSON           `json:"parent"`
	Left                 refJSON           `json:"left"`
}

type pageJSON struct {
	JournalDay   int    `json:"journalDay"`
	Name         string `json:"name"`
	OriginalName string `json:"originalName"`
	ID           int    `json:"id"`
}

type refJSON struct {
	ID int `json:"id"`
}

// buildBlockContent constructs the Logseq block content string.
func buildBlockContent(block Block, uuid string) string {
	var parts []string

	if block.Marker != "" {
		parts = append(parts, block.Marker)
	}

	if block.Priority != "" {
		parts = append(parts, "[#"+block.Priority+"]")
	}

	for _, tag := range block.Tags {
		parts = append(parts, "#"+tag)
	}

	parts = append(parts, block.Text)

	return strings.Join(parts, " ") + "\nid:: " + uuid
}

// buildBlockProperties constructs the properties map and order slice for a block.
func buildBlockProperties(block Block, uuid string, now time.Time) (map[string]string, []string) {
	props := map[string]string{"id": uuid}
	propsOrder := []string{"id"}

	if block.Groomed != "" {
		groomedTime, err := resolveRelativeDate(block.Groomed, now)
		if err == nil {
			props["groomed"] = logseqext.FormatLogseqDate(groomedTime)

			propsOrder = append(propsOrder, "groomed")
		}
	}

	for k, v := range block.ExtraProps {
		props[k] = v
		propsOrder = append(propsOrder, k)
	}

	return props, propsOrder
}

// buildJournalPage derives the page JSON for a block.
func buildJournalPage(block Block, now time.Time, pageID int) pageJSON {
	journalDate := now

	if block.JournalDay != "" {
		parsed, err := time.Parse("2006-01-02", block.JournalDay)
		if err == nil {
			journalDate = parsed
		}
	}

	pageName := strings.ToLower(journalDate.Format("Monday, 02-01-2006"))

	return pageJSON{
		JournalDay:   toJournalDayInt(journalDate),
		Name:         pageName,
		OriginalName: pageName,
		ID:           pageID,
	}
}

// resolveJournalDayInt resolves a relative date string to a Logseq journalDay int.
// Returns 0 if rel is empty or invalid.
func resolveJournalDayInt(rel string, now time.Time) int {
	if rel == "" {
		return 0
	}

	t, err := resolveRelativeDate(rel, now)
	if err != nil {
		return 0
	}

	return toJournalDayInt(t)
}

// buildTagIDs assigns stable numeric IDs to all distinct tags across blocks.
func buildTagIDs(blocks []Block) map[string]int {
	tagIDs := make(map[string]int)
	nextID := 1

	for _, block := range blocks {
		for _, tag := range block.Tags {
			if _, ok := tagIDs[tag]; !ok {
				tagIDs[tag] = nextID
				nextID++
			}
		}
	}

	return tagIDs
}

// buildBlockRefs builds the refs and pathRefs slices for a block.
func buildBlockRefs(block Block, tagIDs map[string]int, pageID int) ([]refJSON, []refJSON) {
	refs := make([]refJSON, 0, len(block.Tags))

	for _, tag := range block.Tags {
		refs = append(refs, refJSON{ID: tagIDs[tag]})
	}

	pathRefs := make([]refJSON, 0, len(refs)+1)
	pathRefs = append(pathRefs, refJSON{ID: pageID})
	pathRefs = append(pathRefs, refs...)

	return refs, pathRefs
}

// buildAPIResponse generates a JSON array string from blocks, matching the Logseq API response format.
// now is used to resolve relative Scheduled/Deadline/Groomed date strings.
func buildAPIResponse(blocks []Block, slugToUUIDMap map[string]string, now time.Time) string {
	tagIDs := buildTagIDs(blocks)
	results := make([]blockJSON, 0, len(blocks))

	for i, block := range blocks {
		uuid := slugToUUIDMap[block.Slug]
		blockID := blockIDBase + i
		pageID := pageIDBase + i

		content := buildBlockContent(block, uuid)
		props, propsOrder := buildBlockProperties(block, uuid, now)
		page := buildJournalPage(block, now, pageID)
		refs, pathRefs := buildBlockRefs(block, tagIDs, pageID)

		results = append(results, blockJSON{
			UUID:                 uuid,
			Marker:               block.Marker,
			Content:              content,
			Deadline:             resolveJournalDayInt(block.Deadline, now),
			Scheduled:            resolveJournalDayInt(block.Scheduled, now),
			Page:                 page,
			Properties:           props,
			PropertiesOrder:      propsOrder,
			PropertiesTextValues: props,
			Refs:                 refs,
			PathRefs:             pathRefs,
			Format:               "markdown",
			ID:                   blockID,
			Parent:               refJSON{ID: pageID},
			Left:                 refJSON{ID: pageID},
		})
	}

	data, err := json.Marshal(results)
	if err != nil {
		panic(fmt.Sprintf("buildAPIResponse: marshal failed: %v", err))
	}

	return string(data)
}

// ExportResolveRelativeDate exposes resolveRelativeDate for white-box testing.
var ExportResolveRelativeDate = resolveRelativeDate //nolint:gochecknoglobals

// ExportBuildAPIResponse exposes buildAPIResponse for white-box testing.
var ExportBuildAPIResponse = buildAPIResponse //nolint:gochecknoglobals
