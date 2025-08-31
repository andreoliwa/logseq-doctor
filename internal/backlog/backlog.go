package backlog

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/pkg/set"
	"github.com/fatih/color"
)

type Result struct {
	FocusRefsFromPage *set.Set[string]
	ShowQuickCapture  bool
}

type Backlog interface {
	Graph() *logseq.Graph
	ProcessAll(partialNames []string) error
	ProcessOne(pageTitle string, funcQueryRefs func() (*internal.CategorizedTasks, error)) (*Result, error)
}

type backlogImpl struct {
	graph        *logseq.Graph
	api          internal.LogseqAPI
	configReader ConfigReader
}

func NewBacklog(graph *logseq.Graph, api internal.LogseqAPI, reader ConfigReader) Backlog {
	return &backlogImpl{graph: graph, api: api, configReader: reader}
}

func (b *backlogImpl) Graph() *logseq.Graph {
	return b.graph
}

func (b *backlogImpl) ProcessAll(partialNames []string) error { //nolint:cyclop
	config, err := b.configReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	allFocusTasks := internal.NewCategorizedTasks()
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
			func() (*internal.CategorizedTasks, error) {
				return queryTasksFromPages(b.graph, b.api, backlogConfig.InputPages)
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

	_, err = b.ProcessOne(config.FocusPage, func() (*internal.CategorizedTasks, error) {
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
	funcQueryRefs func() (*internal.CategorizedTasks, error)) (*Result, error) {
	page := internal.OpenPage(b.graph, pageTitle)

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
	obsoleteBlockRefs := existingBlockRefs.Diff(allValidRefs)

	result, err := insertAndRemoveRefs(b.graph, pageTitle, newBlockRefs, obsoleteBlockRefs,
		blockRefsFromQuery.Overdue)
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
func queryTasksFromPages(graph *logseq.Graph, api internal.LogseqAPI,
	pageTitles []string) (*internal.CategorizedTasks, error) {
	tasks := internal.NewCategorizedTasks()
	finder := internal.NewLogseqFinder(graph)

	if len(pageTitles) <= 1 {
		return queryTasksFromPagesSequential(api, pageTitles, &tasks, finder)
	}

	return queryTasksFromPagesConcurrent(api, pageTitles, &tasks, finder)
}

// queryTasksFromPagesSequential processes pages sequentially (original implementation).
func queryTasksFromPagesSequential(api internal.LogseqAPI,
	pageTitles []string, tasks *internal.CategorizedTasks,
	finder internal.LogseqFinder) (*internal.CategorizedTasks, error) {
	for _, pageTitle := range pageTitles {
		jsonTasks, err := queryTasksFromSinglePage(api, pageTitle, finder)
		if err != nil {
			return nil, err
		}

		fmt.Printf(" %s: ", internal.PageColor(pageTitle))
		fmt.Print(FormatCount(len(jsonTasks), "task", "tasks"))

		addTasksToCategories(jsonTasks, tasks)
	}

	return tasks, nil
}

// queryTasksFromPagesConcurrent processes pages concurrently using goroutines.
func queryTasksFromPagesConcurrent(api internal.LogseqAPI,
	pageTitles []string, tasks *internal.CategorizedTasks,
	finder internal.LogseqFinder) (*internal.CategorizedTasks, error) {
	type pageResult struct {
		pageTitle string
		jsonTasks []internal.TaskJSON
		err       error
	}

	resultChan := make(chan pageResult, len(pageTitles))

	for _, pageTitle := range pageTitles {
		go func(title string) {
			jsonTasks, err := queryTasksFromSinglePage(api, title, finder)
			resultChan <- pageResult{pageTitle: title, jsonTasks: jsonTasks, err: err}
		}(pageTitle)
	}

	for i := 0; i < len(pageTitles); i++ {
		result := <-resultChan

		if result.err != nil {
			return nil, result.err
		}

		// Print results in the order they complete (may be different from input order)
		fmt.Printf(" %s: ", internal.PageColor(result.pageTitle))
		fmt.Print(FormatCount(len(result.jsonTasks), "task", "tasks"))

		addTasksToCategories(result.jsonTasks, tasks)
	}

	return tasks, nil
}

// queryTasksFromSinglePage queries tasks from a single page and returns the JSON tasks.
func queryTasksFromSinglePage(api internal.LogseqAPI, pageTitle string,
	finder internal.LogseqFinder) ([]internal.TaskJSON, error) {
	query := finder.FindFirstQuery(pageTitle)
	if query == "" {
		query = defaultQuery(pageTitle)
	}

	jsonStr, err := api.PostQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query Logseq API: %w", err)
	}

	jsonTasks, err := internal.ExtractTasksFromJSON(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tasks: %w", err)
	}

	return jsonTasks, nil
}

// addTasksToCategories adds tasks to the appropriate categories in CategorizedTasks.
func addTasksToCategories(jsonTasks []internal.TaskJSON, tasks *internal.CategorizedTasks) {
	for _, task := range jsonTasks {
		if task.Overdue() {
			tasks.Overdue.Add(task.UUID)
		}

		if task.Doing() {
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

func insertAndRemoveRefs( //nolint:cyclop,funlen,gocognit
	graph *logseq.Graph, pageTitle string, newBlockRefs, obsoleteBlockRefs,
	overdueBlockRefs *set.Set[string],
) (*Result, error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var firstBlock, dividerNewTasks, dividerOverdue, dividerFocus *content.Block

	deletedCount := 0
	movedCount := 0
	unpinnedCount := 0
	result := &Result{
		FocusRefsFromPage: set.NewSet[string](),
		ShowQuickCapture:  false,
	}
	pinnedBlockRefs := set.NewSet[string]()

	for i, block := range page.Blocks() {
		if i == 0 {
			firstBlock = block
		}

		// Remove refs marked for deletion or overdue tasks
		block.Children().FindDeep(func(node content.Node) bool {
			if text, ok := node.(*content.Text); ok {
				switch {
				case strings.Contains(text.Value, "New tasks"):
					dividerNewTasks = block
				case strings.Contains(text.Value, "Overdue tasks"):
					dividerOverdue = block
				case text.Value == "Focus":
					dividerFocus = block
				}
			}

			if blockRef, ok := node.(*content.BlockRef); ok { //nolint:nestif
				shouldDelete := false

				switch {
				case obsoleteBlockRefs.Contains(blockRef.ID):
					shouldDelete = true

					deletedCount++
				case overdueBlockRefs.Contains(blockRef.ID):
					if nextChildHasPin(node) {
						pinnedBlockRefs.Add(blockRef.ID)
					} else {
						shouldDelete = true

						movedCount++
					}
				default:
					// Here we have an existing task that's not overdue.
					// Find the pin text node and remove it.
					if nextChildHasPin(node) {
						nextChild := node.NextSibling()
						if nextChild != nil {
							nextChild.RemoveSelf()

							unpinnedCount++
						}
					}
				}

				if shouldDelete {
					// Block ref's parents are: paragraph and block
					// TODO: handle cases when the block ref is nested under another block ref.
					//  This will remove the obsolete block and its children.
					//  Should I show a warning message to the user and prevent the block from being deleted?
					blockRef.Parent().Parent().RemoveSelf()
				} else if dividerFocus == nil {
					// Keep adding tasks to the focus section until the divider is found
					result.FocusRefsFromPage.Add(blockRef.ID)
				}

				return false
			}

			return false
		})
	}

	//  Sections: Focus / Overdue / New tasks

	// Insert (new and moved) overdue tasks after the focus section and before the new ones
	//  If they are moved to the top, all overdue tasks will go to the focus page, and this misses the point.
	//  The user should decide manually which tasks should have focus.
	for _, blockRef := range overdueBlockRefs.ValuesSorted() {
		if pinnedBlockRefs.Contains(blockRef) {
			continue
		}

		if dividerOverdue == nil {
			dividerOverdue = content.NewBlock(content.NewParagraph(
				content.NewText("📅 Overdue tasks "),
				content.NewPageLink("quick capture"),
			))
			AddSibling(page, dividerOverdue, firstBlock, dividerFocus)
		}

		overdueTask := content.NewBlock(content.NewParagraph(
			content.NewBlockRef(blockRef),
			content.NewText("📅📌"),
		))
		dividerOverdue.AddChild(overdueTask)
	}

	save := false

	// Insert new tasks
	if newBlockRefs.Size() > 0 {
		for _, blockRef := range newBlockRefs.ValuesSorted() {
			if overdueBlockRefs.Contains(blockRef) {
				// Don't add overdue tasks again as new tasks
				continue
			}

			if dividerNewTasks == nil {
				dividerNewTasks = content.NewBlock(content.NewParagraph(
					content.NewText("New tasks "),
					content.NewPageLink("quick capture"),
				))
				AddSibling(page, dividerNewTasks, firstBlock, dividerOverdue, dividerFocus)
			}

			dividerNewTasks.AddChild(content.NewBlock(content.NewBlockRef(blockRef)))
		}

		color.Green(" %s", FormatCount(newBlockRefs.Size(), "new task", "new tasks"))

		save = true
		result.ShowQuickCapture = true
	}

	save = removeEmptyDividers(save, dividerNewTasks, dividerOverdue)

	if deletedCount > 0 {
		// Remove completed or unreferenced tasks
		color.Red(" %s removed", FormatCount(deletedCount, "task was", "tasks were"))

		save = true
	}

	if movedCount > 0 {
		color.Magenta(" %s moved around", FormatCount(movedCount, "task was", "tasks were"))

		save = true
		result.ShowQuickCapture = true
	}

	if unpinnedCount > 0 {
		color.Cyan(" %s unpinned", FormatCount(unpinnedCount, "task was", "tasks were"))

		save = true
	}

	if save {
		err = transaction.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save transaction: %w", err)
		}
	} else {
		color.Yellow(" no changes")
	}

	if dividerFocus == nil {
		result.FocusRefsFromPage.Clear()
	}

	return result, nil
}

func nextChildHasPin(node content.Node) bool {
	nextChild := node.NextSibling()
	if nextChild != nil {
		if text, ok := nextChild.(*content.Text); ok {
			return strings.Contains(text.Value, "📌")
		}
	}

	return false
}

func AddSibling(page logseq.Page, newBlock, before *content.Block, after ...*content.Block) {
	for _, a := range after {
		if a != nil {
			page.InsertBlockAfter(newBlock, a)

			return
		}
	}

	if before != nil {
		page.InsertBlockBefore(newBlock, before)

		return
	}

	page.AddBlock(newBlock)
}

// removeEmptyDividers removes empty dividers (no blocks under it) and returns true if any were removed.
func removeEmptyDividers(save bool, dividers ...*content.Block) bool {
	for _, divider := range dividers {
		if divider != nil && len(divider.Blocks()) == 0 {
			divider.RemoveSelf()

			save = true
		}
	}

	return save
}
