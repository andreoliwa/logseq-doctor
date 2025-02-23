package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
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

var backlogCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "backlog",
	Short: "Aggregate tasks from multiple pages into a backlog",
	Long: `The backlog command aggregates tasks from one or more pages into a unified backlog.

Each line on the "backlog" page that includes references to other pages or tags generates a separate backlog.
The first argument specifies the name of the backlog page, while tasks are retrieved from all provided pages or tags.
This setup enables users to rearrange tasks using the arrow keys and manage task states (start/stop)
directly within the interface.`,
	Run: func(_ *cobra.Command, _ []string) {
		err := processAllBacklogs()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backlogCmd)
}

func processAllBacklogs() error {
	graph := openGraph("")
	if graph == nil {
		return ErrFailedOpenGraph
	}

	lines, err := linesWithPages(graph)
	if err != nil {
		return err
	}

	allFocusRefs := NewSet[string]()

	for _, pages := range lines {
		focusRefsFromPage, err := processSingleBacklog(graph, "backlog/"+pages[0], func() (*Set[string], error) {
			return queryTasksFromPages(graph, pages)
		})
		if err != nil {
			return err
		}

		allFocusRefs.Update(focusRefsFromPage)
	}

	_, err = processSingleBacklog(graph, "backlog/Focus", func() (*Set[string], error) {
		return allFocusRefs, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func processSingleBacklog(graph *logseq.Graph, pageTitle string,
	queryRefs func() (*Set[string], error)) (*Set[string], error) {
	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page: %w", err)
	}

	existingRefs := refsFromPages(page)

	fmt.Printf("%s: %s\n", PageColor(pageTitle), FormatCount(existingRefs.Size(), "task", "tasks"))

	refsToInsert, err := queryRefs()
	if err != nil {
		return nil, err
	}

	newRefs := refsToInsert.Diff(existingRefs)
	obsoleteRefs := existingRefs.Diff(refsToInsert)

	if newRefs.Size() <= 0 && obsoleteRefs.Size() <= 0 {
		color.Yellow("  no new/deleted tasks found")
	}

	focusRefsFromPage, err := insertAndRemoveRefs(graph, pageTitle, newRefs, obsoleteRefs)
	if err != nil {
		return nil, err
	}

	return focusRefsFromPage, nil
}

func refsFromPages(page logseq.Page) *Set[string] {
	existingRefs := NewSet[string]()

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

func queryTasksFromPages(graph *logseq.Graph, pageTitles []string) (*Set[string], error) {
	refsFromAllQueries := NewSet[string]()

	for _, pageTitle := range pageTitles {
		fmt.Printf("  %s: ", PageColor(pageTitle))

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

		jsonStr, err := queryLogseqAPI(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query Logseq API: %w", err)
		}

		jsonTasks, err := extractTasks(jsonStr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tasks: %w", err)
		}

		fmt.Printf(", queried %s\n", FormatCount(len(jsonTasks), "task", "tasks"))

		for _, t := range jsonTasks {
			refsFromAllQueries.Add(t.UUID)
		}
	}

	return refsFromAllQueries, nil
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
	return fmt.Sprintf("(and [[%s]] (task TODO DOING WAITING))", pageTitle)
}

// queryLogseqAPI sends a query to the Logseq API and returns the result as JSON.
func queryLogseqAPI(query string) (string, error) {
	apiToken := os.Getenv("LOGSEQ_API_TOKEN")

	hostURL := os.Getenv("LOGSEQ_HOST_URL")
	if apiToken == "" || hostURL == "" {
		return "", ErrMissingConfig
	}

	client := &http.Client{} //nolint:exhaustruct

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	ctx := context.Background()
	payload := fmt.Sprintf(`{"method":"logseq.db.q","args":[%s]}`, string(jsonQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostURL+"/api", strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error performing HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %s with payload:\n%s: %w", resp.Status, payload, ErrInvalidResponseStatus)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

type taskJSON struct {
	UUID    string   `json:"uuid"`
	Marker  string   `json:"marker"`
	Content string   `json:"content"`
	Page    pageJSON `json:"page"`
}

type pageJSON struct {
	JournalDay int `json:"journalDay"`
}

func extractTasks(jsonStr string) ([]taskJSON, error) {
	var tasks []taskJSON
	if err := json.Unmarshal([]byte(jsonStr), &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tasks, nil
}

func insertAndRemoveRefs( //nolint:cyclop,funlen,gocognit
	graph *logseq.Graph, pageTitle string, newRefs, obsoleteRefs *Set[string]) (*Set[string], error) {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var firstBlock, dividerNewTasks, dividerFocus, blockAfterFocus, insertPoint *content.Block

	deletedCount := 0
	focusRefs := NewSet[string]()

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
		color.Green("  %s", FormatCount(newRefs.Size(), "new task", "new tasks"))
	}

	if deletedCount > 0 {
		color.Red("  %s removed (completed or unreferenced)", FormatCount(deletedCount, "task was", "tasks were"))
	}

	if dividerFocus == nil {
		focusRefs.Clear()
	}

	return focusRefs, nil
}
