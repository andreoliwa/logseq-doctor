package backlog

import (
	"fmt"
	"strings"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/pkg/utils"
	"github.com/fatih/color"
)

type Backlog interface {
	Graph() *logseq.Graph
	ProcessAll(partialNames []string) error
	ProcessOne(pageTitle string, funcQueryRefs func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error)
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

func (b *backlogImpl) ProcessAll(partialNames []string) error {
	config, err := b.configReader.ReadConfig()
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

		focusRefsFromPage, err := b.ProcessOne(backlogConfig.OutputPage,
			func() (*internal.CategorizedTasks, error) {
				return queryTasksFromPages(b.graph, b.api, backlogConfig.InputPages)
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

	_, err = b.ProcessOne(config.FocusPage, func() (*internal.CategorizedTasks, error) {
		return &allFocusTasks, nil
	})

	return err
}

func (b *backlogImpl) ProcessOne(pageTitle string,
	funcQueryRefs func() (*internal.CategorizedTasks, error)) (*utils.Set[string], error) {
	page := internal.OpenPage(b.graph, pageTitle)

	existingBlockRefs := blockRefsFromPages(page)

	fmt.Printf("%s: %s", internal.PageColor(pageTitle), utils.FormatCount(existingBlockRefs.Size(), "task", "tasks"))

	blockRefsFromQuery, err := funcQueryRefs()
	if err != nil {
		return nil, err
	}

	newBlockRefs := blockRefsFromQuery.All.Diff(existingBlockRefs)
	obsoleteBlockRefs := existingBlockRefs.Diff(blockRefsFromQuery.All)

	focusRefsFromPage, err := insertAndRemoveRefs(b.graph, pageTitle, newBlockRefs, obsoleteBlockRefs,
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

func queryTasksFromPages(graph *logseq.Graph, api internal.LogseqAPI,
	pageTitles []string) (*internal.CategorizedTasks, error) {
	tasks := internal.NewCategorizedTasks()
	finder := internal.NewLogseqFinder(graph)

	for _, pageTitle := range pageTitles {
		fmt.Printf(" %s: ", internal.PageColor(pageTitle))

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

		fmt.Print(utils.FormatCount(len(jsonTasks), "task", "tasks"))

		for _, t := range jsonTasks {
			if t.Overdue() {
				tasks.Overdue.Add(t.UUID)
			}

			tasks.All.Add(t.UUID)
		}
	}

	return &tasks, nil
}

func defaultQuery(pageTitle string) string {
	return fmt.Sprintf("(and [[%s]] (task TODO LATER DOING NOW WAITING))", pageTitle)
}

func insertAndRemoveRefs( //nolint:cyclop,funlen,gocognit
	graph *logseq.Graph, pageTitle string, newBlockRefs, obsoleteBlockRefs,
	overdueBlockRefs *utils.Set[string],
) (*utils.Set[string], error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var firstBlock, dividerNewTasks, dividerOverdue, dividerFocus *content.Block

	deletedCount := 0
	movedCount := 0
	unpinnedCount := 0
	focusBlockRefs := utils.NewSet[string]()
	pinnedBlockRefs := utils.NewSet[string]()

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
					focusBlockRefs.Add(blockRef.ID)
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
				content.NewPageLink("inbox"),
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
					content.NewPageLink("inbox"),
				))
				AddSibling(page, dividerNewTasks, firstBlock, dividerOverdue, dividerFocus)
			}

			dividerNewTasks.AddChild(content.NewBlock(content.NewBlockRef(blockRef)))
		}

		color.Green(" %s", utils.FormatCount(newBlockRefs.Size(), "new task", "new tasks"))

		save = true
	}

	if deletedCount > 0 {
		// Remove completed or unreferenced tasks
		color.Red(" %s removed", utils.FormatCount(deletedCount, "task was", "tasks were"))

		save = true
	}

	if movedCount > 0 {
		color.Magenta(" %s moved around", utils.FormatCount(movedCount, "task was", "tasks were"))

		save = true
	}

	if unpinnedCount > 0 {
		color.Cyan(" %s unpinned", utils.FormatCount(unpinnedCount, "task was", "tasks were"))

		save = true
	}

	if save {
		err = transaction.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save transaction: %w", err)
		}
	} else {
		color.Yellow(" no new/deleted/moved tasks")
	}

	if dividerFocus == nil {
		focusBlockRefs.Clear()
	}

	return focusBlockRefs, nil
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
