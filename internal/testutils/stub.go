package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/backlog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
)

// StubGraph opens the example graph under "testdata" for testing.
func StubGraph(t *testing.T, caseDirName string) *logseq.Graph {
	t.Helper()

	stubGraphDir, err := filepath.Abs(filepath.Join("testdata", "stub-graph"))
	require.NoError(t, err)

	pagesOps := []fs.PathOp{fs.FromDir(filepath.Join(stubGraphDir, "pages"))}

	if caseDirName != "" {
		caseDir, err := filepath.Abs(filepath.Join(stubGraphDir, "pages-cases", caseDirName))
		require.NoError(t, err)

		pagesOps = append(pagesOps, fs.FromDir(caseDir))
	}

	tempDir := fs.NewDir(t, "append-raw",
		fs.WithDir("logseq", fs.FromDir(filepath.Join(stubGraphDir, "logseq"))),
		fs.WithDir("journals", fs.FromDir(filepath.Join(stubGraphDir, "journals"))),
		fs.WithDir("pages", pagesOps...),
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

	path, err := filepath.Abs(filepath.Join("testdata", "stub-api", basename+".jsonl"))
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data), nil
}

func StubBacklog(t *testing.T, configPage, caseDirName string, apiResponses *StubAPIResponses) backlog.Backlog {
	t.Helper()

	graph := StubGraph(t, caseDirName)
	api := newMockLogseqAPI(t, *apiResponses)
	reader := backlog.NewPageConfigReader(graph, configPage)

	return backlog.NewBacklog(graph, api, reader)
}
