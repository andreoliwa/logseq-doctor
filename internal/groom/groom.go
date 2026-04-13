// Package groom implements the grooming business logic for Logseq tasks.
package groom

import (
	"errors"
	"fmt"
	"strings"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	tparse "github.com/karrick/tparse/v2"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

const reGroomDays = 90

const groomSeparator = "--------------------------------------------"

// errEmptyDuration is returned when an empty duration string is provided.
var errEmptyDuration = errors.New("empty duration string")

// ErrBlockIDMissingInFile is returned when a block UUID exists in the Logseq API/DB
// but the id:: property is absent from the .md file. This happens for blocks that
// were assigned a UUID internally by Logseq but have not yet had it written to disk.
// The fix is to open the block in Logseq (which forces the id:: write-back), then re-run groom.
var ErrBlockIDMissingInFile = errors.New("block id:: not found in file")

// CalculateThresholdDate subtracts a human-readable duration from a base time.
// Uses karrick/tparse for calendar-aware math (proper month/year handling).
// Accepts formats like: "5 years", "90 days", "6 months", "1 year", "2 weeks".
func CalculateThresholdDate(base time.Time, olderThan string) (time.Time, error) {
	olderThan = strings.TrimSpace(olderThan)
	if olderThan == "" {
		return time.Time{}, errEmptyDuration
	}

	// tparse expects concatenated format like "5years" — normalize "5 years" → "5years"
	normalized := strings.ReplaceAll(olderThan, " ", "")

	result, err := tparse.AddDuration(base, "-"+normalized)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %w", olderThan, err)
	}

	return result, nil
}

// BuildGroomFilter builds the PocketBase filter for stale tasks.
// thresholdDate is the cutoff: tasks with journal before this date are considered stale.
//
// NOTE: Do NOT add scheduled/deadline filters here. PocketBase date fields store null
// (not empty string) when unset, so `scheduled="` never matches and tasks slip through.
// Future-date filtering is done in Go via HasFutureDate after fetching. See CLAUDE.md.
func BuildGroomFilter(now time.Time, thresholdDate time.Time) string {
	thresholdStr := thresholdDate.Format("2006-01-02 15:04:05.000Z")
	reGroomDate := now.AddDate(0, 0, -reGroomDays).Format("2006-01-02 15:04:05.000Z")

	return fmt.Sprintf(
		"(status='TODO' || status='WAITING')"+
			" && journal<'%s' && journal!=''"+
			" && (groomed=''||groomed<'%s')",
		thresholdStr, reGroomDate,
	)
}

// HasRecentDate reports whether a task should be excluded from the groom queue because
// its scheduled or deadline date is newer than thresholdDate (i.e. not yet stale).
//
// A task with scheduled/deadline older than the threshold is stale enough to review.
// A task with a date newer than the threshold is still active and should be skipped.
// A task with no date is unaffected by this check.
//
// PocketBase date fields store null (not empty string) when unset, so these comparisons
// cannot be done reliably in PocketBase query strings. Filter in Go instead. See CLAUDE.md.
func HasRecentDate(task map[string]any, thresholdDate time.Time) bool {
	threshold := thresholdDate.Format("2006-01-02")

	for _, field := range []string{"scheduled", "deadline"} {
		val, _ := task[field].(string)
		// Compare first 10 chars (YYYY-MM-DD). If the date is >= threshold, the task is recent.
		if len(val) >= len(threshold) && val[:len(threshold)] >= threshold {
			return true
		}
	}

	return false
}

