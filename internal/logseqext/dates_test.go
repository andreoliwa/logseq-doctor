package logseqext_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateYYYYMMDD(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected int
	}{
		{"Jan 1, 2025", time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), 20250101},
		{"Feb 18, 2025", time.Date(2025, time.February, 18, 0, 0, 0, 0, time.UTC), 20250218},
		{"Dec 31, 2023", time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC), 20231231},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := logseqext.DateYYYYMMDD(test.input)
			assert.Equal(t, test.expected, result, "Date conversion failed for %s", test.name)
		})
	}
}

func TestParseLogseqDate_ValidFormat(t *testing.T) {
	result, err := logseqext.ParseLogseqDate("[[Saturday, 21.03.2026]]")
	require.NoError(t, err)
	assert.Equal(t, 2026, result.Year())
	assert.Equal(t, time.March, result.Month())
	assert.Equal(t, 21, result.Day())
}

func TestParseLogseqDate_WithoutBrackets(t *testing.T) {
	result, err := logseqext.ParseLogseqDate("Saturday, 21.03.2026")
	require.NoError(t, err)
	assert.Equal(t, 2026, result.Year())
}

func TestParseLogseqDate_EmptyString(t *testing.T) {
	result, err := logseqext.ParseLogseqDate("")
	require.NoError(t, err)
	assert.True(t, result.IsZero())
}

func TestParseLogseqDate_InvalidFormat(t *testing.T) {
	result, err := logseqext.ParseLogseqDate("[[not-a-date]]")
	require.NoError(t, err)
	assert.True(t, result.IsZero())
}

func TestFormatLogseqDate(t *testing.T) {
	date := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	result := logseqext.FormatLogseqDate(date)
	assert.Equal(t, "[[Saturday, 21.03.2026]]", result)
}

func TestFormatLogseqDate_DifferentDay(t *testing.T) {
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	result := logseqext.FormatLogseqDate(date)
	assert.Equal(t, "[[Monday, 06.01.2025]]", result)
}
