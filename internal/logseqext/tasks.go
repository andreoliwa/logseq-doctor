package logseqext

import (
	"fmt"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

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
