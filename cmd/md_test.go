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

// This file tests the md command using Cobra's recommended constructor pattern.
// We use dependency injection to make the command testable without executing
// the actual external dependencies.

func TestMdCommand_WithDependencyInjection(t *testing.T) {
	graphPath := "/test/graph"
	frozenTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []mdTestCase{
		{
			name:  "basic command with no flags",
			stdin: "test content",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:   nil, // Will be set to mockGraph in test
				Date:    frozenTime,
				Content: "test content",
			},
		},
		{
			name:       "command with parent flag",
			parentFlag: "Project A",
			stdin:      "child task",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:       frozenTime,
				Content:    "child task",
				ParentText: "Project A",
			},
		},
		{
			name:        "command with journal flag",
			journalFlag: "2024-12-25",
			stdin:       "journal entry",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:   nil, // Will be set to mockGraph in test
				Date:    time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
				Content: "journal entry",
			},
		},
		{
			name:        "command with both parent and journal flags",
			journalFlag: "2024-03-10",
			parentFlag:  "meeting notes",
			stdin:       "action item",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:       time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
				Content:    "action item",
				ParentText: "meeting notes",
			},
		},
		{
			name:    "command with key flag",
			keyFlag: "unique identifier",
			stdin:   "updated content",
			expectedOpts: &internal.InsertMarkdownOptions{
				Graph:   nil, // Will be set to mockGraph in test
				Date:    frozenTime,
				Content: "updated content",
				Key:     "unique identifier",
			},
		},
		{
			name:       "command with key and parent flags",
			parentFlag: "Project A",
			keyFlag:    "task-123",
			stdin:      "updated task",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:       frozenTime,
				Content:    "updated task",
				ParentText: "Project A",
				Key:        "task-123",
			},
		},
		{
			name:        "command with key and journal flags",
			journalFlag: "2024-12-25",
			parentFlag:  "",
			keyFlag:     "holiday-note",
			stdin:       "updated holiday note",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:    time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
				Content: "updated holiday note",
				Key:     "holiday-note",
			},
		},
		{
			name:        "command with all flags (journal, parent, key)",
			journalFlag: "2024-03-10",
			parentFlag:  "Sprint 1",
			keyFlag:     "feature-456",
			stdin:       "complete feature",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:       time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
				Content:    "complete feature",
				ParentText: "Sprint 1",
				Key:        "feature-456",
			},
		},
		{
			name:     "command with page flag",
			pageFlag: "Work",
			stdin:    "meeting notes",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:    frozenTime,
				Page:    "Work",
				Content: "meeting notes",
				Key:     "",
			},
		},
		{
			name:       "command with page and parent flags",
			pageFlag:   "Work",
			parentFlag: "Project A",
			stdin:      "child task",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:       frozenTime,
				Page:       "Work",
				Content:    "child task",
				ParentText: "Project A",
			},
		},
		{
			name:     "command with page and key flags",
			pageFlag: "Projects",
			keyFlag:  "feature-123",
			stdin:    "update work item",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:    frozenTime,
				Page:    "Projects",
				Content: "update work item",
				Key:     "feature-123",
			},
		},
		{
			name:       "command with page, parent, and key flags",
			pageFlag:   "Development",
			parentFlag: "Sprint 2",
			keyFlag:    "bug-789",
			stdin:      "fix critical bug",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:       frozenTime,
				Page:       "Development",
				Content:    "fix critical bug",
				ParentText: "Sprint 2",
				Key:        "bug-789",
			},
		},
		{
			name:           "invalid journal date format",
			journalFlag:    "2024/01/15",
			stdin:          "content",
			expectedOpts:   nil, // Should not be called due to error
			expectError:    true,
			expectedErrMsg: "invalid journal date format",
		},
		{
			name:  "insert function returns error",
			stdin: "content",
			expectedOpts: &internal.InsertMarkdownOptions{
				Date:    frozenTime,
				Content: "content",
			},
			insertError:    assert.AnError,
			expectError:    true,
			expectedErrMsg: "assert.AnError general error for testing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runMdCommandTest(t, test, graphPath, frozenTime)
		})
	}
}

type mdTestCaptures struct {
	opts      *internal.InsertMarkdownOptions
	graphPath string
	stdin     string
}

func createMockDeps(
	test mdTestCase, frozenTime time.Time, mockGraph *logseq.Graph, captures *mdTestCaptures,
) *cmd.MdDependencies {
	return &cmd.MdDependencies{
		InsertFn: func(opts *internal.InsertMarkdownOptions) error {
			captures.opts = opts

			return test.insertError
		},
		OpenGraph: func(path string) *logseq.Graph {
			captures.graphPath = path

			return mockGraph
		},
		ReadStdin: func() string {
			captures.stdin = test.stdin

			return test.stdin
		},
		TimeNow: func() time.Time {
			return frozenTime
		},
	}
}

func buildCommandArgs(test mdTestCase) []string {
	args := []string{}
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

	return args
}

func verifyErrorCase(t *testing.T, test mdTestCase, err error, captures *mdTestCaptures) {
	t.Helper()

	require.Error(t, err)

	if test.expectedErrMsg != "" {
		assert.Contains(t, err.Error(), test.expectedErrMsg)
	}

	// If we expect an error due to invalid date format, the insert function shouldn't be called
	if test.expectedOpts == nil {
		assert.Nil(t, captures.opts, "Insert function should not be called when date parsing fails")
	}
}

func verifySuccessCase(
	t *testing.T, test mdTestCase, graphPath string, mockGraph *logseq.Graph, captures *mdTestCaptures,
) {
	t.Helper()

	// Verify all dependencies were called correctly
	assert.Equal(t, graphPath, captures.graphPath, "OpenGraph should be called with correct path")
	assert.Equal(t, test.stdin, captures.stdin, "ReadStdin should return expected content")

	// Verify the insert function was called with expected options
	require.NotNil(t, captures.opts, "Insert function should have been called")
	assert.Equal(t, test.expectedOpts.Date, captures.opts.Date)
	assert.Equal(t, test.expectedOpts.Page, captures.opts.Page)
	assert.Equal(t, test.expectedOpts.Content, captures.opts.Content)
	assert.Equal(t, test.expectedOpts.ParentText, captures.opts.ParentText)
	assert.Equal(t, test.expectedOpts.Key, captures.opts.Key)
	assert.Equal(t, mockGraph, captures.opts.Graph, "Graph should be the mocked graph")
}

type mdTestCase struct {
	name           string
	journalFlag    string
	parentFlag     string
	pageFlag       string
	keyFlag        string
	stdin          string
	expectedOpts   *internal.InsertMarkdownOptions
	insertError    error
	expectError    bool
	expectedErrMsg string
}

func runMdCommandTest(t *testing.T, test mdTestCase, graphPath string, frozenTime time.Time) {
	t.Helper()

	mockGraph := &logseq.Graph{} // Empty graph for testing
	captures := &mdTestCaptures{}

	mockDeps := createMockDeps(test, frozenTime, mockGraph, captures)
	command := cmd.NewMdCmd(mockDeps)

	t.Setenv("LOGSEQ_GRAPH_PATH", graphPath)

	args := buildCommandArgs(test)

	err := command.ParseFlags(args)
	require.NoError(t, err)

	err = command.RunE(command, []string{})

	if test.expectError {
		verifyErrorCase(t, test, err, captures)
	} else {
		require.NoError(t, err)
		verifySuccessCase(t, test, graphPath, mockGraph, captures)
	}
}
