package testutils

import (
	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
	"os"
	"path/filepath"
	"testing"
)

func AssertPagesDontExist(t *testing.T, graph *logseq.Graph, pages []string) {
	t.Helper()

	for _, page := range pages {
		assert.NoFileExists(t, filepath.Join(graph.Directory(), "pages", page+".md"))
	}
}

func assertGoldenContent(t *testing.T, graph *logseq.Graph, journals bool, pages []string) {
	t.Helper()

	// Uses the Windows line ending because the golden file package normalizes line endings to \n
	// Search for GOTESTTOOLS_GOLDEN_NormalizeCRLFToLF in this repo.
	newLine := "\r\n"
	subdir := "pages"

	if journals {
		newLine = ""
		subdir = "journals"
	}

	for _, page := range pages {
		filename := page + ".md"
		newContents, err := os.ReadFile(filepath.Join(graph.Directory(), subdir, filename))
		require.NoError(t, err)
		golden.Assert(t, string(newContents)+newLine, filepath.Join("stub-graph", subdir, filename+".golden"))
	}
}

func AssertGoldenJournals(t *testing.T, graph *logseq.Graph, pages []string) {
	t.Helper()
	assertGoldenContent(t, graph, true, pages)
}

func AssertGoldenPages(t *testing.T, graph *logseq.Graph, pages []string) {
	t.Helper()
	assertGoldenContent(t, graph, false, pages)
}
