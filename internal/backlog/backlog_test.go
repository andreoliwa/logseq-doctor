package backlog_test

import (
	"strings"
	"testing"

	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestEmpty(t *testing.T) {
	back := testutils.StubBacklog(t, "non-existent", "", &testutils.StubAPIResponses{}) //nolint: exhaustruct

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
			name:     "processing specific pages",
			input:    []string{"foo", "bar"},
			expected: "Processing pages with partial names: foo, bar\nSkipping focus page because not all pages were processed",
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
	}{
		{
			name:        "empty backlog pages",
			caseDirName: "new-empty",
		},
		{
			name:        "existing backlog with tasks and divider",
			caseDirName: "new-with-divider",
		},
		{
			name:        "existing backlog with tasks and no divider",
			caseDirName: "new-without-divider",
		},
		{
			name:        "existing backlogs have a focus divider",
			caseDirName: "new-with-focus",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			back := testutils.StubBacklog(t, "bk", test.caseDirName, &testutils.StubAPIResponses{
				Queries: []testutils.QueryArg{
					{Contains: "home"},
					{Contains: "phone"},
				},
			})

			pages := []string{"bk___home", "bk___phone"}

			if strings.Contains(test.caseDirName, "empty") {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestFocus(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
	}{
		{
			name:        "empty focus page is created",
			caseDirName: "focus-empty",
		},
		{
			name:        "focus page already exists",
			caseDirName: "focus-exists",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			back := testutils.StubBacklog(t, "bk", test.caseDirName, &testutils.StubAPIResponses{
				Queries: []testutils.QueryArg{
					{Contains: "home"},
					{Contains: "phone"},
				},
			})

			pages := []string{"bk___Focus"}

			if strings.Contains(test.caseDirName, "empty") {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestDeletedTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
	}{
		{
			name:        "deleted root task",
			caseDirName: "deleted-root",
		},
		{
			name:        "deleted nested task",
			caseDirName: "deleted-nested",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			back := testutils.StubBacklog(t, "bk", test.caseDirName, &testutils.StubAPIResponses{
				Queries: []testutils.QueryArg{
					{Contains: "home"},
					{Contains: "phone"},
				},
			})

			pages := []string{"bk___home", "bk___phone"}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}

func TestOverdueTasks(t *testing.T) {
	tests := []struct {
		name        string
		caseDirName string
	}{
		{
			name:        "overdue tasks before new tasks",
			caseDirName: "overdue-before-new",
		},
		{
			name:        "overdue tasks moved from new section",
			caseDirName: "overdue-moved-from-new",
		},
		{
			name:        "overdue tasks appear on top",
			caseDirName: "overdue-on-top",
		},
		{
			name:        "overdue tasks after focus section",
			caseDirName: "overdue-after-focus",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			back := testutils.StubBacklog(t, "ov", test.caseDirName, &testutils.StubAPIResponses{
				Queries: []testutils.QueryArg{
					{Contains: "computer"},
				},
			})

			pages := []string{"ov___computer"}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, back.Graph(), test.caseDirName, pages)
		})
	}
}
