package logseqext

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// logseqDateFormat is the Go format string for the current graph's journal title format.
// Derived from config.edn `:journal/page-title-format` = "EEEE, dd.MM.yyyy" → "Monday, 02.01.2006".
const logseqDateFormat = "Monday, 02.01.2006"

// journalTitleFormatRe matches the :journal/page-title-format line in config.edn.
var journalTitleFormatRe = regexp.MustCompile(`(?m):journal/page-title-format\s+"([^"]+)"`)

// ReadJournalTitleFormat reads the JS-style date format string used for journal
// page titles from logseq/config.edn (e.g. "EEEE, dd.MM.yyyy").
// Returns the Logseq default "EEE do, MMM yyyy" if the file cannot be read or
// the key is absent. This is a candidate for upstreaming to logseq-go.
func ReadJournalTitleFormat(graphPath string) string {
	const defaultFormat = "EEE do, MMM yyyy"

	if graphPath == "" {
		return defaultFormat
	}

	data, err := os.ReadFile(filepath.Join(graphPath, "logseq", "config.edn"))
	if err != nil {
		return defaultFormat
	}

	matches := journalTitleFormatRe.FindSubmatch(data)
	if len(matches) < 2 { //nolint:mnd
		return defaultFormat
	}

	return string(matches[1])
}

// DateYYYYMMDD returns the current date in YYYYMMDD format.
func DateYYYYMMDD(time time.Time) int {
	currentDate := time.Year()*10000 + int(time.Month())*100 + time.Day()

	return currentDate
}

// ParseLogseqDate parses a Logseq date string like "[[Saturday, 21.03.2026]]" into a time.Time.
// Returns zero time (not error) for empty or unparseable strings.
func ParseLogseqDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	// Strip [[ and ]]
	dateStr = strings.TrimPrefix(dateStr, "[[")
	dateStr = strings.TrimSuffix(dateStr, "]]")

	// Parse returns zero time on failure; parse errors are intentionally ignored per spec.
	parsed, _ := time.Parse(logseqDateFormat, dateStr)

	return parsed, nil
}

// FormatLogseqDate formats a time.Time as a Logseq date string: "[[Saturday, 21.03.2026]]".
func FormatLogseqDate(t time.Time) string {
	return "[[" + t.Format(logseqDateFormat) + "]]"
}

// JournalDayToTime converts a Logseq journalDay integer (YYYYMMDD) to a time.Time.
// Returns zero time for zero input.
func JournalDayToTime(journalDay int) time.Time {
	if journalDay == 0 {
		return time.Time{}
	}

	year := journalDay / JournalDayDivisorYear
	month := (journalDay % JournalDayDivisorYear) / JournalDayDivisorMonth
	day := journalDay % JournalDayDivisorMonth

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
