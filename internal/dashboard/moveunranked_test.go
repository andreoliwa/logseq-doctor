package dashboard_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/dashboard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestGraph creates a minimal Logseq graph directory in a temp dir.
// It writes a logseq/config.edn and a pages/<pageName>.md with the given content.
func makeTestGraph(t *testing.T, pageName, pageContent string) string {
	t.Helper()

	graphDir := t.TempDir()

	logseqDir := filepath.Join(graphDir, "logseq")
	pagesDir := filepath.Join(graphDir, "pages")

	require.NoError(t, os.MkdirAll(logseqDir, 0o755))
	require.NoError(t, os.MkdirAll(pagesDir, 0o755))

	// Minimal config.edn required by logseq-go
	require.NoError(t, os.WriteFile(filepath.Join(logseqDir, "config.edn"), []byte("{}\n"), 0o600))

	// The backlog page
	require.NoError(t, os.WriteFile(filepath.Join(pagesDir, pageName+".md"), []byte(pageContent), 0o600))

	return graphDir
}

const (
	uuid1 = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"
	uuid2 = "b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2"
	uuid3 = "c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3"
)

func TestMoveToUnrankedDoesNothingForEmptyUUIDs(t *testing.T) {
	pageContent := "- ((" + uuid1 + "))\n- ((" + uuid2 + "))\n"
	graphDir := makeTestGraph(t, "my-backlog", pageContent)

	err := dashboard.MoveToUnranked(graphDir, "my-backlog", nil)
	require.NoError(t, err)

	// File unchanged
	got, err := os.ReadFile(filepath.Join(graphDir, "pages", "my-backlog.md"))

	require.NoError(t, err)
	assert.Contains(t, string(got), uuid1)
	assert.Contains(t, string(got), uuid2)
}

func TestMoveToUnrankedMovesTasksAndCreatesHeader(t *testing.T) {
	// Page has three ranked tasks. We move the last two to unranked.
	pageContent := "- ((" + uuid1 + "))\n- ((" + uuid2 + "))\n- ((" + uuid3 + "))\n"
	graphDir := makeTestGraph(t, "my-backlog", pageContent)

	err := dashboard.MoveToUnranked(graphDir, "my-backlog", []string{uuid2, uuid3})
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(graphDir, "pages", "my-backlog.md"))

	require.NoError(t, err)

	resultStr := string(result)

	// The unranked divider must appear.
	assert.Contains(t, resultStr, "🔢 Unranked tasks", "unranked divider should be created")

	// All tasks must still be present.
	assert.Contains(t, resultStr, uuid1)
	assert.Contains(t, resultStr, uuid2)
	assert.Contains(t, resultStr, uuid3)

	// The unranked divider must appear before uuid2 and uuid3
	// (they are children of the divider block).
	dividerIdx := strings.Index(resultStr, "🔢 Unranked tasks")
	task2Idx := strings.Index(resultStr, uuid2)
	task3Idx := strings.Index(resultStr, uuid3)

	assert.Less(t, dividerIdx, task2Idx, "divider should come before uuid2")
	assert.Less(t, dividerIdx, task3Idx, "divider should come before uuid3")
}

func TestMoveToUnrankedPreservesExistingDivider(t *testing.T) {
	// Page already has an unranked divider with one task. We add another.
	pageContent := "- ((" + uuid1 + "))\n- ((" + uuid2 + "))\n- 🔢 Unranked tasks\n\t- ((" + uuid3 + "))\n"
	graphDir := makeTestGraph(t, "my-backlog", pageContent)

	err := dashboard.MoveToUnranked(graphDir, "my-backlog", []string{uuid2})
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(graphDir, "pages", "my-backlog.md"))

	require.NoError(t, err)

	resultStr := string(result)

	// Only one divider header should exist.
	assert.Equal(t, 1, strings.Count(resultStr, "🔢 Unranked tasks"))

	// All three tasks must still be present.
	assert.Contains(t, resultStr, uuid1)
	assert.Contains(t, resultStr, uuid2)
	assert.Contains(t, resultStr, uuid3)
}
