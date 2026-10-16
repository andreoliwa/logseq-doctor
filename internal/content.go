package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-go"
	"github.com/google/uuid"
)

const defaultJournalMode = 0o644

// IsValidMarkdownFile checks if a file is a Markdown file, by looking at its extension, not its content.
func IsValidMarkdownFile(filePath string) bool {
	if filePath == "" {
		return false
	}

	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return false
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// AppendRawMarkdownToJournal inserts outlined roots chronologically, preserving raw bytes.
// Keep source slices instead of serializing: the serializer escapes literal brackets
// such as "[something]", which has broken links on this path in the past.
func AppendRawMarkdownToJournal(graph *logseq.Graph, date time.Time, rawMarkdown string) (int, error) {
	if rawMarkdown == "" {
		return 0, nil
	}

	path, err := graph.JournalPath(date)
	if err != nil {
		return 0, fmt.Errorf("error getting journal path: %w", err)
	}

	path, err = resolveJournalPath(path)
	if err != nil {
		return 0, err
	}

	snapshot, err := readJournalSnapshot(path)
	if err != nil {
		return 0, err
	}

	original := snapshot.data

	mode := os.FileMode(defaultJournalMode)

	if snapshot.info != nil {
		mode = snapshot.info.Mode().Perm()
	}

	result, err := placeRawContent(string(original), rawMarkdown)
	if err != nil {
		return 0, err
	}

	err = replaceJournal(path, []byte(result), mode, snapshot)
	if err != nil {
		return 0, fmt.Errorf("error writing journal file: %w", err)
	}

	return len(result) - len(original), nil
}

type journalSnapshot struct {
	data []byte
	info os.FileInfo
}

var errJournalChanged = errors.New("journal changed while preparing replacement")

func resolveJournalPath(path string) (string, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return path, nil
	}

	if err != nil {
		return "", fmt.Errorf("inspect journal path: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve journal path: %w", err)
	}

	return resolved, nil
}

func readJournalSnapshot(path string) (journalSnapshot, error) {
	empty := journalSnapshot{data: nil, info: nil}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return empty, nil
	}

	if err != nil {
		return empty, fmt.Errorf("open journal snapshot: %w", err)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return empty, fmt.Errorf("stat journal snapshot: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return empty, fmt.Errorf("read journal snapshot: %w", err)
	}

	return journalSnapshot{data: data, info: info}, nil
}

// checkJournalSnapshot detects changes already visible before rename. External
// editors do not share a lock, so this does not close the final check/rename gap.
func checkJournalSnapshot(path string, expected journalSnapshot) error {
	current, err := readJournalSnapshot(path)
	if err != nil {
		return err
	}

	if expected.info == nil && current.info == nil {
		return nil
	}

	if expected.info == nil || current.info == nil {
		return errJournalChanged
	}

	if !os.SameFile(expected.info, current.info) ||
		expected.info.Mode() != current.info.Mode() ||
		!expected.info.ModTime().Equal(current.info.ModTime()) ||
		!bytes.Equal(expected.data, current.data) {
		return errJournalChanged
	}

	return nil
}

// replaceJournal prepares the complete replacement in the same directory, so a
// failed write or rename cannot leave a partially appended journal.
func replaceJournal(path string, data []byte, mode os.FileMode, expected ...journalSnapshot) error {
	snapshot, err := replacementSnapshot(path, expected)
	if err != nil {
		return err
	}

	temporary := filepath.Join(filepath.Dir(path), ".lqd-content-"+uuid.NewString())

	file, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultJournalMode)
	if err != nil {
		return fmt.Errorf("create journal replacement: %w", err)
	}

	defer os.Remove(file.Name())

	if snapshot.info != nil {
		err = file.Chmod(mode)
	}

	if err == nil {
		_, err = file.Write(data)
	}

	if err == nil {
		err = file.Sync()
	}

	closeErr := file.Close()

	if err != nil {
		return fmt.Errorf("prepare journal replacement: %w", err)
	}

	if closeErr != nil {
		return fmt.Errorf("close journal replacement: %w", closeErr)
	}

	err = checkJournalSnapshot(path, snapshot)
	if err != nil {
		return err
	}

	err = os.Rename(file.Name(), path)
	if err != nil {
		return fmt.Errorf("replace journal file: %w", err)
	}

	return nil
}

func replacementSnapshot(path string, expected []journalSnapshot) (journalSnapshot, error) {
	if len(expected) != 0 {
		return expected[0], nil
	}

	return readJournalSnapshot(path)
}
