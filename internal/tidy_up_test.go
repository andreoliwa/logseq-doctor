package internal_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"
	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/config"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/stretchr/testify/require"

	"github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"

	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
)

func TestSortAndRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty slice", []string{}, []string{}},
		{"one element", []string{"apple"}, []string{"apple"}},
		{"duplicates", []string{"orange", "apple", "banana", "apple"}, []string{"apple", "banana", "orange"}},
		{"sorted unique", []string{"orange", "banana", "apple"}, []string{"apple", "banana", "orange"}},
		{
			"unsorted with duplicates",
			[]string{"orange", "banana", "apple", "apple", "orange"},
			[]string{"apple", "banana", "orange"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := internal.SortAndRemoveDuplicates(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func setupPage(t *testing.T, name string) logseq.Page {
	t.Helper()

	graph := testutils.NewStubGraph(t, "stub-graph")
	page := logseqapi.OpenPage(graph, name)

	return page
}

type resultSetupFileContents struct {
	oldContents string
	goldenPath  string
}

func setupFileContents(t *testing.T, name string) resultSetupFileContents {
	t.Helper()

	subdir := filepath.Join("stub-graph", "pages", name+".md")
	path := filepath.Join("testdata", subdir)

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)

	return resultSetupFileContents{string(bytes), subdir + ".golden"}
}

func TestCheckForbiddenReferences(t *testing.T) {
	policy := config.ForbiddenContentPolicy{
		PageReferences: []string{"quick capture", "inbox"},
		URLSubstrings:  []string{"utm_source"},
		TextSubstrings: []string{"📍"},
	}
	invalid := setupPage(t, "forbidden")
	assert.Equal(t, internal.ChangedPage{"remove 4 forbidden references to pages/tags: Inbox, quick capture", false},
		internal.CheckForbiddenReferences(invalid, policy))

	valid := setupPage(t, "valid")
	assert.Equal(t, internal.ChangedPage{"", false}, internal.CheckForbiddenReferences(valid, policy))
}

func TestCheckForbiddenReferencesUsesConfiguredCategories(t *testing.T) {
	page := setupPage(t, "forbidden")
	policy := config.ForbiddenContentPolicy{
		PageReferences: []string{"inbox"},
		URLSubstrings:  []string{"example.com"},
		TextSubstrings: []string{"pin"},
	}

	assert.Equal(t, internal.ChangedPage{"remove 2 forbidden references to pages/tags: Inbox", false},
		internal.CheckForbiddenReferences(page, policy))
}

func TestCheckForbiddenReferencesMatchesURLsAndText(t *testing.T) {
	page := setupPage(t, "forbidden-categories")
	policy := config.ForbiddenContentPolicy{
		URLSubstrings:  []string{"utm_source"},
		TextSubstrings: []string{"📍"},
	}

	result := internal.CheckForbiddenReferences(page, policy)

	assert.False(t, result.Changed)
	assert.Contains(t, result.Msg, "https://example.com/?utm_source=newsletter")
	assert.Contains(t, result.Msg, "Pinned 📍")
}

func TestCheckRunningTasks(t *testing.T) {
	invalid := setupPage(t, "running")
	assert.Equal(t, internal.ChangedPage{"stop 2 running task(s): DOING, IN-PROGRESS", false},
		internal.CheckRunningTasks(invalid))

	valid := setupPage(t, "valid")
	assert.Equal(t, internal.ChangedPage{"", false}, internal.CheckRunningTasks(valid))
}

func TestRemoveDoubleSpaces(t *testing.T) {
	invalid := setupPage(t, "spaces")
	assert.Equal(t,
		internal.ChangedPage{"4 double spaces fixed: 'Link   With  Spaces  ', 'Regular   text with  spaces'," +
			" 'some  page   title  with  spaces', 'some  tag with   spaces'", true},
		internal.RemoveDoubleSpaces(invalid))

	// TODO: compare the saved Markdown file with the golden file
	//  I tested manually, and it works, but I need to do something like:
	// actual := os.ReadFile(invalid.Path())
	// expected := os.ReadFile(invalid.Path() + ".golden")
	// assert.Equal(t, expected, actual)

	valid := setupPage(t, "valid")
	assert.Equal(t, internal.ChangedPage{"", false}, internal.RemoveDoubleSpaces(valid))
}

func TestRemoveUnnecessaryBracketsFromTags(t *testing.T) {
	invalid := setupFileContents(t, "tag-brackets")
	changed := internal.RemoveUnnecessaryBracketsFromTags(invalid.oldContents)
	assert.Equal(t, "unnecessary tag brackets removed", changed.Msg)
	golden.Assert(t, changed.NewContents, invalid.goldenPath)

	valid := setupFileContents(t, "valid")
	assert.Equal(t, internal.ChangedContents{"", ""}, internal.RemoveUnnecessaryBracketsFromTags(valid.oldContents))
}

func TestRemoveEmptyBullets(t *testing.T) {
	invalid := setupPage(t, "empty-bullets")
	assert.Equal(t,
		internal.ChangedPage{"6 empty bullets removed", true},
		internal.RemoveEmptyBullets(invalid))

	// TODO: compare the saved Markdown file with the golden file. I tested manually and it works

	valid := setupPage(t, "valid")
	assert.Equal(t, internal.ChangedPage{"", false}, internal.RemoveEmptyBullets(valid))
}

func TestNormalizeLeadingTimes(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	transaction := graph.NewTransaction()
	page, err := transaction.OpenPage("timecodes")
	require.NoError(t, err)
	root, err := logseq.ParseBlock("- **9h00 Meeting**\n" +
		"  id:: 12345678-1234-1234-1234-123456789abc\n" +
		"  custom:: keep\n" +
		"  10h00 continuation\n" +
		"  - TODO [#A] *10;05* Follow-up\n" +
		"- 25:00 Invalid\n" +
		"- Meet at 9h00\n" +
		"- `9h00` Code\n" +
		"- [9h00](url)")
	require.NoError(t, err)

	for _, block := range root.Blocks() {
		page.AddBlock(block)
	}

	result := internal.NormalizeLeadingTimes(page)
	require.True(t, result.Changed)
	require.Equal(t, "2 leading times normalized", result.Msg)

	first, err := logseq.AsString(page.Blocks()[0])
	require.NoError(t, err)
	require.Contains(t, first, "09:00 **Meeting**")
	require.Contains(t, first, "10h00 continuation")
	require.Contains(t, first, content.TaskStringTodo+" [#A] 10:05 Follow-up")
	require.Contains(t, first, "custom:: keep")
	require.Equal(t, internal.ChangedPage{}, internal.NormalizeLeadingTimes(page))
}

func TestTidyUpTimeNormalizationSavesOnlyChanges(t *testing.T) {
	graph := testutils.NewStubGraph(t, "stub-graph")
	transaction := graph.NewTransaction()
	page, err := transaction.OpenPage("timecodes")
	require.NoError(t, err)
	block, err := logseq.ParseBlock("9h00 Meeting\n10h00 continuation")
	require.NoError(t, err)
	page.AddBlock(block)
	require.NoError(t, transaction.Save())

	journalPath, err := graph.JournalPath(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	path := filepath.Join(filepath.Dir(filepath.Dir(journalPath)), "pages", "timecodes.md")
	require.Equal(t, 1, internal.TidyUpOneFile(graph, path, config.ForbiddenContentPolicy{}))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "- 09:00 Meeting\n  10h00 continuation", string(data))
	// An old timestamp makes an unwanted second save observable without sleeps.
	old := time.Unix(1000000000, 0)
	require.NoError(t, os.Chtimes(path, old, old))
	require.Equal(t, 0, internal.TidyUpOneFile(graph, path, config.ForbiddenContentPolicy{}))
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, old, info.ModTime())
}
