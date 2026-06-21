package logseqext

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
)

// ErrNoParagraph is returned when SetPriority is called on a block with no paragraph or heading.
var ErrNoParagraph = errors.New("block has no paragraph to insert priority into")

// JournalDayDivisorYear is used to extract the year from a journalDay integer (YYYYMMDD).
const JournalDayDivisorYear = 10000

// JournalDayDivisorMonth is used to extract the month from a journalDay integer (YYYYMMDD).
const JournalDayDivisorMonth = 100

// PropertyCancelled is the Logseq block property key written when a task is cancelled.
const PropertyCancelled = "cancelled"

// BlockProperties returns the Properties node for a task block.
// logseq-go's block.Properties() only checks the first child — task blocks have a
// Paragraph first, so it would prepend a new empty Properties node before the TODO line.
// The parser creates an empty Properties placeholder at position 0; the real properties
// (id::, collapsed::, etc.) appear after the first Paragraph.
// This helper finds the first Properties that comes after a Paragraph in the content.
func BlockProperties(block *content.Block) *content.Properties {
	seenParagraph := false

	for _, node := range block.Content() {
		if _, ok := node.(*content.Paragraph); ok {
			seenParagraph = true

			continue
		}

		if p, ok := node.(*content.Properties); ok && seenParagraph {
			return p
		}
	}

	// No properties after a paragraph — fall back to block.Properties() which creates one.
	// This path is taken for blocks that truly have no properties yet.
	return block.Properties()
}

// FindBlockByIDProperty finds a block in a page by its id:: property value.
// It searches Properties nodes within block content because logseq-go's block.Properties()
// only checks the first child — task blocks have a paragraph first, then properties.
func FindBlockByIDProperty(page logseq.Page, uuid string) *content.Block {
	return page.Blocks().FindDeep(func(block *content.Block) bool {
		return blockHasIDProperty(block, uuid)
	})
}

// blockHasIDProperty checks whether a block's id:: property matches the given UUID.
// Searches inside content nodes to handle task blocks where properties follow the paragraph.
func blockHasIDProperty(block *content.Block, uuid string) bool {
	found := false

	block.Content().FindDeep(func(node content.Node) bool {
		props, ok := node.(*content.Properties)
		if !ok {
			return false
		}

		for _, v := range props.Get("id") {
			if t, ok := v.(*content.Text); ok && strings.Contains(t.Value, uuid) {
				found = true

				return true
			}
		}

		return false
	})

	return found
}

// BlockContentText extracts the text content from a block's content nodes.
func BlockContentText(block *content.Block) string {
	var text string

	block.Content().FindDeep(func(node content.Node) bool {
		if textNode, ok := node.(*content.Text); ok {
			text = textNode.Value

			return true
		}

		return false
	})

	return text
}

// headingTaskKeywords lists task status strings that logseq-go embeds as Text inside heading nodes.
//
//nolint:gochecknoglobals
var headingTaskKeywords = content.TaskStatusStrings()

// findTaskMarker returns the TaskMarker node in a block, or nil if absent.
func findTaskMarker(block *content.Block) *content.TaskMarker {
	var taskMarker *content.TaskMarker

	block.Content().FindDeep(func(node content.Node) bool {
		if marker, ok := node.(*content.TaskMarker); ok {
			taskMarker = marker

			return true
		}

		return false
	})

	return taskMarker
}

// replaceHeadingTaskKeyword replaces the leading task keyword in a heading block's first Text node.
func replaceHeadingTaskKeyword(block *content.Block, newKeyword string) {
    // TODO: heading blocks parsed by [[logseq-go]] embed the task keyword as plain Text rather than TaskMarker
	var heading *content.Heading

	block.Content().FindDeep(func(node content.Node) bool {
		if h, ok := node.(*content.Heading); ok {
			heading = h

			return true
		}

		return false
	})

	if heading == nil {
		return
	}

	firstText, ok := heading.FirstChild().(*content.Text)
	if !ok {
		return
	}

	for _, kw := range headingTaskKeywords {
		if strings.HasPrefix(firstText.Value, kw+" ") {
			firstText.Value = newKeyword + firstText.Value[len(kw):]

			return
		}
	}
}

