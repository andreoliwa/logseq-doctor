package backlog

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/pkg/set"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/fatih/color"
)

type Result struct {
	FocusRefsFromPage *set.Set[string]
	ShowQuickCapture  bool
}

type Backlog interface {
	Graph() *logseq.Graph
	ProcessAll(partialNames []string) error
	ProcessOne(pageTitle string, funcQueryRefs func() (*logseqapi.CategorizedTasks, error)) (*Result, error)
}

type backlogImpl struct {
	graph        *logseq.Graph
	logseqAPI    logseqapi.LogseqAPI
	configReader ConfigReader
	currentTime  func() time.Time
}

func NewBacklog(graph *logseq.Graph, logseqAPI logseqapi.LogseqAPI, reader ConfigReader,
	currentTime func() time.Time) Backlog {
	return &backlogImpl{graph: graph, logseqAPI: logseqAPI, configReader: reader, currentTime: currentTime}
}

func (b *backlogImpl) Graph() *logseq.Graph {
	return b.graph
}

func (b *backlogImpl) ProcessAll(partialNames []string) error { //nolint:cyclop
	config, err := b.configReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	allFocusTasks := logseqapi.NewCategorizedTasks()
	processAllPages := len(partialNames) == 0
	showQuickCapture := false

	if processAllPages {
		fmt.Println("Processing all pages in the backlog")
	} else {
		fmt.Printf("Processing pages with partial names: %s\n", strings.Join(partialNames, ", "))
	}

	for _, backlogConfig := range config.Backlogs {
		processThisPage := processAllPages

		for _, partialName := range partialNames {
			if strings.Contains(strings.ToLower(backlogConfig.BacklogPage), strings.ToLower(partialName)) {
				processThisPage = true

				break
			}
		}

		if !processThisPage {
			continue
		}

		result, err := b.ProcessOne(backlogConfig.BacklogPage,
			func() (*logseqapi.CategorizedTasks, error) {
				return queryTasksFromPages(b.graph, b.logseqAPI, backlogConfig.InputPages, b.currentTime)
			})
		if err != nil {
			return err
		}

		allFocusTasks.All.Update(result.FocusRefsFromPage)

		if result.ShowQuickCapture {
			showQuickCapture = true
		}
	}

	if showQuickCapture {
		// The original idea was to wait for a keypress while the user checks the quick capture page,
		// so they can move tasks to the focus section. But new focus tasks would not be detected at this point,
		// we would have to refresh the list of focus tasks after the keypress.
		defer printQuickCaptureURL(b.graph)
	}

	if !processAllPages {
		color.Yellow("Skipping focus page because not all pages were processed")

		return nil
	}

	_, err = b.ProcessOne(config.FocusPage, func() (*logseqapi.CategorizedTasks, error) {
		return &allFocusTasks, nil
	})

	return err
}

func printQuickCaptureURL(graph *logseq.Graph) {
	basename := filepath.Base(graph.Directory())

	fmt.Print("\nCheck new content: ")
	color.Red("logseq://graph/%s?page=quick+capture\n", basename)
}

func (b *backlogImpl) ProcessOne(pageTitle string,
	funcQueryRefs func() (*logseqapi.CategorizedTasks, error)) (*Result, error) {
	page := logseqapi.OpenPage(b.graph, pageTitle)

	existingBlockRefs := blockRefsFromPages(page)

	fmt.Printf("%s: %s", internal.PageColor(pageTitle), FormatCount(existingBlockRefs.Size(), "task", "tasks"))

	blockRefsFromQuery, err := funcQueryRefs()
	if err != nil {
		return nil, err
	}

	newBlockRefs := blockRefsFromQuery.All.Diff(existingBlockRefs)

	// Calculate obsolete refs, but exclude DOING tasks from removal
	// DOING tasks should be preserved even if they're not in the All set
	allValidRefs := set.NewSet[string]()
	allValidRefs.Update(blockRefsFromQuery.All)
	allValidRefs.Update(blockRefsFromQuery.Doing)
	allValidRefs.Update(blockRefsFromQuery.FutureScheduled)
	obsoleteBlockRefs := existingBlockRefs.Diff(allValidRefs)

	result, err := insertAndRemoveRefs(b.graph, pageTitle, newBlockRefs, obsoleteBlockRefs,
		blockRefsFromQuery.Overdue, blockRefsFromQuery.FutureScheduled)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func blockRefsFromPages(page logseq.Page) *set.Set[string] {
	existingRefs := set.NewSet[string]()

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if ref, ok := n.(*content.BlockRef); ok {
				existingRefs.Add(ref.ID)
			}

			return false
		})
	}

	return existingRefs
}

