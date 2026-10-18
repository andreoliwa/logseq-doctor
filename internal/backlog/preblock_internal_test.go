package backlog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreoliwa/logseq-doctor/pkg/set"
	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/stretchr/testify/require"
)

func backlogPrefixGraph(t *testing.T, prefix string) *logseq.Graph {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "logseq"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "pages"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "pages/Example.md"), []byte(prefix+"\n\n- bullet\n"), 0o600))

	graph, err := logseq.Open(context.Background(), root, logseq.WithPreserveOutput())
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, graph.Close()) })

	return graph
}

func TestProcessAllBlocksSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "metadata")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	state := newPageState()

	empty := set.NewSet[string]()
	processAllBlocks(page, state, empty, empty, empty)
	require.Same(t, page.Blocks()[1], state.firstBlock)
}

func TestBlockRefsFromPagesSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "((11111111-1111-1111-1111-111111111111))")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	require.Zero(t, blockRefsFromPages(page).Size())
}

func TestCollectScheduledRefsSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "Scheduled ((11111111-1111-1111-1111-111111111111))")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	state := newPageState()
	collectScheduledRefs(page, state)
	require.Zero(t, state.scheduledBlockRefs.Size())
}

func TestCollectTriagedRefsSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "Triaged ((11111111-1111-1111-1111-111111111111))")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	state := newPageState()
	collectTriagedRefs(page, state)
	require.Nil(t, state.dividerTriaged)
	require.Zero(t, state.triagedBlockRefs.Size())
}

func TestNormalizeHeaderTextSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "overdue")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	require.False(t, NormalizeHeaderText(page))
}

func TestFindFirstSectionDividerSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "overdue")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	require.Nil(t, FindFirstSectionDivider(page))
}

func TestRemoveBlockRefFromRegularAreaSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "((11111111-1111-1111-1111-111111111111))")
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	RemoveBlockRefFromRegularArea(page, "11111111-1111-1111-1111-111111111111")
	require.Len(t, page.Blocks(), 2)
	require.True(t, page.Blocks()[0].IsPreBlock())
}

func TestReadConfigSkipsPagePrefix(t *testing.T) {
	graph := backlogPrefixGraph(t, "[[Metadata]]")
	reader := NewPageConfigReader(graph, "Example")
	configuration, err := reader.ReadConfig()
	require.NoError(t, err)
	require.Empty(t, configuration.Backlogs)
}
