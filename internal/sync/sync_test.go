package lqdsync_test

import (
    "fmt"
    "strings"
    "testing"
    "time"

    logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
    "github.com/andreoliwa/logseq-doctor/internal/logseqext"
    lqdsync "github.com/andreoliwa/logseq-doctor/internal/sync"
    "github.com/stretchr/testify/assert"
)

func TestCalculateRanks_SingleBacklog(t *testing.T) {
    backlogs := map[string][]string{
        "self": {"uuid-1", "uuid-2", "uuid-3"},
    }

    ranks := lqdsync.CalculateRanks(backlogs, []string{"self"})

    assert.Len(t, ranks, 3)
    assert.Equal(t, "self", ranks["uuid-1"].BacklogName)
    assert.Equal(t, 1, ranks["uuid-1"].BacklogIndex)
    assert.Equal(t, 1, ranks["uuid-1"].Rank)
    assert.Equal(t, 3, ranks["uuid-3"].Rank)
}

func TestCalculateRanks_MultipleBacklogs_EarlierOverrides(t *testing.T) {
    backlogs := map[string][]string{
        "self": {"uuid-1", "uuid-shared"},
        "fun":  {"uuid-shared", "uuid-2"},
    }
    order := []string{"self", "fun"}

    ranks := lqdsync.CalculateRanks(backlogs, order)

    assert.Equal(t, "self", ranks["uuid-shared"].BacklogName)
    assert.Equal(t, 1, ranks["uuid-shared"].BacklogIndex)
    assert.Equal(t, 2, ranks["uuid-shared"].Rank)
}

func TestCalculateRanks_Empty(t *testing.T) {
    ranks := lqdsync.CalculateRanks(map[string][]string{}, []string{})
    assert.Empty(t, ranks)
}

func TestExtractDirectTags(t *testing.T) {
    tests := []struct {
        input    string
        expected []string
    }{
        {"TODO #travel Plan trip", []string{"travel"}},
        {"TODO Buy [[groceries]] for #meal-prep", []string{"groceries", "meal-prep"}},
        {"TODO #[[tag with spaces]] and #simple", []string{"tag with spaces", "simple"}},
        {"TODO Check [this link](https://example.com#section)", []string{}},
        {"", nil},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := logseqext.ExtractDirectTags(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestBuildRefLookup(t *testing.T) {
    tasks := []logseqapi.TaskJSON{
        {
            UUID: "task-1", Content: "TODO #travel Plan trip",
            Page: logseqapi.PageJSON{ID: 100, OriginalName: "Saturday, 01.01.2025"},
            Refs: []logseqapi.RefJSON{{ID: 200}},
        },
        {
            UUID: "task-2", Content: "TODO #travel Book hotel",
            Page: logseqapi.PageJSON{ID: 100, OriginalName: "Saturday, 01.01.2025"},
            Refs: []logseqapi.RefJSON{{ID: 200}},
        },
    }

    lookup := logseqapi.BuildRefLookup(tasks)

    assert.Equal(t, "Saturday, 01.01.2025", lookup[100])
    assert.Equal(t, "travel", lookup[200])
}

func TestEnrichTasksWithAncestorTags(t *testing.T) {
    tasks := []logseqapi.TaskJSON{
        {
            UUID: "child-task", Content: "TODO Do subtask",
            Page:     logseqapi.PageJSON{ID: 100},
            Refs:     []logseqapi.RefJSON{{ID: 100}},
            PathRefs: []logseqapi.RefJSON{{ID: 100}, {ID: 200}, {ID: 300}},
        },
    }
    refLookup := map[int]string{100: "journal-page", 200: "travel", 300: "planning"}

    tagsByUUID := logseqapi.EnrichTasksWithAncestorTags(tasks, refLookup)

    assert.Contains(t, tagsByUUID["child-task"], "#planning")
    assert.Contains(t, tagsByUUID["child-task"], "#travel")
    assert.NotContains(t, tagsByUUID["child-task"], "journal-page")
}

func testPageJSON(journalDay int) logseqapi.PageJSON {
    return logseqapi.PageJSON{JournalDay: journalDay}
}

func TestTaskToRecord_Basic(t *testing.T) {
    task := logseqapi.TaskJSON{
        UUID:    "abc-123",
        Marker:  "TODO",
        Content: "TODO Buy groceries\nid:: abc-123",
        Page:    testPageJSON(20250315),
    }

    now := func() time.Time { return time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC) }
    record := lqdsync.TaskToRecord(task, nil, "", now)

    assert.Equal(t, "abc-123", record["id"])
    assert.Equal(t, "Buy groceries", record["name"])
    assert.Equal(t, "TODO", record["status"])
    assert.True(t, strings.HasPrefix(fmt.Sprint(record["journal"]), "2025-03-15"), "journal should be 2025-03-15")
    assert.Equal(t, 0, record["rank"])
    assert.Empty(t, record["backlog_name"])
}

func TestTaskToRecord_WithRankAndDates(t *testing.T) {
    task := logseqapi.TaskJSON{
        UUID:      "def-456",
        Marker:    "DOING",
        Content:   "DOING #travel Plan trip\nid:: def-456",
        Page:      testPageJSON(20250101),
        Scheduled: 20250305,
        Deadline:  20250412,
    }

    rank := &lqdsync.RankInfo{BacklogName: "fun", BacklogIndex: 3, Rank: 5}
    now := func() time.Time { return time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC) }
    record := lqdsync.TaskToRecord(task, rank, "#travel", now)

    assert.Equal(t, "DOING", record["status"])
    assert.Equal(t, "fun", record["backlog_name"])
    assert.Equal(t, 3, record["backlog_index"])
    assert.Equal(t, 5, record["rank"])
    assert.Equal(t, true, record["overdue"])
    assert.True(t, strings.HasPrefix(fmt.Sprint(record["sort_date"]), "2025-03-05"), "sort_date should be 2025-03-05")
}

func TestTaskToRecord_WithGroomed(t *testing.T) {
    task := logseqapi.TaskJSON{
        UUID:    "ghi-789",
        Marker:  "TODO",
        Content: "TODO Some task",
        Page:    testPageJSON(20250101),
        PropertiesTextValues: map[string]string{
            "groomed": "[[Saturday, 21.03.2026]]",
        },
    }

    now := func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) }
    record := lqdsync.TaskToRecord(task, nil, "", now)

    assert.Equal(t, "2026-03-21 00:00:00.000Z", record["groomed"])
}

