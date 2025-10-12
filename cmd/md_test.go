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

// This file tests the md command using Cobra's recommended constructor pattern.
// We use dependency injection to make the command testable without executing
// the actual external dependencies.

func TestMdCommand_WithDependencyInjection(t *testing.T) { //nolint:funlen
	graphPath := "/test/graph"
	frozenTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		journalFlag    string
		parentFlag     string
		stdin          string
		expectedOpts   *internal.InsertMarkdownOptions
		insertError    error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "basic command with no flags",
			journalFlag: "",
			parentFlag:  "",
			stdin:       "test content",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:      nil, // Will be set to mockGraph in test
				Content:    "test content",
				ParentText: "",
				Date:       frozenTime,
			},
			insertError:    nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with parent flag",
			journalFlag: "",
			parentFlag:  "Project A",
			stdin:       "child task",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:      nil, // Will be set to mockGraph in test
				Content:    "child task",
				ParentText: "Project A",
				Date:       frozenTime,
			},
			insertError:    nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with journal flag",
			journalFlag: "2024-12-25",
			parentFlag:  "",
			stdin:       "journal entry",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:      nil, // Will be set to mockGraph in test
				Content:    "journal entry",
				ParentText: "",
				Date:       time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			},
			insertError:    nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:        "command with both flags",
			journalFlag: "2024-03-10",
			parentFlag:  "meeting notes",
			stdin:       "action item",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:      nil, // Will be set to mockGraph in test
				Content:    "action item",
				ParentText: "meeting notes",
				Date:       time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
			},
			insertError:    nil,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "invalid journal date format",
			journalFlag:    "2024/01/15",
			parentFlag:     "",
			stdin:          "content",
			expectedOpts:   nil, // Should not be called due to error
			insertError:    nil,
			expectError:    true,
			expectedErrMsg: "invalid journal date format",
		},
		{
			name:        "insert function returns error",
			journalFlag: "",
			parentFlag:  "",
			stdin:       "content",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:      nil, // Will be set to mockGraph in test
				Content:    "content",
				ParentText: "",
				Date:       frozenTime,
			},
			insertError:    assert.AnError,
			expectError:    true,
			expectedErrMsg: "assert.AnError general error for testing",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockGraph := &logseq.Graph{} // Empty graph for testing

			var capturedOpts *internal.InsertMarkdownOptions

			var capturedGraphPath string

			var capturedStdin string

			mockDeps := &cmd.MdDependencies{
				InsertFn: func(opts *internal.InsertMarkdownOptions) error {
					capturedOpts = opts

					return testCase.insertError
				},
				OpenGraph: func(path string) *logseq.Graph {
					capturedGraphPath = path

					return mockGraph
				},
				ReadStdin: func() string {
					capturedStdin = testCase.stdin

					return testCase.stdin
				},
				TimeNow: func() time.Time {
					return frozenTime
				},
			}

			command := cmd.NewMdCmd(mockDeps)

			t.Setenv("LOGSEQ_GRAPH_PATH", graphPath)

			args := []string{}
			if testCase.journalFlag != "" {
				args = append(args, "--journal", testCase.journalFlag)
			}

			if testCase.parentFlag != "" {
				args = append(args, "--parent", testCase.parentFlag)
			}

			err := command.ParseFlags(args)
			require.NoError(t, err)

			err = command.RunE(command, []string{})

			if testCase.expectError {
				require.Error(t, err)

				if testCase.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), testCase.expectedErrMsg)
				}

				// If we expect an error due to invalid date format, the insert function shouldn't be called
				if testCase.expectedOpts == nil {
					assert.Nil(t, capturedOpts, "Insert function should not be called when date parsing fails")
				}
			} else {
				require.NoError(t, err)

				// Verify all dependencies were called correctly
				assert.Equal(t, graphPath, capturedGraphPath, "OpenGraph should be called with correct path")
				assert.Equal(t, testCase.stdin, capturedStdin, "ReadStdin should return expected content")

				// Verify the insert function was called with expected options
				require.NotNil(t, capturedOpts, "Insert function should have been called")
				assert.Equal(t, testCase.expectedOpts.Content, capturedOpts.Content)
				assert.Equal(t, testCase.expectedOpts.ParentText, capturedOpts.ParentText)
				assert.Equal(t, testCase.expectedOpts.Date, capturedOpts.Date)
				assert.Equal(t, mockGraph, capturedOpts.Graph, "Graph should be the mocked graph")
			}
		})
	}
}
