package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/groom"
	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const groomDefaultLimit = 10
const groomDefaultOlderThan = "1 year"
const groomFetchMultiplier = 5 // fetch 5× the limit to absorb tasks filtered out by HasFutureDate

// errGroomNoCollection is returned when the lqd_tasks collection does not exist in PocketBase.
var errGroomNoCollection = errors.New("no tasks found. Run 'lqd sync --init' first")

// GroomDependencies holds all injectable dependencies for the groom command.
// This enables unit testing without connecting to PocketBase, Logseq, or a terminal.
type GroomDependencies struct {
	TimeNow func() time.Time
}

// NewGroomCmd creates a new groom command with the specified dependencies.
// If deps is nil, it uses default (production) implementations.
func NewGroomCmd(deps *GroomDependencies) *cobra.Command {
	if deps == nil {
		deps = &GroomDependencies{
			TimeNow: time.Now,
		}
	}

	var (
		olderThan string
		limit     int
	)

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "groom",
		Short: "Interactively review and groom stale tasks",
		Long:  "Queries PocketBase for old ungroomed tasks and presents them one at a time for action.",
		Run: func(_ *cobra.Command, _ []string) {
			runGroomWith(deps.TimeNow(), olderThan, limit)
		},
	}

	cmd.Flags().StringVar(&olderThan, "older-than", groomDefaultOlderThan,
		"Task age threshold (e.g. \"5 years\", \"90 days\")")
	cmd.Flags().IntVar(&limit, "limit", groomDefaultLimit, "Maximum tasks to review")

	return cmd
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(NewGroomCmd(nil))
}

// runGroomWith is the testable core of runGroom.
func runGroomWith(now time.Time, olderThan string, limit int) {
	thresholdDate, err := groom.CalculateThresholdDate(now, olderThan)
	if err != nil {
		fmt.Printf("Invalid --older-than value: %v\n", err)
		os.Exit(1)
	}

	pbClient, tasks, err := fetchGroomTasks(now, thresholdDate, limit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if tasks == nil {
		return
	}

	graph, api, backlogConfig, err := openGroomResources()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(groomStyles.warning.Render("Note: avoid editing tasks in Logseq while grooming."))

	pbUpdater := func(recordID string, groomedAt time.Time) error {
		return pbClient.UpdateRecord("lqd_tasks", recordID, map[string]any{
			"groomed": groomedAt.UTC().Format("2006-01-02 15:04:05.000Z"),
		})
	}

	counts := processGroomTasks(tasks, now, graph, api, backlogConfig, pbUpdater)

	allTasks, _ := pbClient.FetchRecords("lqd_tasks", groom.BuildGroomFilter(now, thresholdDate), "")
	remaining := len(allTasks)

	fmt.Print(groom.FormatGroomSummary(counts, remaining, olderThan))
}

// fetchGroomTasks initialises PocketBase, checks the collection, and fetches matching tasks.
// Returns (nil, nil, nil) with a printed message when there are no tasks.
func fetchGroomTasks(now, thresholdDate time.Time, limit int) (*pocketbase.Client, []map[string]any, error) {
	pbURL := os.Getenv("POCKETBASE_URL")
	if pbURL == "" {
		pbURL = "http://127.0.0.1:8090"
	}

	pbClient, err := pocketbase.NewClient(pbURL, os.Getenv("POCKETBASE_USERNAME"), os.Getenv("POCKETBASE_PASSWORD"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to PocketBase: %w", err)
	}

	exists, err := pbClient.CollectionExists("lqd_tasks")
	if err != nil || !exists {
		return nil, nil, errGroomNoCollection
	}

	filter := groom.BuildGroomFilter(now, thresholdDate)

	// Fetch more than the limit to account for tasks filtered out by HasFutureDate.
	// PocketBase date fields store null (not empty string) when unset, so scheduled/deadline
	// comparisons in the query string are unreliable — we filter in Go instead. See CLAUDE.md.
	fetchLimit := limit * groomFetchMultiplier

	rawTasks, err := pbClient.FetchRecords("lqd_tasks", filter, "journal", fetchLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query tasks: %w", err)
	}

	tasks := make([]map[string]any, 0, len(rawTasks))

	for _, t := range rawTasks {
		if !groom.HasFutureDate(t, now) {
			tasks = append(tasks, t)
		}
	}

	if len(tasks) > limit {
		tasks = tasks[:limit]
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found matching criteria.")

		return pbClient, nil, nil
	}

	return pbClient, tasks, nil
}

// openGroomResources opens the Logseq graph, API, and reads the backlog config.
func openGroomResources() (*logseq.Graph, api.LogseqAPI, *backlog.Config, error) {
	path := os.Getenv("LOGSEQ_GRAPH_PATH")
	graph := api.OpenGraphFromPath(path)
	api := api.NewLogseqAPI(path, os.Getenv("LOGSEQ_HOST_URL"), os.Getenv("LOGSEQ_API_TOKEN"))

	configReader := backlog.NewPageConfigReader(graph, "backlog")

	backlogConfig, err := configReader.ReadConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read backlog config: %w", err)
	}

	return graph, api, backlogConfig, nil
}

// logseqGraphName extracts the graph name from the graph path for deep links.
func logseqGraphName() string {
	path := os.Getenv("LOGSEQ_GRAPH_PATH")
	if path == "" {
		return "my-graph"
	}

	// Use the last component of the path as the graph name.
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}

	return path
}

