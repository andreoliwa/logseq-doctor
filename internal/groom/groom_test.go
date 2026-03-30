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

func TestHasRecentDate(t *testing.T) {
	// threshold = 1 year ago from "now" 2026-03-28 → 2025-03-28
	threshold := time.Date(2025, 3, 28, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     map[string]any
		expected bool
	}{
		{"no dates", map[string]any{}, false},
		{"empty scheduled", map[string]any{"scheduled": ""}, false},
		{"empty deadline", map[string]any{"deadline": ""}, false},
		// older than threshold → include in groom
		{"scheduled older than threshold", map[string]any{"scheduled": "2020-01-01 00:00:00.000Z"}, false},
		{"scheduled just before threshold", map[string]any{"scheduled": "2025-03-27 00:00:00.000Z"}, false},
		// at or newer than threshold → exclude from groom
		{"scheduled on threshold day", map[string]any{"scheduled": "2025-03-28 00:00:00.000Z"}, true},
		{"scheduled newer than threshold", map[string]any{"scheduled": "2025-12-01 00:00:00.000Z"}, true},
		{"deadline newer than threshold", map[string]any{"deadline": "2026-06-01 00:00:00.000Z"}, true},
		// any field newer → exclude
		{"old scheduled new deadline", map[string]any{
			"scheduled": "2020-01-01 00:00:00.000Z",
			"deadline":  "2026-06-01 00:00:00.000Z",
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, groom.HasRecentDate(tt.task, threshold))
		})
	}
}

func TestGroomAction_IsValid(t *testing.T) {
	tests := []struct {
		input      string
		hasBacklog bool
		valid      bool
	}{
		{"x", true, true},
		{"f", true, true},
		{"a", true, true},
		{"b", true, true},
		{"c", true, true},
		{"s", true, true},
		{"q", true, true},
		// Case-insensitive
		{"X", true, true},
		{"A", false, true},
		{"B", false, true},
		{"C", false, true},
		// Without backlog: priorities and cancel still available
		{"x", false, true},
		{"a", false, true},
		{"s", false, true},
		{"q", false, true},
		// Focus requires backlog
		{"f", false, false},
		// Invalid keys
		{"k", true, false},
		{"d", true, false},
		{"z", true, false},
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
	assert.Equal(t, "cancel", groom.ParseAction("x", true).Name)
	assert.Equal(t, "focus", groom.ParseAction("f", true).Name)
	assert.Equal(t, "priority-high", groom.ParseAction("a", true).Name)
	assert.Equal(t, "priority-medium", groom.ParseAction("b", true).Name)
	assert.Equal(t, "priority-low", groom.ParseAction("c", true).Name)
	assert.Equal(t, "skip", groom.ParseAction("s", true).Name)
	assert.Equal(t, "quit", groom.ParseAction("q", true).Name)
	assert.Equal(t, "quit", groom.ParseAction("q", false).Name)
}

func TestGroomSummary(t *testing.T) {
	counts := groom.Counts{
		Cancelled:      5,
		Focused:        1,
		PriorityHigh:   2,
		PriorityMedium: 1,
		PriorityLow:    1,
		Skipped:        0,
	}

	output := groom.FormatGroomSummary(counts, 266, "5 years")
	assert.Contains(t, output, "Cancelled:  5")
	assert.Contains(t, output, "Focus:      1")
	assert.Contains(t, output, "High (A):   2")
	assert.Contains(t, output, "Medium (B): 1")
	assert.Contains(t, output, "Low (C):    1")
	assert.Contains(t, output, "Skipped:    0")
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

func TestApplyGroomAction_PriorityHigh_SetsPriorityAndGroomed(t *testing.T) {
	// Verifies that priority action sets [#A] on the block and groomed:: after TODO.
	graph := testutils.NewStubGraph(t, "groom-keep")

	task := map[string]any{"id": "test-block-uuid-0001"}
	groomedDate := "[[Sunday, 22.03.2026]]"
	now := func() time.Time { return time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC) }

	opts := &groom.WriteOpts{CurrentTime: now}

	action := groom.ParseAction("a", false)
	err := groom.ApplyGroomAction(graph, &stubGroomAPI{}, action, task, opts)
	require.NoError(t, err)

	// Read the written journal file from the temp graph directory.
	journalPath := filepath.Join(graph.Directory(), "journals", "2017_03_12.md")
	fileBytes, readErr := os.ReadFile(journalPath)
	require.NoError(t, readErr)

	fileText := string(fileBytes)

	// Verify priority marker was inserted
	assert.Contains(t, fileText, "[#A]")

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
