package backlog

import (
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/pkg/utils"
	"github.com/fatih/color"
	"strings"
)

var dividerNewTasksContent = content.NewBlock(content.NewParagraph( //nolint:gochecknoglobals
	content.NewPageLink("quick capture"),
	content.NewText(" New tasks above this line"),
))
var dividerNewTasksHash = dividerNewTasksContent.GomegaString()  //nolint:gochecknoglobals
var dividerFocusContent = content.NewBlock(content.NewParagraph( //nolint:gochecknoglobals
	content.NewText("Focus"),
))
var dividerFocusHash = dividerFocusContent.GomegaString() //nolint:gochecknoglobals

type Backlog interface {
	ProcessAll(partialNames []string) error
	// TODO: add to interface
	// ProcessOne(pageTitle string, funcQueryRefs func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error)
}

type backlogImpl struct {
	graph                    *logseq.Graph
	funcProcessSingleBacklog func(graph *logseq.Graph, path string,
		query func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error)
	configReader ConfigReader
}

func NewBacklog(graph *logseq.Graph, reader ConfigReader) Backlog {
	return &backlogImpl{graph: graph, configReader: reader, funcProcessSingleBacklog: processSingleBacklog}
}

func (p *backlogImpl) ProcessAll(partialNames []string) error {
	config, err := p.configReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	allFocusTasks := internal.NewCategorizedTasks()
	processAllPages := len(partialNames) == 0

	if processAllPages {
		fmt.Println("Processing all pages in the backlog")
	} else {
		fmt.Printf("Processing pages with partial names: %s\n", strings.Join(partialNames, ", "))
	}

	for _, backlogConfig := range config.Backlogs {
		processThisPage := processAllPages

		for _, partialName := range partialNames {
			if strings.Contains(strings.ToLower(backlogConfig.OutputPage), strings.ToLower(partialName)) {
				processThisPage = true

				break
			}
		}

		if !processThisPage {
			continue
		}

		focusRefsFromPage, err := processSingleBacklog(p.graph, backlogConfig.OutputPage,
			func() (*internal.CategorizedTasks, error) {
				return queryTasksFromPages(p.graph, backlogConfig.InputPages)
			})
		if err != nil {
			return err
		}

		allFocusTasks.All.Update(focusRefsFromPage)
	}

	if !processAllPages {
		color.Yellow("Skipping focus page because not all pages were processed")

		return nil
	}

	_, err = processSingleBacklog(p.graph, "backlog/Focus", func() (*internal.CategorizedTasks, error) {
		return &allFocusTasks, nil
	})

	return err
}

func processSingleBacklog(graph *logseq.Graph, pageTitle string,
	// TODO: add to interface
	// func (p *backlogImpl) ProcessOne(pageTitle string,
	funcQueryRefs func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error) {
	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page: %w", err)
	}

	existingBlockRefs := blockRefsFromPages(page)

	fmt.Printf("%s: %s\n", internal.PageColor(pageTitle), utils.FormatCount(existingBlockRefs.Size(), "task", "tasks"))

	blockRefsFromQuery, err := funcQueryRefs()
	if err != nil {
		return nil, err
	}

	newBlockRefs := blockRefsFromQuery.All.Diff(existingBlockRefs)
	obsoleteBlockRefs := existingBlockRefs.Diff(blockRefsFromQuery.All)

	focusRefsFromPage, err := insertAndRemoveRefs(graph, pageTitle, existingBlockRefs, newBlockRefs, obsoleteBlockRefs,
		blockRefsFromQuery.Overdue)
	if err != nil {
		return nil, err
	}

	return focusRefsFromPage, nil
}

func blockRefsFromPages(page logseq.Page) *utils.Set[string] {
	existingRefs := utils.NewSet[string]()

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

func queryTasksFromPages(graph *logseq.Graph, pageTitles []string) (*internal.CategorizedTasks, error) {
	tasks := internal.NewCategorizedTasks()

	for _, pageTitle := range pageTitles {
		fmt.Printf("  %s: ", internal.PageColor(pageTitle))

		query, err := findFirstQuery(graph, pageTitle)
		if err != nil {
			return nil, err
		}

		if query == "" {
			query = defaultQuery(pageTitle)

			fmt.Print("default")
		} else {
			fmt.Print("found")
		}

		fmt.Printf(" query %s", query)

		jsonStr, err := internal.QueryLogseqAPI(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query Logseq API: %w", err)
		}

		jsonTasks, err := internal.ExtractTasksFromJSON(jsonStr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tasks: %w", err)
		}

		fmt.Printf(", queried %s\n", utils.FormatCount(len(jsonTasks), "task", "tasks"))

		for _, t := range jsonTasks {
			if t.Overdue() {
				tasks.Overdue.Add(t.UUID)
			}

			tasks.All.Add(t.UUID)
		}
	}

	return &tasks, nil
}

func findFirstQuery(graph *logseq.Graph, pageTitle string) (string, error) {
	var query string

	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		return "", fmt.Errorf("failed to open page: %w", err)
	}

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if q, ok := n.(*content.Query); ok {
				query = q.Query
			} else if qc, ok := n.(*content.QueryCommand); ok {
				query = qc.Query
			}

			if query != "" {
				// Stop after finding one query
				return true
			}

			return false
		})
	}

	if query == "" {
		return "", nil
	}

	return replaceCurrentPage(query, pageTitle), nil
}

