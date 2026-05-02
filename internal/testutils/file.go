package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
)

func AssertPagesDontExist(t *testing.T, graph *logseq.Graph, pages []string) {
	t.Helper()

	for _, page := range pages {
		assert.NoFileExists(t, filepath.Join(graph.Directory(), "pages", page+".md"))
	}
}

func assertGoldenContent(t *testing.T, graph *logseq.Graph, journals bool, caseDirName string, pages []string) {
	t.Helper()

	pagesOrJournalsDir := "pages"

	if journals {
		pagesOrJournalsDir = "journals"
	}

	for _, page := range pages {
		filename := page + ".md"
		newContents, err := os.ReadFile(filepath.Join(graph.Directory(), pagesOrJournalsDir, filename))
		require.NoError(t, err)

		content := string(newContents)

		// The end-of-file-fixer hook in .pre-commit-config.yaml adds a trailing newline to all files,
		// including golden files. The backlog system also ensures files end with a newline when it writes them.
		// However, when the backlog system doesn't modify a file (save=false), the original file is preserved as-is.
		// We need to handle both cases: files that already end with newlines and files that don't.
		// For pages (not journals), ensure content ends with exactly one newline to match
		// what the end-of-file-fixer hook expects in golden files
		if !journals {
			// Remove any existing trailing newlines, then add exactly one
			content = strings.TrimRight(content, "\r\n") + "\r\n"
		}

		var goldenPath string

		if caseDirName != "" {
			goldenPath = filepath.Join(caseDirName, pagesOrJournalsDir, filename+".golden")
		} else {
			goldenPath = filepath.Join("stub-graph", pagesOrJournalsDir, filename+".golden")
		}

		golden.Assert(t, content, goldenPath)
	}
}

func AssertGoldenJournals(t *testing.T, graph *logseq.Graph, caseDirName string, pages []string) {
	t.Helper()
	assertGoldenContent(t, graph, true, caseDirName, pages)
}

func AssertGoldenPages(t *testing.T, graph *logseq.Graph, caseDirName string, pages []string) {
	t.Helper()
	assertGoldenContent(t, graph, false, caseDirName, pages)
}
