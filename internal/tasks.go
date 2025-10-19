package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/pkg/set"
)

type TaskJSON struct {
	UUID      string   `json:"uuid"`
	Marker    string   `json:"marker"`
	Content   string   `json:"content"`
	Page      pageJSON `json:"page"`
	Deadline  int      `json:"deadline"`
	Scheduled int      `json:"scheduled"`
}

type pageJSON struct {
	JournalDay int `json:"journalDay"`
}

func ExtractTasksFromJSON(jsonStr string) ([]TaskJSON, error) {
	var tasks []TaskJSON

	err := json.Unmarshal([]byte(jsonStr), &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tasks, nil
}

// Overdue checks if the task is overdue based on deadline or scheduled date.
func (t TaskJSON) Overdue(currentTime func() time.Time) bool {
	currentDate := DateYYYYMMDD(currentTime())

	return (t.Deadline > 0 && t.Deadline <= currentDate) || (t.Scheduled > 0 && t.Scheduled <= currentDate)
}

// Doing checks if the task has the DOING marker.
func (t TaskJSON) Doing() bool {
	return t.Marker == "DOING"
}

// FutureScheduled checks if the task is scheduled for the future (tomorrow onwards) and it's not overdue.
func (t TaskJSON) FutureScheduled(currentTime func() time.Time) bool {
	currentDate := DateYYYYMMDD(currentTime())

	return t.Scheduled > currentDate && !t.Overdue(currentTime)
}

type CategorizedTasks struct {
	All             *set.Set[string]
	Overdue         *set.Set[string]
	Doing           *set.Set[string]
	FutureScheduled *set.Set[string]
}

func NewCategorizedTasks() CategorizedTasks {
	return CategorizedTasks{
		All:             set.NewSet[string](),
		Overdue:         set.NewSet[string](),
		Doing:           set.NewSet[string](),
		FutureScheduled: set.NewSet[string](),
	}
}

// AddTaskOptions contains options for adding a task to Logseq.
type AddTaskOptions struct {
	Graph       *logseq.Graph
	Date        time.Time
	Description string // Task description/content
	Page        string // Page name to add the task to (empty = journal)
	BlockText   string // Partial text to search for in parent blocks
	Key         string // Unique key to search for existing task (case-insensitive)
	Name        string // Short description of the task
}

// AddTask adds a task to Logseq.
// If Key is provided, it searches for an existing task containing that key (case-insensitive)
// and updates it. Otherwise, creates a new task.
// If Page is provided, adds to that page. Otherwise, adds to journal for Date.
// If BlockText is provided, adds as a child of the first block containing that text.
func AddTask(_ *AddTaskOptions) error {
	// TODO: Implement task addition logic
	// This is a placeholder for the command structure
	panic("AddTask not yet implemented")
}
