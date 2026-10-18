package buildhooks_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type buildFixture struct {
	checkout    string
	noticeLog   string
	goLog       string
	environment []string
}

func newBuildFixture(t *testing.T, scenario string) buildFixture {
	t.Helper()

	root := t.TempDir()
	checkout := filepath.Join(root, "doctor with spaces")
	library := filepath.Join(root, "library with spaces")

	if scenario == "default sibling" {
		library = filepath.Join(root, "logseq-go")
	}

	require.NoError(t, os.MkdirAll(checkout, 0o700))
	require.NoError(t, os.MkdirAll(library, 0o700))

	source, err := os.ReadFile("../../Makefile")
	require.NoError(t, err)
	require.Contains(t, string(source), "install: build")

	directory, err := os.OpenRoot(checkout)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, directory.Close()) })

	require.NoError(t, directory.WriteFile("Makefile", source, 0o600))

	counter := filepath.Join(root, "notice-calls")
	goCounter := filepath.Join(root, "go-calls")

	libraryMakefile := "check-upstream:\n\t@printf 'notice\\n' >> \"$(NOTICE_LOG)\"\n"
	if scenario == "failure" {
		libraryMakefile += "\t@exit 1\n"
	}

	require.NoError(t, os.WriteFile(filepath.Join(library, "Makefile"), []byte(libraryMakefile), 0o600))

	if scenario == "missing" {
		library += "-missing"
	}

	bin := filepath.Join(root, "bin")
	require.NoError(t, os.MkdirAll(bin, 0o700))
	// The system shell reads private fixture scripts instead of executing generated files.
	require.NoError(t, os.Symlink("/bin/sh", filepath.Join(bin, "go")))

	for _, verb := range []string{"mod", "build"} {
		script := "printf '" + verb + " %s\\n' \"$*\" >> \"$GO_LOG\"\n"
		require.NoError(t, directory.WriteFile(verb, []byte(script), 0o600))
	}

	environment := []string{
		"PATH=" + bin + ":/usr/bin:/bin", "HOME=" + root,
		"NOTICE_LOG=" + counter, "GO_LOG=" + goCounter, "CI=",
	}
	if scenario != "default sibling" {
		environment = append(environment, "LOGSEQ_GO_DIR="+library)
	}

	if scenario == "ci" {
		environment = append(environment, "CI=1")
	}

	return buildFixture{checkout: checkout, noticeLog: counter, goLog: goCounter, environment: environment}
}

func runBuildFixture(t *testing.T, scenario string, fixture buildFixture) {
	t.Helper()

	for range 2 {
		command := exec.CommandContext(t.Context(), "make", "-s", "build")
		command.Dir = fixture.checkout
		command.Env = fixture.environment
		output, err := command.CombinedOutput()
		require.NoError(t, err, string(output))

		if scenario == "missing" || scenario == "failure" {
			require.Contains(t, string(output), "warning:")
		}
	}

	notices, err := os.ReadFile(fixture.noticeLog)

	if scenario == "missing" || scenario == "ci" {
		require.True(t, os.IsNotExist(err))
	} else {
		require.NoError(t, err)
		require.Equal(t, 2, strings.Count(string(notices), "notice"))
	}

	builds, err := os.ReadFile(fixture.goLog)
	require.NoError(t, err)
	require.Equal(t, 2, strings.Count(string(builds), "mod tidy"))
	require.Contains(t, string(builds), "build ./cmd/lqd")
}

func TestBuildChecksUpstreamWithoutBlocking(t *testing.T) {
	for _, scenario := range []string{"normal", "missing", "failure", "ci", "default sibling"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := newBuildFixture(t, scenario)
			runBuildFixture(t, scenario, fixture)
		})
	}
}
