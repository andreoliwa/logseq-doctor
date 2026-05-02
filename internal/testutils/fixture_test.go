package testutils_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugToUUID_Deterministic(t *testing.T) {
	blocks := []testutils.Block{
		testutils.Task("todo-home-1", "TODO", "Clean windows"),
	}
	m1, _ := testutils.ExportBuildSlugMap(blocks)
	m2, _ := testutils.ExportBuildSlugMap(blocks)
	assert.Equal(t, m1["todo-home-1"], m2["todo-home-1"], "same slug must always produce same UUID")
}

func TestSlugToUUID_RoundTrip(t *testing.T) {
	blocks := []testutils.Block{
		testutils.Task("todo-home-1", "TODO", "Clean windows"),
		testutils.Task("todo-phone-1", "TODO", "Call dentist"),
	}

	s2u, u2s := testutils.ExportBuildSlugMap(blocks)

	for slug, uuid := range s2u {
		assert.Equal(t, slug, u2s[uuid], "UUID→slug must reverse slug→UUID")
	}
}

func TestBuildSlugMap_CollisionPanics(t *testing.T) {
	// Force a collision by using the same slug twice.
	blocks := []testutils.Block{
		testutils.Task("dup", "TODO", "First"),
		testutils.Task("dup", "TODO", "Second"),
	}

	assert.Panics(t, func() { testutils.ExportBuildSlugMap(blocks) })
}

func TestExpandSlugs_ReplacesKnownSlugs(t *testing.T) {
	blocks := []testutils.Block{testutils.Task("todo-home-1", "TODO", "text")}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	uuid := s2u["todo-home-1"]

	result := testutils.ExportExpandSlugs("- (( todo-home-1 ))", s2u)
	assert.Equal(t, "- (("+uuid+"))", result)
}

func TestExpandSlugs_LeavesUnknownUnchanged(t *testing.T) {
	result := testutils.ExportExpandSlugs("- (( unknown-slug ))", map[string]string{})
	assert.Equal(t, "- (( unknown-slug ))", result)
}

func TestExpandSlugs_MultipleSlugsOnOneLine(t *testing.T) {
	blocks := []testutils.Block{
		testutils.Task("slug-a", "TODO", "a"),
		testutils.Task("slug-b", "TODO", "b"),
	}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	input := "(( slug-a )) and (( slug-b ))"
	result := testutils.ExportExpandSlugs(input, s2u)
	assert.Contains(t, result, "(("+s2u["slug-a"]+"))")
	assert.Contains(t, result, "(("+s2u["slug-b"]+"))")
}

func TestCollapseSlugs_ReversesExpand(t *testing.T) {
	blocks := []testutils.Block{testutils.Task("todo-home-1", "TODO", "text")}
	s2u, u2s := testutils.ExportBuildSlugMap(blocks)
	uuid := s2u["todo-home-1"]

	expanded := "- ((" + uuid + "))"
	collapsed := testutils.ExportCollapseSlugs(expanded, u2s)
	assert.Equal(t, "- (( todo-home-1 ))", collapsed)
}

func TestCollapseSlugs_LeavesUnknownUUIDsUnchanged(t *testing.T) {
	input := "- ((67c48cb3-ee20-48fe-9a2b-3f63dff412ff))"
	result := testutils.ExportCollapseSlugs(input, map[string]string{})
	assert.Equal(t, input, result)
}

func TestNewFixture_SlugMapPopulated(t *testing.T) {
	f := testutils.NewFixture(t,
		testutils.Task("todo-home-1", "TODO", "Clean windows", testutils.WithTags("home")),
	)
	require.NotNil(t, f)
	uuid := testutils.ExportFixtureUUID(f, "todo-home-1")
	assert.NotEmpty(t, uuid)
}

func TestAdd_AppendsBlocks(t *testing.T) {
	f := testutils.NewFixture(t,
		testutils.Task("slug-a", "TODO", "A"),
	)
	f.Add(testutils.Task("slug-b", "TODO", "B"))
	assert.NotEmpty(t, testutils.ExportFixtureUUID(f, "slug-a"))
	assert.NotEmpty(t, testutils.ExportFixtureUUID(f, "slug-b"))
}

