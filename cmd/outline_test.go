package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/andreoliwa/logseq-doctor/cmd"
	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errUnexpectedFile is a sentinel used in test fakes when an unknown path is requested.
var errUnexpectedFile = errors.New("unexpected file")

// This file tests the outline command via the constructor+DI pattern.
// All I/O and conversion dependencies are injected as fake functions.

// newOutlineTestDeps returns a base set of no-op dependencies and a shared output buffer.
// Tests override only the fields they need.
func newOutlineTestDeps(t *testing.T) (*cmd.OutlineDependencies, *bytes.Buffer) {
	t.Helper()

	out := &bytes.Buffer{}

	deps := &cmd.OutlineDependencies{
		Convert: func(input string, _ internal.OutlineOptions) string {
			return input
		},
		ReadFile: func(_ string) (string, error) {
			return "", nil
		},
		WriteFile: func(_ string, _ string) error {
			return nil
		},
		Stat: func(_ string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		},
		Rename: func(_, _ string) error {
			return nil
		},
		Remove: func(_ string) error {
			return nil
		},
		Stdin: func() string {
			return ""
		},
		Out: out,
	}

	return deps, out
}

func TestOutline_StdinMode(t *testing.T) {
	deps, out := newOutlineTestDeps(t)

	convertCalled := false
	deps.Stdin = func() string { return "# Hello\n" }
	deps.Convert = func(input string, _ internal.OutlineOptions) string {
		convertCalled = true

		return input + "converted"
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.True(t, convertCalled, "Convert should have been called")
	assert.Contains(t, out.String(), "# Hello\n")
	assert.Contains(t, out.String(), "converted")
}

func TestOutline_SingleFileToStdout(t *testing.T) {
	deps, out := newOutlineTestDeps(t)

	deps.ReadFile = func(path string) (string, error) {
		assert.Equal(t, "a.md", path)

		return "content", nil
	}

	writeFileCalled := false
	deps.WriteFile = func(_ string, _ string) error {
		writeFileCalled = true

		return nil
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"a.md"})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.False(t, writeFileCalled, "WriteFile should NOT be called in stdout mode")
	assert.Equal(t, "content", out.String())
}

func TestOutline_MultipleFilesToStdout(t *testing.T) {
	deps, out := newOutlineTestDeps(t)

	fileContents := map[string]string{
		"a.md": "alpha",
		"b.md": "beta",
	}

	deps.ReadFile = func(path string) (string, error) {
		c, ok := fileContents[path]
		if !ok {
			return "", fmt.Errorf("%w: %s", errUnexpectedFile, path)
		}

		return c, nil
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"a.md", "b.md"})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "alpha")
	assert.Contains(t, out.String(), "beta")
	assert.Equal(t, "alphabeta", out.String())
}

func TestOutline_InPlace(t *testing.T) {
	deps, out := newOutlineTestDeps(t)

	deps.ReadFile = func(_ string) (string, error) {
		return "original", nil
	}

	deps.Convert = func(_ string, _ internal.OutlineOptions) string {
		return "converted"
	}

	var writtenPath, writtenData string

	deps.WriteFile = func(path string, data string) error {
		writtenPath = path
		writtenData = data

		return nil
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"-i", "a.md"})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "a.md", writtenPath, "WriteFile should write to the same path")
	assert.Equal(t, "converted", writtenData)
	assert.Empty(t, out.String(), "nothing should be written to Out in --in-place mode")
}

func TestOutline_MoveTo_Success(t *testing.T) {
	deps, _ := newOutlineTestDeps(t)

	deps.ReadFile = func(_ string) (string, error) {
		return "content", nil
	}

	deps.Stat = func(_ string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	var writtenPath, writtenData string

	deps.WriteFile = func(path string, data string) error {
		writtenPath = path
		writtenData = data

		return nil
	}

	removeCalled := false

	deps.Remove = func(path string) error {
		removeCalled = true

		assert.Equal(t, "a.md", path)

		return nil
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"--move-to", "/dst", "a.md"})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "/dst/a.md", writtenPath)
	assert.Equal(t, "content", writtenData)
	assert.True(t, removeCalled, "Remove should be called for the source file")
}

func TestOutline_MoveTo_DestExists(t *testing.T) {
	deps, _ := newOutlineTestDeps(t)

	deps.ReadFile = func(_ string) (string, error) {
		return "content", nil
	}

	// Stat returns nil error => file exists.
	deps.Stat = func(_ string) (os.FileInfo, error) {
		return nil, nil //nolint:nilnil
	}

	writeFileCalled := false
	deps.WriteFile = func(_ string, _ string) error {
		writeFileCalled = true

		return nil
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"--move-to", "/dst", "a.md"})

	err := outlineCmd.Execute()
	require.ErrorIs(t, err, cmd.ErrDestinationExists)
	assert.False(t, writeFileCalled, "WriteFile should NOT be called when destination exists")
}

func TestOutline_MutualExclusion(t *testing.T) {
	deps, _ := newOutlineTestDeps(t)

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"-i", "--move-to", "/dst", "a.md"})

	err := outlineCmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, cmd.ErrMutuallyExclusive)
}

func TestOutline_KeepBreaksForwarded(t *testing.T) {
	deps, _ := newOutlineTestDeps(t)

	deps.ReadFile = func(_ string) (string, error) {
		return "content", nil
	}

	var capturedOpts internal.OutlineOptions

	deps.Convert = func(input string, opts internal.OutlineOptions) string {
		capturedOpts = opts

		return input
	}

	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetArgs([]string{"--keep-breaks", "a.md"})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.True(t, capturedOpts.KeepBreaks, "KeepBreaks should be true when --keep-breaks is passed")
}

func TestNewOutlineCmd_NilDeps(t *testing.T) {
	outlineCmd := cmd.NewOutlineCmd(nil)

	assert.Equal(t, "outline [file ...]", outlineCmd.Use)

	inPlace := outlineCmd.Flags().Lookup("in-place")
	require.NotNil(t, inPlace, "flag --in-place should be registered")

	moveTo := outlineCmd.Flags().Lookup("move-to")
	require.NotNil(t, moveTo, "flag --move-to should be registered")

	keepBreaks := outlineCmd.Flags().Lookup("keep-breaks")
	require.NotNil(t, keepBreaks, "flag --keep-breaks should be registered")
}

// TestOutline_StdinMode_NoArgs verifies that the command routes to stdin mode
// when called with an explicit writer (no real stdin needed).
func TestOutline_StdinMode_NoArgs(t *testing.T) {
	deps, out := newOutlineTestDeps(t)
	deps.Stdin = func() string { return "stdin-content\n" }
	deps.Convert = func(input string, _ internal.OutlineOptions) string { return "[" + input + "]" }

	// Suppress cobra's own stderr output during Execute by providing a discard writer.
	outlineCmd := cmd.NewOutlineCmd(deps)
	outlineCmd.SetErr(io.Discard)
	outlineCmd.SetArgs([]string{})

	err := outlineCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "[stdin-content\n]", out.String())
}
