package cmd_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the sync command using Cobra's recommended constructor pattern.
// The full sync loop (PocketBase, Logseq API, file I/O) is not unit-testable,
// so these tests cover command structure, flags, and dependency injection wiring.

func TestNewSyncCmd_Structure(t *testing.T) {
	syncCmd := cmd.NewSyncCmd(nil)

	require.NotNil(t, syncCmd)
	assert.Equal(t, "sync", syncCmd.Use)
	assert.NotEmpty(t, syncCmd.Short)
	assert.NotEmpty(t, syncCmd.Long)
}

func TestNewSyncCmd_DefaultFlags(t *testing.T) {
	syncCmd := cmd.NewSyncCmd(nil)

	initFlag := syncCmd.Flags().Lookup("init")
	require.NotNil(t, initFlag)
	assert.Equal(t, "false", initFlag.DefValue)
}

func TestNewSyncCmd_WithCustomDeps(t *testing.T) {
	frozenTime := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	called := false

	deps := &cmd.SyncDependencies{
		TimeNow: func() time.Time {
			called = true

			return frozenTime
		},
	}

	syncCmd := cmd.NewSyncCmd(deps)

	require.NotNil(t, syncCmd)
	// Deps are wired but Run is only called when the command executes.
	assert.False(t, called, "TimeNow should not be called at construction time")
}

func TestNewSyncCmd_NilDepsUsesDefaults(t *testing.T) {
	// Passing nil deps should not panic — defaults should be used.
	syncCmd := cmd.NewSyncCmd(nil)

	assert.NotNil(t, syncCmd)
}
