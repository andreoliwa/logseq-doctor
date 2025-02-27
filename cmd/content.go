package cmd

import (
	"bufio"
	"github.com/andreoliwa/lsd/pkg"
	"log"
	"os"
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
		scanner := bufio.NewScanner(os.Stdin)
		var stdin string
		for scanner.Scan() {
			stdin += scanner.Text() + "\n"
		}
		if err := scanner.Err(); err != nil {
			log.Fatalln("Error reading input:", err)
		}
		pkg.AppendRawMarkdownToJournal("", time.Now(), stdin)
	},
}

func init() {
	// TODO: Future flags for this command could be --append (the default when not informed) and --prepend.
	rootCmd.AddCommand(contentCmd)
}
