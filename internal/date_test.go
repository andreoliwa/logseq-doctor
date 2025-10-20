package internal_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal"

	"github.com/stretchr/testify/assert"
)

func TestDateYYYYMMDD(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected int
	}{
		{"Jan 1, 2025", time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), 20250101},
		{"Feb 18, 2025", time.Date(2025, time.February, 18, 0, 0, 0, 0, time.UTC), 20250218},
		{"Dec 31, 2023", time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC), 20231231},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := internal.DateYYYYMMDD(test.input)
			assert.Equal(t, test.expected, result, "Date conversion failed for %s", test.name)
		})
	}
}