func TestAdd_DuplicateSlugPanics(t *testing.T) {
	f := testutils.NewFixture(t, testutils.Task("slug-a", "TODO", "A"))
	assert.Panics(t, func() { f.Add(testutils.Task("slug-a", "TODO", "B")) })
}

func TestAdd_Chaining(t *testing.T) {
	f := testutils.NewFixture(t, testutils.Task("a", "TODO", "A"))
	result := f.Add(testutils.Task("b", "TODO", "B"))
	assert.Same(t, f, result, "Add must return the same fixture pointer")
}

func TestResolveRelativeDate_Empty(t *testing.T) {
	result, err := testutils.ExportResolveRelativeDate("", time.Now())
	require.NoError(t, err)
	assert.True(t, result.IsZero())
}

func TestResolveRelativeDate_Zero(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := testutils.ExportResolveRelativeDate("0", now)
	require.NoError(t, err)
	assert.Equal(t, now, result)
}

func TestResolveRelativeDate_PlusDays(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := testutils.ExportResolveRelativeDate("+3d", now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 4, 16, 0, 0, 0, 0, time.UTC), result)
}

func TestResolveRelativeDate_MinusWeek(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := testutils.ExportResolveRelativeDate("-1w", now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 4, 6, 0, 0, 0, 0, time.UTC), result)
}

func TestResolveRelativeDate_Invalid(t *testing.T) {
	_, err := testutils.ExportResolveRelativeDate("badvalue", time.Now())
	assert.Error(t, err)
}

func TestBuildAPIResponse_MarkerAndContent(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	blocks := []testutils.Block{
		testutils.Task("todo-home-1", "TODO", "Clean windows", testutils.WithTags("home")),
	}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	resp := testutils.ExportBuildAPIResponse(blocks, s2u, now)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resp), &parsed))
	require.Len(t, parsed, 1)
	assert.Equal(t, "TODO", parsed[0]["marker"])
	assert.Contains(t, parsed[0]["content"], "TODO #home Clean windows")
	assert.Contains(t, parsed[0]["content"], "id:: "+s2u["todo-home-1"])
}

func TestBuildAPIResponse_DeadlineEncoding(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	blocks := []testutils.Block{
		testutils.Task("deadline-1", "TODO", "Buy laptop", testutils.WithDeadline("+3d")),
	}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	resp := testutils.ExportBuildAPIResponse(blocks, s2u, now)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resp), &parsed))
	assert.InDelta(t, float64(20250416), parsed[0]["deadline"], 0)
}

func TestBuildAPIResponse_ScheduledEncoding(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	blocks := []testutils.Block{
		testutils.Task("sched-1", "TODO", "Plan menu", testutils.WithScheduled("-1w")),
	}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	resp := testutils.ExportBuildAPIResponse(blocks, s2u, now)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resp), &parsed))
	assert.InDelta(t, float64(20250406), parsed[0]["scheduled"], 0)
}

func TestBuildAPIResponse_PriorityInContent(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	blocks := []testutils.Block{
		testutils.Task("prio-1", "TODO", "Urgent", testutils.WithPriority("A"), testutils.WithTags("work")),
	}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	resp := testutils.ExportBuildAPIResponse(blocks, s2u, now)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resp), &parsed))
	assert.Contains(t, parsed[0]["content"], "TODO [#A] #work Urgent")
}

func TestBuildAPIResponse_GroomedProperty(t *testing.T) {
	now := time.Date(2025, 4, 13, 0, 0, 0, 0, time.UTC)
	blocks := []testutils.Block{
		testutils.Task("groomed-1", "TODO", "Done last week", testutils.WithGroomed("-7d")),
	}
	s2u, _ := testutils.ExportBuildSlugMap(blocks)
	resp := testutils.ExportBuildAPIResponse(blocks, s2u, now)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(resp), &parsed))
	props, ok := parsed[0]["properties"].(map[string]any)
	require.True(t, ok, "properties must be a map")
	assert.Contains(t, props["groomed"], "[[")
	assert.Contains(t, props["groomed"], "2025")
}
