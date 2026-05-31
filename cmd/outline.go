package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/spf13/cobra"
)

// ErrMutuallyExclusive is returned when --in-place and --move-to are both set.
var ErrMutuallyExclusive = errors.New("--move-to and --in-place cannot be used together")

// ErrDestinationExists is returned when the destination file already exists during --move-to.
var ErrDestinationExists = errors.New("destination already exists")

// OutlineDependencies holds injectable dependencies for the outline command.
type OutlineDependencies struct {
	Convert   func(input string, opts internal.OutlineOptions) string
	ReadFile  func(path string) (string, error)
	WriteFile func(path string, data string) error
	Stat      func(path string) (os.FileInfo, error)
	Rename    func(oldpath, newpath string) error
	Remove    func(path string) error
	Stdin     func() string
	Out       io.Writer
}

func fillOutlineDeps(deps *OutlineDependencies) {
	if deps.Convert == nil {
		deps.Convert = internal.FlatMarkdownToOutline
	}

	if deps.ReadFile == nil {
		deps.ReadFile = func(path string) (string, error) {
			data, err := os.ReadFile(path)

			return string(data), err
		}
	}

	if deps.WriteFile == nil {
		deps.WriteFile = func(path string, data string) error {
			return os.WriteFile(path, []byte(data), 0o600) //nolint:mnd
		}
	}

	if deps.Stat == nil {
		deps.Stat = os.Stat
	}

	if deps.Rename == nil {
		deps.Rename = os.Rename
	}

	if deps.Remove == nil {
		deps.Remove = os.Remove
	}

	if deps.Stdin == nil {
		deps.Stdin = internal.ReadFromStdin
	}

	if deps.Out == nil {
		deps.Out = os.Stdout
	}
}

// NewOutlineCmd creates the outline command with injectable dependencies.
// If deps is nil, production defaults are used. Individual nil fields also fall back
// to their defaults, so tests can inject only the fields they need.
func NewOutlineCmd(deps *OutlineDependencies) *cobra.Command {
	if deps == nil {
		deps = &OutlineDependencies{} //nolint:exhaustruct
	}

	fillOutlineDeps(deps)

	var inPlace bool

	var moveTo string

	var keepBreaks bool

	outlineCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "outline [file ...]",
		Short: "Convert flat Markdown to a Logseq bullet outline",
		Long: `Convert flat Markdown to a Logseq bullet outline.

When no files are given, reads from stdin and writes to stdout.
When files are given, writes converted output to stdout by default.
Use --in-place to overwrite files, or --move-to to write to a destination directory.

Examples:
  printf '# H\n\n- a\n' | lqd outline
  lqd outline file.md
  lqd outline --in-place file.md
  lqd outline --move-to /tmp file.md`,
		Args: cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runOutline(deps, args, inPlace, moveTo, keepBreaks)
		},
	}

	outlineCmd.Flags().BoolVarP(&inPlace, "in-place", "i", false, "Overwrite each file with its converted output")
	outlineCmd.Flags().StringVar(&moveTo, "move-to", "", "Write converted output to this directory; remove original")
	outlineCmd.Flags().BoolVar(&keepBreaks, "keep-breaks", false, "Preserve blank lines as empty bullet lines")

	return outlineCmd
}

// runOutline handles the outline command's stdout-only branch (no flags).
func runOutlineToStdout(deps *OutlineDependencies, path string, opts internal.OutlineOptions) error {
	content, err := deps.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	fmt.Fprint(deps.Out, deps.Convert(content, opts))

	return nil
}

// runOutlineInPlace converts a file and writes the result back to the same path.
func runOutlineInPlace(deps *OutlineDependencies, path string, opts internal.OutlineOptions) error {
	content, err := deps.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	converted := deps.Convert(content, opts)

	err = deps.WriteFile(path, converted)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// runOutlineMoveTo converts a file and writes the result to a destination directory, then removes the source.
func runOutlineMoveTo(deps *OutlineDependencies, path, moveTo string, opts internal.OutlineOptions) error {
	content, err := deps.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	dest := filepath.Join(moveTo, filepath.Base(path))

	_, statErr := deps.Stat(dest)
	if statErr == nil {
		return fmt.Errorf("%w: %s", ErrDestinationExists, dest)
	}

	converted := deps.Convert(content, opts)

	err = deps.WriteFile(dest, converted)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dest, err)
	}

	err = deps.Remove(path)
	if err != nil {
		return fmt.Errorf("removing %s: %w", path, err)
	}

	return nil
}

// runOutline is the core logic for the outline command.
func runOutline(deps *OutlineDependencies, args []string, inPlace bool, moveTo string, keepBreaks bool) error {
	if inPlace && moveTo != "" {
		return ErrMutuallyExclusive
	}

	opts := internal.OutlineOptions{KeepBreaks: keepBreaks}

	// Stdin mode: no file arguments.
	if len(args) == 0 {
		fmt.Fprint(deps.Out, deps.Convert(deps.Stdin(), opts))

		return nil
	}

	// File loop.
	for _, path := range args {
		var err error

		switch {
		case inPlace:
			err = runOutlineInPlace(deps, path, opts)
		case moveTo != "":
			err = runOutlineMoveTo(deps, path, moveTo, opts)
		default:
			err = runOutlineToStdout(deps, path, opts)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(NewOutlineCmd(nil))
}
