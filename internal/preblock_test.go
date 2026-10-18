package internal_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-doctor/internal/config"
	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/stretchr/testify/require"
)

func tidyPrefixPage(t *testing.T, prefix string) logseq.Page {
	t.Helper()

	graph, _ := testutils.NewMarkdownGraph(t, prefix+"\n\n- bullet\n")
	page, err := graph.OpenPage("timeline")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	return page
}

func TestCheckForbiddenReferencesSkipsPagePrefix(t *testing.T) {
	page := tidyPrefixPage(t, "[[Blocked]] metadata")

	policy := config.ForbiddenContentPolicy{PageReferences: []string{"Blocked"}, URLSubstrings: nil, TextSubstrings: nil}
	result := internal.CheckForbiddenReferences(page, policy)
	require.Empty(t, result.Msg)
}

func TestCheckRunningTasksSkipsPagePrefix(t *testing.T) {
	page := tidyPrefixPage(t, "DOING metadata")
	result := internal.CheckRunningTasks(page)
	require.Empty(t, result.Msg)
}

func TestRemoveDoubleSpacesSkipsPagePrefix(t *testing.T) {
	page := tidyPrefixPage(t, "metadata  only")
	result := internal.RemoveDoubleSpaces(page)
	require.False(t, result.Changed)
	require.Empty(t, result.Msg)
}

func TestRemoveEmptyBulletsSkipsPagePrefix(t *testing.T) {
	page := tidyPrefixPage(t, "metadata")
	prefix := page.Blocks()[0]

	for _, node := range prefix.Children() {
		node.RemoveSelf()
	}

	result := internal.RemoveEmptyBullets(page)
	require.False(t, result.Changed)
	require.True(t, page.Blocks()[0].IsPreBlock())
}

func TestNormalizeLeadingTimesSkipsPagePrefix(t *testing.T) {
	page := tidyPrefixPage(t, "9h00 metadata")
	result := internal.NormalizeLeadingTimes(page)
	require.False(t, result.Changed)
	require.Empty(t, result.Msg)
}
