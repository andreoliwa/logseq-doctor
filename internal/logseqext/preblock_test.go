package logseqext_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/stretchr/testify/require"
)

func preblockPage(t *testing.T, prefix, body string) (*logseq.Graph, logseq.Page) {
	t.Helper()

	root, pages := newTestGraph(t)
	err := os.WriteFile(filepath.Join(pages, "Example.md"), []byte(prefix+"\n\n"+body), 0o600)
	require.NoError(t, err)

	graph := openTestGraph(t, root)
	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	return graph, page
}

func TestFindBlockByIDPropertySkipsPagePrefix(t *testing.T) {
	_, page := preblockPage(t, "id:: page-id", "- TODO bullet\n  id:: task-id\n")
	require.Nil(t, logseqext.FindBlockByIDProperty(page, "page-id"))
	require.NotNil(t, logseqext.FindBlockByIDProperty(page, "task-id"))
}

func TestFindBlockContainingTextSkipsPagePrefix(t *testing.T) {
	graph, page := preblockPage(t, "TODO metadata-only", "- TODO bullet\n  id:: task-id\n")
	require.Nil(t, logseqext.FindBlockContainingText(page, "metadata-only"))

	bullet := logseqext.FindBlockContainingText(page, "bullet")
	require.NotNil(t, bullet)

	rendered, err := graph.AsString(bullet)
	require.NoError(t, err)
	require.Contains(t, rendered, "TODO bullet")
}

func TestFindBlockByKeySkipsPagePrefix(t *testing.T) {
	_, page := preblockPage(t, "TODO metadata-only", "- TODO bullet\n  id:: task-id\n")
	require.Nil(t, logseqext.FindBlockByKey(page, nil, "metadata-only"))
	require.NotNil(t, logseqext.FindBlockByKey(page, nil, "bullet"))
}

func TestFindTaskMarkerByKeySkipsPagePrefix(t *testing.T) {
	_, page := preblockPage(t, "TODO metadata-only", "- TODO bullet\n")
	require.Nil(t, logseqext.FindTaskMarkerByKey(page, nil, "metadata-only"))
	require.NotNil(t, logseqext.FindTaskMarkerByKey(page, nil, "bullet"))
}

func TestFindFirstQuerySkipsPagePrefix(t *testing.T) {
	graph, _ := preblockPage(t, "{{query (page-property tags metadata)}}", "- TODO bullet\n  id:: task-id\n")
	finder := logseqext.NewLogseqFinder(graph)
	require.Empty(t, finder.FindFirstQuery("Example"))
}

func TestExtractBlockRefUUIDsSkipsPagePrefix(t *testing.T) {
	_, page := preblockPage(t, "((11111111-1111-1111-1111-111111111111))", "- TODO bullet\n  id:: task-id\n")
	require.Empty(t, logseqext.ExtractBlockRefUUIDs(page))
}

func TestExtractSectionedBlockRefUUIDsSkipsPagePrefix(t *testing.T) {
	_, page := preblockPage(t, "((11111111-1111-1111-1111-111111111111))", "- TODO bullet\n  id:: task-id\n")
	require.Empty(t, logseqext.ExtractSectionedBlockRefUUIDs(page, nil))
}

func TestLaterTimeAnchorSkipsPagePrefix(t *testing.T) {
	preblock := content.NewPreBlock(content.NewParagraph(content.NewText("12:00 metadata")))
	incoming, err := logseq.ParseBlock("10:00 incoming")
	require.NoError(t, err)

	bullet, err := logseq.ParseBlock("11:00 bullet")
	require.NoError(t, err)
	require.Same(t, bullet, logseqext.LaterTimeAnchor(incoming, content.BlockList{preblock, bullet}))
}

func TestCollectDoingTasksSkipsPagePrefix(t *testing.T) {
	graph, _ := preblockPage(t, "DOING metadata-only", "- DOING bullet\n")
	path := filepath.Join(graph.Directory(), "pages/Example.md")
	count, err := logseqext.SyncDoingTasksInFiles(graph, []string{path})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, strings.SplitN(string(data), "- DOING bullet", 2)[0], "id::")
	require.Contains(t, string(data), "- DOING bullet\n  id::")
}
