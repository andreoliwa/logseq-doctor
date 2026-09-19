package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTidyUpPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
[tidy-up.forbidden]
page_references = ["quick capture", "inbox"]
url_substrings = ["utm_source"]
text_substrings = ["📍"]
`), 0o600))

	policy, err := config.LoadTidyUpPolicy(path)

	require.NoError(t, err)
	assert.Equal(t, []string{"quick capture", "inbox"}, policy.PageReferences)
	assert.Equal(t, []string{"utm_source"}, policy.URLSubstrings)
	assert.Equal(t, []string{"📍"}, policy.TextSubstrings)
}

func TestLoadTidyUpPolicyMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	_, err := config.LoadTidyUpPolicy(path)

	require.ErrorContains(t, err, "tidy-up config file not found")
	require.ErrorContains(t, err, path)
}

func TestLoadTidyUpPolicyInvalidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte("[tidy-up.forbidden\n"), 0o600))

	_, err := config.LoadTidyUpPolicy(path)

	require.Error(t, err)
	assert.ErrorContains(t, err, "read tidy-up config file")
}

func TestTidyUpConfigPath(t *testing.T) {
	path, err := config.TidyUpConfigPath()

	require.NoError(t, err)

	actual := filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path))
	assert.Equal(t, filepath.Join("lqd", "config.toml"), actual)
}
