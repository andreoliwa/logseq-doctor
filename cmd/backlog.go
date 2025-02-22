package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"strings"
)

var backlogCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "backlog [backlogPage] [queryPages...]",
	Short: "Aggregate tasks from multiple pages into a backlog",
	Long: `The backlog command aggregates tasks from one or more pages into a single interactive backlog.
The first argument defines the name of the backlog page, while tasks are queried from all provided pages.
This backlog allows users to rearrange tasks with arrow keys and manage task states (start/stop)
directly within the interface.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		err := updateBacklog(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backlogCmd)
}

// TODO: refactor this function after all main features are implemented and tested.
func updateBacklog(pages []string) error { //nolint:cyclop,funlen
	graph := openGraph("")
	if graph == nil {
		return ErrFailedOpenGraph
	}

	pageName := "backlog/" + pages[0]

	page, err := graph.OpenPage(pageName)
	if err != nil {
		return fmt.Errorf("failed to open page: %w", err)
	}

	if page == nil {
		return fmt.Errorf("page %s: %w", pageName, ErrPageNotFound)
	}

	refsFromPage := NewSet[string]()

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if ref, ok := n.(*content.BlockRef); ok {
				refsFromPage.Add(ref.ID)
			}

			return false
		})
	}

	query := buildQuery(pages)

	jsonStr, err := queryJSON(query)
	if err != nil {
		return fmt.Errorf("failed to query Logseq API: %w", err)
	}

	jsonTasks, err := extractTasks(jsonStr)
	if err != nil {
		return fmt.Errorf("failed to extract tasks: %w", err)
	}

	refsFromQuery := NewSet[string]()
	for _, e := range jsonTasks {
		refsFromQuery.Add(e.UUID)
	}

	newRefs := NewSet[string]()

	for _, ref := range refsFromQuery.Values() {
		if !refsFromPage.Contains(ref) {
			newRefs.Add(ref)
		}
	}

	if newRefs.Size() == 0 {
		fmt.Printf("\033[33m%s: no new tasks found\033[0m\n", page.Title())

		return nil
	}

	transaction := graph.NewTransaction()

	page, err = transaction.OpenPage(pageName)
	if err != nil {
		return fmt.Errorf("failed to open page for transaction: %w", err)
	}

	var first *content.Block
	if len(page.Blocks()) == 0 {
		first = nil
	} else {
		first = page.Blocks()[0]
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

	fmt.Printf("\033[92m%s: updated with %d new task(s)\033[0m\n", page.Title(), newRefs.Size())

	return nil
}

func buildQuery(tagsOrPages []string) string {
	if len(tagsOrPages) == 0 {
		return ""
	}

	var condition string
	if len(tagsOrPages) == 1 {
		condition = fmt.Sprintf("[[%s]]", tagsOrPages[0])
	} else {
		withBrackets := make([]string, len(tagsOrPages))
		for i, page := range tagsOrPages {
			withBrackets[i] = fmt.Sprintf("[[%s]]", page)
		}

		pages := strings.Join(withBrackets, " ")
		condition = "(or " + pages + ")"
	}

	query := "(and " + condition + " (task TODO DOING WAITING))"

	return query
}

// queryJSON sends a query to the Logseq API and returns the result as JSON.
func queryJSON(query string) (string, error) {
	apiToken := os.Getenv("LOGSEQ_API_TOKEN")

	hostURL := os.Getenv("LOGSEQ_HOST_URL")
	if apiToken == "" || hostURL == "" {
		return "", ErrMissingConfig
	}

	client := &http.Client{} //nolint:exhaustruct
	payload := strings.NewReader(fmt.Sprintf(`{"method":"logseq.db.q","args":["%s"]}`, query))

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostURL+"/api", payload)
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
		return "", fmt.Errorf("status %s: %w", resp.Status, ErrQueryLogseqAPI)
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
