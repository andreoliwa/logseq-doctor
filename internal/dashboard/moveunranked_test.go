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

const testBacklogPage = "my-backlog"

// makeTestGraph creates a minimal Logseq graph directory in a temp dir.
// It writes logseq/config.edn and pages/my-backlog.md with the given content.
func makeTestGraph(t *testing.T, pageContent string) string {
	t.Helper()

	graphDir := t.TempDir()

	logseqDir := filepath.Join(graphDir, "logseq")
	pagesDir := filepath.Join(graphDir, "pages")

	require.NoError(t, os.MkdirAll(logseqDir, 0o755))
	require.NoError(t, os.MkdirAll(pagesDir, 0o755))

	// Minimal config.edn required by logseq-go
	require.NoError(t, os.WriteFile(filepath.Join(logseqDir, "config.edn"), []byte("{}\n"), 0o600))

	// The backlog page
	require.NoError(t, os.WriteFile(filepath.Join(pagesDir, testBacklogPage+".md"), []byte(pageContent), 0o600))

	return graphDir
}

const (
	uuid1 = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"
	uuid2 = "b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2"
	uuid3 = "c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3"
)

func TestMoveToUnrankedDoesNothingForEmptyUUIDs(t *testing.T) {
	pageContent := "- ((" + uuid1 + "))\n- ((" + uuid2 + "))\n"
	graphDir := makeTestGraph(t, pageContent)

	err := dashboard.MoveToUnranked(graphDir, testBacklogPage, nil)
	require.NoError(t, err)

	// File unchanged
	got, err := os.ReadFile(filepath.Join(graphDir, "pages", testBacklogPage+".md"))

	require.NoError(t, err)
	assert.Contains(t, string(got), uuid1)
	assert.Contains(t, string(got), uuid2)
}

func TestMoveToUnrankedMovesTasksAndCreatesHeader(t *testing.T) {
	// Page has three ranked tasks. We move the last two to unranked.
	pageContent := "- ((" + uuid1 + "))\n- ((" + uuid2 + "))\n- ((" + uuid3 + "))\n"
	graphDir := makeTestGraph(t, pageContent)

	err := dashboard.MoveToUnranked(graphDir, testBacklogPage, []string{uuid2, uuid3})
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(graphDir, "pages", testBacklogPage+".md"))

	require.NoError(t, err)

	resultStr := string(result)

	// The unranked divider must appear.
	assert.Contains(t, resultStr, "⤵️ Unranked tasks", "unranked divider should be created")

	// All tasks must still be present.
	assert.Contains(t, resultStr, uuid1)
	assert.Contains(t, resultStr, uuid2)
	assert.Contains(t, resultStr, uuid3)

	// The unranked divider must appear before uuid2 and uuid3
	// (they are children of the divider block).
	dividerIdx := strings.Index(resultStr, "⤵️ Unranked tasks")
	task2Idx := strings.Index(resultStr, uuid2)
	task3Idx := strings.Index(resultStr, uuid3)

	assert.Less(t, dividerIdx, task2Idx, "divider should come before uuid2")
	assert.Less(t, dividerIdx, task3Idx, "divider should come before uuid3")
}

func TestMoveToUnrankedFromSectionHeader(t *testing.T) {
	// uuid2 lives under ✨ New tasks (a section header child), not at the top level.
	// Moving it to unranked must find it there.
	pageContent := "- ((" + uuid1 + "))\n- ✨ New tasks\n\t- ((" + uuid2 + "))\n"
	graphDir := makeTestGraph(t, pageContent)

	err := dashboard.MoveToUnranked(graphDir, testBacklogPage, []string{uuid2})
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(graphDir, "pages", testBacklogPage+".md"))
	require.NoError(t, err)

	resultStr := string(result)

	assert.Contains(t, resultStr, "⤵️ Unranked tasks", "unranked divider should be created")
	assert.Contains(t, resultStr, uuid1)
	assert.Contains(t, resultStr, uuid2)

	dividerIdx := strings.Index(resultStr, "⤵️ Unranked tasks")
	task2Idx := strings.Index(resultStr, uuid2)
	assert.Less(t, dividerIdx, task2Idx, "uuid2 should be under the unranked divider")
}

func TestMoveToUnrankedPreservesExistingDivider(t *testing.T) {
	// Page already has an unranked divider with one task. We add another.
	pageContent := "- ((" + uuid1 + "))\n- ((" + uuid2 + "))\n- ⤵️ Unranked tasks\n\t- ((" + uuid3 + "))\n"
	graphDir := makeTestGraph(t, pageContent)

	err := dashboard.MoveToUnranked(graphDir, testBacklogPage, []string{uuid2})
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(graphDir, "pages", testBacklogPage+".md"))

	require.NoError(t, err)

	resultStr := string(result)

	// Only one divider header should exist.
	assert.Equal(t, 1, strings.Count(resultStr, "⤵️ Unranked tasks"))

	// All three tasks must still be present.
	assert.Contains(t, resultStr, uuid1)
	assert.Contains(t, resultStr, uuid2)
	assert.Contains(t, resultStr, uuid3)
}
