package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// tidyUpCmd represents the tidyUp command
var tidyUpCmd = &cobra.Command{
	Use:   "tidy-up",
	Short: "Tidy up your Markdown files.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		changes := make([]string, 0)
		for _, f := range args {
			if !isValidMarkdownFile(f) {
				fmt.Println(f + ": skipping, not a Markdown file")
			} else {
				changes = append(changes, "fake Golang success message")
				fmt.Println(f + ": " + strings.Join(changes, ", "))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(tidyUpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tidyUpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tidyUpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
