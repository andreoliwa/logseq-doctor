package logseqext_test

import (
	"os"
	"path/filepath"
	"strings"

	"testing"
	"time"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	logseq "github.com/andreoliwa/logseq-go/v2"

	"github.com/andreoliwa/logseq-go/v2/content"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTasksFromJSON(t *testing.T) {
	filePath := filepath.Join("testdata", "sample-api-response.json")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read test JSON file")

	tasks, err := logseqapi.ExtractTasksFromJSON(string(data))
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

func newTask(deadline, scheduled int) logseqapi.TaskJSON {
	return logseqapi.TaskJSON{
		Deadline:  deadline,
		Scheduled: scheduled,
	}
}

func TestOverdue(t *testing.T) {
	currentDate := logseqext.DateYYYYMMDD(time.Now())

	tests := []struct {
		name     string
		task     logseqapi.TaskJSON
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
			assert.Equal(t, test.expected, logseqapi.TaskOverdue(test.task, time.Now), "Overdue check failed for %s", test.name)
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
			task := logseqapi.TaskJSON{Marker: test.marker}
			assert.Equal(t, test.expected, logseqapi.TaskDoing(task), "Doing check failed for %s", test.name)
		})
	}
}

func TestAddTaskToPageOrJournal(t *testing.T) {
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
			graph := testutils.NewStubGraph(t, "stub-graph")

			// Determine the target date
			var targetDate time.Time

			if test.journal != "" {
				parsedDate, err := time.Parse("2006-01-02", test.journal)
				require.NoError(t, err)

				targetDate = parsedDate
			} else {
				targetDate = frozenTime
			}

			opts := &logseqext.AddTaskOptions{
				Graph:     graph,
				Date:      targetDate,
				Page:      test.page,
				BlockText: "",
				Key:       "",
				Name:      test.taskName,
				TimeNow:   func() time.Time { return targetDate },
			}

			err := logseqext.AddTask(opts)
			require.NoError(t, err)

			if test.page != "" {
				testutils.AssertGoldenPages(t, graph, "stub-graph", []string{test.expectedFile})
			} else {
				testutils.AssertGoldenJournals(t, graph, "stub-graph", []string{test.expectedFile})
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
			name:         "--parent provided, and it doesn't exist",
			taskName:     "Clean the room",
			page:         "block-not-found",
			blockText:    "non-existent block",
			expectedFile: "block-not-found",
		},
		{
			name:         "--parent provided, and it exists with children",
			taskName:     "Clean the room",
			page:         "block-with-children",
			blockText:    "Parent block with children",
			expectedFile: "block-with-children",
		},
		{
			name:         "add task under deeply nested block",
			taskName:     "New task",
			page:         "add-nested-task",
			blockText:    "Level 3",
			expectedFile: "add-nested-task",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.NewStubGraph(t, "stub-graph")

			opts := &logseqext.AddTaskOptions{
				Graph:     graph,
				Date:      frozenTime,
				Page:      test.page,
				BlockText: test.blockText,
				Key:       "",
				Name:      test.taskName,
				TimeNow:   func() time.Time { return frozenTime },
			}

			err := logseqext.AddTask(opts)
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, graph, "stub-graph", []string{test.expectedFile})
		})
	}
}

func TestAddOrUpdateTaskByKey(t *testing.T) {
	defaultFrozenTime := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	//nolint:gosmopolitan // Test data uses Local time
	doingTime := time.Date(2025, 11, 1, 19, 29, 57, 0, time.Local)

	tests := []struct {
		name         string
		taskName     string
		page         string
		blockText    string
		key          string
		expectedFile string
		frozenTime   *time.Time // Optional: if nil, uses defaultFrozenTime
	}{
		{
			name:         "--key provided, and it doesn't exist",
			taskName:     "Clean the room",
			page:         "key-not-found",
			blockText:    "",
			key:          "groceries",
			expectedFile: "key-not-found",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "--key provided, and task exists with children, properties and log book",
			taskName:     "Clean the room",
			page:         "key-exists-with-children",
			blockText:    "",
			key:          "groceries",
			expectedFile: "key-exists-with-children",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "--key provided but --parent is not provided",
			taskName:     "Clean the room",
			page:         "key-search-entire-page",
			blockText:    "",
			key:          "groceries",
			expectedFile: "key-search-entire-page",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "--key and --parent provided: search for key within block and its children",
			taskName:     "Clean the room",
			page:         "key-search-within-block",
			blockText:    "Parent block",
			key:          "groceries",
			expectedFile: "key-search-within-block",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "--key and --parent provided, task is deeply nested",
			taskName:     "Clean the room",
			page:         "key-deeply-nested",
			blockText:    "Parent block",
			key:          "groceries",
			expectedFile: "key-deeply-nested",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "add task to a page with filters and aliases",
			taskName:     "My new task",
			page:         "add-task-alias-filters",
			blockText:    "",
			key:          "",
			expectedFile: "add-task-alias-filters",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "update existing task on a page with filters and aliases",
			taskName:     "My updated task with a key",
			page:         "update-task-alias-filters",
			blockText:    "",
			key:          "key",
			expectedFile: "update-task-alias-filters",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "update DOING task",
			taskName:     "My boring chores for this week",
			page:         "update-doing-task",
			blockText:    "",
			key:          "week",
			expectedFile: "update-doing-task",
			frozenTime:   &doingTime,
		},
		{
			name:         "update task with key in a link",
			taskName:     "Updated task name",
			page:         "key-in-link",
			blockText:    "",
			key:          "key inside link title",
			expectedFile: "key-in-link",
			frozenTime:   &defaultFrozenTime,
		},
		{
			name:         "update task with key in the URL of a link",
			taskName:     "Updated task name",
			page:         "key-in-link",
			blockText:    "",
			key:          "key-in-the-url-itself",
			expectedFile: "key-in-link",
			frozenTime:   &defaultFrozenTime,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graph := testutils.NewStubGraph(t, "stub-graph")

			testFrozenTime := defaultFrozenTime
			if test.frozenTime != nil {
				testFrozenTime = *test.frozenTime
			}

			opts := &logseqext.AddTaskOptions{
				Graph:     graph,
				Date:      testFrozenTime,
				Page:      test.page,
				BlockText: test.blockText,
				Key:       test.key,
				Name:      test.taskName,
				TimeNow:   func() time.Time { return testFrozenTime },
			}

			err := logseqext.AddTask(opts)
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, graph, "stub-graph", []string{test.expectedFile})
		})
	}
}

