package pkg

import (
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/pkg/utils"
	"github.com/fatih/color"
	"strings"
)

const backlogName = "backlog"

var dividerNewTasksContent = content.NewBlock(content.NewParagraph( //nolint:gochecknoglobals
	content.NewPageLink("quick capture"),
	content.NewText(" New tasks above this line"),
))
var dividerNewTasksHash = dividerNewTasksContent.GomegaString()  //nolint:gochecknoglobals
var dividerFocusContent = content.NewBlock(content.NewParagraph( //nolint:gochecknoglobals
	content.NewText("Focus"),
))
var dividerFocusHash = dividerFocusContent.GomegaString() //nolint:gochecknoglobals

type Backlog struct {
	Graph                    *logseq.Graph
	FuncProcessSingleBacklog func(graph *logseq.Graph, path string,
		query func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error)
}

func NewBacklog(graph *logseq.Graph) *Backlog {
	return &Backlog{
		Graph:                    graph,
		FuncProcessSingleBacklog: processSingleBacklog,
	}
}

func (p *Backlog) ProcessBacklogs(partialNames []string) error {
	lines, err := linesWithPages(p.Graph)
	if err != nil {
		return err
	}

	allFocusTasks := internal.NewCategorizedTasks()
	processAllPages := len(partialNames) == 0

	if processAllPages {
		fmt.Println("Processing all pages in the backlog")
	} else {
		fmt.Printf("Processing pages with partial names: %s\n", strings.Join(partialNames, ", "))
	}

	for _, pages := range lines {
		title := pages[0]
		processThisPage := processAllPages

		for _, partialName := range partialNames {
			if strings.Contains(strings.ToLower(title), strings.ToLower(partialName)) {
				processThisPage = true

				break
			}
		}

		if !processThisPage {
			continue
		}

		focusRefsFromPage, err := p.FuncProcessSingleBacklog(p.Graph, "backlog/"+title,
			func() (*internal.CategorizedTasks, error) {
				return queryTasksFromPages(p.Graph, pages)
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

	_, err = p.FuncProcessSingleBacklog(p.Graph, "backlog/Focus", func() (*internal.CategorizedTasks, error) {
		return &allFocusTasks, nil
	})

	return err
}

func processSingleBacklog(graph *logseq.Graph, pageTitle string,
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

	if newBlockRefs.Size() <= 0 && obsoleteBlockRefs.Size() <= 0 {
		color.Yellow("  no new/deleted tasks found")
	}

	focusRefsFromPage, err := insertAndRemoveRefs(graph, pageTitle, newBlockRefs, obsoleteBlockRefs)
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

func linesWithPages(graph *logseq.Graph) ([][]string, error) {
	page, err := graph.OpenPage(backlogName)
	if err != nil {
		return nil, fmt.Errorf("failed to open backlog page: %w", err)
	}

	var lines [][]string

	for _, block := range page.Blocks() {
		var pageTitles []string

		block.Children().FindDeep(func(n content.Node) bool {
			if pageLink, ok := n.(*content.PageLink); ok {
				pageTitles = append(pageTitles, pageLink.To)
			} else if tag, ok := n.(*content.Hashtag); ok {
				pageTitles = append(pageTitles, tag.To)
			}

			return false
		})

		if len(pageTitles) > 0 {
			lines = append(lines, pageTitles)
		}
	}

	if len(lines) == 0 {
		fmt.Println("no pages found in the backlog")
	}

	return lines, nil
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
	graph *logseq.Graph, pageTitle string, newRefs, obsoleteRefs *utils.Set[string]) (*utils.Set[string], error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var firstBlock, dividerNewTasks, dividerFocus, blockAfterFocus, insertPoint *content.Block

	deletedCount := 0
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
		} else if blockHash == dividerFocusHash {
			dividerFocus = block

			nodeAfterFocus := block.NextSibling()
			if nodeAfterFocus != nil {
				converted, ok := nodeAfterFocus.(*content.Block)
				if ok {
					blockAfterFocus = converted
				}
			}
		}

		// Remove refs marked for deletion
		block.Children().FindDeep(func(node content.Node) bool {
			if ref, ok := node.(*content.BlockRef); ok {
				if obsoleteRefs.Contains(ref.ID) {
					// Block ref's parents are: paragraph and block
					// TODO: handle cases when the block ref is nested under another block ref.
					//  This will remove the obsolete block and its children.
					//  Should I show a warning message to the user and prevent the block from being deleted?
					node.Parent().Parent().RemoveSelf()

					deletedCount++
				} else if dividerFocus == nil {
					// Keep adding tasks to the focus section until the divider is found
					focusRefs.Add(ref.ID)
				}
			}

			return false
		})
	}

	if blockAfterFocus != nil {
		insertPoint = blockAfterFocus
	} else {
		insertPoint = firstBlock
	}

	// Insert new tasks before the first one
	for _, ref := range newRefs.Values() {
		newTask := content.NewBlock(content.NewBlockRef(ref))
		if insertPoint == nil {
			page.AddBlock(newTask)
		} else {
			page.InsertBlockBefore(newTask, insertPoint)
		}
	}

	// Will only add a divider block if there are new tasks to add
	if dividerNewTasks == nil && newRefs.Size() > 0 {
		if insertPoint == nil {
			page.AddBlock(dividerNewTasksContent)
		} else {
			page.InsertBlockBefore(dividerNewTasksContent, insertPoint)
		}
	}

	err = transaction.Save()
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	if newRefs.Size() > 0 {
		color.Green("  %s", utils.FormatCount(newRefs.Size(), "new task", "new tasks"))
	}

	if deletedCount > 0 {
		color.Red("  %s removed (completed or unreferenced)", utils.FormatCount(deletedCount, "task was", "tasks were"))
	}

	if dividerFocus == nil {
		focusRefs.Clear()
	}

	return focusRefs, nil
}
