package lqdsync

import (
	"fmt"
	"time"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
)

// RankInfo holds backlog rank data for a task.
type RankInfo struct {
	BacklogName  string
	BacklogIndex int
	Rank         int
}

// CalculateRanks processes backlogs in reverse order and assigns ranks.
// Tasks in earlier backlogs override later backlogs.
func CalculateRanks(backlogs map[string][]string, backlogOrder []string) map[string]RankInfo {
	rankMap := make(map[string]RankInfo)

	for idx := len(backlogOrder) - 1; idx >= 0; idx-- {
		backlogName := backlogOrder[idx]
		uuids, ok := backlogs[backlogName]

		if !ok {
			continue
		}

		for rank, uuid := range uuids {
			rankMap[uuid] = RankInfo{
				BacklogName:  backlogName,
				BacklogIndex: idx + 1,
				Rank:         rank + 1,
			}
		}
	}

	return rankMap
}

// yyyymmddToDateOnly converts a YYYYMMDD integer to a plain date string (YYYY-MM-DD).
// Returns empty string for zero values.
// Used for journal dates where Python sends date.isoformat() without timezone.
func yyyymmddToDateOnly(dateInt int) string {
	if dateInt == 0 {
		return ""
	}

	const (
		yearDivisor  = 10000
		monthDivisor = 100
	)

	year := dateInt / yearDivisor
	month := (dateInt % yearDivisor) / monthDivisor
	day := dateInt % monthDivisor

	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// yyyymmddToLocalISO converts a YYYYMMDD integer to an RFC3339 datetime string with local timezone offset.
// Returns empty string for zero values.
// Used for scheduled/deadline dates where Python uses datetime.strptime().astimezone().isoformat()
// which includes the local timezone offset, causing PocketBase to shift to UTC on storage.
func yyyymmddToLocalISO(dateInt int) string {
	if dateInt == 0 {
		return ""
	}

	const (
		yearDivisor  = 10000
		monthDivisor = 100
	)

	year := dateInt / yearDivisor
	month := (dateInt % yearDivisor) / monthDivisor
	day := dateInt % monthDivisor

	// time.Local matches Python's datetime.strptime().astimezone() (local midnight).
	localTime := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local) //nolint:gosmopolitan

	return localTime.Format(time.RFC3339)
}

// syncUpdateFields returns the fields checked by recordChanged to detect updates.
// rank is intentionally excluded — the UI owns rank after record creation.
func syncUpdateFields() []string {
	return []string{
		"name", "status", "tags", "journal", "scheduled", "deadline",
		"overdue", "backlog_name", "backlog_index", "sort_date", "groomed",
	}
}

const (
	hoursPerDay    = 24
	rankSeedFactor = 1000 // rank is seeded as position × rankSeedFactor on first sync
)

// isOverdue checks if a task is past its scheduled or deadline date.
// Parses RFC3339 date strings (produced by yyyymmddLocalISO) and compares as dates.
func isOverdue(scheduledISO, deadlineISO string, currentTime func() time.Time) bool {
	today := currentTime().Truncate(hoursPerDay * time.Hour)

	return isDateBeforeToday(scheduledISO, today) || isDateBeforeToday(deadlineISO, today)
}

// isDateBeforeToday returns true if the RFC3339 date string is non-empty and before today.
func isDateBeforeToday(dateISO string, today time.Time) bool {
	if dateISO == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, dateISO)
	if err != nil {
		return false
	}

	return t.Truncate(hoursPerDay * time.Hour).Before(today)
}

// determineSortDate picks the best sort date: scheduled > deadline > today (fallback).
func determineSortDate(scheduledISO, deadlineISO, today string) string {
	if scheduledISO != "" {
		return scheduledISO
	}

	if deadlineISO != "" {
		return deadlineISO
	}

	return today
}

// parseGroomedDate extracts the groomed date from task properties.
func parseGroomedDate(task logseqapi.TaskJSON) string {
	if task.PropertiesTextValues == nil {
		return ""
	}

	groomedStr, ok := task.PropertiesTextValues["groomed"]
	if !ok {
		return ""
	}

	groomedTime, _ := logseqext.ParseLogseqDate(groomedStr)
	if groomedTime.IsZero() {
		return ""
	}

	return groomedTime.Format(pocketbase.DateFormat)
}

// extractRankFields returns backlog name, index, and rank from a RankInfo pointer.
func extractRankFields(rank *RankInfo) (string, int, int) {
	if rank == nil {
		return "", 0, 0
	}

	return rank.BacklogName, rank.BacklogIndex, rank.Rank
}

// TaskToRecord converts a TaskJSON + optional RankInfo to a PocketBase record map.
func TaskToRecord(
	task logseqapi.TaskJSON, rank *RankInfo, enrichedTags string, currentTime func() time.Time,
) map[string]any {
	journalISO := yyyymmddToDateOnly(task.Page.JournalDay)
	scheduledISO := yyyymmddToLocalISO(task.Scheduled)
	deadlineISO := yyyymmddToLocalISO(task.Deadline)
	today := currentTime().Format("2006-01-02")
	sortDate := determineSortDate(scheduledISO, deadlineISO, today)
	overdue := isOverdue(scheduledISO, deadlineISO, currentTime)
	groomedISO := parseGroomedDate(task)

	backlogName, backlogIndex, rankValue := extractRankFields(rank)

	return map[string]any{
		"id":            task.UUID,
		"name":          logseqext.CleanTaskName(task.Content, task.Marker),
		"status":        task.Marker,
		"tags":          enrichedTags,
		"journal":       journalISO,
		"scheduled":     scheduledISO,
		"deadline":      deadlineISO,
		"overdue":       overdue,
		"backlog_name":  backlogName,
		"backlog_index": backlogIndex,
		"rank":          rankValue * rankSeedFactor,
		"sort_date":     sortDate,
		"groomed":       groomedISO,
	}
}

// DiffRecords compares existing PB records with desired records.
// Returns slices of records to create, update, and IDs to delete.
func DiffRecords(
	existing, desired []map[string]any,
) ([]map[string]any, []map[string]any, []string) {
	var toCreate, toUpdate []map[string]any

	var toDelete []string

	existingByID := indexRecordsByID(existing)
	desiredByID := indexRecordsByID(desired)

	for recordID, desiredRecord := range desiredByID {
		if existingRecord, exists := existingByID[recordID]; exists {
			if recordChanged(existingRecord, desiredRecord) {
				toUpdate = append(toUpdate, desiredRecord)
			}
		} else {
			toCreate = append(toCreate, desiredRecord)
		}
	}

	for recordID := range existingByID {
		if _, exists := desiredByID[recordID]; !exists {
			toDelete = append(toDelete, recordID)
		}
	}

	return toCreate, toUpdate, toDelete
}

// indexRecordsByID builds a map from record "id" field to the full record.
func indexRecordsByID(records []map[string]any) map[string]map[string]any {
	indexed := make(map[string]map[string]any, len(records))

	for _, record := range records {
		recordID, _ := record["id"].(string)
		indexed[recordID] = record
	}

	return indexed
}

// recordChanged checks if any sync-relevant fields differ between two records.
func recordChanged(existing, desired map[string]any) bool {
	for _, field := range syncUpdateFields() {
		if fmt.Sprint(existing[field]) != fmt.Sprint(desired[field]) {
			return true
		}
	}

	return false
}
