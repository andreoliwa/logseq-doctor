package backlog_test

import (
	"testing"

	"github.com/andreoliwa/lsd/internal/backlog"

	"github.com/stretchr/testify/assert"
)

func TestFormatCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		singular string
		plural   string
		expected string
	}{
		{"singular case", 1, "apple", "apples", "1 apple"},
		{"plural case", 2, "apple", "apples", "2 apples"},
		{"zero case", 0, "apple", "apples", "0 apples"},
		{"negative case", -1, "apple", "apples", "-1 apples"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := backlog.FormatCount(test.count, test.singular, test.plural)
			assert.Equal(t, test.expected, result)
		})
	}
}
