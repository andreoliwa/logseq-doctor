package internal_test

import (
	"testing"

	"github.com/andreoliwa/lsd/internal"
	"github.com/andreoliwa/lsd/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestFindFirstQuery(t *testing.T) {
	graph := testutils.StubGraph(t, "")
	finder := internal.NewLogseqFinder(graph)

	tests := []struct {
		name      string
		pageTitle string
		expected  string
	}{
		{
			name:      "no query",
			pageTitle: "query-none",
			expected:  "",
		},
		{
			name:      "one query",
			pageTitle: "query-one",
			expected: "{:title \"Who is using this account?\"\n  :query (property :payment-method [[query-one]])\n" +
				"  :collapsed? false}",
		},
		{
			name:      "multiple queries",
			pageTitle: "query-multiple",
			expected:  "(and (or [[home]] [[phone]]) (task TODO DOING WAITING))",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := finder.FindFirstQuery(test.pageTitle)

			assert.Equal(t, test.expected, query)
		})
	}
}
