package backlog

// sectionNewTasksText is the detection substring for the "New tasks" section.
// Existing pages may use "New tasks" without the emoji prefix,
// so detection must match the plain text portion.
const sectionNewTasksText = "New tasks"

// PageQuickCapture is the page link included in section dividers.
const PageQuickCapture = "quick capture"

// Section header constants used in backlog pages.
// These are matched by text content when scanning blocks.
const (
	SectionFocus     = "Focus"
	SectionOverdue   = "📅 Overdue tasks"
	SectionNewTasks  = "🆕 " + sectionNewTasksText
	SectionSomeday   = "📦 Someday"
	SectionScheduled = "⏰ Scheduled tasks"
)
