package cmd_test

import (
	"testing"
	"time"

	"github.com/andreoliwa/lqd/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateFromJournalFlag(t *testing.T) { //nolint:funlen
	frozenTime := time.Date(2025, 4, 5, 3, 0, 0, 0, time.UTC)
	timeNow := func() time.Time {
		return frozenTime
	}

	tests := []struct {
		name           string
		journalFlag    string
		expectedDate   time.Time
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "empty flag returns current time",
			journalFlag:    "",
			expectedDate:   frozenTime,
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "valid date format",
			journalFlag:    "2024-12-25",
			expectedDate:   time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "another valid date",
			journalFlag:    "2025-01-15",
			expectedDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "invalid date format - wrong separator",
			journalFlag:    "2024/12/25",
			expectedDate:   time.Time{},
			expectError:    true,
			expectedErrMsg: "invalid journal date format. Use YYYY-MM-DD",
		},
		{
			name:           "invalid date format - missing day",
			journalFlag:    "2024-12",
			expectedDate:   time.Time{},
			expectError:    true,
			expectedErrMsg: "invalid journal date format. Use YYYY-MM-DD",
		},
		{
			name:           "invalid date format - wrong order",
			journalFlag:    "12-25-2024",
			expectedDate:   time.Time{},
			expectError:    true,
			expectedErrMsg: "invalid journal date format. Use YYYY-MM-DD",
		},
		{
			name:           "invalid date format - not a date",
			journalFlag:    "not-a-date",
			expectedDate:   time.Time{},
			expectError:    true,
			expectedErrMsg: "invalid journal date format. Use YYYY-MM-DD",
		},
		{
			name:           "invalid date - invalid month",
			journalFlag:    "2024-13-01",
			expectedDate:   time.Time{},
			expectError:    true,
			expectedErrMsg: "invalid journal date format. Use YYYY-MM-DD",
		},
		{
			name:           "invalid date - invalid day",
			journalFlag:    "2024-02-30",
			expectedDate:   time.Time{},
			expectError:    true,
			expectedErrMsg: "invalid journal date format. Use YYYY-MM-DD",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := cmd.ParseDateFromJournalFlag(test.journalFlag, timeNow)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedDate, result)
			}
		})
	}
}
