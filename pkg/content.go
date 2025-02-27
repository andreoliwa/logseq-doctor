package pkg

import (
	"fmt"
	"github.com/andreoliwa/lsd/internal"
	"log"
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
func AppendRawMarkdownToJournal(graphDir string, date time.Time, rawMarkdown string) int { //nolint:funlen,cyclop
	if rawMarkdown == "" {
		return 0
	}

	graph := internal.OpenGraphFromDirOrEnv(graphDir)

	path, _ := graph.JournalPath(date)

	var originalContents string

	var empty bool

	_, err := os.Stat(path)
	if err == nil {
		bytes, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Fatalln("error reading journal file:", readErr)
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
		log.Fatalln("error opening journal file:", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln("error closing journal file:", err)
		}
	}(file)

	size, err := file.WriteString(rawMarkdown)
	if err != nil {
		fmt.Println(fmt.Errorf("error writing to journal file: %w", err))

		return 0
	}

	return size
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