// FormatGroomTask formats a PB task record for terminal display.
func FormatGroomTask(task map[string]any, index, total int, now time.Time) string {
	name, _ := task["name"].(string)
	journalStr, _ := task["journal"].(string)
	backlogName, _ := task["backlog_name"].(string)
	backlogIndex, _ := task["backlog_index"].(float64)
	tags, _ := task["tags"].(string)

	var buf strings.Builder

	fmt.Fprintf(&buf, "\n%s\n", groomSeparator)
	fmt.Fprintf(&buf, " Task %d/%d\n", index, total)
	fmt.Fprintf(&buf, "%s\n", groomSeparator)
	fmt.Fprintf(&buf, " %s\n\n", name)

	journalDate := strings.Split(journalStr, " ")[0]
	age := FormatTaskAge(journalStr, now)
	fmt.Fprintf(&buf, " Created:  %s  (%s)\n", journalDate, age)

	if backlogName != "" {
		fmt.Fprintf(&buf, " Backlog:  %s (#%d)\n", backlogName, int(backlogIndex))
	} else {
		fmt.Fprintf(&buf, " Backlog:  (none)\n")
	}

	if tags != "" {
		fmt.Fprintf(&buf, " Tags:     %s\n", tags)
	}

	fmt.Fprintf(&buf, "%s\n", groomSeparator)

	hasBacklog := backlogName != ""
	if hasBacklog {
		fmt.Fprintf(&buf, " [k]eep  [c]ancel  [f]ocus  [d]efer  [s]kip  [q]uit\n")
	} else {
		fmt.Fprintf(&buf, " [k]eep  [c]ancel  [s]kip  [q]uit\n")
		fmt.Fprintf(&buf, " (no backlog — focus/defer unavailable)\n")
	}

	return buf.String()
}

const (
	daysPerYear  = 365
	daysPerMonth = 30
)

// FormatTaskAge returns a human-readable age string like "9 years ago".
func FormatTaskAge(isoDate string, now time.Time) string {
	if isoDate == "" {
		return "unknown"
	}

	parsed, err := time.Parse("2006-01-02 15:04:05.000Z", isoDate)
	if err != nil {
		return "unknown"
	}

	diff := now.Sub(parsed)
	days := int(diff.Hours() / 24) //nolint:mnd // 24 hours in a day

	switch {
	case days >= daysPerYear:
		years := days / daysPerYear
		if years == 1 {
			return "1 year ago"
		}

		return fmt.Sprintf("%d years ago", years)
	case days >= daysPerMonth:
		months := days / daysPerMonth
		if months == 1 {
			return "1 month ago"
		}

		return fmt.Sprintf("%d months ago", months)
	default:
		if days == 1 {
			return "1 day ago"
		}

		return fmt.Sprintf("%d days ago", days)
	}
}

// Groom action name constants. Use these instead of string literals whenever
// comparing or switching on GroomAction.Name to avoid typos and enable refactoring.
const (
	GroomActionCancel         = "cancel"
	GroomActionFocus          = "focus"
	GroomActionPriorityHigh   = "priority-high"
	GroomActionPriorityMedium = "priority-medium"
	GroomActionPriorityLow    = "priority-low"
	GroomActionSkip           = "skip"
	GroomActionQuit           = "quit"
)

// Logseq block property key constants written by groom actions.
const (
	GroomPropertyGroomed   = "groomed"
	GroomPropertyCancelled = "cancelled"
)

// Action represents a user's grooming decision.
type Action struct {
	Name         string
	SetsGroomed  bool
	RequiresFile bool
	Priority     content.PriorityValue
}

// groomActions maps single-letter inputs to their corresponding actions.
//
//nolint:gochecknoglobals // lookup table used by ParseAction
var groomActions = map[string]*Action{
	"x": {Name: GroomActionCancel, SetsGroomed: true, RequiresFile: true},
	"f": {Name: GroomActionFocus, SetsGroomed: true, RequiresFile: true},
	"a": {Name: GroomActionPriorityHigh, SetsGroomed: true, RequiresFile: true, Priority: content.PriorityHigh},
	"b": {Name: GroomActionPriorityMedium, SetsGroomed: true, RequiresFile: true, Priority: content.PriorityMedium},
	"c": {Name: GroomActionPriorityLow, SetsGroomed: true, RequiresFile: true, Priority: content.PriorityLow},
	"s": {Name: GroomActionSkip, SetsGroomed: false, RequiresFile: false},
	"q": {Name: GroomActionQuit, SetsGroomed: false, RequiresFile: false},
}

// ParseAction converts a single-letter input to an Action.
// Returns nil if the input is invalid or requires backlog when none exists.
func ParseAction(input string, hasBacklog bool) *Action {
	action, ok := groomActions[strings.ToLower(input)]
	if !ok {
		return nil
	}

	if !hasBacklog && input == "f" {
		return nil
	}

	return action
}

// Counts tracks how many tasks received each action.
type Counts struct {
	Cancelled      int
	Focused        int
	PriorityHigh   int
	PriorityMedium int
	PriorityLow    int
	Skipped        int
}

