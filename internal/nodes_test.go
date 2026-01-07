package internal_test

import (
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/stretchr/testify/assert"
)

func TestIsAncestor(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*content.Block, *content.Block)
		expected bool
	}{
		{
			name: "block is same as ancestor",
			setup: func() (*content.Block, *content.Block) {
				block := content.NewBlock(content.NewText("test"))

				return block, block
			},
			expected: true,
		},
		{
			name: "direct parent relationship",
			setup: func() (*content.Block, *content.Block) {
				parent := content.NewBlock(content.NewText("parent"))
				child := content.NewBlock(content.NewText("child"))
				parent.AddChild(child)

				return child, parent
			},
			expected: true,
		},
		{
			name: "grandparent relationship",
			setup: func() (*content.Block, *content.Block) {
				grandparent := content.NewBlock(content.NewText("grandparent"))
				parent := content.NewBlock(content.NewText("parent"))
				child := content.NewBlock(content.NewText("child"))

				grandparent.AddChild(parent)
				parent.AddChild(child)

				return child, grandparent
			},
			expected: true,
		},
		{
			name: "great-grandparent relationship",
			setup: func() (*content.Block, *content.Block) {
				greatGrandparent := content.NewBlock(content.NewText("great-grandparent"))
				grandparent := content.NewBlock(content.NewText("grandparent"))
				parent := content.NewBlock(content.NewText("parent"))
				child := content.NewBlock(content.NewText("child"))

				greatGrandparent.AddChild(grandparent)
				grandparent.AddChild(parent)
				parent.AddChild(child)

				return child, greatGrandparent
			},
			expected: true,
		},
		{
			name: "no relationship - different branches",
			setup: func() (*content.Block, *content.Block) {
				root := content.NewBlock(content.NewText("root"))
				branch1 := content.NewBlock(content.NewText("branch1"))
				branch2 := content.NewBlock(content.NewText("branch2"))
				child1 := content.NewBlock(content.NewText("child1"))
				child2 := content.NewBlock(content.NewText("child2"))

				root.AddChild(branch1)
				root.AddChild(branch2)
				branch1.AddChild(child1)
				branch2.AddChild(child2)

				// Return child1 and branch2 (different branches, no ancestor relationship)
				return child1, branch2
			},
			expected: false,
		},
		{
			name: "no relationship - sibling blocks",
			setup: func() (*content.Block, *content.Block) {
				parent := content.NewBlock(content.NewText("parent"))
				sibling1 := content.NewBlock(content.NewText("sibling1"))
				sibling2 := content.NewBlock(content.NewText("sibling2"))

				parent.AddChild(sibling1)
				parent.AddChild(sibling2)

				return sibling1, sibling2
			},
			expected: false,
		},
		{
			name: "reverse relationship - child is not ancestor of parent",
			setup: func() (*content.Block, *content.Block) {
				parent := content.NewBlock(content.NewText("parent"))
				child := content.NewBlock(content.NewText("child"))
				parent.AddChild(child)

				return parent, child
			},
			expected: false,
		},
		{
			name: "orphaned block - no parent",
			setup: func() (*content.Block, *content.Block) {
				orphan := content.NewBlock(content.NewText("orphan"))
				other := content.NewBlock(content.NewText("other"))

				return orphan, other
			},
			expected: false,
		},
		{
			name: "nil ancestor",
			setup: func() (*content.Block, *content.Block) {
				block := content.NewBlock(content.NewText("block"))

				return block, nil
			},
			expected: false,
		},
		{
			name: "complex hierarchy - deep nesting",
			setup: func() (*content.Block, *content.Block) {
				// Create a 5-level deep hierarchy
				level0 := content.NewBlock(content.NewText("level0"))
				level1 := content.NewBlock(content.NewText("level1"))
				level2 := content.NewBlock(content.NewText("level2"))
				level3 := content.NewBlock(content.NewText("level3"))
				level4 := content.NewBlock(content.NewText("level4"))

				level0.AddChild(level1)
				level1.AddChild(level2)
				level2.AddChild(level3)
				level3.AddChild(level4)

				return level4, level0
			},
			expected: true,
		},
		{
			name: "complex hierarchy - middle ancestor",
			setup: func() (*content.Block, *content.Block) {
				// Create a 5-level deep hierarchy, check middle level
				level0 := content.NewBlock(content.NewText("level0"))
				level1 := content.NewBlock(content.NewText("level1"))
				level2 := content.NewBlock(content.NewText("level2"))
				level3 := content.NewBlock(content.NewText("level3"))
				level4 := content.NewBlock(content.NewText("level4"))

				level0.AddChild(level1)
				level1.AddChild(level2)
				level2.AddChild(level3)
				level3.AddChild(level4)

				return level4, level2
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block, ancestor := test.setup()
			result := internal.IsAncestor(block, ancestor)
			assert.Equal(t, test.expected, result, "IsAncestor result mismatch for test case: %s", test.name)
		})
	}
}

func TestIsAncestor_EdgeCases(t *testing.T) {
	t.Run("nil block", func(t *testing.T) {
		ancestor := content.NewBlock(content.NewText("ancestor"))
		result := internal.IsAncestor(nil, ancestor)
		assert.False(t, result, "nil block should return false")
	})

	t.Run("both nil", func(t *testing.T) {
		result := internal.IsAncestor(nil, nil)
		assert.False(t, result, "both nil should return false")
	})
}
