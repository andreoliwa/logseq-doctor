package api_test

import (
	"testing"
	"time"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestBuildRefLookup_PageNames(t *testing.T) {
	tasks := []logseqapi.TaskJSON{
		{
			UUID:    "task-1",
			Content: "TODO Plan trip",
			Page:    logseqapi.PageJSON{ID: 100, OriginalName: "Saturday, 01.01.2025"},
			Refs:    []logseqapi.RefJSON{{ID: 100}},
		},
	}

	lookup := logseqapi.BuildRefLookup(tasks)

	assert.Equal(t, "Saturday, 01.01.2025", lookup[100])
}

func TestBuildRefLookup_FallsBackToName(t *testing.T) {
	tasks := []logseqapi.TaskJSON{
		{
			UUID:    "task-1",
			Content: "TODO Plan",
			Page:    logseqapi.PageJSON{ID: 100, Name: "saturday-01-01-2025"},
		},
	}

	lookup := logseqapi.BuildRefLookup(tasks)

	assert.Equal(t, "saturday-01-01-2025", lookup[100])
}

func TestBuildRefLookup_TagHeuristic(t *testing.T) {
	// Unresolved refs get names from the most frequently co-occurring direct tag.
	tasks := []logseqapi.TaskJSON{
		{
			UUID:    "task-1",
			Content: "TODO #travel Plan trip",
			Page:    logseqapi.PageJSON{ID: 100, OriginalName: "Saturday, 01.01.2025"},
			Refs:    []logseqapi.RefJSON{{ID: 200}},
		},
		{
			UUID:    "task-2",
			Content: "TODO #travel Book hotel",
			Page:    logseqapi.PageJSON{ID: 101, OriginalName: "Sunday, 02.01.2025"},
			Refs:    []logseqapi.RefJSON{{ID: 200}},
		},
	}

	lookup := logseqapi.BuildRefLookup(tasks)

	assert.Equal(t, "travel", lookup[200])
}

func TestBuildRefLookup_EmptyTasks(t *testing.T) {
	lookup := logseqapi.BuildRefLookup(nil)
	assert.Empty(t, lookup)
}

func TestEnrichTasksWithAncestorTags_AddsAncestors(t *testing.T) {
	tasks := []logseqapi.TaskJSON{
		{
			UUID:     "child-task",
			Content:  "TODO Do subtask",
			Page:     logseqapi.PageJSON{ID: 100},
			Refs:     []logseqapi.RefJSON{{ID: 100}},
			PathRefs: []logseqapi.RefJSON{{ID: 100}, {ID: 200}, {ID: 300}},
		},
	}
	refLookup := map[int]string{100: "journal-page", 200: "travel", 300: "planning"}

	tagsByUUID := logseqapi.EnrichTasksWithAncestorTags(tasks, refLookup)

	assert.Contains(t, tagsByUUID["child-task"], "#planning")
	assert.Contains(t, tagsByUUID["child-task"], "#travel")
}

func TestEnrichTasksWithAncestorTags_ExcludesDirectRefs(t *testing.T) {
	tasks := []logseqapi.TaskJSON{
		{
			UUID:     "task-1",
			Content:  "TODO Do task",
			Page:     logseqapi.PageJSON{ID: 100},
			Refs:     []logseqapi.RefJSON{{ID: 100}},
			PathRefs: []logseqapi.RefJSON{{ID: 100}},
		},
	}
	refLookup := map[int]string{100: "journal-page"}

	tagsByUUID := logseqapi.EnrichTasksWithAncestorTags(tasks, refLookup)

	// The page ref is in both Refs and PathRefs, so it's excluded as a direct ref.
	assert.NotContains(t, tagsByUUID["task-1"], "journal-page")
}

func TestEnrichTasksWithAncestorTags_DirectTagsIncluded(t *testing.T) {
	tasks := []logseqapi.TaskJSON{
		{
			UUID:     "task-1",
			Content:  "TODO #mytag Do task",
			Page:     logseqapi.PageJSON{ID: 100},
			PathRefs: []logseqapi.RefJSON{{ID: 100}},
		},
	}
	refLookup := map[int]string{100: "some-page"}

	tagsByUUID := logseqapi.EnrichTasksWithAncestorTags(tasks, refLookup)

	assert.Contains(t, tagsByUUID["task-1"], "#mytag")
}

func TestTaskFutureScheduled(t *testing.T) {
	// Fixed reference: Jan 1, 2025
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	now := func() time.Time { return fixedTime }

	tests := []struct {
		name      string
		scheduled int
		deadline  int
		expected  bool
	}{
		{"future scheduled, no deadline", 20250201, 0, true},
		{"today scheduled (overdue)", 20250101, 0, false},
		{"past scheduled (overdue)", 20241231, 0, false},
		{"no scheduled", 0, 0, false},
		{"future scheduled but overdue deadline", 20250201, 20241231, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := logseqapi.TaskJSON{Scheduled: test.scheduled, Deadline: test.deadline}
			assert.Equal(t, test.expected, logseqapi.TaskFutureScheduled(task, now))
		})
	}
}