func TestTaskToRecord_SortDatePrecedence(t *testing.T) {
    now := func() time.Time { return time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC) }

    task1 := logseqapi.TaskJSON{UUID: "a", Marker: "TODO", Content: "TODO x", Page: testPageJSON(20250101)}
    firstRecord := lqdsync.TaskToRecord(task1, nil, "", now)
    // No scheduled/deadline: falls back to today (matching Python's behaviour).
    assert.Equal(t, "2025-04-13", firstRecord["sort_date"])

    task2 := logseqapi.TaskJSON{
        UUID: "b", Marker: "TODO", Content: "TODO x", Page: testPageJSON(20250101), Deadline: 20250601,
    }
    secondRecord := lqdsync.TaskToRecord(task2, nil, "", now)
    assert.True(t, strings.HasPrefix(fmt.Sprint(secondRecord["sort_date"]), "2025-06-01"),
        "sort_date should be 2025-06-01")

    task3 := logseqapi.TaskJSON{
        UUID: "c", Marker: "TODO", Content: "TODO x",
        Page: testPageJSON(20250101), Scheduled: 20250501, Deadline: 20250601,
    }
    thirdRecord := lqdsync.TaskToRecord(task3, nil, "", now)
    assert.True(t,
        strings.HasPrefix(fmt.Sprint(thirdRecord["sort_date"]), "2025-05-01"),
        "sort_date should be 2025-05-01")
}

func TestDiffRecords(t *testing.T) {
    existing := []map[string]any{
        {"id": "keep-same", "name": "Same Task", "status": "TODO"},
        {"id": "will-update", "name": "Old Name", "status": "TODO"},
        {"id": "will-delete", "name": "Gone Task", "status": "TODO"},
    }

    desired := []map[string]any{
        {"id": "keep-same", "name": "Same Task", "status": "TODO"},
        {"id": "will-update", "name": "New Name", "status": "DOING"},
        {"id": "new-task", "name": "Brand New", "status": "TODO"},
    }

    toCreate, toUpdate, toDelete := lqdsync.DiffRecords(existing, desired)

    assert.Len(t, toCreate, 1)
    assert.Equal(t, "new-task", toCreate[0]["id"])

    assert.Len(t, toUpdate, 1)
    assert.Equal(t, "will-update", toUpdate[0]["id"])
    assert.Equal(t, "New Name", toUpdate[0]["name"])

    assert.Len(t, toDelete, 1)
    assert.Equal(t, "will-delete", toDelete[0])
}
