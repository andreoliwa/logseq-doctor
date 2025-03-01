package testutils

import (
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"gotest.tools/v3/fs"
	"path/filepath"
	"testing"
)

func OpenTestGraph(t *testing.T) *logseq.Graph {
	t.Helper()

	graphDir := filepath.Join("testdata", "graph")
	tempDir := fs.NewDir(t, "append-raw",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(graphDir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(graphDir, "journals"))))

	return internal.OpenGraphFromDirOrEnv(tempDir.Path())
}
