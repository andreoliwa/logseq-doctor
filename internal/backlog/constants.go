package backlog

import (
	"strings"

	"github.com/andreoliwa/logseq-go/content"
)

// Header represents a backlog section divider with an emoji and a text label.
// The canonical display form is "emoji text" (e.g. "🔢 Unranked tasks").
type Header struct {
	Emoji string
	Text  string
}

// String returns the canonical display form: "emoji text".
func (h Header) String() string { return h.Emoji + " " + h.Text }

// Matches reports whether blockText contains h.Text, case-insensitively.
// This allows detecting existing headers regardless of emoji or capitalisation.
func (h Header) Matches(blockText string) bool {
	return strings.Contains(strings.ToLower(blockText), strings.ToLower(h.Text))
}

// NewParagraph returns a paragraph node with the canonical header text followed
// by a [[quick capture]] page link. Use this when creating a new section divider
// so the user can identify blocks inserted by lqd backlog.
func (h Header) NewParagraph() *content.Paragraph {
	return content.NewParagraph(
		content.NewText(h.String()+" "),
		content.NewPageLink(quickCapturePageName),
	)
}

// quickCapturePageName is the Logseq page linked in newly created section dividers
// so the user can identify dividers inserted by lqd backlog.
const quickCapturePageName = "quick capture"

// Backlog section header definitions.
// Detection uses Header.Matches (case-insensitive text search).
// Creation uses Header.String() so the canonical emoji+text is always written.
//
//nolint:gochecknoglobals // named constants for well-known headers
var (
	HeaderFocus     = Header{"🎯", "Focus"}
	HeaderOverdue   = Header{"📅", "Overdue tasks"}
	HeaderNewTasks  = Header{"🆕", "New tasks"}
	HeaderTriaged   = Header{"🏷️", "Triaged tasks"}
	HeaderScheduled = Header{"⏰", "Scheduled tasks"}
	HeaderUnranked  = Header{"🔢", "Unranked tasks"}
)

// allHeaders is the full list used to normalise section dividers on write-back.
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
	SectionRanked   = 1 // manually ordered, above the 🔢 Unranked tasks divider
	SectionUnranked = 2 // under 🔢 Unranked tasks, 📅 Overdue, ⏰ Scheduled, 🆕 New tasks, 🏷️ Triaged
	SectionOrphan   = 3 // not referenced in any backlog page
)
