package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/api"
	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/spf13/cobra"
)

// ContentDependencies isolates journal I/O from command flag handling.
type ContentDependencies struct {
	AppendFn  func(*logseq.Graph, time.Time, string) (int, error)
	OpenGraph func(string) *logseq.Graph
	// ReadStdin optionally overrides raw command input for callers and tests.
	ReadStdin func() string
	TimeNow   func() time.Time
}

// NewContentCmd constructs the raw Markdown insertion command.
func NewContentCmd(deps *ContentDependencies) *cobra.Command {
	if deps == nil {
		deps = &ContentDependencies{
			AppendFn:  internal.AppendRawMarkdownToJournal,
			OpenGraph: api.OpenGraphFromPath,
			ReadStdin: nil, // Read raw bytes from Cobra input instead of the line-oriented shared reader.
			TimeNow:   time.Now,
		}
	}

	var journal string

	command := &cobra.Command{ //nolint:exhaustruct_v5
		Use:   "content",
		Short: "Insert raw Markdown content into a Logseq journal",
		Long: `Insert raw Markdown content into a Logseq journal from stdin.

Explicit root bullets are placed independently before the first later timed
sibling. Equal times retain arrival order; untimed roots append. Children and
supporting lines stay with their root. Ref-only anchors use their first timed
direct child and remain one unit. Existing and incoming source bytes are retained,
adding only needed insertion separators. Times and statuses are not normalized.
Unoutlined text keeps raw append behavior. Empty input is a no-op.
Unsafe joins that would change root ownership fail without changing the journal.

Examples:
  printf '%s\n' '- 09:00 Call' '- 11:00 Review' | lqd content --journal 2026-10-16
  echo 'Raw note' | lqd content`,
		RunE: func(command *cobra.Command, _ []string) error {
			date, err := ParseDateFromJournalFlag(journal, deps.TimeNow)
			if err != nil {
				return err
			}

			raw, err := readContentInput(command, deps.ReadStdin)
			if err != nil {
				return err
			}

			graph := deps.OpenGraph(os.Getenv("LOGSEQ_GRAPH_PATH"))
			_, err = deps.AppendFn(graph, date, raw)

			return err
		},
	}
	addJournalFlag(command, &journal)

	return command
}

func readContentInput(command *cobra.Command, readStdin func() string) (string, error) {
	if readStdin != nil {
		return readStdin(), nil
	}

	raw, err := io.ReadAll(command.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read raw Markdown from stdin: %w", err)
	}

	return string(raw), nil
}

var contentCmd = NewContentCmd(nil) //nolint:gochecknoglobals

func init() { rootCmd.AddCommand(contentCmd) }
