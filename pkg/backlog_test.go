package pkg_test

import (
	"bytes"
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/pkg"
	"github.com/fatih/color"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures both stdout and stderr.
func captureOutput(function func()) string {
	// Create pipes
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	read, write, _ := os.Pipe()

	// Set stdout and stderr to the pipe
	os.Stdout = write
	os.Stderr = write

	// Disable color to avoid ANSI escape sequences in captured output
	color.NoColor = true
	color.Output = os.Stderr

	// Create a channel to read output asynchronously
	outC := make(chan string)

	// Start a goroutine to read output
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, read)
		outC <- buf.String()
	}()

	// Run the function
	function()

	// Close the writer to signal EOF
	_ = write.Close()

	// Restore stdout and stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Return captured output
	return <-outC
}

func TestProcessBacklog(t *testing.T) {
	mockGraph := &logseq.Graph{}
	processor := &pkg.Backlog{
		Graph: mockGraph,
		FuncProcessSingleBacklog: func(_ *logseq.Graph, _ string,
			_ func() (*pkg.Set[string], error)) (*pkg.Set[string], error) {
			return pkg.NewSet[string](), nil
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
			output := captureOutput(func() {
				_ = processor.ProcessBacklogs(test.input) // Ignore error handling for now
			})
			fmt.Printf("Captured output:\n%s\n", output)

			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output %q not found in: %q", test.expected, output)
			}
		})
	}
}
