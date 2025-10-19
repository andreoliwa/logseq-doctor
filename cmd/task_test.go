package cmd_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/cmd"
	"github.com/andreoliwa/lsd/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the task command using Cobra's recommended constructor pattern.
// We use dependency injection to make the command testable without executing
// the actual external dependencies.

func TestNewTaskCmd(t *testing.T) {
	cmd := cmd.NewTaskCmd()

	require.NotNil(t, cmd)
	assert.Equal(t, "task", cmd.Use)
	assert.Equal(t, "Manage tasks in Logseq", cmd.Short)
	assert.Contains(t, cmd.Long, "Manage tasks in your Logseq graph")
}

func TestTaskAddCommand_WithDependencyInjection(t *testing.T) { //nolint:funlen,maintidx
	graphPath := "/test/graph"
	frozenTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		args           []string
		journalFlag    string
		blockFlag      string
		pageFlag       string
		keyFlag        string
		nameFlag       string
		expectedOpts   *internal.AddTaskOptions
		addTaskError   error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "basic command with task description only",
			args:        []string{"Review pull request"},
			journalFlag: "",
			blockFlag:   "",
			pageFlag:    "",
			keyFlag:     "",
			nameFlag:    "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        frozenTime,
				Description: "Review pull request",
				Page:        "",
				BlockText:   "",
				Key:         "",
				Name:        "",
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with page flag",
			args:        []string{"Call client"},
			journalFlag: "",
			blockFlag:   "",
			pageFlag:    "Work",
			keyFlag:     "",
			nameFlag:    "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        frozenTime,
				Description: "Call client",
				Page:        "Work",
				BlockText:   "",
				Key:         "",
				Name:        "",
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with journal flag",
			args:        []string{"Buy groceries"},
			journalFlag: "2024-12-25",
			blockFlag:   "",
			pageFlag:    "",
			keyFlag:     "",
			nameFlag:    "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
				Description: "Buy groceries",
				Page:        "",
				BlockText:   "",
				Key:         "",
				Name:        "",
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with block flag",
			args:        []string{"Meeting notes"},
			journalFlag: "",
			blockFlag:   "Project A",
			pageFlag:    "",
			keyFlag:     "",
			nameFlag:    "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        frozenTime,
				Description: "Meeting notes",
				Page:        "",
				BlockText:   "Project A",
				Key:         "",
				Name:        "",
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with key and name flags",
			args:        []string{"Water plants"},
			journalFlag: "",
			blockFlag:   "",
			pageFlag:    "",
			keyFlag:     "water plants",
			nameFlag:    "Water plants in living room",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        frozenTime,
				Description: "Water plants",
				Page:        "",
				BlockText:   "",
				Key:         "water plants",
				Name:        "Water plants in living room",
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with all flags",
			args:        []string{"Complete task"},
			journalFlag: "2024-03-10",
			blockFlag:   "Sprint 1",
			pageFlag:    "Development",
			keyFlag:     "task-123",
			nameFlag:    "Complete the feature",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
				Description: "Complete task",
				Page:        "Development",
				BlockText:   "Sprint 1",
				Key:         "task-123",
				Name:        "Complete the feature",
			},
			addTaskError:   nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "invalid journal date format",
			args:           []string{"Task description"},
			journalFlag:    "2024/01/15",
			blockFlag:      "",
			pageFlag:       "",
			keyFlag:        "",
			nameFlag:       "",
			expectedOpts:   nil, // Should not be called due to error
			addTaskError:   nil,
			expectError:    true,
			expectedErrMsg: "invalid journal date format",
		},
		{
			name:        "AddTask function returns error",
			args:        []string{"Task with error"},
			journalFlag: "",
			blockFlag:   "",
			pageFlag:    "",
			keyFlag:     "",
			nameFlag:    "",
			expectedOpts: &internal.AddTaskOptions{
				Graph:       nil, // Will be set to mockGraph in test
				Date:        frozenTime,
				Description: "Task with error",
				Page:        "",
				BlockText:   "",
				Key:         "",
				Name:        "",
			},
			addTaskError:   assert.AnError,
			expectError:    true,
			expectedErrMsg: "assert.AnError general error for testing",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockGraph := &logseq.Graph{} // Empty graph for testing

			var capturedOpts *internal.AddTaskOptions

			var capturedGraphPath string

			mockDeps := &cmd.TaskAddDependencies{
				AddTaskFn: func(opts *internal.AddTaskOptions) error {
					capturedOpts = opts

					return testCase.addTaskError
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

			args := testCase.args
			if testCase.journalFlag != "" {
				args = append(args, "--journal", testCase.journalFlag)
			}

			if testCase.blockFlag != "" {
				args = append(args, "--block", testCase.blockFlag)
			}

			if testCase.pageFlag != "" {
				args = append(args, "--page", testCase.pageFlag)
			}

			if testCase.keyFlag != "" {
				args = append(args, "--key", testCase.keyFlag)
			}

			if testCase.nameFlag != "" {
				args = append(args, "--name", testCase.nameFlag)
			}

			err := command.ParseFlags(args[1:]) // Skip the task description arg
			require.NoError(t, err)

			err = command.RunE(command, args[:1]) // Pass only the task description

			if testCase.expectError {
				require.Error(t, err)

				if testCase.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), testCase.expectedErrMsg)
				}

				// If we expect an error due to invalid date format, the AddTask function shouldn't be called
				if testCase.expectedOpts == nil {
					assert.Nil(t, capturedOpts, "AddTask function should not be called when date parsing fails")
				}
			} else {
				require.NoError(t, err)

				// Verify all dependencies were called correctly
				assert.Equal(t, graphPath, capturedGraphPath, "OpenGraph should be called with correct path")

				// Verify the AddTask function was called with expected options
				require.NotNil(t, capturedOpts, "AddTask function should have been called")
				assert.Equal(t, testCase.expectedOpts.Description, capturedOpts.Description)
				assert.Equal(t, testCase.expectedOpts.Page, capturedOpts.Page)
				assert.Equal(t, testCase.expectedOpts.BlockText, capturedOpts.BlockText)
				assert.Equal(t, testCase.expectedOpts.Key, capturedOpts.Key)
				assert.Equal(t, testCase.expectedOpts.Name, capturedOpts.Name)
				assert.Equal(t, testCase.expectedOpts.Date, capturedOpts.Date)
				assert.Equal(t, mockGraph, capturedOpts.Graph, "Graph should be the mocked graph")
			}
		})
	}
}

func TestTaskAddCommand_WithNilDependencies(t *testing.T) {
	// Test that NewTaskAddCmd works with nil dependencies (uses defaults)
	cmd := cmd.NewTaskAddCmd(nil)

	require.NotNil(t, cmd)
	assert.Equal(t, "add [task description]", cmd.Use)
	assert.Equal(t, "Add a new task to Logseq", cmd.Short)
	assert.Contains(t, cmd.Long, "Add a new task to your Logseq graph")
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
