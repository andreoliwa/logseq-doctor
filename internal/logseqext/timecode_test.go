package logseqext_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/require"
)

func TestEffectiveTime(t *testing.T) {
	const ref = "((12345678-1234-1234-1234-123456789abc))"
	for _, test := range []struct {
		name, input string
		minutes     int
		timed       bool
	}{
		{"direct", "- 10:30 Parent\n  - 12:06 Child", 630, true},
		{"ref", "- " + ref + "\n  - 12:06 Child", 726, true},
		{"first not earliest", "- " + ref + "\n  - Notes\n  - *12h20* First\n  - 12:00 Second", 740, true},
		{"properties", "- " + ref + "\n  collapsed:: true\n  - *12h06* Child", 726, true},
		{"grandchild", "- " + ref + "\n  - Notes\n    - 12:06 Deeper", 0, false},
		{"ordinary", "- Notes\n  - 12:06 Child", 0, false},
		{"mixed", "- Notes about " + ref + "\n  - 12:06 Child", 0, false},
		{"no children", "- " + ref, 0, false},
		{"task ref", "- " + content.TaskStringTodo + " " + ref + "\n  - 12:06 Child", 0, false},
		{"priority ref", "- [#A] " + ref + "\n  - 12:06 Child", 0, false},
		{"ref child", "- " + ref + "\n  - " + ref + "\n    - 12:06 Child", 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, err := logseq.ParseBlock(test.input)
			require.NoError(t, err)

			value, ok := logseqext.EffectiveTime(root.Blocks()[0])
			require.Equal(t, test.timed, ok)

			if ok {
				require.Equal(t, test.minutes, value.Minutes())
			}
		})
	}
}
