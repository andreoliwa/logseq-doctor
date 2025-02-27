package cmd_test

import (
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreoliwa/lsd/cmd"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
)

func TestAppendRawMarkdownToJournal(t *testing.T) {
	graphDir := filepath.Join("testdata", "graph")
	tempDir := fs.NewDir(t, "append-raw",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(graphDir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(graphDir, "journals"))))

	now := time.Now()
	assert.Equal(t, 0, cmd.AppendRawMarkdownToJournal(tempDir.Path(), now, ""))

	contentToAppend, err := os.ReadFile(filepath.Join("testdata", "append-raw-journal.md"))
	require.NoError(t, err)

	testCase := func(day int, expectedFilename string) func(*testing.T) {
		return func(*testing.T) {
			date := time.Date(2024, 12, day, 0, 0, 0, 0, time.UTC)
			cmd.AppendRawMarkdownToJournal(tempDir.Path(), date, string(contentToAppend))
			modifiedContents, err := os.ReadFile(filepath.Join(tempDir.Path(), "journals",
				expectedFilename+".md"))
			require.NoError(t, err)
			golden.Assert(t, string(modifiedContents), filepath.Join("graph", "journals",
				expectedFilename+".md.golden"))
		}
	}

	t.Run("Journal exists and has content", testCase(24, "2024_12_24"))
	t.Run("Journal doesn't exist", testCase(25, "2024_12_25"))
	t.Run("Journal exists but it's an empty file", testCase(26, "2024_12_25"))
}
