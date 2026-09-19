package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/config"
	"github.com/andreoliwa/logseq-go"
	"github.com/spf13/cobra"
)

var errForbiddenContent = errors.New("tidy-up found forbidden content")

// TidyUpDependencies supplies dependencies for the tidy-up command.
type TidyUpDependencies struct {
	OpenGraphFromPath func(string) *logseq.Graph
	LoadPolicy        func() (config.ForbiddenContentPolicy, error)
	TidyUpOneFile     func(*logseq.Graph, string, config.ForbiddenContentPolicy) int
}

// NewTidyUpCmd creates the tidy-up command.
func NewTidyUpCmd(deps *TidyUpDependencies) *cobra.Command {
	if deps == nil {
		deps = &TidyUpDependencies{
			OpenGraphFromPath: api.OpenGraphFromPath,
			LoadPolicy: func() (config.ForbiddenContentPolicy, error) {
				path, err := config.TidyUpConfigPath()
				if err != nil {
					return config.ForbiddenContentPolicy{}, fmt.Errorf("get tidy-up config path: %w", err)
				}

				return config.LoadTidyUpPolicy(path)
			},
			TidyUpOneFile: internal.TidyUpOneFile,
		}
	}

	return &cobra.Command{ //nolint:exhaustruct_v5
		Use:           "tidy-up file1.md [file2.md ...]",
		Short:         "Tidy up your Markdown files.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `Tidy up your Markdown files, checking for invalid content and fixing some of them automatically.

- Check forbidden references and configured URL and text strings
- Check running tasks (DOING)
- Check double spaces`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			policy, err := deps.LoadPolicy()
			if err != nil {
				_, _ = fmt.Fprintln(command.ErrOrStderr(), err)

				return err
			}

			graph := deps.OpenGraphFromPath(os.Getenv("LOGSEQ_GRAPH_PATH"))
			for _, path := range args {
				if deps.TidyUpOneFile(graph, path, policy) != 0 {
					return errForbiddenContent
				}
			}

			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(NewTidyUpCmd(nil))
}
