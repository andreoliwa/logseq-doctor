package cmd_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/cmd"
	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the task command using Cobra's recommended constructor pattern.
// We use dependency injection to make the command testable without executing
// the actual external dependencies.

func TestNewTaskCmd(t *testing.T) {
	taskCmd := cmd.NewTaskCmd()

	require.NotNil(t, taskCmd)
	assert.Equal(t, "task", taskCmd.Use)
	assert.Equal(t, "Manage tasks in Logseq", taskCmd.Short)
	assert.Contains(t, taskCmd.Long, "Manage tasks in your Logseq graph")
}

func TestTaskAddCommand_WithDependencyInjection(t *testing.T) { //nolint:maintidx
	graphPath := "/test/graph"
	frozenTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		args           []string
		journalFlag    string
		parentFlag     string
		pageFlag       string
		keyFlag        string
		expectedOpts   *internal.AddTaskOptions
		addTaskError   error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "basic command with task description only",
			args:        []string{"Review pull request"},
			journalFlag: "",
			parentFlag:  "",
			pageFlag:    "",
			keyFlag:     "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      frozenTime,
				Page:      "",
				BlockText: "",
				Key:       "",
				Name:      "Review pull request",
				TimeNow:   func() time.Time { return frozenTime },
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with page flag",
			args:        []string{"Call client"},
			journalFlag: "",
			parentFlag:  "",
			pageFlag:    "Work",
			keyFlag:     "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      frozenTime,
				Page:      "Work",
				BlockText: "",
				Key:       "",
				Name:      "Call client",
				TimeNow:   func() time.Time { return frozenTime },
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with journal flag",
			args:        []string{"Buy groceries"},
			journalFlag: "2024-12-25",
			parentFlag:  "",
			pageFlag:    "",
			keyFlag:     "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
				Page:      "",
				BlockText: "",
				Key:       "",
				Name:      "Buy groceries",
				TimeNow: func() time.Time {
					return time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
				},
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with parent flag",
			args:        []string{"Meeting notes"},
			journalFlag: "",
			parentFlag:  "Project A",
			pageFlag:    "",
			keyFlag:     "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      frozenTime,
				Page:      "",
				BlockText: "Project A",
				Key:       "",
				Name:      "Meeting notes",
				TimeNow:   func() time.Time { return frozenTime },
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with key flag",
			args:        []string{"Water plants in living room"},
			journalFlag: "",
			parentFlag:  "",
			pageFlag:    "",
			keyFlag:     "water plants",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      frozenTime,
				Page:      "",
				BlockText: "",
				Key:       "water plants",
				Name:      "Water plants in living room",
				TimeNow:   func() time.Time { return frozenTime },
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with all flags",
			args:        []string{"Complete the feature"},
			journalFlag: "2024-03-10",
			parentFlag:  "Sprint 1",
			pageFlag:    "Development",
			keyFlag:     "task-123",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
				Page:      "Development",
				BlockText: "Sprint 1",
				Key:       "task-123",
				Name:      "Complete the feature",
				TimeNow: func() time.Time {
					return time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
				},
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "invalid journal date format",
			args:           []string{"Task description"},
			journalFlag:    "2024/01/15",
			parentFlag:     "",
			pageFlag:       "",
			keyFlag:        "",
			expectedOpts:   nil, // Should not be called due to error
			addTaskError:   nil,
			expectError:    true,
			expectedErrMsg: "invalid journal date format",
		},
		{
			name:        "AddTask function returns error",
			args:        []string{"Task with error"},
			journalFlag: "",
			parentFlag:  "",
			pageFlag:    "",
			keyFlag:     "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:     nil, // Will be set to mockGraph in test
				Date:      frozenTime,
				Page:      "",
				BlockText: "",
				Key:       "",
				Name:      "Task with error",
				TimeNow:   func() time.Time { return frozenTime },
			},
			addTaskError:   assert.AnError,
			expectError:    true,
			expectedErrMsg: "assert.AnError general error for testing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockGraph := &logseq.Graph{} // Empty graph for testing

			var capturedOpts *internal.AddTaskOptions

			var capturedGraphPath string

			mockDeps := &cmd.TaskAddDependencies{
				AddTaskFn: func(opts *internal.AddTaskOptions) error {
					capturedOpts = opts

					return test.addTaskError
				},
				OpenGraph: func(path string) *logseq.Graph {
					capturedGraphPath = path

					return mockGraph
				},
				TimeNow: func() time.Time {
					return frozenTime
				},
			}

			command := cmd.NewTaskAddCmd(mockDeps)

			t.Setenv("LOGSEQ_GRAPH_PATH", graphPath)

			args := test.args
			if test.journalFlag != "" {
				args = append(args, "--journal", test.journalFlag)
			}

			if test.parentFlag != "" {
				args = append(args, "--parent", test.parentFlag)
			}

			if test.pageFlag != "" {
				args = append(args, "--page", test.pageFlag)
			}

			if test.keyFlag != "" {
				args = append(args, "--key", test.keyFlag)
			}

			err := command.ParseFlags(args[1:]) // Skip the task description arg
			require.NoError(t, err)

			err = command.RunE(command, args[:1]) // Pass only the task description

			if test.expectError {
				require.Error(t, err)

				if test.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), test.expectedErrMsg)
				}

				// If we expect an error due to invalid date format, the AddTask function shouldn't be called
				if test.expectedOpts == nil {
					assert.Nil(t, capturedOpts, "AddTask function should not be called when date parsing fails")
				}
			} else {
				require.NoError(t, err)

				// Verify all dependencies were called correctly
				assert.Equal(t, graphPath, capturedGraphPath, "OpenGraph should be called with correct path")

				// Verify the AddTask function was called with expected options
				require.NotNil(t, capturedOpts, "AddTask function should have been called")
				assert.Equal(t, test.expectedOpts.Page, capturedOpts.Page)
				assert.Equal(t, test.expectedOpts.BlockText, capturedOpts.BlockText)
				assert.Equal(t, test.expectedOpts.Key, capturedOpts.Key)
				assert.Equal(t, test.expectedOpts.Name, capturedOpts.Name)
				assert.Equal(t, test.expectedOpts.Date, capturedOpts.Date)
				assert.Equal(t, mockGraph, capturedOpts.Graph, "Graph should be the mocked graph")
			}
		})
	}
}

func TestTaskAddCommand_WithNilDependencies(t *testing.T) {
	// Test that NewTaskAddCmd works with nil dependencies (uses defaults)
	taskAddCmd := cmd.NewTaskAddCmd(nil)

	require.NotNil(t, taskAddCmd)
	assert.Equal(t, "add [task description]", taskAddCmd.Use)
	assert.Equal(t, "Add a new task to Logseq", taskAddCmd.Short)
	assert.Contains(t, taskAddCmd.Long, "Add a new task to your Logseq graph")
}

func TestTaskAddCommand_RequiresArgument(t *testing.T) {
	mockDeps := &cmd.TaskAddDependencies{
		AddTaskFn: func(_ *internal.AddTaskOptions) error {
			t.Fatal("AddTaskFn should not be called when args validation fails")

			return nil
		},
		OpenGraph: func(_ string) *logseq.Graph {
			return &logseq.Graph{}
		},
		TimeNow: time.Now,
	}

	command := cmd.NewTaskAddCmd(mockDeps)

	// Try to execute without any arguments - this should fail validation
	command.SetArgs([]string{})
	err := command.Execute()

	// Should fail because MinimumNArgs(1) is set
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least 1 arg(s)")
}
