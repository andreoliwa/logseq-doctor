package logseqext

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/google/uuid"
)

// ErrDoingTaskNotFound is returned when a task moves while its ID is being assigned.
var ErrDoingTaskNotFound = errors.New("DOING task not found")

// SyncDoingTasks adds block references for every DOING task to the DOING page.
// It returns the number of DOING tasks found.
func SyncDoingTasks(graph *logseq.Graph) (int, error) {
	return syncDoingTasks(graph, nil)
}

// SyncDoingTasksInFiles adds block references for DOING tasks in file paths to the DOING page.
// It returns the number of DOING tasks found.
func SyncDoingTasksInFiles(graph *logseq.Graph, paths []string) (int, error) {
	return syncDoingTasks(graph, paths)
}

func syncDoingTasks(graph *logseq.Graph, paths []string) (int, error) {
	tasks, err := findDoingTasks(graph, paths)
	if err != nil {
		return 0, fmt.Errorf("failed to find DOING tasks: %w", err)
	}

	existingRefs, err := pageBlockRefs(graph, content.TaskStringDoing)
	if err != nil {
		return 0, fmt.Errorf("failed to read DOING page: %w", err)
	}

	transaction := graph.NewTransaction()

	changed, err := ensureDoingTaskIDs(transaction, tasks)
	if err != nil {
		return 0, err
	}

	changed, err = addDoingTaskRefs(transaction, tasks, existingRefs, changed)
	if err != nil {
		return 0, err
	}

	if changed {
		saveErr := transaction.Save()
		if saveErr != nil {
			return 0, fmt.Errorf("failed to save DOING task references: %w", saveErr)
		}
	}

	return len(tasks), nil
}

func ensureDoingTaskIDs(transaction *logseq.Transaction, tasks []doingTask) (bool, error) {
	changed := false

	for index := range tasks {
		if tasks[index].id != "" {
			continue
		}

		page, err := transaction.OpenViaPath(tasks[index].path)
		if err != nil {
			return false, fmt.Errorf("failed to open task page %s: %w", tasks[index].path, err)
		}

		block := blockAtLocation(page.Blocks(), tasks[index].location)
		if block == nil {
			return false, fmt.Errorf("%w in %s", ErrDoingTaskNotFound, tasks[index].path)
		}

		tasks[index].id = EnsureBlockID(block)
		changed = true
	}

	return changed, nil
}

func addDoingTaskRefs(transaction *logseq.Transaction, tasks []doingTask,
	existingRefs map[string]struct{}, changed bool) (bool, error) {
	newRefs := newBlockRefs(tasks, existingRefs)
	if len(newRefs) == 0 {
		return changed, nil
	}

	page, err := transaction.OpenPage(content.TaskStringDoing)
	if err != nil {
		return false, fmt.Errorf("failed to open DOING page: %w", err)
	}

	// Prepend in reverse so the task scan order is preserved at the top of the page.
	for index := range slices.Backward(newRefs) {
		page.PrependBlock(content.NewBlock(content.NewParagraph(content.NewBlockRef(newRefs[index]))))
	}

	return true, nil
}

type doingTask struct {
	path     string
	location []int
	id       string
}

func findDoingTasks(graph *logseq.Graph, paths []string) ([]doingTask, error) {
	if paths != nil {
		return findDoingTasksInFiles(graph, paths)
	}

	paths, err := markdownFiles(graph.Directory())
	if err != nil {
		return nil, err
	}

	return findDoingTasksInFiles(graph, paths)
}

func markdownFiles(graphPath string) ([]string, error) {
	paths := make([]string, 0)

	err := filepath.WalkDir(graphPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		paths = append(paths, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk graph: %w", err)
	}

	return paths, nil
}

func findDoingTasksInFiles(graph *logseq.Graph, paths []string) ([]doingTask, error) {
	tasks := make([]doingTask, 0)

	for _, path := range paths {
		page, err := graph.OpenViaPath(path)
		if err != nil || page == nil {
			continue
		}

		collectDoingTasks(page.Blocks(), path, nil, &tasks)
	}

	return tasks, nil
}

func collectDoingTasks(blocks content.BlockList, path string, parentLocation []int, tasks *[]doingTask) {
	for index, block := range blocks {
		location := append(append([]int{}, parentLocation...), index)
		if blockIsDoing(block) {
			*tasks = append(*tasks, doingTask{
				path:     path,
				location: location,
				id:       blockID(block),
			})
		}

		collectDoingTasks(block.Blocks(), path, location, tasks)
	}
}

func blockIsDoing(block *content.Block) bool {
	isDoing := false

	block.Content().FindDeep(func(node content.Node) bool {
		marker, ok := node.(*content.TaskMarker)
		if ok && marker.Status == content.TaskStatusDoing {
			isDoing = true

			return true
		}

		return false
	})

	return isDoing
}

func pageBlockRefs(graph *logseq.Graph, pageTitle string) (map[string]struct{}, error) {
	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("open page: %w", err)
	}

	refs := make(map[string]struct{})

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(node content.Node) bool {
			if ref, ok := node.(*content.BlockRef); ok {
				refs[ref.ID] = struct{}{}
			}

			return false
		})
	}

	return refs, nil
}

func newBlockRefs(tasks []doingTask, existingRefs map[string]struct{}) []string {
	refs := make([]string, 0)
	seen := make(map[string]struct{})

	for _, task := range tasks {
		if _, exists := existingRefs[task.id]; exists {
			continue
		}

		if _, exists := seen[task.id]; exists {
			continue
		}

		seen[task.id] = struct{}{}
		refs = append(refs, task.id)
	}

	return refs
}

func blockAtLocation(blocks content.BlockList, location []int) *content.Block {
	if len(location) == 0 || location[0] >= len(blocks) {
		return nil
	}

	block := blocks[location[0]]

	for _, index := range location[1:] {
		children := block.Blocks()
		if index >= len(children) {
			return nil
		}

		block = children[index]
	}

	return block
}

// EnsureBlockID returns a block's ID, adding one when needed.
func EnsureBlockID(block *content.Block) string {
	if taskID := blockID(block); taskID != "" {
		return taskID
	}

	taskID := uuid.NewString()
	BlockProperties(block).Set("id", content.NewText(taskID))

	return taskID
}

func blockID(block *content.Block) string {
	var taskID string

	block.Content().FindDeep(func(node content.Node) bool {
		properties, ok := node.(*content.Properties)
		if !ok {
			return false
		}

		for _, value := range properties.Get("id") {
			if text, ok := value.(*content.Text); ok {
				taskID = text.Value

				return true
			}
		}

		return false
	})

	return taskID
}