func TestTimedTaskInsertion(t *testing.T) {
	const ref = "((12345678-1234-1234-1234-123456789abc))"

	todo := content.TaskStringTodo
	for _, test := range []struct{ name, input, description, parent, output string }{
		{
			"notes and tasks",
			"- 09:00 Breakfast\n- " +
				todo + " 11:00 Review",
			"**10h00** Call",
			"",
			"- 09:00 Breakfast\n- " +
				todo + " 10:00 Call\n- " +
				todo + " 11:00 Review",
		},
		{
			"equal and untimed",
			"- 10:00 First\n- Notes\n- 10:00 Second\n- 11:00 Later",
			"10:00 Call",
			"",
			"- 10:00 First\n- Notes\n- 10:00 Second\n- " +
				todo + " 10:00 Call\n- 11:00 Later",
		},
		{
			"untimed keeps position",
			"- 09:00 First\n- Notes\n- 11:00 Later",
			"10:00 Call",
			"",
			"- 09:00 First\n- Notes\n- " +
				todo + " 10:00 Call\n- 11:00 Later",
		},
		{"no later anchor", "- 11:00 First\n- Notes", "12:00 Call", "", "- 11:00 First\n- Notes\n- " + todo + " 12:00 Call"},
		{
			"disorder",
			"- 11:00 First\n- 09:00 Earlier\n- 12:00 Last",
			"10:00 Call",
			"",
			"- " +
				todo + " 10:00 Call\n- 11:00 First\n- 09:00 Earlier\n- 12:00 Last",
		},
		{"untimed appends", "- 11:00 First\n- Notes", "Buy milk", "", "- 11:00 First\n- Notes\n- " + todo + " Buy milk"},
		{"invalid appends", "- 11:00 First", "25:00 Invalid", "", "- 11:00 First\n- " + todo + " 25:00 Invalid"},
		{
			"formatted description and untouched anchor",
			"- **11h00** Review",
			"**9h00 Call**",
			"",
			"- " +
				todo + " 09:00 **Call**\n- **11h00** Review",
		},
		{
			"parent scope",
			"- 08:00 Outside\n- Parent\n\t- 09:00 First\n\t- 11:00 Last\n- 09:30 Outside",
			"10:00 Call",
			"Parent",
			"- 08:00 Outside\n- Parent\n\t- 09:00 First\n\t- " +
				todo + " 10:00 Call\n\t- 11:00 Last\n- 09:30 Outside",
		},
		{"missing parent", "- 11:00 Last", "10:00 Call", "Missing", "- " + todo + " 10:00 Call\n- 11:00 Last"},
		{
			"ref anchors outer timeline",
			"- " +
				todo + " 11:25 Review\n- " +
				ref + "\n\t- 12:06 Child\n- " +
				content.TaskStringDone + " 12:27 Lunch",
			"12:00 Call",
			"",
			"- " +
				todo + " 11:25 Review\n- " +
				todo + " 12:00 Call\n- " +
				ref + "\n\t- 12:06 Child\n- " +
				content.TaskStringDone + " 12:27 Lunch",
		},
		{
			"ref is indivisible",
			"- 11:25 Earlier\n- " +
				ref + "\n\t- 12:00 First\n\t- 12:20 Second\n- 12:27 Later",
			"12:10 Call",
			"",
			"- 11:25 Earlier\n- " +
				ref + "\n\t- 12:00 First\n\t- 12:20 Second\n- " +
				todo + " 12:10 Call\n- 12:27 Later",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph, path := testutils.NewMarkdownGraph(t, test.input)
			require.NoError(t, logseqext.AddTask(&logseqext.AddTaskOptions{
				Graph: graph, Page: "timeline",
				Name: test.description, BlockText: test.parent,
			}))

			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, test.output, string(data))
		})
	}
}

