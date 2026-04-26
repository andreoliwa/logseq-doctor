package backlog_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/stretchr/testify/assert"
)

func TestHeader_String(t *testing.T) {
	tests := []struct {
		header   backlog.Header
		expected string
	}{
		{backlog.HeaderFocus, "🎯 Focus tasks"},
		{backlog.HeaderOverdue, "📅 Overdue tasks"},
		{backlog.HeaderNewTasks, "✨ New tasks"},
		{backlog.HeaderTriaged, "🏷️ Triaged tasks"},
		{backlog.HeaderScheduled, "⏰ Scheduled tasks"},
		{backlog.HeaderUnranked, "⤵️ Unranked tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.header.String())
		})
	}
}

func TestHeader_Matches(t *testing.T) {
	tests := []struct {
		name      string
		header    backlog.Header
		blockText string
		want      bool
	}{
		{"bare keyword matches", backlog.HeaderFocus, "focus", true},
		{"bare keyword case-insensitive", backlog.HeaderFocus, "FOCUS", true},
		{"canonical form matches", backlog.HeaderFocus, "🎯 Focus tasks", true},
		{"old form without emoji matches", backlog.HeaderNewTasks, "New tasks", true},
		{"old emoji + old label matches", backlog.HeaderNewTasks, "🆕 New tasks", true},
		{"wrong header no match", backlog.HeaderFocus, "scheduled", false},
		{"empty string no match", backlog.HeaderFocus, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.header.Matches(tt.blockText))
		})
	}
}
