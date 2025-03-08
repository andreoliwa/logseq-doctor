package testutils

import (
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
	"path/filepath"
	"testing"
)

// OpenExampleGraph opens the example graph under "testdata" for testing.
func OpenExampleGraph(t *testing.T) *logseq.Graph {
	t.Helper()

	dir, err := filepath.Abs(filepath.Join("testdata", "example-graph"))
	require.NoError(t, err)

	tempDir := fs.NewDir(t, "append-raw",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(dir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(dir, "journals"))),
		fs.WithDir("pages", fs.FromDir(filepath.Join(dir, "pages"))),
	)

	return internal.OpenGraphFromDirOrEnv(tempDir.Path())
}
