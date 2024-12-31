package cmd

import (
	"context"
	"github.com/andreoliwa/logseq-go"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "lsdg",
	Short: "Logseq Doctor (Go) heals your Markdown files for Logseq",
	Long: `Logseq Doctor (Go) heals your Markdown files for Logseq.

Convert flat Markdown to Logseq outline, clean up Markdown,
prevent invalid content, and more stuff to come.

"lsd" is the CLI tool originally written in Python; "lsdg" is the Go version.
The intention is to slowly convert everything to Go.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.logseq-doctor.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func openGraph(dir string) *logseq.Graph {
	if dir == "" {
		dir = os.Getenv("LOGSEQ_GRAPH_PATH")
		if dir == "" {
			log.Fatalln("LOGSEQ_GRAPH_PATH environment variable is not set.")
		}
	}

	ctx := context.Background()

	graph, err := logseq.Open(ctx, dir)
	if err != nil {
		log.Fatalln("error opening graph: %w", err)
	}

	return graph
}