// SetTaskCanceled changes the task marker to CANCELED and sets the cancelled:: date property.
// An optional date argument overrides the default (today). The cancelled:: property is always set.
func SetTaskCanceled(block *content.Block, date ...time.Time) error {
	taskMarker := findTaskMarker(block)

	if taskMarker == nil {
		replaceHeadingTaskKeyword(block, content.TaskStatusCanceled.String())
	} else {
		_, err := taskMarker.WithStatus(content.TaskStatusCanceled)
		if err != nil {
			return fmt.Errorf("failed to change task status to canceled: %w", err)
		}
	}

	cancelDate := time.Now()
	if len(date) > 0 {
		cancelDate = date[0]
	}

	BlockProperties(block).Set(PropertyCancelled, content.NewText(FormatLogseqDate(cancelDate)))

	return nil
}

// SetTaskWaiting changes the task marker to WAITING.
func SetTaskWaiting(block *content.Block) error {
	taskMarker := findTaskMarker(block)

	if taskMarker == nil {
		replaceHeadingTaskKeyword(block, content.TaskStatusWaiting.String())

		return nil
	}

	_, err := taskMarker.WithStatus(content.TaskStatusWaiting)
	if err != nil {
		return fmt.Errorf("failed to change task status to waiting: %w", err)
	}

	return nil
}

// SetTaskTodo changes the task marker to TODO.
func SetTaskTodo(block *content.Block) error {
	taskMarker := findTaskMarker(block)

	if taskMarker == nil {
		replaceHeadingTaskKeyword(block, content.TaskStatusTodo.String())

		return nil
	}

	_, err := taskMarker.WithStatus(content.TaskStatusTodo)
	if err != nil {
		return fmt.Errorf("failed to change task status to todo: %w", err)
	}

	return nil
}

// SetPriority sets or replaces the priority marker ([#A]/[#B]/[#C]) on a block.
// If a Priority node exists, it is updated in place. Otherwise, a new Priority node
// is inserted after the TaskMarker (or at the start of the first paragraph or heading).
func SetPriority(block *content.Block, priority content.PriorityValue) error {
	var existingPriority *content.Priority

	var taskMarker *content.TaskMarker

	var firstContainer content.HasChildren

	block.Content().FindDeep(func(node content.Node) bool {
		switch typedNode := node.(type) {
		case *content.Paragraph:
			if firstContainer == nil {
				firstContainer = typedNode
			}
		case *content.Heading:
			if firstContainer == nil {
				firstContainer = typedNode
			}
		case *content.Priority:
			existingPriority = typedNode
		case *content.TaskMarker:
			taskMarker = typedNode
		}

		return false
	})

	if existingPriority != nil {
		existingPriority.WithPriority(priority)
		// Clear OriginalText so the serializer uses the Priority field, not the stale parsed text.
		existingPriority.OriginalText = ""

		return nil
	}

	if firstContainer == nil {
		return ErrNoParagraph
	}

	return insertPriority(content.NewPriority(priority), taskMarker, firstContainer)
}

// insertPriority places a new Priority node after the TaskMarker (or task keyword text) in the container.
// Falls back to prepending when no task anchor is found.
func insertPriority(
	newPriority *content.Priority, taskMarker *content.TaskMarker, container content.HasChildren,
) error {
	if taskMarker != nil {
		// Use the TaskMarker's own parent, not container,
		// because a block can have multiple paragraphs (e.g. SCHEDULED line).
		taskMarker.Parent().InsertChildAfter(newPriority, taskMarker)

		return nil
	}

	// Heading blocks: parser embeds the task keyword in a Text node rather than a TaskMarker.
	// Find that text node and split it so the priority sits right after the keyword.
	firstText, ok := container.FirstChild().(*content.Text)
	if !ok {
		container.PrependChild(newPriority)

		return nil
	}

	for _, kw := range headingTaskKeywords {
		if strings.HasPrefix(firstText.Value, kw+" ") {
			// Split "KEYWORD rest" into "KEYWORD " + [#X] + " rest".
			remainder := firstText.Value[len(kw)+1:]
			firstText.Value = kw + " "
			container.InsertChildAfter(content.NewText(remainder), firstText)
			container.InsertChildAfter(newPriority, firstText)

			return nil
		}
	}

	container.PrependChild(newPriority)

	return nil
}

