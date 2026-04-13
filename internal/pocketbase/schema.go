package pocketbase

import (
	"time"

	"github.com/andreoliwa/logseq-go/content"
)

// taskStatusValues are the task status values stored in the PocketBase select field.
//
//nolint:gochecknoglobals // constant derived from logseq-go
var taskStatusValues = []string{
	content.TaskStringTodo, content.TaskStringDoing, content.TaskStringDone,
	content.TaskStringWaiting, content.TaskStringCanceled,
}

// DateFormat is the ISO date format used for PocketBase date/datetime record fields.
const DateFormat = "2006-01-02 15:04:05.000Z"

// FormatDateLocal parses a PocketBase UTC datetime string and returns it in local time
// as "YYYY-MM-DD HH:MM". Returns the raw string if parsing fails.
func FormatDateLocal(utcStr string) string {
	t, err := time.Parse(DateFormat, utcStr)
	if err != nil {
		return utcStr
	}

	return t.Local().Format("2006-01-02 15:04") //nolint:gosmopolitan
}

// idMaxLength is UUID (36) + underscore (1) + backlog name (up to 50) = 87.
const idMaxLength = float64(87)

// LqdTasksSchema returns the PocketBase collection schema for lqd_tasks.
// Go code is the source of truth — not PB migrations.
func LqdTasksSchema() map[string]any {
	return map[string]any{
		"name":   "lqd_tasks",
		"type":   "base",
		"fields": lqdTasksFields(),
	}
}

func lqdTasksFields() []map[string]any {
	return append(lqdTasksIdentityFields(), lqdTasksDataFields()...)
}

func lqdTasksIdentityFields() []map[string]any {
	return []map[string]any{
		{
			"name":    "id",
			"type":    "text",
			"pattern": "^[-a-z0-9_]+$",
			"max":     idMaxLength,
		},
		{
			// task_uuid holds the raw Logseq block UUID so JS can build deep links
			// even after the record id became a composite uuid_backlog key.
			"name": "task_uuid",
			"type": "text",
		},
		{
			"name":     "name",
			"type":     "text",
			"required": true,
		},
		{
			"name":     "status",
			"type":     "select",
			"required": true,
			"values":   taskStatusValues,
		},
	}
}

func lqdTasksDataFields() []map[string]any {
	return []map[string]any{
		{"name": "tags", "type": "text"},
		{"name": "journal", "type": "date"},
		{"name": "scheduled", "type": "date"},
		{"name": "deadline", "type": "date"},
		{"name": "overdue", "type": "bool"},
		{"name": "backlog_name", "type": "text"},
		{"name": "backlog_index", "type": "number"},
		{"name": "section", "type": "number"},
		{"name": "rank", "type": "number"},
		{"name": "sort_date", "type": "date"},
		{"name": "groomed", "type": "date"},
	}
}
