package groom_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal/groom"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateThresholdDate(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"5 years", time.Date(2021, 3, 21, 0, 0, 0, 0, time.UTC)},
		{"1 year", time.Date(2025, 3, 21, 0, 0, 0, 0, time.UTC)},
		{"6 months", time.Date(2025, 9, 21, 0, 0, 0, 0, time.UTC)},
		{"90 days", time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC)},
		{"1 day", time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := groom.CalculateThresholdDate(now, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateThresholdDate_Invalid(t *testing.T) {
	now := time.Now()

	_, err := groom.CalculateThresholdDate(now, "")
	assert.Error(t, err)
}

func TestBuildGroomFilter(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	thresholdDate := time.Date(2021, 3, 21, 0, 0, 0, 0, time.UTC)

	filter := groom.BuildGroomFilter(now, thresholdDate)

	assert.Contains(t, filter, "status='TODO'")
	assert.Contains(t, filter, "status='WAITING'")
	assert.Contains(t, filter, "journal<'2021-03-21")
	assert.Contains(t, filter, "journal!=''")
	assert.Contains(t, filter, "groomed=''||groomed<'")
	// scheduled/deadline are NOT in the filter: PocketBase null dates don't match `field=''`
	assert.NotContains(t, filter, "scheduled")
	assert.NotContains(t, filter, "deadline")
}

func TestHasFutureDate(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     map[string]any
		expected bool
	}{
		{"no dates", map[string]any{}, false},
		{"empty scheduled", map[string]any{"scheduled": ""}, false},
		// past dates → include in groom (overdue/forgotten)
		{"past scheduled (PB format)", map[string]any{"scheduled": "2023-07-27 00:00:00.000Z"}, false},
		{"today scheduled", map[string]any{"scheduled": "2026-03-21 00:00:00.000Z"}, false},
		// future dates → exclude (still actively planned)
		{"future scheduled (PB format)", map[string]any{"scheduled": "2026-04-01 00:00:00.000Z"}, true},
		{"future scheduled (RFC3339)", map[string]any{"scheduled": "2026-04-01T00:00:00+02:00"}, true},
		{"empty deadline", map[string]any{"deadline": ""}, false},
		{"past deadline", map[string]any{"deadline": "2025-01-01"}, false},
		{"future deadline", map[string]any{"deadline": "2026-12-31"}, true},
		{"past scheduled future deadline", map[string]any{"scheduled": "2025-01-01", "deadline": "2026-12-31"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, groom.HasFutureDate(tt.task, now))
		})
	}
}

func TestGroomAction_IsValid(t *testing.T) {
	tests := []struct {
		input      string
		hasBacklog bool
		valid      bool
	}{
		{"k", true, true},
		{"c", true, true},
		{"f", true, true},
		{"d", true, true},
		{"s", true, true},
		{"k", false, true},
		{"c", false, true},
		{"s", false, true},
		{"q", false, true},
		{"q", true, true},
		{"f", false, false},
		{"d", false, false},
		{"x", true, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_backlog_%v", tt.input, tt.hasBacklog), func(t *testing.T) {
			action := groom.ParseAction(tt.input, tt.hasBacklog)
			if tt.valid {
				assert.NotNil(t, action)
			} else {
				assert.Nil(t, action)
			}
		})
	}
}

func TestGroomAction_Names(t *testing.T) {
	assert.Equal(t, "keep", groom.ParseAction("k", true).Name)
	assert.Equal(t, "cancel", groom.ParseAction("c", true).Name)
	assert.Equal(t, "focus", groom.ParseAction("f", true).Name)
	assert.Equal(t, "defer", groom.ParseAction("d", true).Name)
	assert.Equal(t, "skip", groom.ParseAction("s", true).Name)
	assert.Equal(t, "quit", groom.ParseAction("q", true).Name)
	assert.Equal(t, "quit", groom.ParseAction("q", false).Name)
}

func TestGroomSummary(t *testing.T) {
	counts := groom.Counts{
		Kept:      3,
		Cancelled: 5,
		Focused:   1,
		Deferred:  1,
		Skipped:   0,
	}

	output := groom.FormatGroomSummary(counts, 266, "5 years")
	assert.Contains(t, output, "Kept:      3")
	assert.Contains(t, output, "Cancelled: 5")
	assert.Contains(t, output, "Focus:     1")
	assert.Contains(t, output, "Deferred:  1")
	assert.Contains(t, output, "Skipped:   0")
	assert.Contains(t, output, "older than 5 years: 266")
}

// stubGroomAPI satisfies LogseqAPI for groom write-back tests (no HTTP needed).
// Returns a datascript response pointing to the 2017-03-12 journal page.
type stubGroomAPI struct{}

func (s *stubGroomAPI) PostQuery(_ string) (string, error)                     { return "[]", nil }
func (s *stubGroomAPI) UpsertBlockProperty(_ string, _ string, _ string) error { return nil }
func (s *stubGroomAPI) PostDatascriptQuery(_ string) (string, error) {
	page := `{"id":1,"journal-day":20170312,"original-name":"Sunday, 12.03.2017"}`

	return `[[{"uuid":"test-block-uuid-0001","page":` + page + `}]]`, nil
}

func TestApplyGroomAction_Keep_GroomedPropertyAfterTODO(t *testing.T) {
	// Reproduces: groomed:: appearing before the TODO line instead of after it.
	// Root cause: block.Properties().Set() prepends a new Properties node when the
	// first child is a Paragraph (task blocks always have paragraph first).
	graph := testutils.NewStubGraph(t, "groom-keep")

	task := map[string]any{"id": "test-block-uuid-0001"}
	groomedDate := "[[Sunday, 22.03.2026]]"
	now := func() time.Time { return time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC) }

	opts := &groom.WriteOpts{CurrentTime: now}

	action := groom.ParseAction("k", false)
	err := groom.ApplyGroomAction(graph, &stubGroomAPI{}, action, task, opts)
	require.NoError(t, err)

	// Read the written journal file from the temp graph directory.
	journalPath := filepath.Join(graph.Directory(), "journals", "2017_03_12.md")
	fileBytes, readErr := os.ReadFile(journalPath)
	require.NoError(t, readErr)

	fileText := string(fileBytes)

	// groomed:: must appear AFTER the TODO line, not before it.
	todoIdx := indexOf(fileText, "TODO")
	groomedIdx := indexOf(fileText, "groomed::")

	require.NotEqual(t, -1, todoIdx, "TODO line not found in file")
	require.NotEqual(t, -1, groomedIdx, "groomed:: not found in file")
	assert.Greater(t, groomedIdx, todoIdx, "groomed:: should appear after TODO, not before it")
	assert.Contains(t, fileText, groomedDate)
}

func indexOf(s, substr string) int {
	for i := range s {
		if len(s)-i >= len(substr) && s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}

func TestApplyGroomAction_Cancel(t *testing.T) {
	t.Skip("Implement with test graph fixtures — follow backlog_test.go pattern")
}
