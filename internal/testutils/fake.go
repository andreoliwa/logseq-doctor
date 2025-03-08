package testutils

import (
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/backlog"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
	"path/filepath"
	"testing"
)

// FakeGraph opens the example graph under "testdata" for testing.
func FakeGraph(t *testing.T) *logseq.Graph {
	t.Helper()

	dir, err := filepath.Abs(filepath.Join("testdata", "fake-graph"))
	require.NoError(t, err)

	tempDir := fs.NewDir(t, "append-raw",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(dir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(dir, "journals"))),
		fs.WithDir("pages", fs.FromDir(filepath.Join(dir, "pages"))),
	)

	return internal.OpenGraphFromPath(tempDir.Path())
}

func FakeBacklog(t *testing.T, rootPage string) backlog.Backlog {
	t.Helper()

	graph := FakeGraph(t)
	api := internal.NewLogseqAPI("", "fake-host", "fake-token")
	reader := backlog.NewPageConfigReader(graph, rootPage)

	return backlog.NewBacklog(graph, api, reader)
}
