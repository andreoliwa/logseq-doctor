package cmd

import (
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/pkg"
	"log"
	"time"

	"github.com/spf13/cobra"
)

// contentCmd represents the content command.
var contentCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "content",
	Short: "Append raw Markdown content to Logseq",
	Long: `Append raw Markdown content to Logseq.

Pipe your content via stdin.
For now, it will be appended at the end of the current journal page.`,
	Run: func(_ *cobra.Command, _ []string) {
		graph := internal.OpenGraphFromDirOrEnv("")
		stdin := internal.ReadFromStdin()
		_, err := pkg.AppendRawMarkdownToJournal(graph, time.Now(), stdin)
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	// TODO: Future flags for this command could be --append (the default when not informed) and --prepend.
	rootCmd.AddCommand(contentCmd)
}
