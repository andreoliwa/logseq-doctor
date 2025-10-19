package backlog_test

import (
	"testing"

	"github.com/andreoliwa/lqd/internal/backlog"
	"github.com/andreoliwa/lqd/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageConfigReader_emptyBacklog(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	reader := backlog.NewPageConfigReader(graph, "non-existing")

	result, err := reader.ReadConfig()
	require.NoError(t, err)

	expected := backlog.Config{
		FocusPage: "non-existing/Focus",
		Backlogs:  nil,
	}
	assert.Equal(t, &expected, result)
}

func TestPageConfigReader_ReadConfig(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	prefix := "config"
	reader := backlog.NewPageConfigReader(graph, prefix)

	result, err := reader.ReadConfig()
	require.NoError(t, err)

	expected := backlog.Config{
		FocusPage: "config/Focus",
		Backlogs: []backlog.SingleBacklogConfig{
			{
				BacklogPage: prefix + "/computer",
				Icon:        "",
				InputPages:  []string{"computer", "Android", "iOS"},
			},
			{
				BacklogPage: prefix + "/house",
				Icon:        "",
				InputPages:  []string{"house"},
			},
			{
				BacklogPage: prefix + "/work",
				Icon:        "",
				InputPages:  []string{"work", "office"},
			},
			{
				BacklogPage: prefix + "/start",
				Icon:        "",
				InputPages:  []string{"pages", "same", "line"},
			},
			{
				BacklogPage: prefix + "/middle",
				Icon:        "",
				InputPages:  []string{"link", "anywhere", "line"},
			},
			{
				BacklogPage: prefix + "/end",
				Icon:        "",
				InputPages:  []string{"link", "also", "last one"},
			},
			{
				BacklogPage: prefix + "/start-again",
				Icon:        "",
				InputPages:  []string{"page", "tag"},
			},
			{
				BacklogPage: prefix + "/middle-again",
				Icon:        "",
				InputPages:  []string{"page-again", "tag"},
			},
			{
				BacklogPage: prefix + "/end-again",
				Icon:        "",
				InputPages:  []string{"tag", "page"},
			},
		},
	}
	assert.Equal(t, &expected, result)
}
