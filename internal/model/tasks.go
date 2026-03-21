package model

import (
	"github.com/andreoliwa/logseq-doctor/pkg/set"
)

// RefJSON represents a reference entry from the Logseq API response.
type RefJSON struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

// PageJSON holds page-level metadata from the Logseq API.
type PageJSON struct {
	ID           int    `json:"id"`
	JournalDay   int    `json:"journalDay"`
	Name         string `json:"name"`
	OriginalName string `json:"originalName"`
}

// TaskJSON represents a task block from the Logseq HTTP API.
type TaskJSON struct {
	UUID                 string            `json:"uuid"`
	Marker               string            `json:"marker"`
	Content              string            `json:"content"`
	Page                 PageJSON          `json:"page"`
	Deadline             int               `json:"deadline"`
	Scheduled            int               `json:"scheduled"`
	Refs                 []RefJSON         `json:"refs"`
	PathRefs             []RefJSON         `json:"pathRefs"`
	PropertiesTextValues map[string]string `json:"propertiesTextValues"`
}

// CategorizedTasks holds sets of task UUIDs grouped by category.
type CategorizedTasks struct {
	All             *set.Set[string]
	Overdue         *set.Set[string]
	Doing           *set.Set[string]
	FutureScheduled *set.Set[string]
}

// NewCategorizedTasks creates a new CategorizedTasks with initialized sets.
func NewCategorizedTasks() CategorizedTasks {
	return CategorizedTasks{
		All:             set.NewSet[string](),
		Overdue:         set.NewSet[string](),
		Doing:           set.NewSet[string](),
		FutureScheduled: set.NewSet[string](),
	}
}