// queryTasksFromPages queries Logseq API for tasks from specified pages.
// It uses concurrent processing for multiple pages and sequential processing for a single page.
func queryTasksFromPages(graph *logseq.Graph, logseqAPI logseqapi.LogseqAPI,
	pageTitles []string, currentTime func() time.Time) (*logseqapi.CategorizedTasks, error) {
	tasks := logseqapi.NewCategorizedTasks()
	finder := logseqext.NewLogseqFinder(graph)

	if len(pageTitles) <= 1 {
		return queryTasksFromPagesSequential(logseqAPI, pageTitles, &tasks, finder, currentTime)
	}

	return queryTasksFromPagesConcurrent(logseqAPI, pageTitles, &tasks, finder, currentTime)
}

// queryTasksFromPagesSequential processes pages sequentially (original implementation).
func queryTasksFromPagesSequential(logseqAPI logseqapi.LogseqAPI,
	pageTitles []string, tasks *logseqapi.CategorizedTasks,
	finder logseqext.LogseqFinder, currentTime func() time.Time) (*logseqapi.CategorizedTasks, error) {
	for _, pageTitle := range pageTitles {
		jsonTasks, err := queryTasksFromSinglePage(logseqAPI, pageTitle, finder)
		if err != nil {
			return nil, err
		}

		fmt.Printf(" %s: ", internal.PageColor(pageTitle))
		fmt.Print(FormatCount(len(jsonTasks), "task", "tasks"))

		addTasksToCategories(jsonTasks, tasks, currentTime)
	}

	return tasks, nil
}

// queryTasksFromPagesConcurrent processes pages concurrently using goroutines.
func queryTasksFromPagesConcurrent(logseqAPI logseqapi.LogseqAPI,
	pageTitles []string, tasks *logseqapi.CategorizedTasks,
	finder logseqext.LogseqFinder, currentTime func() time.Time) (*logseqapi.CategorizedTasks, error) {
	type pageResult struct {
		pageTitle string
		jsonTasks []logseqapi.TaskJSON
		err       error
	}

	resultChan := make(chan pageResult, len(pageTitles))

	for _, pageTitle := range pageTitles {
		go func(title string) {
			jsonTasks, err := queryTasksFromSinglePage(logseqAPI, title, finder)
			resultChan <- pageResult{pageTitle: title, jsonTasks: jsonTasks, err: err}
		}(pageTitle)
	}

	for range pageTitles {
		result := <-resultChan

		if result.err != nil {
			return nil, result.err
		}

		// Print results in the order they complete (may be different from input order)
		fmt.Printf(" %s: ", internal.PageColor(result.pageTitle))
		fmt.Print(FormatCount(len(result.jsonTasks), "task", "tasks"))

		addTasksToCategories(result.jsonTasks, tasks, currentTime)
	}

	return tasks, nil
}

// queryTasksFromSinglePage queries tasks from a single page and returns the JSON tasks.
func queryTasksFromSinglePage(logseqAPI logseqapi.LogseqAPI, pageTitle string,
	finder logseqext.LogseqFinder) ([]logseqapi.TaskJSON, error) {
	query := finder.FindFirstQuery(pageTitle)
	if query == "" {
		query = defaultQuery(pageTitle)
	}

	jsonStr, err := logseqAPI.PostQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query Logseq API: %w", err)
	}

	jsonTasks, err := logseqapi.ExtractTasksFromJSON(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tasks: %w", err)
	}

	return jsonTasks, nil
}

// addTasksToCategories adds tasks to the appropriate categories in CategorizedTasks.
func addTasksToCategories(jsonTasks []logseqapi.TaskJSON, tasks *logseqapi.CategorizedTasks,
	currentTime func() time.Time) {
	for _, task := range jsonTasks {
		if logseqapi.TaskOverdue(task, currentTime) {
			tasks.Overdue.Add(task.UUID)
		}

		if logseqapi.TaskFutureScheduled(task, currentTime) {
			tasks.FutureScheduled.Add(task.UUID)
		}

		if logseqapi.TaskDoing(task) {
			tasks.Doing.Add(task.UUID)
		} else {
			// Only add non-DOING tasks to All set
			// DOING tasks should not be added to backlog as new tasks
			tasks.All.Add(task.UUID)
		}
	}
}

func defaultQuery(pageTitle string) string {
	return fmt.Sprintf("(and [[%s]] (task TODO LATER DOING NOW WAITING))", pageTitle)
}
