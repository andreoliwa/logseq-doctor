package internal_test

import (
	"context"
	"github.com/andreoliwa/lsd/internal"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andreoliwa/logseq-go"
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

	ctx := context.Background()

	graph, err := logseq.Open(ctx, filepath.Join("testdata", "graph"))
	if err != nil {
		t.Fatal(err)
	}

	page, err := graph.OpenPage(name)
	if err != nil {
		t.Fatal(err)
	}

	return page
}

type resultSetupFileContents struct {
	oldContents string
	goldenPath  string
}

func setupFileContents(t *testing.T, name string) resultSetupFileContents {
	t.Helper()

	subdir := filepath.Join("graph", "pages", name+".md")
	path := filepath.Join("testdata", subdir)

	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return resultSetupFileContents{string(bytes), subdir + ".golden"}
}

func TestCheckForbiddenReferences(t *testing.T) {
	invalid := setupPage(t, "forbidden")
	assert.Equal(t, internal.ChangedPage{"remove 4 forbidden references to pages/tags: Inbox, quick capture", false},
		internal.CheckForbiddenReferences(invalid))

	valid := setupPage(t, "valid")
	assert.Equal(t, internal.ChangedPage{"", false}, internal.CheckForbiddenReferences(valid))
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
