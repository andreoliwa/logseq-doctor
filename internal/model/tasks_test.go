package model_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewCategorizedTasks(t *testing.T) {
	tasks := model.NewCategorizedTasks()

	assert.NotNil(t, tasks.All, "All set should be initialized")
	assert.NotNil(t, tasks.Overdue, "Overdue set should be initialized")
	assert.NotNil(t, tasks.Doing, "Doing set should be initialized")
	assert.NotNil(t, tasks.FutureScheduled, "FutureScheduled set should be initialized")

	assert.Equal(t, 0, tasks.All.Size(), "All set should be empty")
	assert.Equal(t, 0, tasks.Overdue.Size(), "Overdue set should be empty")
	assert.Equal(t, 0, tasks.Doing.Size(), "Doing set should be empty")
	assert.Equal(t, 0, tasks.FutureScheduled.Size(), "FutureScheduled set should be empty")
}

func TestTaskJSONNewFields(t *testing.T) {
	task := model.TaskJSON{
		UUID:    "test-uuid",
		Marker:  "TODO",
		Content: "Test task",
		Page: model.PageJSON{
			ID:           1,
			JournalDay:   20250101,
			Name:         "test-page",
			OriginalName: "Test Page",
		},
		Refs:                 []model.RefJSON{{ID: 1, Name: "ref1"}},
		PathRefs:             []model.RefJSON{{ID: 2, Name: "ref2"}},
		PropertiesTextValues: map[string]string{"key": "value"},
	}

	assert.Equal(t, "test-uuid", task.UUID)
	assert.Equal(t, 1, task.Page.ID)
	assert.Equal(t, "test-page", task.Page.Name)
	assert.Equal(t, "Test Page", task.Page.OriginalName)
	assert.Len(t, task.Refs, 1)
	assert.Equal(t, "ref1", task.Refs[0].Name)
	assert.Len(t, task.PathRefs, 1)
	assert.Equal(t, "value", task.PropertiesTextValues["key"])
}
