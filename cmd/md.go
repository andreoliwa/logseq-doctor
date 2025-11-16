package cmd

import (
	"os"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-go"
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
  echo "New task" | lqd md
  echo "Child task" | lqd md --parent "Project A"
  echo "Another task" | lqd md --parent "meeting notes"`,
		RunE: func(_ *cobra.Command, _ []string) error {
			graphPath := os.Getenv("LOGSEQ_GRAPH_PATH")
			stdin := deps.ReadStdin()
			graph := deps.OpenGraph(graphPath)

			targetDate, err := ParseDateFromJournalFlag(journalFlag, deps.TimeNow)
			if err != nil {
				return err
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

	addJournalFlag(cmd, &journalFlag)
	addParentFlag(cmd, &parentFlag, "Markdown content")
	// TODO: addPageFlag(cmd, &pageFlag, "Markdown content")

	return cmd
}

// mdCmd represents the md command using the default dependencies.
var mdCmd = NewMdCmd(nil) //nolint:gochecknoglobals

func init() {
	rootCmd.AddCommand(mdCmd)
}
