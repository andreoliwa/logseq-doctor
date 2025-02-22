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

var backlogCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "backlog",
	Short: "Aggregate tasks from multiple pages into a backlog",
	Long: `The backlog command aggregates tasks from one or more pages into a unified backlog.

Each line on the "backlog" page that includes references to other pages or tags generates a separate backlog.
The first argument specifies the name of the backlog page, while tasks are retrieved from all provided pages or tags.
This setup enables users to rearrange tasks using the arrow keys and manage task states (start/stop)
directly within the interface.`,
	Run: func(_ *cobra.Command, _ []string) {
		err := updateBacklog()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backlogCmd)
}

func updateBacklog() error {
	graph := openGraph("")
	if graph == nil {
		return ErrFailedOpenGraph
	}

	lines, err := linesFromBacklog(graph)
	if err != nil {
		return err
	}

	for _, pages := range lines {
		pageTitle := "backlog/" + pages[0]

		page, err := graph.OpenPage(pageTitle)
		if err != nil {
			return fmt.Errorf("failed to open page: %w", err)
		}

		existingRefs := refsFromPages(page)

		fmt.Printf("%s: %s\n", PageColor(pageTitle), FormatCount(existingRefs.Size(), "task", "tasks"))

		queriedRefs, err := queryTasksFromPages(graph, pages, existingRefs)
		if err != nil {
			return err
		}

		newRefs := queriedRefs.Diff(existingRefs)
		obsoleteRefs := existingRefs.Diff(queriedRefs)

		if newRefs.Size() > 0 || obsoleteRefs.Size() > 0 {
			err = saveBacklog(graph, pageTitle, newRefs, obsoleteRefs)
			if err != nil {
				return err
			}
		} else {
			color.Yellow("  no new/deleted tasks found")
		}
	}

	return nil
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

func linesFromBacklog(graph *logseq.Graph) ([][]string, error) {
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

func queryTasksFromPages(graph *logseq.Graph, pageTitles []string, existingRefs *Set[string]) (*Set[string], error) {
	refsFromAllQueries := NewSet[string]()

	for _, pageTitle := range pageTitles {
		fmt.Printf("  %s: ", PageColor(pageTitle))

		query, err := findFirstQuery(graph, pageTitle)
		if err != nil {
			return nil, err
		}

		if query == "" {
			query = defaultQuery(pageTitle)
		}

		jsonStr, err := queryLogseqAPI(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query Logseq API: %w", err)
		}

		jsonTasks, err := extractTasks(jsonStr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tasks: %w", err)
		}

		fmt.Printf("    queried %s, ", FormatCount(len(jsonTasks), "task", "tasks"))

		newCount := 0

		for _, t := range jsonTasks {
			if !existingRefs.Contains(t.UUID) {
				newCount++
			}

			refsFromAllQueries.Add(t.UUID)
		}

		formatted := FormatCount(newCount, "new task", "new tasks")
		if newCount == 0 {
			fmt.Printf("found %s\n", formatted)
		} else {
			color.Green("found %s", formatted)
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

	fmt.Printf("found query %s\n", query)

	return replaceCurrentPage(query, pageTitle), nil
}

// replaceCurrentPage replaces the current page placeholder in the query with the actual page name.
func replaceCurrentPage(query, pageTitle string) string {
	return strings.ReplaceAll(query, "<% current page %>", "[["+pageTitle+"]]")
}

func defaultQuery(pageTitle string) string {
	query := fmt.Sprintf("(and [[%s]] (task TODO DOING WAITING))", pageTitle)
	fmt.Printf("default query %s\n", query)

	return query
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

func saveBacklog( //nolint:cyclop,funlen
	graph *logseq.Graph, pageTitle string, newRefs, obsoleteRefs *Set[string]) error {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageTitle)
	if err != nil {
		return fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var first *content.Block

	dividerBlock := content.NewBlock(content.NewParagraph(
		content.NewPageLink("quick capture"),
		content.NewText(" New tasks above this line"),
	))
	hasDivider := false
	deletedCount := 0

	for i, block := range page.Blocks() {
		if i == 0 {
			first = block
		}

		// Will only add a divider block if there are new tasks to add
		// TODO: add AsMarkdown() or ContentHash() or Hash() to content.Block, to make it possible to compare blocks
		//  Or fix the message "the operator == is not defined on NodeList"
		if block.GomegaString() == dividerBlock.GomegaString() {
			hasDivider = true
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
				}
			}

			return false
		})
	}

	for _, ref := range newRefs.Values() {
		newBlock := content.NewBlock(content.NewBlockRef(ref))
		if first == nil {
			page.AddBlock(newBlock)
		} else {
			page.InsertBlockBefore(newBlock, first)
		}
	}

	if !hasDivider && newRefs.Size() > 0 {
		if first == nil {
			page.AddBlock(dividerBlock)
		} else {
			page.InsertBlockBefore(dividerBlock, first)
		}
	}

	err = transaction.Save()
	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	if newRefs.Size() > 0 {
		color.Green("  updated with a total of %s", FormatCount(newRefs.Size(), "new task", "new tasks"))
	}

	if deletedCount > 0 {
		color.Red("  %s removed (completed or unreferenced)", FormatCount(deletedCount, "task was", "tasks were"))
	}

	return nil
}
