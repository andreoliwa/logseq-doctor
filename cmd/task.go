package cmd

import (
	"os"
	"time"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lqd/internal"
	"github.com/spf13/cobra"
)

// TaskAddDependencies holds all the dependencies for the task add command.
type TaskAddDependencies struct {
	AddTaskFn func(*internal.AddTaskOptions) error
	OpenGraph func(string) *logseq.Graph
	TimeNow   func() time.Time
}

// NewTaskCmd creates the parent task command.
func NewTaskCmd() *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "task",
		Short: "Manage tasks in Logseq",
		Long:  `Manage tasks in your Logseq graph. Use subcommands to add, list, or modify tasks.`,
	}

	return cmd
}

// NewTaskAddCmd creates a new task add subcommand with the specified dependencies.
// If deps is nil, it uses default implementations.
func NewTaskAddCmd(deps *TaskAddDependencies) *cobra.Command {
	if deps == nil {
		deps = &TaskAddDependencies{
			AddTaskFn: internal.AddTask,
			OpenGraph: internal.OpenGraphFromPath,
			TimeNow:   time.Now,
		}
	}

	var journalFlag, blockFlag, pageFlag, keyFlag string

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "add [task description]",
		Short: "Add a new task to Logseq",
		Long: `Add a new task to your Logseq graph.

The task will be added to the specified page or to today's journal by default.
If --key is provided, searches for an existing task containing that key (case-insensitive)
and updates it. Otherwise, creates a new task.

Examples:
  lqd task add "Review pull request"
  lqd task add "Call client" --page "Work"
  lqd task add "Buy groceries" --journal "2024-12-25"
  lqd task add "Water plants in living room" --key "water plants"
  lqd task add "Meeting notes" --block "Project A"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			graphPath := os.Getenv("LOGSEQ_GRAPH_PATH")
			graph := deps.OpenGraph(graphPath)

			targetDate, err := ParseDateFromJournalFlag(journalFlag, deps.TimeNow)
			if err != nil {
				return err
			}

			opts := &internal.AddTaskOptions{
				Graph:     graph,
				Date:      targetDate,
				Page:      pageFlag,
				BlockText: blockFlag,
				Key:       keyFlag,
				Name:      args[0],
			}

			return deps.AddTaskFn(opts)
		},
	}

	addJournalFlag(cmd, &journalFlag)
	addBlockFlag(cmd, &blockFlag, "task")
	addPageFlag(cmd, &pageFlag, "task")

	cmd.Flags().StringVarP(&keyFlag, "key", "k", "",
		"Unique key, will be used to update an existing task")

	return cmd
}

// taskCmd represents the task command using the default dependencies.
var taskCmd = NewTaskCmd() //nolint:gochecknoglobals

func init() {
	taskCmd.AddCommand(NewTaskAddCmd(nil))
	rootCmd.AddCommand(taskCmd)
}
