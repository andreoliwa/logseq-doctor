package cmd_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/andreoliwa/logseq-doctor/cmd"
	"github.com/andreoliwa/logseq-doctor/internal/config"
	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errConfigNotFound = errors.New("tidy-up config file not found")

func TestNewTidyUpCmdStructure(t *testing.T) {
	command := cmd.NewTidyUpCmd(nil)

	require.NotNil(t, command)
	assert.Equal(t, "tidy-up file1.md [file2.md ...]", command.Use)
	assert.True(t, command.SilenceErrors)
	assert.True(t, command.SilenceUsage)
}

func TestNewTidyUpCmdStopsBeforeOpeningGraphWhenConfigFails(t *testing.T) {
	graphOpened := false
	command := cmd.NewTidyUpCmd(&cmd.TidyUpDependencies{
		OpenGraphFromPath: func(string) *logseq.Graph {
			graphOpened = true

			return nil
		},
		LoadPolicy: func() (config.ForbiddenContentPolicy, error) {
			return config.ForbiddenContentPolicy{}, errConfigNotFound
		},
		TidyUpOneFile: func(*logseq.Graph, string, config.ForbiddenContentPolicy) int {
			return 0
		},
	})

	var stderr bytes.Buffer
	command.SetErr(&stderr)
	command.SetArgs([]string{"page.md"})

	err := command.Execute()

	require.ErrorContains(t, err, "config file not found")
	assert.Contains(t, stderr.String(), "config file not found")
	assert.False(t, graphOpened)
}

func TestNewTidyUpCmdUsesLoadedPolicy(t *testing.T) {
	policy := config.ForbiddenContentPolicy{TextSubstrings: []string{"📍"}}
	called := false
	command := cmd.NewTidyUpCmd(&cmd.TidyUpDependencies{
		OpenGraphFromPath: func(string) *logseq.Graph { return nil },
		LoadPolicy:        func() (config.ForbiddenContentPolicy, error) { return policy, nil },
		TidyUpOneFile: func(_ *logseq.Graph, path string, actual config.ForbiddenContentPolicy) int {
			called = true

			assert.Equal(t, "page.md", path)
			assert.Equal(t, policy, actual)

			return 0
		},
	})
	command.SetArgs([]string{"page.md"})

	require.NoError(t, command.Execute())
	assert.True(t, called)
}

func TestNewTidyUpCmdReturnsErrorForForbiddenContent(t *testing.T) {
	command := cmd.NewTidyUpCmd(&cmd.TidyUpDependencies{
		OpenGraphFromPath: func(string) *logseq.Graph { return nil },
		LoadPolicy:        func() (config.ForbiddenContentPolicy, error) { return config.ForbiddenContentPolicy{}, nil },
		TidyUpOneFile:     func(*logseq.Graph, string, config.ForbiddenContentPolicy) int { return 1 },
	})

	var stderr bytes.Buffer
	command.SetErr(&stderr)
	command.SetArgs([]string{"page.md"})

	err := command.Execute()

	require.Error(t, err)
	require.ErrorContains(t, err, "forbidden content")
	assert.Empty(t, stderr.String())
}
