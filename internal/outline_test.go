package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
)

func TestFlatMarkdownToOutlineGolden(t *testing.T) {
	inputBytes, err := os.ReadFile(filepath.Join("testdata", "outline", "dirty.md"))
	require.NoError(t, err)

	result := internal.FlatMarkdownToOutline(string(inputBytes), internal.OutlineOptions{})

	golden.Assert(t, result, filepath.Join("outline", "dirty.md.golden"))
}

func TestFlatMarkdownToOutline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "test_links",
			input: "#  Header\n\n" +
				"-  [Link only](https://example.com)\n" +
				"-   Text before, then [a link](https://link.com), then text after\n" +
				"- Only text before, [link a the end](https://endlink.com)\n",
			want: "- # Header\n" +
				"  - [Link only](https://example.com)\n" +
				"  - Text before, then [a link](https://link.com), then text after\n" +
				"  - Only text before, [link a the end](https://endlink.com)\n",
		},
		{
			name: "test_flat_paragraphs_without_header",
			input: "Some flat paragraph.\n\n" +
				"[Link only](https://example.com).\n" +
				"Text before, then [a link](https://link.com), then text after.\n\n" +
				"Only text before, [link a the end](https://endlink.com).\n",
			want: "- Some flat paragraph.\n" +
				"- [Link only](https://example.com).\n" +
				"- Text before, then [a link](https://link.com), then text after.\n" +
				"- Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "test_flat_paragraphs_with_header",
			input: "# Some sneaky header\n" +
				"Some flat paragraph.\n\n" +
				"[Link only](https://example.com).\n" +
				"Text before, then [a link](https://link.com), then text after.\n\n" +
				"Only text before, [link a the end](https://endlink.com).\n",
			want: "- # Some sneaky header\n" +
				"  - Some flat paragraph.\n" +
				"  - [Link only](https://example.com).\n" +
				"  - Text before, then [a link](https://link.com), then text after.\n" +
				"  - Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "test_flat_paragraphs_with_deeper_headers",
			input: "## Some sneaky h2 without h1\n" +
				"Some flat paragraph.\n\n" +
				"[Link only](https://example.com).\n" +
				"Text before, then [a link](https://link.com), then text after.\n\n" +
				"Only text before, [link a the end](https://endlink.com).\n",
			want: "  - ## Some sneaky h2 without h1\n" +
				"    - Some flat paragraph.\n" +
				"    - [Link only](https://example.com).\n" +
				"    - Text before, then [a link](https://link.com), then text after.\n" +
				"    - Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "test_nested_lists_single_level",
			input: "# Header\n\n" +
				"- Parent\n" +
				"  - Child 1\n" +
				"  - Child 2\n",
			want: "- # Header\n" +
				"  - Parent\n" +
				"    - Child 1\n" +
				"    - Child 2\n",
		},
		{
			name: "test_nested_lists_multiple_levels",
			input: "# Header\n\n" +
				"- Parent\n" +
				"  - Child 1\n" +
				"    - Grand child 1.1\n" +
				"    - Grand child 1.2\n" +
				"    - Grand child 1.3\n" +
				"  - Child 2\n" +
				"    - Grand child 2.1\n" +
				"      - ABC\n",
			want: "- # Header\n" +
				"  - Parent\n" +
				"    - Child 1\n" +
				"      - Grand child 1.1\n" +
				"      - Grand child 1.2\n" +
				"      - Grand child 1.3\n" +
				"    - Child 2\n" +
				"      - Grand child 2.1\n" +
				"        - ABC\n",
		},
		{
			name: "test_thematic_break_setext_heading_with_frontmatter",
			input: "---\n" +
				"date: 2021-10-29T09:41:12.490Z\n" +
				"dateCreated: 2021-10-14T20:48:58.837Z\n" +
				"---\n\n" +
				"# Some title\n\n" +
				"Line1\n" +
				"Line2\n",
			want: "---\n" +
				"date: 2021-10-29T09:41:12.490Z\n" +
				"dateCreated: 2021-10-14T20:48:58.837Z\n" +
				"---\n" +
				"- # Some title\n" +
				"  - Line1\n" +
				"  - Line2\n",
		},
		{
			name: "test_ordered_lists",
			input: "# Header\n\n" +
				"1. First\n" +
				"2. Second\n",
			want: "- # Header\n" +
				"  - First\n" +
				"  - Second\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.FlatMarkdownToOutline(tc.input, internal.OutlineOptions{})
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFlatMarkdownToOutlineIdempotency(t *testing.T) {
	idempotentInputs := []struct {
		name  string
		input string
	}{
		{
			name: "links_output",
			input: "- # Header\n" +
				"  - [Link only](https://example.com)\n" +
				"  - Text before, then [a link](https://link.com), then text after\n" +
				"  - Only text before, [link a the end](https://endlink.com)\n",
		},
		{
			name: "flat_paragraphs_without_header_output",
			input: "- Some flat paragraph.\n" +
				"- [Link only](https://example.com).\n" +
				"- Text before, then [a link](https://link.com), then text after.\n" +
				"- Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "flat_paragraphs_with_header_output",
			input: "- # Some sneaky header\n" +
				"  - Some flat paragraph.\n" +
				"  - [Link only](https://example.com).\n" +
				"  - Text before, then [a link](https://link.com), then text after.\n" +
				"  - Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "nested_lists_single_level_output",
			input: "- # Header\n" +
				"  - Parent\n" +
				"    - Child 1\n" +
				"    - Child 2\n",
		},
		{
			name: "golden_output",
			input: "- # Header 1\n" +
				"  - Item 1\n" +
				"  - Item 2\n" +
				"  - ## Header 2\n" +
				"    - Item 3\n" +
				"    - ### Header 3\n" +
				"      - Item 4\n",
		},
	}

	for _, tc := range idempotentInputs {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.FlatMarkdownToOutline(tc.input, internal.OutlineOptions{})
			assert.Equal(t, tc.input, got, "idempotency: second pass must return identical output")
		})
	}
}
