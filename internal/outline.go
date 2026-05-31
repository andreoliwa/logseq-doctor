package internal

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	outlineIndentUnit = "  " // 2 spaces per indent level
	outlineBullet     = "-"
)

// OutlineOptions configures the flat-Markdown-to-outline conversion.
type OutlineOptions struct {
	// KeepBreaks preserves blank lines between blocks as empty "- " bullet lines.
	KeepBreaks bool
}

// outlineLine formats a single outline bullet at the given nesting level.
func outlineLine(level int, lineText string) string {
	return strings.Repeat(outlineIndentUnit, level) + outlineBullet + " " + lineText + "\n"
}

// stripFrontmatter splits YAML frontmatter from the body.
// If the input starts with "---\n" and contains a closing "\n---\n",
// the frontmatter (including fences) is returned separately from the body.
// This must run before goldmark sees the input so that "---" is not parsed
// as a thematic break or setext heading underline.
func stripFrontmatter(input string) (string, string) {
	const fence = "---\n"

	if !strings.HasPrefix(input, fence) {
		return "", input
	}

	rest := input[len(fence):]

	inner, after, found := strings.Cut(rest, "\n---\n")
	if !found {
		return "", input
	}

	// frontmatter includes both fences
	return fence + inner + "\n---\n", after
}

// FlatMarkdownToOutline converts flat Markdown to a Logseq bullet outline.
// It strips YAML frontmatter before parsing and prepends it unchanged to the result.
// It is idempotent: if the conversion is a no-op, the original input is returned.
func FlatMarkdownToOutline(input string, opts OutlineOptions) string {
	frontmatter, body := stripFrontmatter(input)
	result := frontmatter + convertOnce(body, opts)

	// Idempotency: if the result is identical to the input, return the original.
	if result == input {
		return input
	}

	return result
}

// converter walks the goldmark AST and builds the outline string.
type converter struct {
	source       []byte
	currentLevel int
	sb           strings.Builder
	keepBreaks   bool
}

// convertOnce parses and converts a Markdown body (without frontmatter) to outline form.
func convertOnce(body string, opts OutlineOptions) string {
	source := []byte(body)
	reader := text.NewReader(source)
	doc := goldmark.DefaultParser().Parse(reader)

	conv := &converter{ //nolint:exhaustruct
		source:     source,
		keepBreaks: opts.KeepBreaks,
	}

	_ = ast.Walk(doc, conv.walk)

	return conv.sb.String()
}

// walk is the goldmark AST visitor function.
//
//nolint:cyclop
func (c *converter) walk(node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch node.Kind() {
	case ast.KindHeading:
		if entering {
			heading := node.(*ast.Heading) //nolint:forcetypeassert

			if c.keepBreaks && node.HasBlankPreviousLines() {
				c.sb.WriteString(outlineLine(c.currentLevel, ""))
			}

			c.currentLevel = heading.Level
			hashes := strings.Repeat("#", heading.Level)
			headingText := c.inlineText(node)
			c.sb.WriteString(outlineLine(heading.Level-1, hashes+" "+headingText))

			return ast.WalkSkipChildren, nil
		}

	case ast.KindParagraph:
		if entering {
			if c.keepBreaks && node.HasBlankPreviousLines() {
				c.sb.WriteString(outlineLine(c.currentLevel, ""))
			}

			paraText := c.inlineText(node)

			for line := range strings.SplitSeq(strings.TrimRight(paraText, "\n"), "\n") {
				if line != "" {
					c.sb.WriteString(outlineLine(c.currentLevel, line))
				}
			}

			return ast.WalkSkipChildren, nil
		}

	case ast.KindList:
		// Lists do not emit their own bullets; only list items do.
		if entering && c.keepBreaks && node.HasBlankPreviousLines() {
			c.sb.WriteString(outlineLine(c.currentLevel, ""))
		}

	case ast.KindListItem:
		if entering {
			c.handleListItem(node)

			return ast.WalkSkipChildren, nil
		}

	case ast.KindThematicBreak:
		if entering {
			c.sb.WriteString("---\n")
		}
	}

	return ast.WalkContinue, nil
}

// handleListItem emits the list item text and manages level for nested lists.
func (c *converter) handleListItem(node ast.Node) {
	itemText := c.listItemText(node)

	nestedList := findNestedList(node)

	if nestedList == nil {
		// Leaf item: emit at current level.
		c.sb.WriteString(outlineLine(c.currentLevel, itemText))

		return
	}

	// Item has a nested list: emit item text at current level, then recurse.
	c.sb.WriteString(outlineLine(c.currentLevel, itemText))
	c.currentLevel++
	c.walkNestedList(nestedList)
	c.currentLevel--
}

// findNestedList returns the first *ast.List child of a list item, or nil.
func findNestedList(item ast.Node) *ast.List {
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok {
			return list
		}
	}

	return nil
}

// listItemText extracts the text content from a list item's first paragraph or inline block.
func (c *converter) listItemText(item ast.Node) string {
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.Kind() {
		case ast.KindParagraph, ast.KindTextBlock:
			return strings.TrimRight(c.inlineText(child), "\n")

		case ast.KindHeading:
			// A heading nested inside a list item (e.g. idempotency re-parse of "- # Header").
			heading := child.(*ast.Heading) //nolint:forcetypeassert
			hashes := strings.Repeat("#", heading.Level)

			return hashes + " " + strings.TrimRight(c.inlineText(child), "\n")
		}
	}

	return ""
}

// walkNestedList recursively processes a nested list and its items.
func (c *converter) walkNestedList(list *ast.List) {
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		c.handleListItem(item)
	}
}

// inlineText collects the rendered text of all inline children of a node.
// It replicates the Python LogseqRenderer.render_inner / render_link behavior.
func (c *converter) inlineText(node ast.Node) string {
	var builder strings.Builder

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		c.collectInline(child, &builder)
	}

	return builder.String()
}

// collectInline recursively renders inline nodes into builder.
func (c *converter) collectInline(node ast.Node, builder *strings.Builder) {
	switch node.Kind() {
	case ast.KindText:
		textNode := node.(*ast.Text) //nolint:forcetypeassert
		builder.Write(textNode.Value(c.source))

		if textNode.SoftLineBreak() {
			builder.WriteByte('\n')
		}

	case ast.KindLink:
		linkNode := node.(*ast.Link) //nolint:forcetypeassert

		builder.WriteByte('[')

		// Render link text from children.
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			c.collectInline(child, builder)
		}

		builder.WriteString("](")
		builder.Write(linkNode.Destination)
		builder.WriteByte(')')

	default:
		// For all other inline nodes (strong, emphasis, code, etc.) render children.
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			c.collectInline(child, builder)
		}
	}
}