// FormatGroomSummary formats the end-of-session summary.
func FormatGroomSummary(counts Counts, remaining int, olderThan string) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "\n%s\n", groomSeparator)
	fmt.Fprintf(&buf, " Grooming complete!\n")
	fmt.Fprintf(&buf, "%s\n", groomSeparator)
	fmt.Fprintf(&buf, " Cancelled:  %d\n", counts.Cancelled)
	fmt.Fprintf(&buf, " Focus:      %d\n", counts.Focused)
	fmt.Fprintf(&buf, " High (A):   %d\n", counts.PriorityHigh)
	fmt.Fprintf(&buf, " Medium (B): %d\n", counts.PriorityMedium)
	fmt.Fprintf(&buf, " Low (C):    %d\n", counts.PriorityLow)
	fmt.Fprintf(&buf, " Skipped:    %d\n", counts.Skipped)
	fmt.Fprintf(&buf, "%s\n", groomSeparator)
	fmt.Fprintf(&buf, " Remaining ungroomed tasks older than %s: %d\n", olderThan, remaining)
	fmt.Fprintf(&buf, "%s\n", groomSeparator)

	return buf.String()
}

// WriteOpts holds config-derived values needed for write-back operations.
// These are passed from cmd/ to avoid circular imports with internal/backlog.
type WriteOpts struct {
	FocusPageTitle       string
	BacklogPageTitle     string
	TriagedSectionText   string
	ScheduledSectionText string
	CurrentTime          func() time.Time
}

// EnsureBlockOnDisk checks whether a task's block UUID is present in the Logseq .md file.
// If the id:: property is missing from disk (Logseq lazy-writes UUIDs), it calls
// logseq.Editor.upsertBlockProperty via the HTTP API to force Logseq to write it immediately.
//
// Background: Logseq assigns UUIDs to blocks internally but only writes them to .md files
// when triggered (e.g. a backlink is created, the block is edited, or the Editor API writes it).
// Without this, groom cannot locate or modify the block by its UUID and would skip the task.
//
// The caller should print the log message to stdout so the user knows what happened.
//
// Returns (exists bool, upserted bool):
//   - exists=true means the block is on disk (possibly after the upsert triggered a write).
//   - upserted=true means the API call was made (Logseq may need a moment to flush; a second
//     groom run will reliably find it even if the file hasn't been updated within this process).
func EnsureBlockOnDisk(graph *logseq.Graph, groomAPI logseqapi.LogseqAPI, task map[string]any) (bool, bool) {
	uuid, _ := task["task_uuid"].(string)
	if uuid == "" {
		return false, false
	}

	blockInfo, err := logseqapi.FindBlockByUUID(groomAPI, uuid)
	if err != nil {
		return false, false
	}

	transaction := graph.NewTransaction()

	page, err := openPageForBlock(transaction, blockInfo)
	if err != nil {
		return false, false
	}

	if logseqext.FindBlockByIDProperty(page, uuid) != nil {
		return true, false
	}

	// Block not on disk yet — call the Logseq Editor API to force the write-back.
	// logseq.Editor.upsertBlockProperty triggers Logseq to persist the id:: property
	// to the .md file. We pass the UUID as both key and value because that's how Logseq
	// stores the block identity (id:: <uuid>).
	upsertErr := groomAPI.UpsertBlockProperty(uuid, "id", uuid)
	if upsertErr != nil {
		// API unavailable or failed — can't force the write. Skip gracefully.
		return false, false
	}

	// Re-open the page to pick up the freshly-written file.
	transaction2 := graph.NewTransaction()

	page2, err := openPageForBlock(transaction2, blockInfo)
	if err != nil {
		return false, true // upserted but can't verify yet
	}

	return logseqext.FindBlockByIDProperty(page2, uuid) != nil, true
}

