package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
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
		pageName := "backlog/" + pages[0]

		page, err := graph.OpenPage(pageName)
		if err != nil {
			return fmt.Errorf("failed to open page: %w", err)
		}

		existingTasks := refsFromPages(page)

		fmt.Printf("%s: %s from %s\n", pageName,
			FormatCount(existingTasks.Size(), "task", "tasks"),
			strings.Join(pages, ", "))

		queriedTasks, err := queryTasksFromPages(graph, pages)
		if err != nil {
			return err
		}

		newRefs := NewSet[string]()

		for _, ref := range queriedTasks.Values() {
			if !existingTasks.Contains(ref) {
				newRefs.Add(ref)
			}
		}

		if newRefs.Size() == 0 {
			fmt.Printf("\033[33m  no new tasks found\033[0m\n")

			continue
		}

		err = saveBacklog(graph, pageName, newRefs)
		if err != nil {
			return err
		}
	}

	return nil
}

func refsFromPages(page logseq.Page) *Set[string] {
	existingTasks := NewSet[string]()

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if ref, ok := n.(*content.BlockRef); ok {
				existingTasks.Add(ref.ID)
			}

			return false
		})
	}

	return existingTasks
}

func linesFromBacklog(graph *logseq.Graph) ([][]string, error) {
	page, err := graph.OpenPage(backlogName)
	if err != nil {
		return nil, fmt.Errorf("failed to open backlog page: %w", err)
	}

	var lines [][]string

	for _, block := range page.Blocks() {
		var pages []string

		block.Children().FindDeep(func(n content.Node) bool {
			if pageLink, ok := n.(*content.PageLink); ok {
				pages = append(pages, pageLink.To)
			} else if tag, ok := n.(*content.Hashtag); ok {
				pages = append(pages, tag.To)
			}

			return false
		})

		if len(pages) > 0 {
			lines = append(lines, pages)
		}
	}

	if len(lines) == 0 {
		fmt.Println("no pages found in the backlog")
	}

	return lines, nil
}

func queryTasksFromPages(graph *logseq.Graph, pages []string) (*Set[string], error) {
	refsFromAllQueries := NewSet[string]()

	for _, page := range pages {
		query, err := findFirstQuery(graph, page)
		if err != nil {
			return nil, err
		}

		if query == "" {
			query = defaultQuery(page)
			fmt.Printf("  default query: %s\n", query)
		} else {
			fmt.Printf("  found query: %s\n", query)
		}

		jsonStr, err := queryLogseqAPI(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query Logseq API: %w", err)
		}

		jsonTasks, err := extractTasks(jsonStr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tasks: %w", err)
		}

		for _, t := range jsonTasks {
			refsFromAllQueries.Add(t.UUID)
		}
	}

	return refsFromAllQueries, nil
}

func findFirstQuery(graph *logseq.Graph, pageName string) (string, error) {
	var query string

	page, err := graph.OpenPage(pageName)
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

	return replaceCurrentPage(query, pageName), nil
}

// replaceCurrentPage replaces the current page placeholder in the query with the actual page name.
// replaceCurrentPage replaces the current page placeholder in the query with the actual page name.
func replaceCurrentPage(query, pageName string) string {
	return strings.ReplaceAll(query, "<% current page %>", "[["+pageName+"]]")
}

func defaultQuery(page string) string {
	return fmt.Sprintf("(and [[%s]] (task TODO DOING WAITING))", page)
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

func saveBacklog(graph *logseq.Graph, pageName string, newRefs *Set[string]) error {
	transaction := graph.NewTransaction()

	page, err := transaction.OpenPage(pageName)
	if err != nil {
		return fmt.Errorf("failed to open page for transaction: %w", err)
	}

	blocks := page.Blocks()

	var first *content.Block
	if len(blocks) == 0 {
		first = nil
	} else {
		first = blocks[0]
	}

	for _, ref := range newRefs.Values() {
		block := content.NewBlock(content.NewBlockRef(ref))
		if first == nil {
			page.AddBlock(block)
		} else {
			page.InsertBlockBefore(block, first)
		}
	}

	divider := content.NewBlock(content.NewParagraph(
		content.NewPageLink("quick capture"),
		content.NewText(" New tasks above this line"),
	))

	if first == nil {
		page.AddBlock(divider)
	} else {
		page.InsertBlockBefore(divider, first)
	}

	err = transaction.Save()
	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	fmt.Printf("\033[92m  updated with %s\033[0m\n", FormatCount(newRefs.Size(), "new task", "new tasks"))

	return nil
}
