package cmd

import (
	"fmt"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/backlog"
	"github.com/spf13/cobra"
	"os"
)

var backlogCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "backlog [partial page names]",
	Short: "Aggregate tasks from multiple pages into a backlog",
	Long: `The backlog command aggregates tasks from one or more pages into a unified backlog.

If partial page names are provided, only page titles that contain the provided names are processed.

Each line on the "backlog" page that includes references to other pages or tags generates a separate backlog.
The first page in the line determines the name of the backlog page.
Tasks are retrieved from all provided pages or tags.
This setup enables users to rearrange tasks using the arrow keys and manage task states (start/stop)
directly within the interface.`,
	Run: func(_ *cobra.Command, args []string) {
		path := os.Getenv("LOGSEQ_GRAPH_PATH")
		api := internal.NewLogseqAPI(path,
			os.Getenv("LOGSEQ_HOST_URL"), os.Getenv("LOGSEQ_API_TOKEN"))
		graph := internal.OpenGraphFromPath(path)
		reader := backlog.NewPageConfigReader(graph, "backlog")
		proc := backlog.NewBacklog(graph, api, reader)

		err := proc.ProcessAll(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backlogCmd)
}