// ApplyGroomAction applies a groom action to a Logseq block.
func ApplyGroomAction(
	graph *logseq.Graph, groomAPI logseqapi.LogseqAPI, action *Action,
	task map[string]any, opts *WriteOpts,
) error {
	if action.Name == GroomActionSkip {
		return nil
	}

	uuid, _ := task["task_uuid"].(string)
	groomedDate := logseqext.FormatLogseqDate(opts.CurrentTime())

	blockInfo, err := logseqapi.FindBlockByUUID(groomAPI, uuid)
	if err != nil {
		return fmt.Errorf("failed to find block %s: %w", uuid, err)
	}

	transaction := graph.NewTransaction()

	page, err := openPageForBlock(transaction, blockInfo)
	if err != nil {
		return fmt.Errorf("failed to open page for block %s: %w", uuid, err)
	}

	block := logseqext.FindBlockByIDProperty(page, uuid)
	if block == nil {
		return fmt.Errorf("%w: %s", ErrBlockIDMissingInFile, uuid)
	}

	applyErr := applyActionToBlock(transaction, action, block, groomedDate, uuid, opts)
	if applyErr != nil {
		return applyErr
	}

	saveErr := transaction.Save()
	if saveErr != nil {
		return fmt.Errorf("failed to save transaction: %w", saveErr)
	}

	return nil
}

// openPageForBlock opens the appropriate page (journal or regular) for a block.
func openPageForBlock(transaction *logseq.Transaction, blockInfo *logseqapi.BlockQueryInfo) (logseq.Page, error) {
	if blockInfo.IsJournal {
		page, err := transaction.OpenJournal(blockInfo.JournalDate)
		if err != nil {
			return nil, fmt.Errorf("failed to open journal page: %w", err)
		}

		return page, nil
	}

	page, err := transaction.OpenPage(blockInfo.PageName)
	if err != nil {
		return nil, fmt.Errorf("failed to open page %s: %w", blockInfo.PageName, err)
	}

	return page, nil
}

// applyActionToBlock applies the specific groom action to a block.
func applyActionToBlock(
	transaction *logseq.Transaction, action *Action,
	block *content.Block, groomedDate, uuid string, opts *WriteOpts,
) error {
	switch action.Name {
	case GroomActionCancel:
		return applyCancelAction(block, groomedDate)
	case GroomActionFocus:
		return applyFocusAction(transaction, block, groomedDate, uuid, opts.FocusPageTitle)
	case GroomActionPriorityHigh, GroomActionPriorityMedium, GroomActionPriorityLow:
		return applyPriorityAction(transaction, block, groomedDate, uuid, action.Priority, opts)
	}

	return nil
}

// applyCancelAction sets the task as canceled with a cancelled:: date property.
// The groomed:: property is intentionally omitted — cancelled:: is sufficient and avoids clutter.
func applyCancelAction(block *content.Block, groomedDate string) error {
	cancelErr := logseqext.SetTaskCanceled(block)
	if cancelErr != nil {
		return fmt.Errorf("failed to set task canceled: %w", cancelErr)
	}

	logseqext.BlockProperties(block).Set(GroomPropertyCancelled, content.NewText(groomedDate))

	return nil
}

// applyFocusAction marks the block groomed and adds a reference to the Focus page.
func applyFocusAction(
	transaction *logseq.Transaction, block *content.Block,
	groomedDate, uuid, focusPageTitle string,
) error {
	logseqext.BlockProperties(block).Set(GroomPropertyGroomed, content.NewText(groomedDate))

	focusErr := backlog.AddBlockRefToFocusPage(transaction, focusPageTitle, uuid)
	if focusErr != nil {
		return fmt.Errorf("failed to add block ref to focus page: %w", focusErr)
	}

	return nil
}

// applyPriorityAction sets priority on the block, marks it groomed,
// and adds a reference to the Triaged section of the backlog page (if the task has a backlog).
func applyPriorityAction(
	transaction *logseq.Transaction, block *content.Block,
	groomedDate, uuid string, priority content.PriorityValue, opts *WriteOpts,
) error {
	priorityErr := logseqext.SetPriority(block, priority)
	if priorityErr != nil {
		return fmt.Errorf("failed to set priority: %w", priorityErr)
	}

	logseqext.BlockProperties(block).Set(GroomPropertyGroomed, content.NewText(groomedDate))

	if opts.BacklogPageTitle != "" {
		triagedErr := backlog.MoveBlockRefToTriagedSection(
			transaction, opts.BacklogPageTitle, uuid, opts.TriagedSectionText, opts.ScheduledSectionText,
		)
		if triagedErr != nil {
			return fmt.Errorf("failed to add block ref to triaged section: %w", triagedErr)
		}
	}

	return nil
}