// groomStyles holds the lipgloss styles for the groom TUI.
var groomStyles = struct { //nolint:gochecknoglobals
	separator lipgloss.Style
	header    lipgloss.Style
	taskName  lipgloss.Style
	label     lipgloss.Style
	value     lipgloss.Style
	age       lipgloss.Style
	actions   lipgloss.Style
	prompt    lipgloss.Style
	success   lipgloss.Style
	warning   lipgloss.Style
	errStyle  lipgloss.Style
	link      lipgloss.Style
}{
	separator: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	header:    lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true),
	taskName:  lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true),
	label:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
	value:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	age:       lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Italic(true),
	actions:   lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true),
	prompt:    lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true),
	success:   lipgloss.NewStyle().Foreground(lipgloss.Color("82")),
	warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
	errStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	link:      lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Underline(true),
}

const groomSep = "────────────────────────────────────────────"
const groomDefaultTermWidth = 80
const groomNameMargin = 2 // leading space + right margin

// processGroomTasks presents tasks one at a time in a plain scrolling terminal loop.
// Each task card is printed, the user presses a key, the result is printed, then the
// next task scrolls into view. No alternate screen — every action is permanently visible.
func processGroomTasks(
	tasks []map[string]any, now time.Time,
	graph *logseq.Graph, api api.LogseqAPI, backlogConfig *backlog.Config,
	pbUpdater func(recordID string, groomedAt time.Time) error,
) groom.Counts {
	var counts groom.Counts

	termWidth := terminalWidth()

	displayed := 0

	for _, task := range tasks {
		// Pre-screen: ensure the block UUID is present in the .md file before showing
		// the task card. Logseq lazy-writes UUIDs; EnsureBlockOnDisk calls the Logseq
		// Editor API to force the write when needed, then re-checks the file.
		exists, upserted := groom.EnsureBlockOnDisk(graph, api, task)

		if upserted {
			taskID, _ := task["id"].(string)
			fmt.Println(groomStyles.warning.Render(
				" → Triggered Logseq to write id:: for block " + taskID,
			))
		}

		if !exists {
			taskID, _ := task["id"].(string)
			fmt.Println(groomStyles.warning.Render(
				" ⚠ Skipped (id:: not on disk after upsert): " + taskID + " — will be available on next run.",
			))

			counts.Skipped++

			continue
		}

		displayed++
		printTaskCard(task, displayed, len(tasks), now, termWidth)

		quit := groomHandleTask(graph, api, backlogConfig, pbUpdater, task, now, &counts)
		if quit {
			break
		}
	}

	return counts
}

