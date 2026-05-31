package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-go/content"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

// ErrBlockNotFoundViaAPI is returned when a block UUID query returns no results from the Logseq API.
var ErrBlockNotFoundViaAPI = errors.New("block not found via API")

// BlockQueryInfo holds the result of a UUID lookup via Logseq API.
type BlockQueryInfo struct {
	PageName    string
	JournalDate time.Time
	IsJournal   bool
}

// FindBlockByUUID queries the Logseq HTTP API to find a block by UUID.
// Uses PostDatascriptQuery (logseq.db.datascriptQuery) because the pull syntax
// required here is not supported by logseq.db.q (PostQuery).
// The nested {:block/page [*]} expands page attributes; without it, page is just {id: N}.
func FindBlockByUUID(api LogseqAPI, uuid string) (*BlockQueryInfo, error) {
	query := fmt.Sprintf(`[:find (pull ?b [* {:block/page [*]}]) :where [?b :block/uuid #uuid "%s"]]`, uuid)

	jsonStr, err := api.PostDatascriptQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query block by UUID: %w", err)
	}

	return parseBlockQueryResponse(jsonStr, uuid)
}

// parseBlockQueryResponse parses the JSON response from a block UUID query.
func parseBlockQueryResponse(jsonStr, uuid string) (*BlockQueryInfo, error) {
	if jsonStr == "null" || jsonStr == "" {
		return nil, fmt.Errorf("%w: %s", ErrBlockNotFoundViaAPI, uuid)
	}

	var results [][]map[string]any

	unmarshalErr := json.Unmarshal([]byte(jsonStr), &results)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse UUID query response: %w", unmarshalErr)
	}

	if len(results) == 0 || len(results[0]) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrBlockNotFoundViaAPI, uuid)
	}

	block := results[0][0]
	page, _ := block["page"].(map[string]any)

	return extractBlockInfo(page), nil
}

// extractBlockInfo extracts page name and journal info from a page map.
// Logseq's datascript API returns hyphenated keys: "journal-day", "original-name".
func extractBlockInfo(page map[string]any) *BlockQueryInfo {
	info := &BlockQueryInfo{} //nolint:exhaustruct // fields set conditionally below

	if journalDay, ok := page["journal-day"].(float64); ok && journalDay > 0 {
		info.IsJournal = true
		dayInt := int(journalDay)
		info.JournalDate = time.Date(
			dayInt/logseqext.JournalDayDivisorYear,
			time.Month((dayInt%logseqext.JournalDayDivisorYear)/logseqext.JournalDayDivisorMonth),
			dayInt%logseqext.JournalDayDivisorMonth,
			0, 0, 0, 0, time.UTC,
		)
	}

	if name, ok := page["original-name"].(string); ok {
		info.PageName = name
	}

	return info
}

// BuildTaskListQuery assembles the Logseq Datalog query for listing tasks.
// It matches the Python `lqdpy tasks` query format exactly.
func BuildTaskListQuery(tags []string, includeCanceled, includeDone bool) string {
	condition := ""

	switch {
	case len(tags) == 1:
		condition = " [[" + tags[0] + "]]"
	case len(tags) > 1:
		parts := make([]string, len(tags))
		for i, t := range tags {
			parts[i] = "[[" + t + "]]"
		}

		condition = " (or " + strings.Join(parts, " ") + ")"
	}

	statuses := "TODO DOING WAITING NOW LATER"
	if includeCanceled {
		statuses += " " + content.TaskStringCanceled
	}

	if includeDone {
		statuses += " " + content.TaskStringDone
	}

	return "(and" + condition + " (task " + statuses + "))"
}

// SortTasksByDate sorts tasks in place by (JournalDay, Content) ascending,
// matching Python's Block.sort_by_date behavior.
func SortTasksByDate(tasks []TaskJSON) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Page.JournalDay != tasks[j].Page.JournalDay {
			return tasks[i].Page.JournalDay < tasks[j].Page.JournalDay
		}

		return tasks[i].Content < tasks[j].Content
	})
}
