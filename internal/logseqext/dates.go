package logseqext

import (
	"strings"
	"time"
)

// logseqDateFormat is the Go format string for the current graph's journal title format.
// Derived from config.edn `:journal/page-title-format` = "%A, %d.%m.%Y" → "Monday, 02.01.2006".
// TODO: Read this from the graph config via logseq-go's ConvertDateFormat() instead of hardcoding.
const logseqDateFormat = "Monday, 02.01.2006"

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
