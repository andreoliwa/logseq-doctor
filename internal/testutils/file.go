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

func assertGoldenContent(t *testing.T, graph *logseq.Graph, journals bool, caseDirName string, pages []string) {
	t.Helper()

	// Uses the Windows line ending because the golden file package normalizes line endings to \n
	// Search for GOTESTTOOLS_GOLDEN_NormalizeCRLFToLF in this repo.
	newLine := "\r\n"
	pagesOrJournalsDir := "pages"

	if journals {
		newLine = ""
		pagesOrJournalsDir = "journals"
	}

	for _, page := range pages {
		filename := page + ".md"
		newContents, err := os.ReadFile(filepath.Join(graph.Directory(), pagesOrJournalsDir, filename))
		require.NoError(t, err)

		var subDir string
		if caseDirName != "" {
			subDir = filepath.Join("pages-cases", caseDirName)
		} else {
			subDir = pagesOrJournalsDir
		}

		golden.Assert(t, string(newContents)+newLine, filepath.Join("stub-graph", subDir, filename+".golden"))
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
