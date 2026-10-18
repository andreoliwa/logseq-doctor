package dashboard

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/stretchr/testify/require"
)

func TestCollectBlocksToMoveSkipsPagePrefix(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "logseq"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "pages"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0o600))

	source := "((11111111-1111-1111-1111-111111111111))\n\n- bullet\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "pages/Example.md"), []byte(source), 0o600))

	graph, err := logseq.Open(context.Background(), root, logseq.WithPreserveOutput())
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, graph.Close()) })

	page, err := graph.OpenPage("Example")
	require.NoError(t, err)
	require.True(t, page.Blocks()[0].IsPreBlock())

	selected := collectBlocksToMove(page, map[string]bool{"11111111-1111-1111-1111-111111111111": true})
	require.Empty(t, selected)
}
