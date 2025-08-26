package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"github.com/spf13/cobra"
)

// MdDependencies holds all the dependencies for the md command.
type MdDependencies struct {
	InsertFn  func(*internal.InsertMarkdownOptions) error
	OpenGraph func(string) *logseq.Graph
	ReadStdin func() string
	TimeNow   func() time.Time
}

// NewMdCmd creates a new md command with the specified dependencies.
// If deps is nil, it uses default implementations.
func NewMdCmd(deps *MdDependencies) *cobra.Command {
	if deps == nil {
		deps = &MdDependencies{
			InsertFn:  internal.InsertMarkdownToJournal,
			OpenGraph: internal.OpenGraphFromPath,
			ReadStdin: internal.ReadFromStdin,
			TimeNow:   time.Now,
		}
	}

	var journalFlag, parentFlag string

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "md",
		Short: "Add Markdown content to Logseq using the DOM",
		Long: `Add Markdown content to Logseq using the DOM.

Pipe your Markdown content via stdin.
If --parent is provided, the content will be added as a child block under the first block
containing the specified text. Otherwise, it will be appended at the end of the journal page.

Examples:
  echo "New task" | lsd md
  echo "Child task" | lsd md --parent "Project A"
  echo "Another task" | lsd md -p "meeting notes"`,
		RunE: func(_ *cobra.Command, _ []string) error {
			graphPath := os.Getenv("LOGSEQ_GRAPH_PATH")
			stdin := deps.ReadStdin()
			graph := deps.OpenGraph(graphPath)

			var targetDate time.Time
			if journalFlag != "" {
				parsedDate, err := time.Parse("2006-01-02", journalFlag)
				if err != nil {
					return fmt.Errorf("invalid journal date format. Use YYYY-MM-DD: %w", err)
				}
				targetDate = parsedDate
			} else {
				targetDate = deps.TimeNow()
			}

			opts := &internal.InsertMarkdownOptions{
				Graph:      graph,
				Date:       targetDate,
				Content:    stdin,
				ParentText: parentFlag,
			}

			return deps.InsertFn(opts)
		},
	}

	cmd.Flags().StringVarP(&journalFlag, "journal", "j", "", "Journal date in YYYY-MM-DD format (default: today)")
	cmd.Flags().StringVarP(&parentFlag, "parent", "p", "",
		"Partial text of a block that will be the parent of the added Markdown")

	return cmd
}

// mdCmd represents the md command using the default dependencies.
var mdCmd = NewMdCmd(nil) //nolint:gochecknoglobals

func init() {
	rootCmd.AddCommand(mdCmd)
}
