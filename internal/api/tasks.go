package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/andreoliwa/logseq-go/content"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/pkg/set"
)

// TaskUUID is a type alias for task block UUIDs, making it clear when a string represents a Logseq block UUID.
type TaskUUID = string

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
	UUID                 TaskUUID          `json:"uuid"`
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
	All             *set.Set[TaskUUID]
	Overdue         *set.Set[TaskUUID]
	Doing           *set.Set[TaskUUID]
	FutureScheduled *set.Set[TaskUUID]
	TaskLookup      map[TaskUUID]TaskJSON
}

// NewCategorizedTasks creates a new CategorizedTasks with initialized sets.
func NewCategorizedTasks() CategorizedTasks {
	return CategorizedTasks{
		All:             set.NewSet[TaskUUID](),
		Overdue:         set.NewSet[TaskUUID](),
		Doing:           set.NewSet[TaskUUID](),
		FutureScheduled: set.NewSet[TaskUUID](),
		TaskLookup:      make(map[TaskUUID]TaskJSON),
	}
}

// ExtractTasksFromJSON parses a JSON string into a slice of TaskJSON.
func ExtractTasksFromJSON(jsonStr string) ([]TaskJSON, error) {
	var tasks []TaskJSON

	err := json.Unmarshal([]byte(jsonStr), &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tasks, nil
}

// TaskOverdue checks if the task is overdue based on deadline or scheduled date.
func TaskOverdue(t TaskJSON, currentTime func() time.Time) bool {
	currentDate := logseqext.DateYYYYMMDD(currentTime())

	return (t.Deadline > 0 && t.Deadline <= currentDate) || (t.Scheduled > 0 && t.Scheduled <= currentDate)
}

// TaskDoing checks if the task has the DOING marker.
func TaskDoing(t TaskJSON) bool {
	return t.Marker == content.TaskStringDoing
}

// TaskFutureScheduled checks if the task is scheduled for the future (tomorrow onwards) and it's not overdue.
func TaskFutureScheduled(t TaskJSON, currentTime func() time.Time) bool {
	currentDate := logseqext.DateYYYYMMDD(currentTime())

	return t.Scheduled > currentDate && !TaskOverdue(t, currentTime)
}
