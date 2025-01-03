package cmd

import (
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
	// TODO: dynamically generate the long description based on the functions in the code.
	Long: `Tidy up your Markdown files, checking for invalid content and fixing some of them automatically.

- Check for forbidden references to pages/tags
- Check for running tasks (DOING)
- Check for double spaces`,
	// TODO: add help for the Markdown files accepted as arguments, "lsdg tidy-up [flags] file1.md [file2.md ...]".
	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		graph := openGraph("")

		exitCode := 0
		for _, path := range args {
			if !isValidMarkdownFile(path) {
				fmt.Printf("%s: skipping, not a Markdown file\n", path)
			} else {
				// Some fixes still need modifications directly on the file contents.
				// We will do them first, and apply each function on top of the previously modified contents.
				bytes, err := os.ReadFile(path)
				if err != nil {
					log.Fatalf("%s: error reading file contents: %s\n", path, err)
				}
				currentFileContents := string(bytes)

				fileInfo, err := os.Stat(path)
				if err != nil {
					log.Fatalf("%s: error getting file info: %s\n", path, err)
				}

				messages := make([]string, 0)

				write := false
				for _, f := range []func(string) changedContents{
					removeUnnecessaryBracketsFromTags,
				} {
					result := f(currentFileContents)
					if result.msg != "" {
						messages = append(messages, result.msg)
						// Pass the new contents to the next function.
						currentFileContents = result.newContents
						write = true
					}
				}
				if write {
					err := os.WriteFile(path, []byte(currentFileContents), fileInfo.Mode())
					if err != nil {
						log.Fatalf("%s: error writing file contents: %s\n", path, err)
					}
				}

				// Now we will apply the functions that modify the Markdown through a Page and a transaction.
				transaction := graph.NewTransaction()
				commit := false

				page, err := transaction.OpenViaPath(path)
				if err != nil {
					log.Fatalf("%s: error opening file via path: %s\n", path, err)
				}
				if page == nil {
					log.Fatalf("%s: error opening file via path: page is nil\n", path)
				}

				for _, f := range []func(logseq.Page) changedPage{
					checkForbiddenReferences, checkRunningTasks, removeDoubleSpaces, removeEmptyBullets,
				} {
					result := f(page)
					if result.msg != "" {
						messages = append(messages, result.msg)
					}
					if result.changed {
						commit = true
					}
				}

				if len(messages) > 0 {
					exitCode = 1
					for _, msg := range messages {
						fmt.Printf("%s: %s\n", path, msg)
					}
				}

				if commit {
					// Only one transaction per file, to avoid saving files that were not modified but were opened.
					// logseq-go rewrites Markdown, and it modified brackets in content without need.
					// I'll avoid using .Save() without need until these bugs are fixed.
					err = transaction.Save()
					if err != nil {
						log.Fatalf("error saving transaction: %s\n", err)
					}
				}
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

// changedContents is the result of a check function that modifies file contents directly without a transaction.
type changedContents struct {
	msg         string
	newContents string
}

// changedPage is the result of a check function that modifies Markdown through a Page and a transaction.
type changedPage struct {
	msg     string
	changed bool
}

// checkForbiddenReferences checks if a page has forbidden references to other pages or tags.
func checkForbiddenReferences(page logseq.Page) changedPage {
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

		return changedPage{fmt.Sprintf("remove %d forbidden references to pages/tags: %s",
			count, strings.Join(unique, ", ")), false}
	}

	return changedPage{"", false}
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
func checkRunningTasks(page logseq.Page) changedPage {
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

		return changedPage{fmt.Sprintf("stop %d running task(s): %s", count, strings.Join(unique, ", ")), false}
	}

	return changedPage{"", false}
}

func removeDoubleSpaces(page logseq.Page) changedPage {
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

		return changedPage{fmt.Sprintf("%d double spaces fixed: %s", count, strings.Join(unique, ", ")), fixed}
	}

	return changedPage{"", fixed}
}

// removeUnnecessaryBracketsFromTags removes unnecessary brackets from hashtags.
// logseq-go rewrites tags correctly when saving the transaction, removing unnecessary brackets.
// But, when reading the file, the AST doesn't provide the information if a tag has brackets or not.
// So I would have to rewrite the file to fix them, and I don't want to do it every time there is a tag without spaces.
// Also, as of 2024-12-30, logseq-go has a bug when reading properties with spaces in values,
// which causes them to be partially removed from the file, destroying data. I will report it soon.
func removeUnnecessaryBracketsFromTags(oldContents string) changedContents {
	re := regexp.MustCompile(`#\[\[([^ ]*?)]]`)

	newContents := re.ReplaceAllString(oldContents, "#$1")
	if newContents != oldContents {
		return changedContents{"unnecessary tag brackets removed", newContents}
	}

	return changedContents{"", ""}
}

func removeEmptyBullets(page logseq.Page) changedPage {
	removed := 0
	// TODO: add methods Find* and Filter* to logseq.Page in logseq-go and replace this for loop with page.FindDeep()
	//  - Create an interface for it with all 4 methods Find(), FindDeep(), Filter(), FilterDeep()?
	//  - Reuse this interface in existing methods in logseq-go?
	//    They have different signatures: node content.Node, block *content.Block
	//  - Use generics for this interface or create 2 interfaces?
	for _, block := range page.Blocks() {
		if block.FirstChild() == nil {
			block.RemoveSelf()

			removed++
		} else {
			block.Children().FindDeep(func(node content.Node) bool {
				if nestedBlock, ok := node.(*content.Block); ok {
					if nestedBlock.FirstChild() == nil {
						nestedBlock.RemoveSelf()

						removed++
					}
				}

				return false
			})
		}
	}

	if removed > 0 {
		return changedPage{fmt.Sprintf("%d empty bullets removed", removed), true}
	}

	return changedPage{"", false}
}