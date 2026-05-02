package backlog_test

import (
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/require"
)

func homePhoneFixture(t *testing.T) *testutils.TaskFixture {
	t.Helper()

	todo := content.TaskStringTodo

	return testutils.NewFixture(t,
		testutils.Task("home-clean-windows", todo, "Clean windows before spring", testutils.WithTags("home")),
		testutils.Task("home-vacuum-carpets", todo, "Vacuum carpets and couch", testutils.WithTags("home")),
		testutils.Task("home-wash-bed-sheets", todo, "Wash and replace bed sheets", testutils.WithTags("home")),
		testutils.Task("home-clean-basement", todo, "Clean up the basement", testutils.WithTags("home")),
		testutils.Task("phone-backup-photos", todo, "Backup my mobile photos to the cloud",
			testutils.WithTags("phone")),
		testutils.Task("phone-remove-apps", todo, "Remove unused apps to save space", testutils.WithTags("phone")),
		testutils.Task("phone-change-provider", todo, "Change mobile phone provider to a cheaper one",
			testutils.WithTags("phone")),
	)
}

func kitchenWorkFixture(t *testing.T) *testutils.TaskFixture {
	t.Helper()

	kitchen := testutils.WithTags("kitchen")
	work := testutils.WithTags("work")

	todo := content.TaskStringTodo

	return testutils.NewFixture(t,
		testutils.Task("kitchen-supermarket", todo, "Go to the supermarket and refill fridge",
			kitchen),
		testutils.Task("kitchen-unfreeze-meat", todo, "Unfreeze meat for my recipe",
			kitchen),
		testutils.Task("kitchen-clean-oven", todo, "Clean the oven (future deadline)",
			kitchen, testutils.WithDeadline("+39d")),
		testutils.Task("kitchen-buy-frying-pan", todo, "Buy a new frying pan",
			kitchen),
		testutils.Task("kitchen-plan-menu", todo, "Plan summer menu",
			kitchen, testutils.WithScheduled("+32d")),
		testutils.Task("kitchen-organize-spice", todo, "Organize spice cabinet",
			kitchen, testutils.WithScheduled("+68d")),
		testutils.Task("work-monthly-report", todo, "Prepare monthly report",
			work),
		testutils.Task("work-tidy-papers", todo, "Tidy up office papers",
			work, testutils.WithDeadline("+26d")),
		testutils.Task("work-quarterly-review", todo, "Prepare quarterly review",
			work, testutils.WithScheduled("+93d")),
		testutils.Task("work-team-building", todo, "Plan team building event",
			work, testutils.WithScheduled("+134d")),
	)
}

func healthFixture(t *testing.T) *testutils.TaskFixture {
	t.Helper()

	return testutils.NewFixture(t,
		testutils.Task("health-doctor", content.TaskStringTodo, "Scheduled appointment with doctor",
			testutils.WithTags("health")),
		testutils.Task("health-dentist", content.TaskStringTodo, "Go to the dentist", testutils.WithTags("health")),
		testutils.Task("health-doing", content.TaskStringDoing, "Taking care of my health",
			testutils.WithTags("health")),
	)
}

func computerFixture(t *testing.T) *testutils.TaskFixture {
	t.Helper()

	todo := content.TaskStringTodo

	return testutils.NewFixture(t,
		testutils.Task("computer-backup-files", todo, "Backup my files",
			testutils.WithTags("computer"), testutils.WithScheduled("-39d")),
		testutils.Task("computer-delete-big-files", todo, "Delete big files taking up space",
			testutils.WithTags("computer"), testutils.WithDeadline("-1d")),
		testutils.Task("computer-clean-desktop", todo, "Clean up desktop files",
			testutils.WithTags("computer")),
		testutils.Task("computer-buy-laptop", todo, "Buy a new laptop in a few years",
			testutils.WithTags("computer"), testutils.WithDeadline("+1084d")),
	)
}

func TestEmpty(t *testing.T) {
	back := testutils.NewFixture(t).FakeBacklog(t, "non-existent", "")

	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "all pages processed",
			input:    []string{}, // Empty slice means process all pages
			expected: "Processing all pages in the backlog",
		},
		{
			name:  "processing specific pages",
			input: []string{"foo", "bar"},
			expected: "Processing pages with partial names: foo, bar\n" +
				"Skipping focus page because not all pages were processed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := testutils.CaptureOutput(func() {
				_ = back.ProcessAll(test.input) // Ignore error handling for now
			})

			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output %q not found in: %q", test.expected, output)
			}
		})
	}
}

func TestNewTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
		pagesExist  bool
	}{
		{name: "empty backlog pages", caseDirName: "new-empty-backlog", pagesExist: false},
		{name: "existing backlog with tasks and divider", caseDirName: "new-with-divider", pagesExist: true},
		{name: "existing backlog with tasks and no divider", caseDirName: "new-without-divider", pagesExist: true},
		{name: "existing backlogs have a focus divider", caseDirName: "new-with-focus", pagesExist: true},
		{name: "remove empty divider", caseDirName: "new-remove-empty-divider", pagesExist: true},
		{name: "new tasks inserted before unranked divider", caseDirName: "new-before-unranked", pagesExist: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := homePhoneFixture(t)
			back := fixture.FakeBacklog(t, "bk", test.caseDirName)
			pages := []string{"bk___home", "bk___phone"}

			if !test.pagesExist {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			fixture.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestFocus(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
		pagesExist  bool
	}{
		{name: "empty focus page is created", caseDirName: "focus-empty", pagesExist: false},
		{name: "focus page already exists", caseDirName: "focus-exists", pagesExist: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := homePhoneFixture(t)
			back := fixture.FakeBacklog(t, "bk", test.caseDirName)
			pages := []string{"bk___Focus"}

			if !test.pagesExist {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			fixture.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestDeletedTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
	}{
		{name: "deleted root task", caseDirName: "deleted-root"},
		{name: "deleted nested task", caseDirName: "deleted-nested"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := homePhoneFixture(t)
			back := fixture.FakeBacklog(t, "bk", test.caseDirName)
			pages := []string{"bk___home", "bk___phone"}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			fixture.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestOverdueTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
	}{
		{name: "overdue tasks before new tasks", caseDirName: "overdue-before-new"},
		{name: "overdue tasks moved from new section", caseDirName: "overdue-moved-from-new"},
		{name: "overdue tasks appear on top", caseDirName: "overdue-on-top"},
		{name: "overdue tasks after focus section", caseDirName: "overdue-after-focus"},
		{name: "overdue tasks moved to existing divider", caseDirName: "overdue-divider"},
		{name: "pinned overdue tasks should not be touched", caseDirName: "overdue-pinned"},
		{name: "remove empty divider", caseDirName: "overdue-remove-empty-divider"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := computerFixture(t)
			back := fixture.FakeBacklog(t, "ov", test.caseDirName)
			pages := []string{"ov___computer"}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			fixture.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestFutureScheduledTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
	}{
		{name: "existing scheduled divider", caseDirName: "scheduled-existing-divider"},
		{name: "non-existing scheduled divider", caseDirName: "scheduled-non-existing-divider"},
		{
			name:        "existing future scheduled task moved to scheduled divider",
			caseDirName: "scheduled-existing-task-moved",
		},
		{name: "new scheduled task added directly to scheduled divider", caseDirName: "scheduled-new-task-direct"},
		{
			name:        "task without scheduled date moved out of scheduled divider to new tasks",
			caseDirName: "scheduled-unscheduled-task",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := kitchenWorkFixture(t)
			back := fixture.FakeBacklog(t, "sch", test.caseDirName)
			pages := []string{"sch___kitchen", "sch___work"}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			fixture.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestTriagedDedup(t *testing.T) {
	fixture := homePhoneFixture(t)
	back := fixture.FakeBacklog(t, "bk", "triaged-dedup")
	pages := []string{"bk___home", "bk___phone"}

	err := back.ProcessAll([]string{})
	require.NoError(t, err)

	fixture.AssertGoldenPages(t, back.Graph(), "triaged-dedup", pages)
}

func TestDoingTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
		pagesExist  bool
	}{
		{name: "DOING tasks not added to empty page", caseDirName: "doing-not-added-empty", pagesExist: false},
		{name: "DOING tasks not added to existing page", caseDirName: "doing-not-added-existing", pagesExist: true},
		{name: "DOING tasks preserved in existing page", caseDirName: "doing-preserved", pagesExist: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := healthFixture(t)
			back := fixture.FakeBacklog(t, "dt", test.caseDirName)
			pages := []string{"dt___health"}

			if !test.pagesExist {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			fixture.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}
