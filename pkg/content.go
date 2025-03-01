package pkg

import (
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"os"
	"strings"
	"time"
)

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

// AppendRawMarkdownToJournal appends raw Markdown content to the journal page for the given date.
// I tried appending blocks with `logseq-go` but there is and with text containing brackets.
// e.g. "[something]" is escaped like "\[something\]" and this breaks links.
func AppendRawMarkdownToJournal(graph *logseq.Graph, date time.Time, //nolint:cyclop
	rawMarkdown string) (int, error) {
	if rawMarkdown == "" {
		return 0, nil
	}

	path, err := graph.JournalPath(date)
	if err != nil {
		return 0, fmt.Errorf("error getting journal path: %w", err)
	}

	var originalContents string

	var empty bool

	_, err = os.Stat(path)
	if err == nil {
		bytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return 0, fmt.Errorf("error reading journal file: %w", readErr)
		}

		originalContents = string(bytes)
		empty = strings.TrimLeft(strings.TrimRight(originalContents, "\n"), "- ") == ""
	}

	if os.IsNotExist(err) {
		originalContents = ""
		empty = true
	}

	newline := "\n"
	if os.PathSeparator == '\\' {
		newline = "\r\n"
	}

	var flags int
	if empty {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	} else {
		flags = os.O_WRONLY | os.O_APPEND

		// Add a newline before the new content if the original content doesn't end with a newline.
		if !strings.HasSuffix(originalContents, newline) {
			rawMarkdown = newline + rawMarkdown
		}
	}

	const perm = 0644

	file, err := os.OpenFile(path, flags, perm)
	if err != nil {
		return 0, fmt.Errorf("error opening journal file: %w", err)
	}
	defer file.Close()

	size, err := file.WriteString(rawMarkdown)
	if err != nil {
		return 0, fmt.Errorf("error writing journal file: %w", err)
	}

	return size, nil
}

// TODO: use and improve this function when appending tasks to the current journal.
// func appendToCurrentJournal(rawMarkdown string) {
//	graph := internal.OpenGraphFromDirOrEnv()
//
//	transaction := graph.NewTransaction()
//
//	date := time.Now()
//	journal, err := transaction.OpenJournal(date)
//	if err != nil {
//		log.Fatalf("error opening journal %s: %s\n", date, err)
//	}
//
//	if journal == nil {
//		log.Fatalf("error opening journal %s: it is nil\n", date)
//	}
//
//	journal.AddBlock(content.NewBlock(content.NewText(rawMarkdown)))
//
//	err = transaction.Save()
//	if err != nil {
//		log.Fatalf("error saving transaction: %s\n", err)
//	}
//}
