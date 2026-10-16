//go:build !windows

package internal //nolint:testpackage // Verify replacement permissions under an isolated process umask.

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJournalNewFileHonorsUmask(t *testing.T) {
	if os.Getenv("LQD_TEST_UMASK") != "1" {
		// Run this test binary in a child process to isolate the process-wide umask.
		command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestJournalNewFileHonorsUmask$") //nolint:gosec

		command.Env = append(os.Environ(), "LQD_TEST_UMASK=1")
		output, err := command.CombinedOutput()
		require.NoError(t, err, string(output))

		return
	}

	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	path := filepath.Join(t.TempDir(), "new.md")
	require.NoError(t, replaceJournal(path, []byte("new"), defaultJournalMode))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}
