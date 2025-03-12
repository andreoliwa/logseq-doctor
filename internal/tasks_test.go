package internal_test

import (
	"github.com/andreoliwa/lsd/internal"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractTasksFromJSON(t *testing.T) {
	filePath := filepath.Join("testdata", "sample-api-response.json")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read test JSON file")

	tasks, err := internal.ExtractTasksFromJSON(string(data))
	require.NoError(t, err, "Failed to unmarshal JSON")
	assert.NotEmpty(t, tasks, "Expected tasks to be non-empty")

	task := tasks[0]
	assert.NotEmpty(t, task.UUID)
	assert.NotEmpty(t, task.Marker)
	assert.NotEmpty(t, task.Content)
	assert.NotZero(t, task.Page.JournalDay)
	assert.Positive(t, task.Deadline)
	assert.Positive(t, task.Scheduled)
}

func newTask(deadline, scheduled int) internal.TaskJSON {
	return internal.TaskJSON{ //nolint:exhaustruct
		Deadline:  deadline,
		Scheduled: scheduled,
	}
}

func TestOverdue(t *testing.T) {
	currentDate := internal.DateYYYYMMDD(time.Now())

	tests := []struct {
		name     string
		task     internal.TaskJSON
		expected bool
	}{
		{"future task", newTask(currentDate+1, currentDate+2), false},
		{"overdue by deadline", newTask(currentDate-1, 0), true},
		{"overdue by schedule", newTask(0, currentDate-1), true},
		{"overdue by today", newTask(currentDate, 0), true},
		{"overdue by scheduled today", newTask(0, currentDate), true},
		{"no deadline or scheduled", newTask(0, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.task.Overdue(), "Overdue check failed for %s", tt.name)
		})
	}
}
