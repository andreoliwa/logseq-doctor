package testutils

import (
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/backlog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// StubGraph opens the example graph under "testdata" for testing.
func StubGraph(t *testing.T) *logseq.Graph {
	t.Helper()

	dir, err := filepath.Abs(filepath.Join("testdata", "stub-graph"))
	require.NoError(t, err)

	tempDir := fs.NewDir(t, "append-raw",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(dir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(dir, "journals"))),
		fs.WithDir("pages", fs.FromDir(filepath.Join(dir, "pages"))),
	)

	return internal.OpenGraphFromPath(tempDir.Path())
}

type mockLogseqAPI struct {
	mock.Mock
	t         *testing.T
	responses *StubAPIResponses
}

func newMockLogseqAPI(t *testing.T, responses StubAPIResponses) *mockLogseqAPI {
	t.Helper()

	api := mockLogseqAPI{t: t, responses: &responses} //nolint:exhaustruct
	api.On("PostQuery", mock.Anything).Return("{}", nil)

	return &api
}

type StubAPIResponses struct {
	Queries []QueryArg
}
type QueryArg struct {
	Contains string
}

func (m *mockLogseqAPI) PostQuery(query string) (string, error) {
	args := m.Called(query)

	// Return predefined test data based on query content.
	for _, q := range m.responses.Queries {
		if strings.Contains(query, q.Contains) {
			return stubJSONResponse(m.t, q.Contains)
		}
	}

	return args.String(0), args.Error(1)
}

func stubJSONResponse(t *testing.T, basename string) (string, error) {
	t.Helper()

	path, err := filepath.Abs(filepath.Join("testdata", "stub-api", basename+".json"))
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data), nil
}

func StubBacklog(t *testing.T, rootPage string, apiResponses *StubAPIResponses) backlog.Backlog {
	t.Helper()

	graph := StubGraph(t)
	api := newMockLogseqAPI(t, *apiResponses)
	reader := backlog.NewPageConfigReader(graph, rootPage)

	return backlog.NewBacklog(graph, api, reader)
}
