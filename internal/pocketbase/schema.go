package pocketbase

// DateFormat is the ISO date format used for PocketBase date/datetime record fields.
const DateFormat = "2006-01-02 15:04:05.000Z"

const idMaxLength = float64(36)

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
	return []map[string]any{
		{
			"name":    "id",
			"type":    "text",
			"pattern": "^[-a-z0-9]+$",
			"max":     idMaxLength,
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
			"values":   []string{"TODO", "DOING", "DONE", "WAITING", "CANCELED"},
		},
		{
			"name": "tags",
			"type": "text",
		},
		{
			"name": "journal",
			"type": "date",
		},
		{
			"name": "scheduled",
			"type": "date",
		},
		{
			"name": "deadline",
			"type": "date",
		},
		{
			"name": "overdue",
			"type": "bool",
		},
		{
			"name": "backlog_name",
			"type": "text",
		},
		{
			"name": "backlog_index",
			"type": "number",
		},
		{
			"name": "rank",
			"type": "number",
		},
		{
			"name": "sort_date",
			"type": "date",
		},
		{
			"name": "groomed",
			"type": "date",
		},
	}
}
