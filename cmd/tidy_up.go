package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/spf13/cobra"
)

// tidyUpCmd represents the tidyUp command.
var tidyUpCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "tidy-up",
	Short: "Tidy up your Markdown files.",
	Long: `Tidy up your Markdown files, checking for invalid content and fixing some of them automatically.

- Check for forbidden references to pages/tags
- Check for running tasks (DOING)
- Check for double spaces`,
	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		dir := os.Getenv("LOGSEQ_GRAPH_PATH")
		if dir == "" {
			log.Fatalln("LOGSEQ_GRAPH_PATH environment variable is not set.")
		}

		ctx := context.Background()
		graph, err := logseq.Open(ctx, dir)
		if err != nil {
			log.Fatalln("error opening graph: %w", err)
		}

		transaction := graph.NewTransaction()
		shouldSave := false

		exitCode := 0
		for _, path := range args {
			if !isValidMarkdownFile(path) {
				fmt.Printf("%s: skipping, not a Markdown file\n", path)
			} else {
				page, err := transaction.OpenViaPath(path)
				if err != nil {
					log.Fatalf("%s: error opening file via path: %s\n", path, err)
				}
				if page == nil {
					log.Fatalf("%s: error opening file via path: page is nil\n", path)
				}

				messages := make([]string, 0)

				functions := []func(logseq.Page) tidyInfo{
					checkForbiddenReferences, checkRunningTasks, removeDoubleSpaces}
				for _, f := range functions {
					info := f(page)
					if info.msg != "" {
						messages = append(messages, info.msg)
					}
					if info.changed {
						shouldSave = true
					}
				}

				if len(messages) > 0 {
					exitCode = 1
					for _, msg := range messages {
						fmt.Printf("%s: %s\n", path, msg)
					}
				}
			}
		}
		if shouldSave {
			err = transaction.Save()
			if err != nil {
				log.Fatalf("error saving transaction: %s\n", err)
			}
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(tidyUpCmd)
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

type tidyInfo struct {
	msg     string
	changed bool
}

// checkForbiddenReferences checks if a page has forbidden references to other pages or tags.
func checkForbiddenReferences(page logseq.Page) tidyInfo {
	all := make([]string, 0)

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			var reference string
			if pageLink, ok := n.(*content.PageLink); ok {
				reference = pageLink.To
			} else if tag, ok := n.(*content.Hashtag); ok {
				reference = tag.To
			}

			// TODO: these values should be read from a config file or env var
			forbidden := false

			switch strings.ToLower(reference) {
			case "quick capture":
				forbidden = true
			case "inbox":
				forbidden = true
			}

			if forbidden {
				all = append(all, reference)
			}

			return false
		})
	}

	if count := len(all); count > 0 {
		unique := sortAndRemoveDuplicates(all)

		return tidyInfo{fmt.Sprintf("remove %d forbidden references to pages/tags: %s",
			count, strings.Join(unique, ", ")), false}
	}

	return tidyInfo{"", false}
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
func checkRunningTasks(page logseq.Page) tidyInfo {
	all := make([]string, 0)

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(n content.Node) bool {
			if task, ok := n.(*content.TaskMarker); ok {
				status := task.Status
				// TODO: convert to strings "DOING"/"IN-PROGRESS" in logseq-go
				if status == content.TaskStatusDoing {
					all = append(all, "DOING")
				}

				if status == content.TaskStatusInProgress {
					all = append(all, "IN-PROGRESS")
				}
			}

			return false
		})
	}

	if count := len(all); count > 0 {
		unique := sortAndRemoveDuplicates(all)

		return tidyInfo{fmt.Sprintf("stop %d running task(s): %s", count, strings.Join(unique, ", ")), false}
	}

	return tidyInfo{"", false}
}

func removeDoubleSpaces(page logseq.Page) tidyInfo {
	all := make([]string, 0)
	doubleSpaceRegex := regexp.MustCompile(`\s{2,}`)
	fixed := false

	for _, block := range page.Blocks() {
		block.Children().FindDeep(func(node content.Node) bool {
			var oldValue string

			if text, ok := node.(*content.Text); ok {
				oldValue = text.Value
			} else if pageLink, ok := node.(*content.PageLink); ok {
				oldValue = pageLink.To
			} else if tag, ok := node.(*content.Hashtag); ok {
				oldValue = tag.To
			}

			if strings.Contains(oldValue, "  ") {
				all = append(all, fmt.Sprintf("'%s'", oldValue))
				newValue := doubleSpaceRegex.ReplaceAllString(oldValue, " ")
				fixed = true

				if text, ok := node.(*content.Text); ok {
					text.Value = newValue
				} else if pageLink, ok := node.(*content.PageLink); ok {
					pageLink.To = newValue
				} else if tag, ok := node.(*content.Hashtag); ok {
					tag.To = newValue
				}
			}

			return false
		})
	}

	if count := len(all); count > 0 {
		unique := sortAndRemoveDuplicates(all)

		return tidyInfo{fmt.Sprintf("fixed %d double spaces: %s", count, strings.Join(unique, ", ")), fixed}
	}

	return tidyInfo{"", fixed}
}
