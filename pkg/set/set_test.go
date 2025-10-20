package set_test

import (
	"github.com/andreoliwa/logseq-doctor/pkg/set"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	set := set.NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(1) // Duplicate should not be added

	assert.Equal(t, 2, set.Size())
	assert.True(t, set.Contains(1))
	assert.True(t, set.Contains(2))
	assert.False(t, set.Contains(3))
}

func TestSet_Remove(t *testing.T) {
	set := set.NewSet[int]()
	set.Add(1)
	set.Add(2)

	set.Remove(1)
	assert.False(t, set.Contains(1))
	assert.True(t, set.Contains(2))

	set.Remove(2)
	assert.False(t, set.Contains(2))
	assert.Equal(t, 0, set.Size())
}

func TestSet_Contains(t *testing.T) {
	set := set.NewSet[string]()
	set.Add("hello")
	set.Add("world")

	assert.True(t, set.Contains("hello"))
	assert.True(t, set.Contains("world"))
	assert.False(t, set.Contains("golang"))
}

func TestSet_Size(t *testing.T) {
	set := set.NewSet[int]()
	assert.Equal(t, 0, set.Size())

	set.Add(10)
	set.Add(20)
	assert.Equal(t, 2, set.Size())

	set.Remove(10)
	assert.Equal(t, 1, set.Size())
}

func TestSet_Values(t *testing.T) {
	set := set.NewSet[int]()
	set.Add(10)
	set.Add(5)
	set.Add(15)

	values := set.Values()
	assert.ElementsMatch(t, []int{5, 10, 15}, values)
}

func TestOrderedSet_ValuesSorted(t *testing.T) {
	set := set.NewSet[int]()
	set.Add(10)
	set.Add(5)
	set.Add(15)

	sortedValues := set.ValuesSorted()
	assert.Equal(t, []int{5, 10, 15}, sortedValues)
}

func TestSet_Diff(t *testing.T) {
	setA := set.NewSet[int]()
	setA.Add(1)
	setA.Add(2)
	setA.Add(3)

	setB := set.NewSet[int]()
	setB.Add(2)
	setB.Add(4)

	setC := set.NewSet[int]()
	setC.Add(3)
	setC.Add(5)

	diffSet := setA.Diff(setB, setC)
	assert.ElementsMatch(t, []int{1}, diffSet.Values()) // 1 should remain
}

func TestSet_Update(t *testing.T) {
	setA := set.NewSet[int]()
	setA.Add(1)
	setA.Add(2)

	setB := set.NewSet[int]()
	setB.Add(3)
	setB.Add(4)

	setC := set.NewSet[int]()
	setC.Add(4)
	setC.Add(5)

	setA.Update(setB, setC)
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5}, setA.Values())
}

func TestSet_Clear(t *testing.T) {
	set := set.NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)

	set.Clear()
	assert.Equal(t, 0, set.Size())
	assert.False(t, set.Contains(1))
	assert.False(t, set.Contains(2))
	assert.False(t, set.Contains(3))
}
