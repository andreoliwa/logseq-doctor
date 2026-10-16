package cmd_test

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/andreoliwa/logseq-doctor/cmd"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/stretchr/testify/require"
)

var (
	errJournalWrite = errors.New("journal write failed")
	errRead         = errors.New("stdin read failed")
)

func TestContentCommandRawInputAndDate(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		args    []string
		date    time.Time
		failure error
	}{
		{"today", nil, time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC), nil},
		{"journal", []string{"--journal", "2026-10-15"}, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), nil},
		{"failure", nil, time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC), errJournalWrite},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			const raw = "- **9h00** New\r\n  - [literal]\r\n- 11:00 Next\n"

			calls := 0
			command := cmd.NewContentCmd(&cmd.ContentDependencies{
				AppendFn: func(_ *logseq.Graph, date time.Time, text string) (int, error) {
					calls++

					require.Equal(t, testCase.date, date)
					require.Equal(t, raw, text)

					return 0, testCase.failure
				},
				OpenGraph: func(string) *logseq.Graph { return nil },
				ReadStdin: func() string { return raw },
				TimeNow:   func() time.Time { return time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC) },
			})
			command.SetArgs(testCase.args)
			err := command.Execute()
			require.ErrorIs(t, err, testCase.failure)
			require.Equal(t, 1, calls)
		})
	}
}

func TestContentCommandRejectsInvalidDate(t *testing.T) {
	command := cmd.NewContentCmd(&cmd.ContentDependencies{
		AppendFn: func(*logseq.Graph, time.Time, string) (int, error) {
			t.Fatal("invalid date reached writer")

			return 0, nil
		},
		OpenGraph: func(string) *logseq.Graph { return nil }, ReadStdin: func() string { return "raw" }, TimeNow: time.Now,
	})
	command.SetArgs([]string{"--journal", "bad-date"})
	require.Error(t, command.Execute())
}

func TestContentCommandReadsRawCommandInput(t *testing.T) {
	for _, testCase := range []struct{ name, raw string }{
		{"CRLF", "- 09:00 First\r\n  detail\r\n"},
		{"no final newline", "- 09:00 Final"},
		{"long line", "- 09:00 " + strings.Repeat("x", 128*1024)},
		{"empty", ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			calls := 0
			command := cmd.NewContentCmd(&cmd.ContentDependencies{
				AppendFn: func(_ *logseq.Graph, _ time.Time, text string) (int, error) {
					calls++

					require.Equal(t, testCase.raw, text)

					return len(text), nil
				},
				OpenGraph: func(string) *logseq.Graph { return nil },
				TimeNow:   time.Now,
			})
			command.SetArgs([]string{})
			command.SetIn(strings.NewReader(testCase.raw))
			require.NoError(t, command.Execute())
			require.Equal(t, 1, calls)
		})
	}
}

func TestContentCommandReadFailureDoesNotWrite(t *testing.T) {
	command := cmd.NewContentCmd(&cmd.ContentDependencies{
		AppendFn: func(*logseq.Graph, time.Time, string) (int, error) {
			t.Fatal("failed input reached writer")

			return 0, nil
		},
		OpenGraph: func(string) *logseq.Graph {
			t.Fatal("failed input opened graph")

			return nil
		},
		TimeNow: time.Now,
	})
	command.SetArgs([]string{})
	command.SetIn(iotest.ErrReader(errRead))
	require.ErrorIs(t, command.Execute(), errRead)
}
