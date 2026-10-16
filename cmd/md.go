package cmd

import (
	"os"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	api "github.com/andreoliwa/logseq-doctor/internal/api"
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

const mdHelp = `Add Markdown content to Logseq using the DOM.

Pipe Markdown via stdin to a named page or today's journal (or --journal DATE).

Explicit root bullets are placed independently without a wrapper. Unbulleted input
stays one block, with its supporting paragraphs and children attached.

Timed blocks go before the first strictly later sibling; equal times retain arrival
order, untimed blocks append, and existing siblings are not sorted. Time spelling,
formatting and supplied task statuses are retained, without generating TODO markers.
A ref-only sibling uses its first timed direct child as its time, not later descendants.

With --parent, compare only the matched parent's direct children; if no parent matches,
fall back to page roots.

With --key, update one matching block (case-insensitive).
Timed key updates move that same block among its actual siblings, retaining its parent,
properties, children and logbook; untimed updates stay in place. Key updates keep the
single-block content-update contract rather than importing a root batch.

This command uses structured Markdown serialization, not raw byte preservation.

Examples:
  echo "New task" | lqd md
  echo "9h00 Meeting notes" | lqd md --page "Work"
  printf -- '- 09:00 Call\n- 11:00 Review\n' | lqd md --journal 2026-10-16
  echo "Child task" | lqd md --parent "Project A"
  echo "Another task" | lqd md --parent "meeting notes"
  echo "10:00 Updated content" | lqd md --key "unique identifier"
  echo "Update work item" | lqd md --page "Projects" --key "feature-123"`

// NewMdCmd creates a new md command with the specified dependencies.
// If deps is nil, it uses default implementations.
func NewMdCmd(deps *MdDependencies) *cobra.Command {
	if deps == nil {
		deps = &MdDependencies{
			InsertFn:  internal.InsertMarkdown,
			OpenGraph: api.OpenGraphFromPath,
			ReadStdin: internal.ReadFromStdin,
			TimeNow:   time.Now,
		}
	}

	var journalFlag, parentFlag, pageFlag, keyFlag string

	cmd := &cobra.Command{ //nolint:exhaustruct_v5
		Use:   "md",
		Short: "Add Markdown content to Logseq using the DOM",
		Long:  mdHelp,
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
				Page:       pageFlag,
				Content:    stdin,
				ParentText: parentFlag,
				Key:        keyFlag,
			}

			return deps.InsertFn(opts)
		},
	}

	addJournalFlag(cmd, &journalFlag)
	addParentFlag(cmd, &parentFlag, "Markdown content")
	addPageFlag(cmd, &pageFlag, "Markdown content")
	addKeyFlag(cmd, &keyFlag)

	return cmd
}

// mdCmd represents the md command using the default dependencies.
var mdCmd = NewMdCmd(nil) //nolint:gochecknoglobals

func init() {
	rootCmd.AddCommand(mdCmd)
}
