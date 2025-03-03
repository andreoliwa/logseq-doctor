package internal_test

import (
	"github.com/andreoliwa/lsd/internal"
	"testing"
	"time"

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.DateYYYYMMDD(tt.input)
			assert.Equal(t, tt.expected, result, "Date conversion failed for %s", tt.name)
		})
	}
}
