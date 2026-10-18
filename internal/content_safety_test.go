package internal //nolint:testpackage // Exercise unsafe-range and atomic-replacement guards directly.

import (
	"os"
	"path/filepath"
	"testing"

	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/stretchr/testify/require"
)

func TestRejectInvalidSourceRanges(t *testing.T) {
	block := content.NewBlock()
	for _, ranges := range [][]logseq.BlockSourceRange{
		{{Block: block, Start: -1, End: 2}},
		{{Block: block, Start: 0, End: 4}},
		{{Block: block, Start: 1, End: 1}},
		{{Block: nil, Start: 0, End: 3}},
		{{Block: block, Start: 0, End: 1}, {Block: block, Start: 2, End: 3}},
		{{Block: block, Start: 0, End: 2}},
	} {
		require.Error(t, validateSourceRanges("abc", ranges))
	}
}

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
