package api

import (
	"sort"
	"strings"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

// BuildRefLookup builds a mapping from Logseq ref ID to human-readable name.
func BuildRefLookup(tasks []TaskJSON) map[int]string {
	refLookup := make(map[int]string)

	for _, task := range tasks {
		populatePageRef(refLookup, task)
	}

	refCandidates := collectRefCandidates(tasks, refLookup)
	resolveRefCandidates(refLookup, refCandidates)

	return refLookup
}

// populatePageRef adds the page reference for a task to the lookup.
func populatePageRef(refLookup map[int]string, task TaskJSON) {
	if task.Page.ID == 0 {
		return
	}

	name := task.Page.OriginalName
	if name == "" {
		name = task.Page.Name
	}

	if name != "" {
		refLookup[task.Page.ID] = name
	}
}

// collectRefCandidates gathers tag candidates for unresolved ref IDs.
func collectRefCandidates(tasks []TaskJSON, refLookup map[int]string) map[int]map[string]int {
	candidates := make(map[int]map[string]int)

	for _, task := range tasks {
		directTags := logseqext.ExtractDirectTags(task.Content)

		for _, ref := range task.Refs {
			if _, resolved := refLookup[ref.ID]; resolved {
				continue
			}

			if _, exists := candidates[ref.ID]; !exists {
				candidates[ref.ID] = make(map[string]int)
			}

			for _, tag := range directTags {
				candidates[ref.ID][tag]++
			}
		}
	}

	return candidates
}

// resolveRefCandidates picks the best tag name for each unresolved ref ID.
func resolveRefCandidates(refLookup map[int]string, candidates map[int]map[string]int) {
	for refID, tagCounts := range candidates {
		bestTag := ""
		bestCount := 0

		for tag, count := range tagCounts {
			if count > bestCount {
				bestTag = tag
				bestCount = count
			}
		}

		if bestTag != "" {
			refLookup[refID] = bestTag
		}
	}
}

// EnrichTasksWithAncestorTags adds inherited tags from pathRefs to each task's tag set.
func EnrichTasksWithAncestorTags(tasks []TaskJSON, refLookup map[int]string) map[string]string {
	tagsByUUID := make(map[string]string)

	for _, task := range tasks {
		directTags := logseqext.ExtractDirectTags(task.Content)
		directRefIDs := buildDirectRefIDSet(task)
		ancestorTags := collectAncestorTags(task, directRefIDs, refLookup)

		allTags := logseqext.UniqueStrings(append(directTags, ancestorTags...))
		sort.Strings(allTags)
		logseqext.NormalizeTagPrefixes(allTags)

		tagsByUUID[task.UUID] = strings.Join(allTags, " ")
	}

	return tagsByUUID
}

// buildDirectRefIDSet creates a set of ref IDs that are direct references (including page).
func buildDirectRefIDSet(task TaskJSON) map[int]bool {
	directRefIDs := make(map[int]bool)

	for _, ref := range task.Refs {
		directRefIDs[ref.ID] = true
	}

	directRefIDs[task.Page.ID] = true

	return directRefIDs
}

// collectAncestorTags gathers tags from pathRefs that are not in the direct ref set.
func collectAncestorTags(task TaskJSON, directRefIDs map[int]bool, refLookup map[int]string) []string {
	var ancestorTags []string

	for _, pathRef := range task.PathRefs {
		if directRefIDs[pathRef.ID] {
			continue
		}

		if name, ok := refLookup[pathRef.ID]; ok {
			ancestorTags = append(ancestorTags, name)
		}
	}

	return ancestorTags
}
