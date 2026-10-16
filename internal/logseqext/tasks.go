package logseqext

import (
	"errors"
	"fmt"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

var errInvalidTaskDescription = errors.New("task description must begin with a paragraph")

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
		err = addNewTask(targetPage, parentBlock, opts.Name)
		if err != nil {
			return err
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

	err := updateTaskBackToTodo(existingTaskMarker, opts.Name)
	if err != nil {
		return err
	}

	RepositionBlock(existingTaskMarker.ParentBlock())

	return nil
}

func updateTaskBackToTodo(taskMarker *content.TaskMarker, newName string) error {
	description, parseErr := logseq.ParseBlock(content.TaskStringTodo + " " + newName)
	if parseErr != nil {
		return fmt.Errorf("parse task description: %w", parseErr)
	}

	content.NormalizeLeadingTime(description)

	_, err := taskMarker.WithStatus(content.TaskStatusTodo)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	paragraph := taskMarker.ParentParagraph()
	if paragraph == nil {
		return nil
	}

	// Replace only inline content after the retained task marker. Save the next
	// node before removal, which clears its sibling links.
	for node := taskMarker.NextSibling(); node != nil; {
		next := node.NextSibling()
		paragraph.RemoveChild(node)
		node = next
	}

	// Add parsed description nodes after the retained task marker.
	replacement, ok := description.FirstChild().(*content.Paragraph)
	if !ok {
		return errInvalidTaskDescription
	}

	for _, node := range replacement.Children()[1:] {
		paragraph.AddChild(node)
	}

	// Keep every supplied description paragraph without replacing the block's
	// existing metadata, logbook or child blocks.
	block := taskMarker.ParentBlock()

	var after content.Node = paragraph
	for _, node := range description.Children()[1:] {
		block.InsertChildAfter(node, after)
		after = node
	}

	return nil
}

func addNewTask(page logseq.Page, parent *content.Block, name string) error {
	block, err := logseq.ParseBlock(content.TaskStringTodo + " " + name)
	if err != nil {
		return fmt.Errorf("parse task description: %w", err)
	}

	content.NormalizeLeadingTime(block)

	PlaceBlock(page, parent, block)

	return nil
}
