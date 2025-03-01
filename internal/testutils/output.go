package testutils

import (
	"bytes"
	"github.com/fatih/color"
	"io"
	"os"
)

// CaptureOutput captures both stdout and stderr. It also works with the "color" package.
func CaptureOutput(function func()) string {
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
