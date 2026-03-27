package logseqext_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/stretchr/testify/assert"
)

func TestCleanTaskName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		marker   string
		expected string
	}{
		{"strip marker", "TODO Buy groceries", "TODO", "Buy groceries"},
		{"strip marker DOING", "DOING Fix the bug", "DOING", "Fix the bug"},
		{"multiline strips first line only", "TODO First line\nid:: abc", "TODO", "First line"},
		{"strip time prefix", "TODO **9:00** Morning standup", "TODO", "Morning standup"},
		{"strip unbolded time prefix", "TODO 9:00 Morning standup", "TODO", "Morning standup"},
		{"no marker", "Buy groceries", "", "Buy groceries"},
		{"trailing whitespace trimmed", "TODO  Spaces  ", "TODO", "Spaces"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := logseqext.CleanTaskName(test.content, test.marker)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty", nil, nil},
		{"no duplicates", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"with duplicates", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"all duplicates", []string{"x", "x", "x"}, []string{"x"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := logseqext.UniqueStrings(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestNormalizeTagPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"already prefixed", []string{"#travel"}, []string{"#travel"}},
		{"no prefix added", []string{"travel"}, []string{"#travel"}},
		{"uppercase lowercased", []string{"#Travel", "Fun"}, []string{"#travel", "#fun"}},
		{"accented chars removed", []string{"#café"}, []string{"#cafe"}},
		{"hyphens stripped", []string{"#meal-prep"}, []string{"#mealprep"}},
		{"spaces stripped", []string{"#tag with spaces"}, []string{"#tagwithspaces"}},
		{"mixed", []string{"travel", "#fun", "Héro"}, []string{"#travel", "#fun", "#hero"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logseqext.NormalizeTagPrefixes(test.input)
			assert.Equal(t, test.expected, test.input)
		})
	}
}

func TestExtractDirectTags_PageRefs(t *testing.T) {
	result := logseqext.ExtractDirectTags("TODO [[grocery shopping]] for [[home]]")
	assert.Equal(t, []string{"grocery shopping", "home"}, result)
}

func TestExtractDirectTags_BracketHashtags(t *testing.T) {
	result := logseqext.ExtractDirectTags("TODO #[[tag with spaces]]")
	assert.Equal(t, []string{"tag with spaces"}, result)
}

func TestExtractDirectTags_MarkdownLinkURLIgnored(t *testing.T) {
	// Hash in URL should not be extracted as a tag.
	result := logseqext.ExtractDirectTags("TODO Check [this link](https://example.com#section)")
	assert.Equal(t, []string{}, result)
}

func TestExtractDirectTags_Deduplication(t *testing.T) {
	result := logseqext.ExtractDirectTags("TODO [[tag]] #tag")
	assert.Equal(t, []string{"tag"}, result)
}

func TestExtractDirectTags_EmptyString(t *testing.T) {
	result := logseqext.ExtractDirectTags("")
	assert.Nil(t, result)
}
