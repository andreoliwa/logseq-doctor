package testutils

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
)

// NewStubGraph creates a test graph using the new directory structure.
// It uses graph-template as the base and loads test data
// from testdata/{subDir}/journals and testdata/{subDir}/pages.
func NewStubGraph(t *testing.T, subDir string) *logseq.Graph {
	t.Helper()

	graphTemplateDir, err := filepath.Abs(filepath.Join("testdata", "graph-template"))
	require.NoError(t, err)

	path, err := filepath.Abs(filepath.Join("testdata", subDir))
	require.NoError(t, err)

	tempDir := fs.NewDir(t, "test-graph",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(graphTemplateDir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(path, "journals"))),
		fs.WithDir("pages", fs.FromDir(filepath.Join(path, "pages"))),
	)

	return logseqapi.OpenGraphFromPath(tempDir.Path())
}

type mockLogseqAPI struct {
	mock.Mock

	tagResponses map[string]string
}

// newMockLogseqAPIFromMap creates a mockLogseqAPI that returns pre-built JSON responses keyed by tag.
func newMockLogseqAPIFromMap(t *testing.T, responses map[string]string) *mockLogseqAPI {
	t.Helper()

	api := mockLogseqAPI{tagResponses: responses} //nolint:exhaustruct
	api.On("PostQuery", mock.Anything).Return("{}", nil)
	api.On("PostDatascriptQuery", mock.Anything).Return("[]", nil)

	return &api
}

func (m *mockLogseqAPI) PostQuery(query string) (string, error) {
	args := m.Called(query)

	for tag, resp := range m.tagResponses {
		if strings.Contains(query, tag) {
			return resp, nil
		}
	}

	return args.String(0), args.Error(1)
}

func (m *mockLogseqAPI) PostDatascriptQuery(query string) (string, error) {
	args := m.Called(query)

	return args.String(0), args.Error(1)
}

func (m *mockLogseqAPI) UpsertBlockProperty(_ string, _ string, _ string) error {
	return nil
}

var testStartTime = time.Now()                                   //nolint:gochecknoglobals
var baselineTime = time.Date(2025, 4, 13, 3, 33, 0, 0, time.UTC) //nolint:gochecknoglobals

func RelativeTime() time.Time {
	elapsed := time.Since(testStartTime)

	return baselineTime.Add(elapsed)
}
