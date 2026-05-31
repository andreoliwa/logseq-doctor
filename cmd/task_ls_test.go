package cmd_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andreoliwa/logseq-doctor/cmd"
	"github.com/andreoliwa/logseq-doctor/internal/api"
)

var errConnectionRefused = errors.New("connection refused")

// mockTaskLsAPI is a test double for api.LogseqAPI used in task ls tests.
type mockTaskLsAPI struct {
	queryResult string
	queryErr    error
	lastQuery   string
}

func (m *mockTaskLsAPI) PostQuery(q string) (string, error) {
	m.lastQuery = q

	return m.queryResult, m.queryErr
}

func (m *mockTaskLsAPI) PostDatascriptQuery(string) (string, error) { return "", nil }

func (m *mockTaskLsAPI) UpsertBlockProperty(_, _, _ string) error { return nil }

// twoTaskJSON is a sample JSON payload with two tasks on different journal days.
// u2 (Dec 1) should sort before u1 (Dec 15) after SortTasksByDate.
const twoTaskJSON = `[` +
	`{"uuid":"u1","content":"TODO buy milk\nscheduled:: [[2024-12-15]]",` +
	`"page":{"journalDay":20241215,"originalName":"Dec 15th, 2024"}},` +
	`{"uuid":"u2","content":"DOING write report",` +
	`"page":{"journalDay":20241201,"originalName":"Dec 1st, 2024"}}` +
	`]`

func newTestDeps(mock *mockTaskLsAPI, buf *bytes.Buffer) *cmd.TaskLsDependencies {
	return &cmd.TaskLsDependencies{
		NewAPI:    func() api.LogseqAPI { return mock },
		GraphName: func() string { return "my-graph" },
		Out:       buf,
	}
}

func TestNewTaskLsCmd(t *testing.T) {
	// Disable color output so ANSI codes don't break string assertions.
	color.NoColor = true

	t.Cleanup(func() { color.NoColor = false })

	tests := []struct {
		name          string
		args          []string
		jsonPayload   string
		queryErr      error
		wantErr       bool
		checkQuery    func(t *testing.T, got string)
		checkOutput   func(t *testing.T, got string)
	}{
		{
			name:        "default no flags returns TODO DOING WAITING statuses",
			args:        []string{},
			jsonPayload: `[]`,
			checkQuery: func(t *testing.T, got string) {
				t.Helper()
				assert.Equal(t, "(and (task TODO DOING WAITING NOW LATER))", got)
			},
		},
		{
			name:        "single tag filter",
			args:        []string{"work"},
			jsonPayload: `[]`,
			checkQuery: func(t *testing.T, got string) {
				t.Helper()
				assert.Equal(t, "(and [[work]] (task TODO DOING WAITING NOW LATER))", got)
			},
		},
		{
			name:        "multiple tag filters use OR",
			args:        []string{"work", "home"},
			jsonPayload: `[]`,
			checkQuery: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "(or [[work]] [[home]])")
			},
		},
		{
			name:        "canceled flag includes CANCELED not DONE",
			args:        []string{"--canceled"},
			jsonPayload: `[]`,
			checkQuery: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "CANCELED")
				assert.NotContains(t, got, "DONE")
			},
		},
		{
			name:        "done flag includes DONE not CANCELED",
			args:        []string{"--done"},
			jsonPayload: `[]`,
			checkQuery: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "DONE")
				assert.NotContains(t, got, "CANCELED")
			},
		},
		{
			name:        "completed flag includes CANCELED and DONE",
			args:        []string{"--completed"},
			jsonPayload: `[]`,
			checkQuery: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "CANCELED")
				assert.Contains(t, got, "DONE")
			},
		},
		{
			name:        "verbose flag prints query before results",
			args:        []string{"--verbose"},
			jsonPayload: `[]`,
			checkOutput: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "Query: (and (task TODO DOING WAITING NOW LATER))")
			},
		},
		{
			name:        "json flag outputs raw JSON without formatting",
			args:        []string{"--json"},
			jsonPayload: twoTaskJSON,
			checkOutput: func(t *testing.T, got string) {
				t.Helper()
				assert.JSONEq(t, twoTaskJSON, strings.TrimSpace(got))
				assert.NotContains(t, got, "§")
			},
		},
		{
			name:        "text output sorted by journalDay ascending",
			args:        []string{},
			jsonPayload: twoTaskJSON,
			checkOutput: func(t *testing.T, got string) {
				t.Helper()
				// u2 (Dec 1, journalDay 20241201) should appear before u1 (Dec 15, journalDay 20241215)
				idxDec1 := strings.Index(got, "Dec 1st, 2024")
				idxDec15 := strings.Index(got, "Dec 15th, 2024")

				assert.Greater(t, idxDec1, -1, "Dec 1st line should be in output")
				assert.Greater(t, idxDec15, -1, "Dec 15th line should be in output")
				assert.Less(t, idxDec1, idxDec15, "Dec 1st should appear before Dec 15th")
			},
		},
		{
			name:        "text output contains section sign separator and block URL",
			args:        []string{},
			jsonPayload: twoTaskJSON,
			checkOutput: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "§")
				assert.Contains(t, got, "logseq://graph/my-graph?block-id=")
			},
		},
		{
			name:        "multi-line content prints first line only",
			args:        []string{},
			jsonPayload: twoTaskJSON,
			checkOutput: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, got, "TODO buy milk")
				assert.NotContains(t, got, "scheduled::")
			},
		},
		{
			name:        "PostQuery error is propagated",
			args:        []string{},
			jsonPayload: "",
			queryErr:    errConnectionRefused,
			wantErr:     true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer

			mock := &mockTaskLsAPI{queryResult: testCase.jsonPayload, queryErr: testCase.queryErr}
			deps := newTestDeps(mock, &buf)

			c := cmd.NewTaskLsCmd(deps)
			c.SetArgs(testCase.args)

			err := c.Execute()

			if testCase.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			if testCase.checkQuery != nil {
				testCase.checkQuery(t, mock.lastQuery)
			}

			if testCase.checkOutput != nil {
				testCase.checkOutput(t, buf.String())
			}
		})
	}
}

func TestNewTaskLsCmd_NilDeps(t *testing.T) {
	// Verify the constructor works with nil deps (uses all defaults).
	c := cmd.NewTaskLsCmd(nil)

	require.NotNil(t, c)
	assert.Equal(t, "ls [tag...]", c.Use)
	assert.Equal(t, "List tasks from Logseq", c.Short)
}
