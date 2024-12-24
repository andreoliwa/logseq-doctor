package cmd

import (
	"context"
	"fmt"
	"github.com/aholstenson/logseq-go"
	"github.com/aholstenson/logseq-go/content"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sort"
	"strings"
)

// tidyUpCmd represents the tidyUp command
var tidyUpCmd = &cobra.Command{
	Use:   "tidy-up",
	Short: "Tidy up your Markdown files.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := os.Getenv("LOGSEQ_GRAPH_PATH")
		if dir == "" {
			log.Fatalln("LOGSEQ_GRAPH_PATH environment variable is not set.")
		}

		ctx := context.Background()
		graph, err := logseq.Open(ctx, dir)
		if err != nil {
			log.Fatalln("error opening graph: %w", err)
			return
		}

		changes := make([]string, 0)
		var what string
		exitCode := 0
		for _, path := range args {
			if !isValidMarkdownFile(path) {
				fmt.Printf("%s: skipping, not a Markdown file\n", path)
			} else {
				page, err := graph.OpenViaPath(path)
				if err != nil {
					log.Fatalf("%s: error opening file via path: %s\n", path, err)
					return
				}
				if page == nil {
					log.Fatalf("%s: error opening file via path: page is nil\n", path)
					return
				}

				what = checkForbiddenReferences(page)
				if what != "" {
					changes = append(changes, what)
				}
				what = checkRunningTasks(page)
				if what != "" {
					changes = append(changes, what)
				}

				if len(changes) > 0 {
					exitCode = 1
					fmt.Printf("%s: %s\n", path, strings.Join(changes, ", "))
				}
			}
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(tidyUpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tidyUpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tidyUpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// isValidMarkdownFile checks if a file is a Markdown file, by looking at its extension, not its content.
func isValidMarkdownFile(filePath string) bool {
	if filePath == "" {
		return false
	}

	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return false
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// checkForbiddenReferences checks if a page has forbidden references to other pages or tags.
func checkForbiddenReferences(page logseq.Page) string {
	all := make([]string, 0)
	for _, block := range page.Blocks() {
		block.Children().FilterDeep(func(n content.Node) bool {
			var to string
			if pageLink, ok := n.(*content.PageLink); ok {
				to = pageLink.To
			} else if tag, ok := n.(*content.Hashtag); ok {
				to = tag.To
			}

			// TODO: these values should be read from a config file or env var
			forbidden := false
			switch strings.ToLower(to) {
			case "quick capture":
				forbidden = true
			case "inbox":
				forbidden = true
			}
			if forbidden {
				all = append(all, to)
			}
			return false
		})
	}
	if len(all) > 0 {
		unique := sortAndRemoveDuplicates(all)
		return fmt.Sprintf("remove these forbidden references to pages/tags: %s", strings.Join(unique, ", "))
	}
	return ""
}

func sortAndRemoveDuplicates(elements []string) []string {
	seen := make(map[string]bool)
	uniqueElements := make([]string, 0)

	for _, element := range elements {
		if !seen[element] {
			seen[element] = true
			uniqueElements = append(uniqueElements, element)
		}
	}
	sort.Strings(uniqueElements)

	return uniqueElements
}

// checkRunningTasks checks if a page has running tasks (DOING, etc.).
func checkRunningTasks(page logseq.Page) string {
	running := false
	for _, block := range page.Blocks() {
		block.Children().FilterDeep(func(n content.Node) bool {
			if task, ok := n.(*content.TaskMarker); ok {
				s := task.Status
				if s == content.TaskStatusDoing || s == content.TaskStatusInProgress {
					running = true
				}
			}
			return false
		})
	}
	if running {
		return "stop the running tasks (DOING/IN-PROGRESS)"
	}
	return ""
}
