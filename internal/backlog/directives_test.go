package backlog_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/require"
)

func directivesFixture(t *testing.T) *testutils.TaskFixture {
	t.Helper()

	todo := content.TaskStringTodo
	waiting := content.TaskStringWaiting

	return testutils.NewFixture(t,
		testutils.Task("task-cancel", todo, "Cancel this task", testutils.WithTags("home")),
		testutils.Task("task-waiting", todo, "Set this task as waiting", testutils.WithTags("home")),
		testutils.Task("task-priority-a", todo, "Set priority A on this task", testutils.WithTags("home")),
		testutils.Task("task-priority-b", todo, "Set priority B on this task", testutils.WithTags("home")),
		testutils.Task("task-priority-c", todo, "Set priority C on this task", testutils.WithTags("home")),
		testutils.Task("task-a-to-b", todo, "Change priority from A to B",
			testutils.WithTags("home"), testutils.WithPriority("A")),
		testutils.Task("task-plain", todo, "Leave this task unchanged", testutils.WithTags("home")),
		testutils.Task("task-waiting-to-todo", waiting, "Change waiting task to todo", testutils.WithTags("home")),
		testutils.Task("task-multi-directive", todo, "Set waiting and priority B", testutils.WithTags("home")),
	)
}

func TestDirectives_StripsDirectivesAndTransformsTasks(t *testing.T) {
	fixture := directivesFixture(t)

	back := fixture.FakeBacklogWithUUIDPages(
		t, "bk", "directives",
		map[string]string{
			"task-cancel":          "home",
			"task-waiting":         "home",
			"task-priority-a":      "home",
			"task-priority-b":      "home",
			"task-priority-c":      "home",
			"task-a-to-b":          "home",
			"task-waiting-to-todo": "home",
			"task-multi-directive": "home",
		},
	)

	err := back.ProcessAll([]string{})
	require.NoError(t, err)

	// Backlog page: all directive prefixes stripped, bare block refs remain.
	fixture.AssertGoldenPages(t, back.Graph(), "directives", []string{"bk___home"})

	// Task source page: each task transformed according to its directive.
	fixture.AssertGoldenPages(t, back.Graph(), "directives", []string{"home"})
}

func TestDirectives_NilAPI_WarnsAndSkips(t *testing.T) {
	// When no Logseq API is available (nil), directives warn and leave the backlog unchanged.
	fixture := directivesFixture(t)

	back := fixture.FakeBacklog(t, "bk", "directives")

	err := back.ProcessAll([]string{})
	require.NoError(t, err)
}