// groomHandleTask reads keypresses for a single task until a valid action is taken.
// Returns true if the user requested quit.
func groomHandleTask(
	graph *logseq.Graph, api api.LogseqAPI, backlogConfig *backlog.Config,
	pbUpdater func(recordID string, groomedAt time.Time) error,
	task map[string]any, now time.Time, counts *groom.Counts,
) bool {
	backlogName, _ := task["backlog_name"].(string)

	for {
		key, err := readKey()
		if err != nil {
			fmt.Println(groomStyles.errStyle.Render("Error reading input: " + err.Error()))

			return true
		}

		if key == "\x03" { // ctrl+c
			return true
		}

		action := groom.ParseAction(key, backlogName != "")
		if action == nil {
			fmt.Println(groomStyles.warning.Render(" Invalid key — try again."))

			continue
		}

		if action.Name == groom.GroomActionQuit {
			return true
		}

		opts := &groom.WriteOpts{
			FocusPageTitle:       backlogConfig.FocusPage,
			BacklogPageTitle:     backlogConfig.FindBacklogPageTitle(backlogName),
			SomedaySectionText:   backlog.SectionSomeday,
			ScheduledSectionText: backlog.SectionScheduled,
			CurrentTime:          time.Now,
		}

		applyErr := groom.ApplyGroomAction(graph, api, action, task, opts)
		if applyErr != nil {
			groomPrintApplyError(applyErr, counts)

			return false
		}

		updateGroomCounts(counts, action.Name)
		fmt.Println(groomStyles.success.Render(" ✓ " + action.Name))
		groomSyncPocketBase(pbUpdater, action, task, now)

		return false
	}
}

func groomPrintApplyError(applyErr error, counts *groom.Counts) {
	counts.Skipped++

	if errors.Is(applyErr, groom.ErrBlockIDMissingInFile) {
		fmt.Println(groomStyles.warning.Render(
			" ⚠ Skipped: block not on disk yet. Open in Logseq, then re-run groom.",
		))

		return
	}

	fmt.Println(groomStyles.errStyle.Render(" ✗ Error: " + applyErr.Error()))
}

func groomSyncPocketBase(
	pbUpdater func(recordID string, groomedAt time.Time) error,
	action *groom.Action, task map[string]any, now time.Time,
) {
	if !action.SetsGroomed || pbUpdater == nil {
		return
	}

	taskID, _ := task["id"].(string)

	pbErr := pbUpdater(taskID, now)
	if pbErr != nil {
		fmt.Println(groomStyles.warning.Render(" ⚠ PB update failed: " + pbErr.Error()))
	}
}

