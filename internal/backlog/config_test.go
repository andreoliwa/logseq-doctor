package backlog_test

import (
	"github.com/andreoliwa/lsd/internal/backlog"
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestPageConfigReader_emptyBacklog(t *testing.T) {
	graph := testutils.FakeGraph(t)
	reader := backlog.NewPageConfigReader(graph, "non-existing")

	result, err := reader.ReadConfig()
	require.NoError(t, err)

	expected := backlog.Config{
		Backlogs: nil,
	}
	assert.Equal(t, &expected, result)
}

func TestPageConfigReader_ReadConfig(t *testing.T) {
	graph := testutils.FakeGraph(t)
	config := "config"
	reader := backlog.NewPageConfigReader(graph, config)

	result, err := reader.ReadConfig()
	require.NoError(t, err)

	expected := backlog.Config{
		Backlogs: []backlog.SingleBacklogConfig{
			{
				Icon:       "",
				InputPages: []string{"computer", "Android", "iOS"},
				OutputPage: config + "/computer",
			},
			{
				Icon:       "",
				InputPages: []string{"house"},
				OutputPage: config + "/house",
			},
			{
				Icon:       "",
				InputPages: []string{"work", "office"},
				OutputPage: config + "/work",
			},
		},
	}
	assert.Equal(t, &expected, result)
}
