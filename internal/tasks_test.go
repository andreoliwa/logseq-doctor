package internal_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.task.Overdue(time.Now), "Overdue check failed for %s", test.name)
		})
	}
}

func TestDoing(t *testing.T) {
	tests := []struct {
		name     string
		marker   string
		expected bool
	}{
		{"DOING task", "DOING", true},
		{"TODO task", "TODO", false},
		{"DONE task", "DONE", false},
		{"LATER task", "LATER", false},
		{"NOW task", "NOW", false},
		{"WAITING task", "WAITING", false},
		{"empty marker", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := internal.TaskJSON{Marker: test.marker} //nolint:exhaustruct
			assert.Equal(t, test.expected, task.Doing(), "Doing check failed for %s", test.name)
		})
	}
}

func TestAddTaskToPageOrJournal(t *testing.T) { //nolint:funlen
	frozenTime := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		taskName     string
		page         string
		journal      string
		expectedFile string
	}{
		{
			name:         "simple task with no flags (use frozen date of Jan 4th 2025)",
			taskName:     "Clean the room",
			page:         "",
			journal:      "",
			expectedFile: "2025_01_04",
		},
		{
			name:         "provided --page exists",
			taskName:     "Clean the room",
			page:         "add-task",
			journal:      "",
			expectedFile: "add-task",
		},
		{
			name:         "provided --page doesn't exist",
			taskName:     "Clean the room",
			page:         "non-existent-page",
			journal:      "",
			expectedFile: "non-existent-page",
		},
		{
			name:         "valid --journal 2025-01-05 provided",
			taskName:     "Clean the room",
			page:         "",
			journal:      "2025-01-05",
			expectedFile: "2025_01_05",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.StubGraph(t, "")

			// Determine the target date
			var targetDate time.Time
			if test.journal != "" {
				parsedDate, err := time.Parse("2006-01-02", test.journal)
				require.NoError(t, err)

				targetDate = parsedDate
			} else {
				targetDate = frozenTime
			}

			opts := &internal.AddTaskOptions{
				Graph:     graph,
				Date:      targetDate, // TODO: accept a raw date string and parse it inside the AddTask function
				Page:      test.page,
				BlockText: "",
				Key:       "",
				Name:      test.taskName,
			}

			err := internal.AddTask(opts)
			require.NoError(t, err)

			if test.page != "" {
				testutils.AssertGoldenPages(t, graph, "", []string{test.expectedFile})
			} else {
				testutils.AssertGoldenJournals(t, graph, "", []string{test.expectedFile})
			}
		})
	}
}

func TestAddTaskUnderBlock(t *testing.T) {
	frozenTime := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		taskName     string
		page         string
		blockText    string
		expectedFile string
	}{
		{
			name:         "--block provided, and it doesn't exist",
			taskName:     "Clean the room",
			page:         "block-not-found",
			blockText:    "non-existent block",
			expectedFile: "block-not-found",
		},
		{
			name:         "--block provided, and it exists with children",
			taskName:     "Clean the room",
			page:         "block-with-children",
			blockText:    "Parent block with children",
			expectedFile: "block-with-children",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.StubGraph(t, "")

			opts := &internal.AddTaskOptions{
				Graph:     graph,
				Date:      frozenTime,
				Page:      test.page,
				BlockText: test.blockText,
				Key:       "",
				Name:      test.taskName,
			}

			err := internal.AddTask(opts)
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, graph, "", []string{test.expectedFile})
		})
	}
}

func TestAddOrUpdateTaskByKey(t *testing.T) { //nolint:funlen
	frozenTime := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		taskName     string
		page         string
		blockText    string
		key          string
		expectedFile string
	}{
		{
			name:         "--key provided, and it doesn't exist",
			taskName:     "Clean the room",
			page:         "key-not-found",
			blockText:    "",
			key:          "groceries",
			expectedFile: "key-not-found",
		},
		{
			name:         "--key provided, and task exists with children, properties and log book",
			taskName:     "Clean the room",
			page:         "key-exists-with-children",
			blockText:    "",
			key:          "groceries",
			expectedFile: "key-exists-with-children",
		},
		{
			name:         "--key provided but --block is not provided",
			taskName:     "Clean the room",
			page:         "key-search-entire-page",
			blockText:    "",
			key:          "groceries",
			expectedFile: "key-search-entire-page",
		},
		{
			name:         "--key and --block provided: search for key within block and its children",
			taskName:     "Clean the room",
			page:         "key-search-within-block",
			blockText:    "Parent block",
			key:          "groceries",
			expectedFile: "key-search-within-block",
		},
		{
			name:         "--key and --block provided, task is deeply nested",
			taskName:     "Clean the room",
			page:         "key-deeply-nested",
			blockText:    "Parent block",
			key:          "groceries",
			expectedFile: "key-deeply-nested",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.StubGraph(t, "")

			opts := &internal.AddTaskOptions{
				Graph:     graph,
				Date:      frozenTime,
				Page:      test.page,
				BlockText: test.blockText,
				Key:       test.key,
				Name:      test.taskName,
			}

			err := internal.AddTask(opts)
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, graph, "", []string{test.expectedFile})
		})
	}
}