// replaceCurrentPage replaces the current page placeholder in the query with the actual page name.
func replaceCurrentPage(query, pageTitle string) string {
	return strings.ReplaceAll(query, "<% current page %>", "[["+pageTitle+"]]")
}

func defaultQuery(pageTitle string) string {
	return fmt.Sprintf("(and [[%s]] (task TODO LATER DOING NOW WAITING))", pageTitle)
}

func insertAndRemoveRefs( //nolint:cyclop,funlen,gocognit
	graph *logseq.Graph, pageTitle string, existingBlockRefs, newBlockRefs, obsoleteBlockRefs,
	overdueBlockRefs *utils.Set[string],
) (*utils.Set[string], error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var firstBlock, dividerNewTasks, dividerFocus *content.Block

	deletedCount := 0
	movedCount := 0
	focusRefs := utils.NewSet[string]()

	for i, block := range page.Blocks() {
		if i == 0 {
			firstBlock = block
		}

		// TODO: add AsMarkdown() or ContentHash() or Hash() to content.Block, to make it possible to compare blocks
		//  Or fix the message "the operator == is not defined on NodeList"
		blockHash := block.GomegaString()
		if blockHash == dividerNewTasksHash {
			dividerNewTasks = block

			continue
		} else if blockHash == dividerFocusHash {
			dividerFocus = block

			continue
		}

		// Remove refs marked for deletion or overdue tasks
		block.Children().FindDeep(func(node content.Node) bool {
			if blockRef, ok := node.(*content.BlockRef); ok {
				shouldDelete := false

				switch {
				case obsoleteBlockRefs.Contains(blockRef.ID):
					shouldDelete = true

					deletedCount++
				case overdueBlockRefs.Contains(blockRef.ID):
					if !nextChildHasPin(node) {
						shouldDelete = true

						existingBlockRefs.Remove(blockRef.ID)

						movedCount++
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
					focusRefs.Add(blockRef.ID)
				}

				return true
			}

			return false
		})
	}

	//  Sections: Focus / Overdue / New tasks

	// Will only add a divider block if there are new tasks to add
	if dividerNewTasks == nil && newBlockRefs.Size() > 0 {
		insertOrAddBlock(page, firstBlock, dividerFocus, dividerNewTasksContent)
	}

	// Insert (new and moved) overdue tasks after the focus section and before the new ones
	//  If they are moved to the top, all overdue tasks will go to the focus page, and this misses the point.
	//  The user should decide manually which tasks should have focus.
	for _, blockRef := range overdueBlockRefs.Values() {
		if existingBlockRefs.Contains(blockRef) {
			continue
		}

		overdueTask := content.NewBlock(content.NewParagraph(
			content.NewText("📅 "),
			content.NewStrong(content.NewText("overdue")),
			content.NewText(" "),
			content.NewBlockRef(blockRef),
			content.NewText(" 📌"),
		))
		insertOrAddBlock(page, firstBlock, dividerFocus, overdueTask)
	}

	// Insert new tasks
	for _, blockRef := range newBlockRefs.Values() {
		if overdueBlockRefs.Contains(blockRef) {
			// Don't add overdue tasks again
			continue
		}

		insertOrAddBlock(page, firstBlock, dividerFocus, content.NewBlock(content.NewBlockRef(blockRef)))
	}

	save := false

	if newBlockRefs.Size() > 0 {
		color.Green("  %s", utils.FormatCount(newBlockRefs.Size(), "new task", "new tasks"))

		save = true
	}

	if deletedCount > 0 {
		color.Red("  %s removed (completed or unreferenced)", utils.FormatCount(deletedCount, "task was", "tasks were"))

		save = true
	}

	if movedCount > 0 {
		color.Magenta("  %s moved around", utils.FormatCount(movedCount, "task was", "tasks were"))

		save = true
	}

	if save {
		err = transaction.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save transaction: %w", err)
		}
	} else {
		color.Yellow("  no new/deleted/moved tasks")
	}

	if dividerFocus == nil {
		focusRefs.Clear()
	}

	return focusRefs, nil
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

func insertOrAddBlock(page logseq.Page, firstBlock *content.Block, dividerFocus *content.Block,
	newTask *content.Block) {
	switch {
	case dividerFocus != nil:
		page.InsertBlockAfter(newTask, dividerFocus)
	case firstBlock != nil:
		page.InsertBlockBefore(newTask, firstBlock)
	default:
		page.AddBlock(newTask)
	}
}