func TestTimedKeyUpdates(t *testing.T) {
	todo := content.TaskStringTodo
	for _, test := range []struct{ name, input, description, output string }{
		{
			"earlier to first",
			"- 10:00 First\n- " +
				todo + " 11:00 key\n- 12:00 Last",
			"*9h30 key*",
			"- " +
				todo + " 09:30 *key*\n- 10:00 First\n- 12:00 Last",
		},
		{
			"later to last",
			"- " +
				todo + " 09:00 key\n- 10:00 First\n- Notes\n- 12:00 Last",
			"13:00 key",
			"- 10:00 First\n- Notes\n- 12:00 Last\n- " +
				todo + " 13:00 key",
		},
		{
			"equal",
			"- " +
				todo + " 09:00 key\n- 10:00 First\n- 10:00 Second\n- 12:00 Last",
			"10:00 key",
			"- 10:00 First\n- 10:00 Second\n- " +
				todo + " 10:00 key\n- 12:00 Last",
		},
		{
			"same position",
			"- 09:00 First\n- " +
				todo + " 10:00 key\n- 12:00 Last",
			"**10h00** key",
			"- 09:00 First\n- " +
				todo + " 10:00 key\n- 12:00 Last",
		},
		{
			"remove time in place",
			"- 09:00 First\n- " +
				todo + " 10:00 key\n- 12:00 Last",
			"Buy milk key",
			"- 09:00 First\n- " +
				todo + " Buy milk key\n- 12:00 Last",
		},
		{
			"nested recursive lookup",
			"- 08:00 Outside\n- Parent\n\t- 10:00 First\n\t- " +
				todo + " 11:00 key\n\t- 12:00 Last\n- 14:00 Outside",
			"09:30 key",
			"- 08:00 Outside\n- Parent\n\t- " +
				todo + " 09:30 key\n\t- 10:00 First\n\t- 12:00 Last\n- 14:00 Outside",
		},
		{"single task self excluded", "- " + todo + " 11:00 key", "09:30 key", "- " + todo + " 09:30 key"},
		{
			"anchor not normalized",
			"- **10h00** First\n- " +
				todo + " 11:00 key",
			"09:30 key",
			"- " +
				todo + " 09:30 key\n- **10h00** First",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			graph, path := testutils.NewMarkdownGraph(t, test.input)
			require.NoError(t, logseqext.AddTask(&logseqext.AddTaskOptions{
				Graph: graph, Page: "timeline",
				Key: "KEY", Name: test.description,
			}))

			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, test.output, string(data))
		})
	}
}

func TestTimedKeyUpdateRetainsMetadata(t *testing.T) {
	const blockID = "12345678-1234-1234-1234-123456789abc"

	input := "- 10:00 First\n- " +
		content.TaskStringTodo + " 11:00 key\n  id:: " +
		blockID + "\n" +
		"  custom:: keep\n" +
		"  :LOGBOOK:\n" +
		"  CLOCK: [2025-01-01 Wed 09:00:00]--[2025-01-01 Wed 10:00:00] =>  01:00:00\n" +
		"  :END:\n" +
		"\t- Child [link](https://example.com)\n" +
		"- 12:00 Last"
	graph, path := testutils.NewMarkdownGraph(t, input)
	before, err := graph.OpenPage("timeline")
	require.NoError(t, err)
	original, err := logseq.AsMarkdownWithOptions(before.Blocks()[1], logseq.WithPreservedMarkdown())
	require.NoError(t, err)
	require.NoError(t, logseqext.AddTask(&logseqext.AddTaskOptions{
		Graph: graph, Page: "timeline",
		Key: "key", Name: "*9h30* key",
	}))
	after, err := graph.OpenPage("timeline")
	require.NoError(t, err)

	actual, err := logseq.AsMarkdownWithOptions(after.Blocks()[0], logseq.WithPreservedMarkdown())
	require.NoError(t, err)
	require.Equal(t, strings.Replace(original, "11:00 key", "09:30 key", 1), actual)
	require.Contains(t, actual, "id:: "+blockID)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "custom:: keep")
	require.Contains(t, string(data), "CLOCK:")
	require.Contains(t, string(data), "Child [link](https://example.com)")
}

func TestKeyUpdatePreservesWholeNewDescription(t *testing.T) {
	graph, path := testutils.NewMarkdownGraph(t, "- "+content.TaskStringTodo+" 11:00 key\n\t- Existing child")
	require.NoError(t, logseqext.AddTask(&logseqext.AddTaskOptions{
		Graph: graph, Page: "timeline", Key: "key",
		Name: "**9h00** key\n\nNew paragraph with **detail**",
	}))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	expected := "- " + content.TaskStringTodo + " 09:00 key\n\n" +
		"  New paragraph with **detail**\n\t- Existing child"
	require.Equal(t, expected, string(data))
}