// priorityRegex matches Logseq priority markers like [#A], [#B], [#C] in content strings.
var priorityRegex = regexp.MustCompile(`\[#([ABC])\]`)

// ParsePriorityFromContent extracts a priority value from a Logseq API content string.
// It looks for [#A], [#B], or [#C] patterns and returns the corresponding PriorityValue.
// Returns PriorityNone if no priority marker is found.
func ParsePriorityFromContent(contentStr string) content.PriorityValue {
	match := priorityRegex.FindStringSubmatch(contentStr)
	if len(match) < 2 { //nolint:mnd
		return content.PriorityNone
	}

	return content.ParsePriorityFromLetter(match[1])
}

// AddSibling inserts newBlock into page relative to the given anchor blocks.
// It inserts after the last non-nil block in after, or before the before block,
// or appends to the page if neither is provided.
func AddSibling(page logseq.Page, newBlock, before *content.Block, after ...*content.Block) {
	for _, a := range after {
		if a != nil {
			page.InsertBlockAfter(newBlock, a)

			return
		}
	}

	if before != nil {
		page.InsertBlockBefore(newBlock, before)

		return
	}

	page.AddBlock(newBlock)
}

// RemoveEmptyBlocks removes blocks that have no child blocks and returns true if any were removed.
func RemoveEmptyBlocks(save bool, blocks ...*content.Block) bool {
	for _, block := range blocks {
		if block != nil && len(block.Blocks()) == 0 {
			block.RemoveSelf()

			save = true
		}
	}

	return save
}

// OpenGraphFromPath opens a Logseq graph from the given directory path.
// Aborts the program if path is empty or the graph cannot be opened,
// to avoid error-handling boilerplate at every call site.
func OpenGraphFromPath(path string) *logseq.Graph {
	if path == "" {
		log.Fatalln("path is empty, maybe the LOGSEQ_GRAPH_PATH environment variable is not set")
	}

	ctx := context.Background()

	graph, err := logseq.Open(ctx, path)
	if err != nil {
		log.Fatalf("error opening graph: %v", err)
	}

	return graph
}

// OpenPage opens a page in the Logseq graph.
// Aborts the program in case of error to avoid boilerplate at every call site.
func OpenPage(graph *logseq.Graph, pageTitle string) logseq.Page {
	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		log.Fatalf("error opening page: %v", err)
	}

	return page
}

// priorityPrefixRegex matches a priority marker at the start of a line for stripping.
var priorityPrefixRegex = regexp.MustCompile(`^\[#[ABC]\]\s*`)

// ExtractBlockRefUUID extracts the UUID from a block ref inside a child block.
func ExtractBlockRefUUID(block *content.Block) string {
	var uuid string

	block.Content().FindDeep(func(node content.Node) bool {
		if ref, ok := node.(*content.BlockRef); ok {
			uuid = ref.ID

			return true
		}

		return false
	})

	return uuid
}

// ExtractFirstLine extracts the first line of task content, stripping the marker and priority.
func ExtractFirstLine(taskContent string) string {
	line, _, _ := strings.Cut(taskContent, "\n")

	for _, status := range content.TaskStatusStrings() {
		line = strings.TrimPrefix(line, status+" ")
	}

	// Strip priority marker
	line = priorityPrefixRegex.ReplaceAllString(line, "")

	return strings.TrimSpace(line)
}
