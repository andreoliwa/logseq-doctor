package backlog //nolint:testpackage

import (
	"github.com/andreoliwa/lsd/internal/testutils"
	"strings"
	"testing"
)

func TestProcessBacklog(t *testing.T) {
	graph := testutils.OpenTestGraph(t)
	backlog := &backlogImpl{
		graph:        graph,
		configReader: NewPageConfigReader(graph, "backlog"),
	}

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
				_ = backlog.ProcessAll(test.input) // Ignore error handling for now
			})

			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output %q not found in: %q", test.expected, output)
			}
		})
	}
}
