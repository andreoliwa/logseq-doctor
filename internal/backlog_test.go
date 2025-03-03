package internal_test

import (
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/andreoliwa/lsd/pkg/utils"
	"strings"
	"testing"
)

func TestProcessBacklog(t *testing.T) {
	backlog := &internal.Backlog{
		Graph: testutils.OpenTestGraph(t),
		FuncProcessSingleBacklog: func(_ *logseq.Graph, _ string,
			_ func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error) {
			return utils.NewSet[string](), nil
		},
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
				_ = backlog.ProcessBacklogs(test.input) // Ignore error handling for now
			})
			fmt.Printf("Captured output:\n%s\n", output)

			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output %q not found in: %q", test.expected, output)
			}
		})
	}
}
