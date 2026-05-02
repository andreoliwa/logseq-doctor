package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
)

// dirPerm is the permission used for directories created by the fixture framework.
const dirPerm = 0o755

// filePerm is the permission used for files created by the fixture framework.
const filePerm = 0o600

// Block is a generic Logseq block fixture.
// Most callers use the Task() constructor which sets Marker.
// The framework is built on Block internally so non-task blocks
// (plain content, property blocks) can be added in the future
// without architectural changes — just add a Block constructor alongside Task().
type Block struct {
	Slug       string
	Marker     string // content.TaskString* from logseq-go (e.g. TaskStringTodo), or "" for non-task blocks
	Text       string
	Tags       []string
	Scheduled  string            // relative date: "+3d", "-1w", "+2m", "+1y", "" for none
	Deadline   string            // same format as Scheduled
	Priority   string            // "A", "B", "C", or ""
	Groomed    string            // relative date for groomed:: property, "" means no property
	JournalDay string            // "2025-03-02" format; defaults to baseline date (2025-04-13)
	ExtraProps map[string]string // arbitrary Logseq block properties
}

// BlockOpt is a functional option for configuring a Block.
type BlockOpt func(*Block)

// Task constructs a Block with Marker set. Use BlockOpts for optional fields.
func Task(slug, status, text string, opts ...BlockOpt) Block {
	b := Block{Slug: slug, Marker: status, Text: text} //nolint:exhaustruct
	for _, o := range opts {
		o(&b)
	}

	return b
}

// WithTags sets the Tags field.
func WithTags(tags ...string) BlockOpt {
	return func(b *Block) { b.Tags = tags }
}

// WithScheduled sets the Scheduled relative date (e.g. "+3d", "-1w").
func WithScheduled(rel string) BlockOpt {
	return func(b *Block) { b.Scheduled = rel }
}

// WithDeadline sets the Deadline relative date.
func WithDeadline(rel string) BlockOpt {
	return func(b *Block) { b.Deadline = rel }
}

// WithPriority sets the Priority ("A", "B", or "C").
func WithPriority(p string) BlockOpt {
	return func(b *Block) { b.Priority = p }
}

// WithGroomed sets the Groomed relative date (e.g. "-7d").
func WithGroomed(rel string) BlockOpt {
	return func(b *Block) { b.Groomed = rel }
}

// WithJournalDay sets the journal page date ("2025-03-02").
func WithJournalDay(day string) BlockOpt {
	return func(b *Block) { b.JournalDay = day }
}

// WithExtraProps sets arbitrary Logseq block properties.
func WithExtraProps(props map[string]string) BlockOpt {
	return func(b *Block) { b.ExtraProps = props }
}

// TaskFixture holds block definitions and generates fake API responses for a test.
type TaskFixture struct {
	t          *testing.T
	blocks     []Block
	slugToUUID map[string]string
	uuidToSlug map[string]string
}

// NewFixture creates a TaskFixture from the given blocks.
func NewFixture(t *testing.T, blocks ...Block) *TaskFixture {
	t.Helper()

	f := &TaskFixture{t: t} //nolint:exhaustruct
	f.slugToUUID, f.uuidToSlug = buildSlugMap(blocks)
	f.blocks = blocks

	return f
}

// Add appends blocks to the fixture and returns it for chaining.
// Panics if any new slug collides with an existing one.
func (f *TaskFixture) Add(blocks ...Block) *TaskFixture {
	f.t.Helper()

	all := append(f.blocks, blocks...) //nolint:gocritic
	f.slugToUUID, f.uuidToSlug = buildSlugMap(all)
	f.blocks = all

	return f
}

// ExportFixtureUUID returns the UUID for the given slug in the fixture (for white-box testing).
func ExportFixtureUUID(f *TaskFixture, slug string) string {
	return f.slugToUUID[slug]
}

// FakeBacklog creates a backlog.Backlog backed by:
// - a temp graph populated from testdata/{caseDirName}/ with slugs expanded to UUIDs
// - a fake Logseq API returning generated JSON grouped by tag
//
// configPage is the backlog config page name (e.g. "bk", "ov").
// caseDirName is the subdirectory under testdata/ for this test case (may be empty).
func (f *TaskFixture) FakeBacklog(t *testing.T, configPage, caseDirName string) backlog.Backlog {
	t.Helper()

	graph := f.fakeGraph(t, caseDirName)
	api := f.fakeAPI(t)
	reader := backlog.NewPageConfigReader(graph, configPage)

	return backlog.NewBacklog(graph, api, reader, RelativeTime)
}

