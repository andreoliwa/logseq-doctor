package backlog_test

import (
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestEmpty(t *testing.T) {
	back := testutils.StubBacklog(t, "non-existent", &testutils.StubAPIResponses{}) //nolint: exhaustruct

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
		name     string
		rootPage string
	}{
		{
			name:     "empty backlog pages",
			rootPage: "new-empty",
		},
		{
			name:     "existing backlog with tasks and divider",
			rootPage: "new-with-divider",
		},
		{
			name:     "existing backlog with tasks and no divider",
			rootPage: "new-without-divider",
		},
		{
			name:     "existing backlogs have a focus divider",
			rootPage: "new-with-focus",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			back := testutils.StubBacklog(t, test.rootPage, &testutils.StubAPIResponses{
				Queries: []testutils.QueryArg{
					{Contains: "home"},
					{Contains: "phone"},
				},
			})

			pages := []string{test.rootPage + "___home", test.rootPage + "___phone"}

			if strings.Contains(test.rootPage, "empty") {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, back.Graph(), pages)
		})
	}
}

func TestFocus(t *testing.T) {
	tests := []struct {
		name     string
		rootPage string
	}{
		{
			name:     "empty focus page is created",
			rootPage: "focus-empty",
		},
		{
			name:     "focus page already exists",
			rootPage: "focus-exists",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			back := testutils.StubBacklog(t, test.rootPage, &testutils.StubAPIResponses{
				Queries: []testutils.QueryArg{
					{Contains: "home"},
					{Contains: "phone"},
				},
			})

			pages := []string{test.rootPage + "___Focus"}

			if strings.Contains(test.rootPage, "empty") {
				testutils.AssertPagesDontExist(t, back.Graph(), pages)
			}

			err := back.ProcessAll([]string{})
			require.NoError(t, err)

			testutils.AssertGoldenPages(t, back.Graph(), pages)
		})
	}
}
