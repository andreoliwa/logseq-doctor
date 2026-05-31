package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-go"
	"github.com/spf13/cobra"
)

// firstLineCount is the split limit used to extract the first line of block content.
const firstLineCount = 2

// TaskAddDependencies holds all the dependencies for the task add command.
type TaskAddDependencies struct {
	AddTaskFn func(*logseqext.AddTaskOptions) error
	OpenGraph func(string) *logseq.Graph
	TimeNow   func() time.Time
}

// TaskLsDependencies holds all the dependencies for the task ls command.
type TaskLsDependencies struct {
	NewAPI    func() api.LogseqAPI
	GraphName func() string
	Out       io.Writer
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
			AddTaskFn: logseqext.AddTask,
			OpenGraph: api.OpenGraphFromPath,
			TimeNow:   time.Now,
		}
	}

	var journalFlag, parentFlag, pageFlag, keyFlag string

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
  lqd task add "Meeting notes" --parent "Project A"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			graphPath := os.Getenv("LOGSEQ_GRAPH_PATH")
			graph := deps.OpenGraph(graphPath)

			targetDate, err := ParseDateFromJournalFlag(journalFlag, deps.TimeNow)
			if err != nil {
				return err
			}

			opts := &logseqext.AddTaskOptions{
				Graph:     graph,
				Date:      targetDate,
				Page:      pageFlag,
				BlockText: parentFlag,
				Key:       keyFlag,
				Name:      args[0],
				TimeNow:   deps.TimeNow,
			}

			return deps.AddTaskFn(opts)
		},
	}

	addJournalFlag(cmd, &journalFlag)
	addParentFlag(cmd, &parentFlag, "task")
	addPageFlag(cmd, &pageFlag, "task")

	cmd.Flags().StringVarP(&keyFlag, "key", "k", "",
		"Unique key, will be used to update an existing task")

	return cmd
}

// taskLsFlags holds the flag values for NewTaskLsCmd.
type taskLsFlags struct {
	canceled  bool
	done      bool
	completed bool
	json      bool
	verbose   bool
}

func runTaskLs(deps *TaskLsDependencies, flags *taskLsFlags, args []string) error {
	out := deps.Out

	if flags.completed {
		flags.canceled = true
		flags.done = true
	}

	query := api.BuildTaskListQuery(args, flags.canceled, flags.done)

	if flags.verbose {
		fmt.Fprintf(out, "Query: %s\n", query)
	}

	client := deps.NewAPI()

	jsonStr, err := client.PostQuery(query)
	if err != nil {
		return fmt.Errorf("failed to query Logseq API: %w", err)
	}

	if flags.json {
		fmt.Fprintln(out, jsonStr)

		return nil
	}

	tasks, err := api.ExtractTasksFromJSON(jsonStr)
	if err != nil {
		return fmt.Errorf("failed to extract tasks: %w", err)
	}

	api.SortTasksByDate(tasks)

	graphName := deps.GraphName()

	green := color.New(color.FgGreen)
	blueBold := color.New(color.FgBlue, color.Bold)

	for _, task := range tasks {
		firstLine := strings.SplitN(task.Content, "\n", firstLineCount)[0]
		url := fmt.Sprintf("logseq://graph/%s?block-id=%s", graphName, task.UUID)
		fmt.Fprintf(out, "%s%s%s\n",
			green.Sprint(task.Page.OriginalName+"§"),
			blueBold.Sprint(url),
			"§"+firstLine,
		)
	}

	return nil
}

// NewTaskLsCmd creates a new task ls subcommand with the specified dependencies.
// If deps is nil, it uses default implementations. Individual nil fields also fall back
// to their defaults, so a test can inject only LogseqAPI + Out and leave GraphName defaulted.
func NewTaskLsCmd(deps *TaskLsDependencies) *cobra.Command {
	if deps == nil {
		deps = &TaskLsDependencies{
			NewAPI:    nil,
			GraphName: nil,
			Out:       nil,
		}
	}

	if deps.NewAPI == nil {
		deps.NewAPI = func() api.LogseqAPI {
			return api.NewLogseqAPI(
				os.Getenv("LOGSEQ_GRAPH_PATH"),
				os.Getenv("LOGSEQ_HOST_URL"),
				os.Getenv("LOGSEQ_API_TOKEN"),
			)
		}
	}

	if deps.GraphName == nil {
		deps.GraphName = logseqGraphName
	}

	if deps.Out == nil {
		deps.Out = os.Stdout
	}

	var flags taskLsFlags

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "ls [tag...]",
		Short: "List tasks from Logseq",
		Long: `List tasks from your Logseq graph via the HTTP API.

Positional arguments filter by tag or page reference. Multiple tags are combined with OR.

Examples:
  lqd task ls
  lqd task ls work
  lqd task ls --done --canceled
  lqd task ls --completed
  lqd task ls --json
  lqd task ls -v work`,
		Args: cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runTaskLs(deps, &flags, args)
		},
	}

	const completedUsage = "Include canceled and done tasks (shorthand for --canceled --done)"

	cmd.Flags().BoolVar(&flags.canceled, "canceled", false, "Include CANCELED tasks")
	cmd.Flags().BoolVar(&flags.done, "done", false, "Include DONE tasks")
	cmd.Flags().BoolVarP(&flags.completed, "completed", "c", false, completedUsage)
	cmd.Flags().BoolVar(&flags.json, "json", false, "Output raw JSON")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Print the Datalog query before results")

	return cmd
}

// taskCmd represents the task command using the default dependencies.
var taskCmd = NewTaskCmd() //nolint:gochecknoglobals

func init() {
	taskCmd.AddCommand(NewTaskAddCmd(nil))
	taskCmd.AddCommand(NewTaskLsCmd(nil))
	rootCmd.AddCommand(taskCmd)
}