// printTaskCard prints a single task card to stdout.
func printTaskCard(task map[string]any, index, total int, now time.Time, termWidth int) {
	name, _ := task["name"].(string)
	journalStr, _ := task["journal"].(string)
	backlogName, _ := task["backlog_name"].(string)
	backlogIndex, _ := task["backlog_index"].(float64)
	tags, _ := task["tags"].(string)
	taskID, _ := task["id"].(string)
	scheduledStr, _ := task["scheduled"].(string)
	deadlineStr, _ := task["deadline"].(string)

	journalDate := strings.Split(journalStr, " ")[0]
	age := groom.FormatTaskAge(journalStr, now)
	hasBacklog := backlogName != ""
	sep := groomStyles.separator.Render(groomSep)

	nameWidth := max(termWidth-groomNameMargin, groomDefaultTermWidth)
	wrappedName := lipgloss.NewStyle().Inherit(groomStyles.taskName).Width(nameWidth).Render(name)

	fmt.Println()
	fmt.Println(sep)
	fmt.Println(groomStyles.header.Render(fmt.Sprintf(" Task %d/%d", index, total)))
	fmt.Println(sep)
	fmt.Println(" " + wrappedName)
	fmt.Println()
	fmt.Println(" " + groomStyles.label.Render("Created:   ") +
		groomStyles.value.Render(journalDate) + "  " + groomStyles.age.Render("("+age+")"))

	if scheduledStr != "" {
		fmt.Println(" " + groomStyles.label.Render("Scheduled: ") +
			groomStyles.warning.Render(strings.Split(scheduledStr, "T")[0]))
	}

	if deadlineStr != "" {
		fmt.Println(" " + groomStyles.label.Render("Deadline:  ") +
			groomStyles.warning.Render(strings.Split(deadlineStr, "T")[0]))
	}

	if hasBacklog {
		fmt.Println(" " + groomStyles.label.Render("Backlog:   ") +
			groomStyles.value.Render(fmt.Sprintf("%s (#%d)", backlogName, int(backlogIndex))))
	} else {
		fmt.Println(" " + groomStyles.label.Render("Backlog:   ") + groomStyles.warning.Render("(none)"))
	}

	if tags != "" {
		fmt.Println(" " + groomStyles.label.Render("Tags:      ") + groomStyles.value.Render(tags))
	}

	if taskID != "" {
		url := "logseq://graph/" + logseqGraphName() + "?block-id=" + taskID
		fmt.Println(" " + groomStyles.link.Render(url))
	}

	fmt.Println(sep)

	if hasBacklog {
		fmt.Println(" " + groomStyles.actions.Render("[k]eep  [c]ancel  [f]ocus  [d]efer  [s]kip  [q]uit"))
	} else {
		fmt.Println(" " + groomStyles.actions.Render("[k]eep  [c]ancel  [s]kip  [q]uit"))
		fmt.Println(" " + groomStyles.warning.Render("(no backlog — focus/defer unavailable)"))
	}
}

// readKey reads a single keypress from /dev/tty without requiring Enter.
func readKey() (string, error) {
	ctx := context.Background()

	tty, err := os.Open("/dev/tty")
	if err != nil {
		return "", fmt.Errorf("open /dev/tty: %w", err)
	}

	defer tty.Close()

	rawCmd := exec.CommandContext(ctx, "stty", "cbreak", "-echo")
	rawCmd.Stdin = tty

	err = rawCmd.Run()
	if err != nil {
		return "", fmt.Errorf("stty cbreak: %w", err)
	}

	defer func() {
		restoreCmd := exec.CommandContext(ctx, "stty", "-cbreak", "echo")
		restoreCmd.Stdin = tty
		_ = restoreCmd.Run()
	}()

	buf := make([]byte, 1)

	_, err = tty.Read(buf)
	if err != nil {
		return "", fmt.Errorf("read key: %w", err)
	}

	return string(buf), nil
}

// terminalWidth returns the current terminal width, falling back to groomDefaultTermWidth.
func terminalWidth() int {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return groomDefaultTermWidth
	}

	defer tty.Close()

	cmd := exec.CommandContext(context.Background(), "stty", "size")
	cmd.Stdin = tty

	out, err := cmd.Output()
	if err != nil {
		return groomDefaultTermWidth
	}

	var rows, cols int

	_, err = fmt.Sscan(string(out), &rows, &cols)
	if err != nil || cols == 0 {
		return groomDefaultTermWidth
	}

	return cols
}

// updateGroomCounts increments the appropriate counter for the given action name.
func updateGroomCounts(counts *groom.Counts, actionName string) {
	switch actionName {
	case groom.GroomActionKeep:
		counts.Kept++
	case groom.GroomActionCancel:
		counts.Cancelled++
	case groom.GroomActionFocus:
		counts.Focused++
	case groom.GroomActionDefer:
		counts.Deferred++
	case groom.GroomActionSkip:
		counts.Skipped++
	}
}
