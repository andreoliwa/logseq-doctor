package internal //nolint:testpackage // Exercise atomic-replacement guards directly.

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceJournalFailureLeavesTargetAndNoTemporaryFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "target")
	require.NoError(t, os.Mkdir(path, 0o750))
	retained := filepath.Join(path, "retained.md")
	require.NoError(t, os.WriteFile(retained, []byte("original"), 0o600))
	require.Error(t, replaceJournal(path, []byte("replacement"), 0o640))

	got, err := os.ReadFile(retained)
	require.NoError(t, err)
	require.Equal(t, "original", string(got))

	entries, err := os.ReadDir(directory)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "target", entries[0].Name())
	require.Error(t, replaceJournal(filepath.Join(directory, "absent", "journal.md"), []byte("replacement"), 0o640))
}
