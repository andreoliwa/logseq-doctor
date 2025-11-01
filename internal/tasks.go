package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/andreoliwa/logseq-doctor/pkg/set"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
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
	Graph     *logseq.Graph
	Date      time.Time
	Page      string           // Page name to add the task to (empty = journal)
	BlockText string           // Partial text to search for in parent blocks
	Key       string           // Unique key to search for existing task (case-insensitive)
	Name      string           // Short name of the task
	TimeNow   func() time.Time // For testing
}

// AddTask adds a task to Logseq.
// If Key is provided, it searches for an existing task containing that key (case-insensitive)
// and updates it. Otherwise, creates a new task.
// If Page is provided, adds to that page. Otherwise, adds to journal for Date.
// If BlockText is provided, adds as a child of the first block containing that text.
func AddTask(opts *AddTaskOptions) error {
	transaction := opts.Graph.NewTransaction()

	var targetPage logseq.Page

	var err error

	if opts.Page != "" {
		targetPage, err = transaction.OpenPage(opts.Page)
	} else {
		targetPage, err = transaction.OpenJournal(opts.Date)
	}

	if err != nil {
		return fmt.Errorf("error opening target page: %w", err)
	}

	var parentBlock *content.Block
	if opts.BlockText != "" {
		parentBlock = FindBlockContainingText(targetPage, opts.BlockText)
		// If parent not found, parentBlock will be nil and task will be added to top level
	}

	var existingTaskMarker *content.TaskMarker
	if opts.Key != "" {
		existingTaskMarker = FindTaskMarkerByKey(targetPage, parentBlock, opts.Key)
	}

	if existingTaskMarker != nil {
		err = updateExistingTask(existingTaskMarker, opts)
		if err != nil {
			return fmt.Errorf("error updating task: %w", err)
		}
	} else {
		newBlockTask := content.NewBlock(content.NewParagraph(
			content.NewTaskMarker(content.TaskStatusTodo),
			content.NewText(opts.Name),
		))

		if parentBlock != nil {
			parentBlock.AddChild(newBlockTask)
		} else {
			targetPage.AddBlock(newBlockTask)
		}
	}

	err = transaction.Save()
	if err != nil {
		return fmt.Errorf("error saving transaction: %w", err)
	}

	return nil
}

func updateExistingTask(existingTaskMarker *content.TaskMarker, opts *AddTaskOptions) error {
	// Override time provider for testing
	if opts.TimeNow != nil {
		existingTaskMarker.SetTimeNow(opts.TimeNow)
	}

	return updateTaskBackToTodo(existingTaskMarker, opts.Name)
}

func updateTaskBackToTodo(taskMarker *content.TaskMarker, newName string) error {
	_, err := taskMarker.WithStatus(content.TaskStatusTodo)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	paragraph := taskMarker.ParentParagraph()
	if paragraph == nil {
		return nil
	}

	// Remove all children after the task marker and replace with the new name
	// First, collect all children after the task marker to remove them
	var nodesToRemove []content.Node

	foundTaskMarker := false

	for node := paragraph.FirstChild(); node != nil; node = node.NextSibling() {
		if node == taskMarker {
			foundTaskMarker = true

			continue
		}

		if foundTaskMarker {
			nodesToRemove = append(nodesToRemove, node)
		}
	}

	// Remove the collected nodes
	for _, node := range nodesToRemove {
		node.RemoveSelf()
	}

	// Add the new name text after the task marker
	paragraph.AddChild(content.NewText(newName))

	return nil
}
