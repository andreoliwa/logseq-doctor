package pocketbase_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLqdTasksSchema_HasRequiredFields(t *testing.T) {
	schema := pocketbase.LqdTasksSchema()

	assert.Equal(t, "lqd_tasks", schema["name"])
	assert.Equal(t, "base", schema["type"])

	fields, ok := schema["fields"].([]map[string]any)
	require.True(t, ok)

	fieldNames := make([]string, 0, len(fields))
	for _, f := range fields {
		name, ok := f["name"].(string)
		require.True(t, ok)

		fieldNames = append(fieldNames, name)
	}

	expectedFields := []string{"name", "status", "tags", "journal", "scheduled", "deadline",
		"overdue", "backlog_name", "backlog_index", "rank", "sort_date", "groomed"}
	for _, expected := range expectedFields {
		assert.Contains(t, fieldNames, expected, "missing field: %s", expected)
	}
}

func TestLqdTasksSchema_IDPattern(t *testing.T) {
	schema := pocketbase.LqdTasksSchema()

	fields, ok := schema["fields"].([]map[string]any)
	require.True(t, ok)

	for _, f := range fields {
		if f["name"] == "id" {
			assert.Equal(t, "^[-a-z0-9]+$", f["pattern"])
			assert.InDelta(t, float64(36), f["max"], 0)

			return
		}
	}

	t.Fatal("id field not found in schema")
}

func TestLqdTasksSchema_StatusValues(t *testing.T) {
	schema := pocketbase.LqdTasksSchema()

	fields, ok := schema["fields"].([]map[string]any)
	require.True(t, ok)

	for _, f := range fields {
		if f["name"] == "status" {
			values, ok := f["values"].([]string)
			require.True(t, ok)
			assert.Contains(t, values, "TODO")
			assert.Contains(t, values, "DOING")
			assert.Contains(t, values, "DONE")
			assert.Contains(t, values, "WAITING")
			assert.Contains(t, values, "CANCELED")

			return
		}
	}

	t.Fatal("status field not found")
}
