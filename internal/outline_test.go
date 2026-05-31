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
				"\t- [Link only](https://example.com)\n" +
				"\t- Text before, then [a link](https://link.com), then text after\n" +
				"\t- Only text before, [link a the end](https://endlink.com)\n",
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
				"\t- Some flat paragraph.\n" +
				"\t- [Link only](https://example.com).\n" +
				"\t- Text before, then [a link](https://link.com), then text after.\n" +
				"\t- Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "test_flat_paragraphs_with_deeper_headers",
			input: "## Some sneaky h2 without h1\n" +
				"Some flat paragraph.\n\n" +
				"[Link only](https://example.com).\n" +
				"Text before, then [a link](https://link.com), then text after.\n\n" +
				"Only text before, [link a the end](https://endlink.com).\n",
			want: "\t- ## Some sneaky h2 without h1\n" +
				"\t\t- Some flat paragraph.\n" +
				"\t\t- [Link only](https://example.com).\n" +
				"\t\t- Text before, then [a link](https://link.com), then text after.\n" +
				"\t\t- Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "test_nested_lists_single_level",
			input: "# Header\n\n" +
				"- Parent\n" +
				"  - Child 1\n" +
				"  - Child 2\n",
			want: "- # Header\n" +
				"\t- Parent\n" +
				"\t\t- Child 1\n" +
				"\t\t- Child 2\n",
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
				"\t- Parent\n" +
				"\t\t- Child 1\n" +
				"\t\t\t- Grand child 1.1\n" +
				"\t\t\t- Grand child 1.2\n" +
				"\t\t\t- Grand child 1.3\n" +
				"\t\t- Child 2\n" +
				"\t\t\t- Grand child 2.1\n" +
				"\t\t\t\t- ABC\n",
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
				"\t- Line1\n" +
				"\t- Line2\n",
		},
		{
			name: "test_ordered_lists",
			input: "# Header\n\n" +
				"1. First\n" +
				"2. Second\n",
			want: "- # Header\n" +
				"\t- First\n" +
				"\t  logseq.order-list-type:: number\n" +
				"\t- Second\n" +
				"\t  logseq.order-list-type:: number\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.FlatMarkdownToOutline(tc.input, internal.OutlineOptions{})
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFlatMarkdownToOutlineGoldenIdempotency(t *testing.T) {
	goldenBytes, err := os.ReadFile(filepath.Join("testdata", "outline", "dirty.md.golden"))
	require.NoError(t, err)

	input := string(goldenBytes)
	result := internal.FlatMarkdownToOutline(input, internal.OutlineOptions{})

	assert.Equal(t, input, result, "idempotency: converting dirty.md.golden again must produce identical output")
}

func TestFlatMarkdownToOutlineIdempotency(t *testing.T) {
	idempotentInputs := []struct {
		name  string
		input string
	}{
		{
			name: "links_output",
			input: "- # Header\n" +
				"\t- [Link only](https://example.com)\n" +
				"\t- Text before, then [a link](https://link.com), then text after\n" +
				"\t- Only text before, [link a the end](https://endlink.com)\n",
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
				"\t- Some flat paragraph.\n" +
				"\t- [Link only](https://example.com).\n" +
				"\t- Text before, then [a link](https://link.com), then text after.\n" +
				"\t- Only text before, [link a the end](https://endlink.com).\n",
		},
		{
			name: "nested_lists_single_level_output",
			input: "- # Header\n" +
				"\t- Parent\n" +
				"\t\t- Child 1\n" +
				"\t\t- Child 2\n",
		},
		{
			name: "golden_output",
			input: "- # Header 1\n" +
				"\t- Item 1\n" +
				"\t- Item 2\n" +
				"\t- ## Header 2\n" +
				"\t\t- Item 3\n" +
				"\t\t- ### Header 3\n" +
				"\t\t\t- Item 4\n" +
				"\t- ## Ordered Section\n" +
				"\t\t- First ordered\n" +
				"\t\t  logseq.order-list-type:: number\n" +
				"\t\t- Second ordered\n" +
				"\t\t  logseq.order-list-type:: number\n" +
				"\t\t- Third ordered\n" +
				"\t\t  logseq.order-list-type:: number\n",
		},
		{
			name: "ordered_list_output",
			input: "- # Header\n" +
				"\t- First\n" +
				"\t  logseq.order-list-type:: number\n" +
				"\t- Second\n" +
				"\t  logseq.order-list-type:: number\n",
		},
	}

	for _, tc := range idempotentInputs {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.FlatMarkdownToOutline(tc.input, internal.OutlineOptions{})
			assert.Equal(t, tc.input, got, "idempotency: second pass must return identical output")
		})
	}
}
