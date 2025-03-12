package utils_test

import (
	"github.com/andreoliwa/lsd/pkg/utils"
	"testing"

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatCount(tt.count, tt.singular, tt.plural)
			assert.Equal(t, tt.expected, result)
		})
	}
}
