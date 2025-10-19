package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// addJournalFlag adds a --journal/-j flag to the command with customizable help text.
func addJournalFlag(cmd *cobra.Command, flagVar *string) {
	cmd.Flags().StringVarP(flagVar, "journal", "j", "", "Journal date in YYYY-MM-DD format (default: today)")
}

// addBlockFlag adds a --block/-b flag to the command with customizable help text.
func addBlockFlag(cmd *cobra.Command, flagVar *string, what string) {
	helpText := "Partial text of a block that will be the parent of the added " + what
	cmd.Flags().StringVarP(flagVar, "block", "b", "", helpText)
}

// addPageFlag adds a --page/-p flag to the command with customizable help text.
func addPageFlag(cmd *cobra.Command, flagVar *string, what string) {
	helpText := fmt.Sprintf("Page name where the %s will be added to (default: today's journal)", what)
	cmd.Flags().StringVarP(flagVar, "page", "p", "", helpText)
}

// ParseDateFromJournalFlag parses the journal flag and returns the target date.
// If journalFlag is empty, it returns the current time from timeNow.
// If journalFlag is not empty, it parses it as YYYY-MM-DD format.
// Returns an error if the date format is invalid.
func ParseDateFromJournalFlag(journalFlag string, timeNow func() time.Time) (time.Time, error) {
	if journalFlag != "" {
		parsedDate, err := time.Parse("2006-01-02", journalFlag)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid journal date format. Use YYYY-MM-DD: %w", err)
		}

		return parsedDate, nil
	}

	return timeNow(), nil
}
