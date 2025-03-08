package cmd

import (
	"github.com/andreoliwa/lsd/internal"
	"github.com/spf13/cobra"
	"os"
)

// tidyUpCmd represents the tidyUp command.
var tidyUpCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "tidy-up file1.md [file2.md ...]",
	Short: "Tidy up your Markdown files.",
	// TODO: dynamically generate the long description based on the functions in the code.
	Long: `Tidy up your Markdown files, checking for invalid content and fixing some of them automatically.

- Check for forbidden references to pages/tags
- Check for running tasks (DOING)
- Check for double spaces`,
	Args: cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		graph := internal.OpenGraphFromPath(os.Getenv("LOGSEQ_GRAPH_PATH"))

		exitCode := 0
		for _, path := range args {
			if internal.TidyUpOneFile(graph, path) != 0 {
				exitCode = 1
			}
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(tidyUpCmd)
}
