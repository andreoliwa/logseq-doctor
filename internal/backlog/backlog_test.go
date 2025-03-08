package backlog_test

import (
	"github.com/andreoliwa/lsd/internal/testutils"
	"strings"
	"testing"
)

func TestBacklogImpl_ProcessEmptyBacklog(t *testing.T) {
	back := testutils.FakeBacklog(t, "non-existent")

	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "All pages processed",
			input:    []string{}, // Empty slice means process all pages
			expected: "Processing all pages in the backlog",
		},
		{
			name:     "Processing specific pages",
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

// TODO: mock API
// func TestBacklogImpl_ProcessAll(t *testing.T) {
//	b := testutils.FakeBacklog(t, "config")
//	err := b.ProcessAll([]string{})
//	require.NoError(t, err)
//}
