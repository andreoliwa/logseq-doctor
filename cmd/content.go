package cmd

import (
	"log"
	"os"
	"time"

	"github.com/andreoliwa/lsd/internal"
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
		graph := internal.OpenGraphFromPath(os.Getenv("LOGSEQ_GRAPH_PATH"))
		stdin := internal.ReadFromStdin()

		var targetDate time.Time
		if journalFlag != "" {
			parsedDate, err := time.Parse("2006-01-02", journalFlag)
			if err != nil {
				log.Fatalln("Invalid journal date format. Use YYYY-MM-DD:", err)
			}
			targetDate = parsedDate
		} else {
			targetDate = time.Now()
		}

		_, err := internal.AppendRawMarkdownToJournal(graph, targetDate, stdin)
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var journalFlag string //nolint:gochecknoglobals

func init() {
	// TODO: Future flags for this command could be --append (the default when not informed) and --prepend.
	contentCmd.Flags().StringVarP(&journalFlag, "journal", "j", "", "Journal date in YYYY-MM-DD format (default: today)")
	rootCmd.AddCommand(contentCmd)
}