// AssertGoldenPages collapses UUIDs back to slugs in each output page, then
// compares against golden files in testdata/{caseDirName}/pages/.
func (f *TaskFixture) AssertGoldenPages(
	t *testing.T, graph *logseq.Graph, caseDirName string, pages []string,
) {
	t.Helper()
	f.assertGoldenFiles(t, graph, false, caseDirName, pages)
}

// AssertGoldenJournals is like AssertGoldenPages but for journals/.
func (f *TaskFixture) AssertGoldenJournals(
	t *testing.T, graph *logseq.Graph, caseDirName string, pages []string,
) {
	t.Helper()
	f.assertGoldenFiles(t, graph, true, caseDirName, pages)
}

func (f *TaskFixture) assertGoldenFiles(
	t *testing.T, graph *logseq.Graph, journals bool, caseDirName string, pages []string,
) {
	t.Helper()

	dir := "pages"
	if journals {
		dir = "journals"
	}

	for _, page := range pages {
		filename := page + ".md"

		data, err := os.ReadFile(filepath.Join(graph.Directory(), dir, filename))
		require.NoError(t, err)

		content := collapseSlugs(string(data), f.uuidToSlug)
		if !journals {
			content = strings.TrimRight(content, "\r\n") + "\r\n"
		}

		goldenPath := filepath.Join(caseDirName, dir, filename+".golden")
		golden.Assert(t, content, goldenPath)
	}
}

// fakeGraph builds a temp graph with slug-expanded .md files from testdata/{caseDirName}/.
// Uses the graph-template as the base and expands slugs in the page files.
func (f *TaskFixture) fakeGraph(t *testing.T, caseDirName string) *logseq.Graph {
	t.Helper()

	graphTemplateDir, err := filepath.Abs(filepath.Join("testdata", "graph-template"))
	require.NoError(t, err)

	tempDir := t.TempDir()
	logseqDir := filepath.Join(tempDir, "logseq")
	require.NoError(t, copyDirTree(filepath.Join(graphTemplateDir, "logseq"), logseqDir))

	pagesDir := filepath.Join(tempDir, "pages")
	require.NoError(t, os.MkdirAll(pagesDir, dirPerm))

	journalsDir := filepath.Join(tempDir, "journals")
	require.NoError(t, os.MkdirAll(journalsDir, dirPerm))

	if caseDirName != "" {
		casePath, err := filepath.Abs(filepath.Join("testdata", caseDirName))
		require.NoError(t, err)

		f.copyExpandedMDs(t, filepath.Join(casePath, "pages"), pagesDir)
		f.copyExpandedMDs(t, filepath.Join(casePath, "journals"), journalsDir)
	}

	return logseqapi.OpenGraphFromPath(tempDir)
}

// copyExpandedMDs copies .md files from src to dst, expanding slugs to UUIDs in each file.
// Silently skips if src does not exist.
func (f *TaskFixture) copyExpandedMDs(t *testing.T, src, dst string) {
	t.Helper()

	entries, err := os.ReadDir(src)
	if os.IsNotExist(err) {
		return
	}

	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		require.NoError(t, err)

		expanded := expandSlugs(string(data), f.slugToUUID)
		require.NoError(t, os.WriteFile(filepath.Join(dst, entry.Name()), []byte(expanded), filePerm)) //nolint:gosec
	}
}

// fakeAPI creates a mockLogseqAPI returning generated JSON grouped by tag.
func (f *TaskFixture) fakeAPI(t *testing.T) *mockLogseqAPI {
	t.Helper()

	byTag := make(map[string][]Block)

	for _, b := range f.blocks {
		for _, tag := range b.Tags {
			byTag[tag] = append(byTag[tag], b)
		}
	}

	now := RelativeTime()
	responses := make(map[string]string, len(byTag))

	for tag, blocks := range byTag {
		responses[tag] = buildAPIResponse(blocks, f.slugToUUID, now)
	}

	return newMockLogseqAPIFromMap(t, responses)
}

// copyDirTree copies a directory tree recursively from src to dst.
func copyDirTree(src, dst string) error {
	err := filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return fmt.Errorf("copyDirTree: %w", relErr)
		}

		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, dirPerm)
		}

		data, readErr := os.ReadFile(path) //nolint:gosec
		if readErr != nil {
			return fmt.Errorf("copyDirTree: %w", readErr)
		}

		return os.WriteFile(target, data, filePerm) //nolint:gosec
	})
	if err != nil {
		return fmt.Errorf("copyDirTree: walk %s: %w", src, err)
	}

	return nil
}
