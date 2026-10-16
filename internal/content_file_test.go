package internal //nolint:testpackage // Exercise snapshot guards without a live editor race.

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/require"
)

func TestContentPreservesJournalSymlink(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"logseq", "journals", "pages"} {
		require.NoError(t, os.Mkdir(filepath.Join(root, directory), 0o750))
	}

	require.NoError(t, os.WriteFile(filepath.Join(root, "logseq", "config.edn"), []byte("{}"), 0o600))

	graph, err := logseq.Open(context.Background(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, graph.Close()) })

	date := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	path, err := graph.JournalPath(date)
	require.NoError(t, err)

	target := filepath.Join(t.TempDir(), "actual.md")
	require.NoError(t, os.WriteFile(target, []byte("- 10:00 Existing\n"), 0o600))
	require.NoError(t, os.Symlink(target, path))

	_, err = AppendRawMarkdownToJournal(graph, date, "- 09:00 New\n")
	require.NoError(t, err)

	linked, err := os.Readlink(path)
	require.NoError(t, err)
	require.Equal(t, target, linked)

	actual, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "- 09:00 New\n- 10:00 Existing\n", string(actual))
}

func TestJournalRejectsChangedSnapshot(t *testing.T) {
	for _, initiallyExists := range []bool{false, true} {
		t.Run(strconv.FormatBool(initiallyExists), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "journal.md")
			if initiallyExists {
				require.NoError(t, os.WriteFile(path, []byte("old"), 0o600))
			}

			snapshot, err := readJournalSnapshot(path)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, []byte("editor change"), 0o600))
			require.Error(t, replaceJournal(path, []byte("replacement"), 0o600, snapshot))

			actual, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, "editor change", string(actual))

			entries, err := os.ReadDir(filepath.Dir(path))
			require.NoError(t, err)
			require.Len(t, entries, 1)
		})
	}
}
