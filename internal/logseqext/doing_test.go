package logseqext_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncDoingTasks(t *testing.T) {
	graph, pagesDir := newTestGraph(t)

	err := os.WriteFile(filepath.Join(pagesDir, "doing-source.md"), []byte(`- DOING First task
  id:: 11111111-1111-1111-1111-111111111111
- Parent
  - DOING Second task
    id:: 22222222-2222-2222-2222-222222222222
- TODO Not running
`), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(pagesDir, "doing.md"), []byte(`- Existing note
  - ((22222222-2222-2222-2222-222222222222))
`), 0o600)
	require.NoError(t, err)

	count, err := logseqext.SyncDoingTasks(graph)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	contents, err := os.ReadFile(filepath.Join(pagesDir, "doing.md"))
	require.NoError(t, err)

	expected := "- ((11111111-1111-1111-1111-111111111111))\n" +
		"- Existing note\n\t- ((22222222-2222-2222-2222-222222222222))"
	assert.Equal(t, expected, string(contents))

	before := string(contents)
	count, err = logseqext.SyncDoingTasks(graph)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	contents, err = os.ReadFile(filepath.Join(pagesDir, "doing.md"))
	require.NoError(t, err)
	assert.Equal(t, before, string(contents))
}

func TestSyncDoingTasksInFiles(t *testing.T) {
	graph, pagesDir := newTestGraph(t)
	matchedPath := filepath.Join(pagesDir, "matched.md")

	err := os.WriteFile(matchedPath, []byte(`- DOING Matched task
  id:: 33333333-3333-3333-3333-333333333333
`), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(pagesDir, "not-matched.md"), []byte(`- DOING Unmatched task
  id:: 44444444-4444-4444-4444-444444444444
`), 0o600)
	require.NoError(t, err)

	count, err := logseqext.SyncDoingTasksInFiles(graph, []string{matchedPath})
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	contents, err := os.ReadFile(filepath.Join(pagesDir, "doing.md"))
	require.NoError(t, err)
	assert.Equal(t, "- ((33333333-3333-3333-3333-333333333333))", string(contents))
}

func TestSyncDoingTasksIncludesJournalTasks(t *testing.T) {
	graph, _ := newTestGraph(t)

	err := os.WriteFile(filepath.Join(graph.Directory(), "journals", "2025_01_01.md"), []byte(`- DOING Journal task
  id:: 33333333-3333-3333-3333-333333333333
`), 0o600)
	require.NoError(t, err)

	count, err := logseqext.SyncDoingTasks(graph)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	contents, err := os.ReadFile(filepath.Join(graph.Directory(), "pages", "doing.md"))
	require.NoError(t, err)
	assert.Equal(t, "- ((33333333-3333-3333-3333-333333333333))", string(contents))
}

func TestSyncDoingTasksAssignsID(t *testing.T) {
	graph, pagesDir := newTestGraph(t)

	err := os.WriteFile(filepath.Join(pagesDir, "task-without-id.md"), []byte("- DOING Task without an ID\n"), 0o600)
	require.NoError(t, err)

	count, err := logseqext.SyncDoingTasks(graph)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	taskContents, err := os.ReadFile(filepath.Join(pagesDir, "task-without-id.md"))
	require.NoError(t, err)

	taskID := regexp.MustCompile(`id:: ([0-9a-f-]+)`).FindStringSubmatch(string(taskContents))
	require.Len(t, taskID, 2)
	assert.Equal(t, "- DOING Task without an ID\n  id:: "+taskID[1], string(taskContents))

	doingContents, err := os.ReadFile(filepath.Join(pagesDir, "doing.md"))
	require.NoError(t, err)
	assert.Equal(t, "- (("+taskID[1]+"))", string(doingContents))
}

func newTestGraph(t *testing.T) (*logseq.Graph, string) {
	t.Helper()

	graphPath := t.TempDir()
	pagesDir := filepath.Join(graphPath, "pages")
	require.NoError(t, os.MkdirAll(filepath.Join(graphPath, "logseq"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(graphPath, "journals"), 0o755))
	require.NoError(t, os.MkdirAll(pagesDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(graphPath, "logseq", "config.edn"), []byte("{}"), 0o600))

	graph, err := logseq.Open(context.Background(), graphPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, graph.Close()) })

	return graph, pagesDir
}
