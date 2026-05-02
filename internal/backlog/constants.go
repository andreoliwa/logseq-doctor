package backlog

import (
	"strings"

	"github.com/andreoliwa/logseq-go/content"
)

// Header represents a backlog section divider.
// Label is the display word(s) without the "tasks" suffix (e.g. "Focus").
// String() always returns "Emoji Label tasks".
type Header struct {
	Emoji string
	Label string
}

// String returns the canonical display form: "emoji label tasks".
func (h Header) String() string { return h.Emoji + " " + h.Label + " tasks" }

// Matches reports whether blockText contains the label, case-insensitively.
func (h Header) Matches(blockText string) bool {
	return strings.Contains(strings.ToLower(blockText), strings.ToLower(h.Label))
}

// NewHeading returns a level-1 heading node with the canonical header text
// followed by a [[quick capture]] page link. Use this when creating a new
// section divider so the user can identify blocks inserted by lqd backlog.
func (h Header) NewHeading() *content.Heading {
	return content.NewHeading(1,
		content.NewText(h.String()+" "),
		content.NewPageLink(quickCapturePageName),
	)
}

// quickCapturePageName is the Logseq page linked in newly created section dividers
// so the user can identify dividers inserted by lqd backlog.
const quickCapturePageName = "quick capture"

// Backlog section header definitions.
// Detection uses Header.Matches (case-insensitive label search).
// Creation uses Header.String() so the canonical emoji+label+tasks is always written.
//
//nolint:gochecknoglobals // named constants for well-known headers
var (
	HeaderFocus     = Header{"🎯", "Focus"}
	HeaderOverdue   = Header{"📅", "Overdue"}
	HeaderNewTasks  = Header{"✨", "New"}
	HeaderTriaged   = Header{"🏷️", "Triaged"}
	HeaderScheduled = Header{"⏰", "Scheduled"}
	HeaderUnranked  = Header{"⤵️", "Unranked"}
)

// allHeaders is the full list used to normalize section dividers on write-back.
//
//nolint:gochecknoglobals // package-level list derived from the Header vars above
var allHeaders = []Header{
	HeaderFocus, HeaderOverdue, HeaderNewTasks,
	HeaderTriaged, HeaderScheduled, HeaderUnranked,
}

// Section values for the PocketBase `section` field.
// Ranked=1, Unranked=2, Orphan=3 so that (backlog_index, section, rank) sorts
// ranked tasks before unranked tasks before orphans within any backlog.
const (
	SectionRanked   = 1 // manually ordered, above the ⤵️ Unranked tasks divider
	SectionUnranked = 2 // under ⤵️ Unranked tasks, 📅 Overdue tasks, ⏰ Scheduled tasks, ✨ New tasks, 🏷️ Triaged tasks
	SectionOrphan   = 3 // not referenced in any backlog page
)
