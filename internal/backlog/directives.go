package backlog

import (
	"fmt"
	"time"

	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/fatih/color"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

type directiveKind int

const (
	directiveCancel directiveKind = iota
	directiveWaiting
	directiveTodo
	directivePriority
)

// blockDirective records a pending task modification detected on a backlog page.
type blockDirective struct {
	UUID          string
	Kind          directiveKind
	Priority      content.PriorityValue // only for directivePriority; PriorityNone for other kinds
	DirectiveNode content.Node          // the TaskMarker or Priority node to remove after apply
	BacklogBlock  *content.Block        // the block on the backlog page containing the BlockRef
}

// detectDirectives inspects all preceding siblings of the BlockRef in its parent Paragraph.
// It walks backwards from the BlockRef collecting TaskMarker and Priority nodes until
// it reaches a non-directive node. Returns all found directives (may be empty).
//
//nolint:cyclop // complexity comes from the inherent number of directive kinds, not poor structure
func detectDirectives(blockRef *content.BlockRef) []blockDirective {
	// blockRef lives inside a Paragraph inside a Block.
	var backlogBlock *content.Block

	if para, ok := blockRef.Parent().(*content.Paragraph); ok {
		if block, ok := para.Parent().(*content.Block); ok {
			backlogBlock = block
		}
	}

	var found []blockDirective

	for prev := blockRef.PreviousSibling(); prev != nil; prev = prev.PreviousSibling() {
		switch node := prev.(type) {
		case *content.TaskMarker:
			switch node.Status { //nolint:exhaustive // only CANCELED/WAITING/TODO are valid directives; others are ignored
			case content.TaskStatusCanceled, content.TaskStatusCancelled:
				found = append(found, blockDirective{
					UUID:          blockRef.ID,
					Kind:          directiveCancel,
					Priority:      content.PriorityNone,
					DirectiveNode: node,
					BacklogBlock:  backlogBlock,
				})
			case content.TaskStatusWaiting, content.TaskStatusWait:
				found = append(found, blockDirective{
					UUID:          blockRef.ID,
					Kind:          directiveWaiting,
					Priority:      content.PriorityNone,
					DirectiveNode: node,
					BacklogBlock:  backlogBlock,
				})
			case content.TaskStatusTodo:
				found = append(found, blockDirective{
					UUID:          blockRef.ID,
					Kind:          directiveTodo,
					Priority:      content.PriorityNone,
					DirectiveNode: node,
					BacklogBlock:  backlogBlock,
				})
			default:
				return found
			}
		case *content.Priority:
			found = append(found, blockDirective{
				UUID:          blockRef.ID,
				Kind:          directivePriority,
				Priority:      node.Priority,
				DirectiveNode: node,
				BacklogBlock:  backlogBlock,
			})
		default:
			return found
		}
	}

	return found
}

// applyDirectives processes all collected directives: modifies the real task block on disk,
// then strips the directive node from the backlog page.
//
// Directives for the same UUID are grouped and applied in a single transaction so the task
// file is opened and saved only once (e.g. WAITING + [#B] on the same block ref).
//
// If a block is not on disk and the Logseq API is available, it forces a UUID write-back.
// If the API is unavailable, it warns and skips.
// Returns true if any directive was successfully applied (meaning the backlog page AST was mutated
// and the caller must save the backlog transaction).
func applyDirectives(
	graph *logseq.Graph,
	logseqAPI logseqapi.LogseqAPI,
	directives []blockDirective,
	currentTime func() time.Time,
) bool {
	groups := groupDirectivesByUUID(directives)
	applied := false

	for gi := range groups {
		if applyDirectiveGroupAndCleanup(graph, logseqAPI, &groups[gi], currentTime) {
			applied = true
		}
	}

	return applied
}

// directiveGroup holds all directives targeting the same task UUID.
type directiveGroup struct {
	items     []*blockDirective
	uuid      string
	backlog   *content.Block
	hasCancel bool
}

// groupDirectivesByUUID groups a flat slice of directives into per-UUID groups,
// preserving insertion order so the task file is opened and saved only once per UUID.
func groupDirectivesByUUID(directives []blockDirective) []directiveGroup {
	seenUUID := make(map[string]int, len(directives))
	groups := make([]directiveGroup, 0, len(directives))

	for i := range directives {
		directive := &directives[i]

		if idx, ok := seenUUID[directive.UUID]; ok {
			groups[idx].items = append(groups[idx].items, directive)

			if directive.Kind == directiveCancel {
				groups[idx].hasCancel = true
			}
		} else {
			seenUUID[directive.UUID] = len(groups)
			groups = append(groups, directiveGroup{
				items:     []*blockDirective{directive},
				uuid:      directive.UUID,
				backlog:   directive.BacklogBlock,
				hasCancel: directive.Kind == directiveCancel,
			})
		}
	}

	return groups
}

// applyDirectiveGroupAndCleanup applies all directives in a group and strips their nodes.
// Returns true if the group was successfully applied.
//
func applyDirectiveGroupAndCleanup(
	graph *logseq.Graph,
	logseqAPI logseqapi.LogseqAPI,
	grp *directiveGroup,
	currentTime func() time.Time,
) bool {
	err := applyDirectiveGroup(graph, logseqAPI, grp.items, currentTime)
	if err != nil {
		kinds := make([]string, len(grp.items))
		for i, item := range grp.items {
			kinds[i] = kindName(item.Kind)
		}

		color.Yellow("[backlog] WARNING: directives %v on block %s: %v",
			kinds, grp.uuid, err)

		return false
	}

	for _, item := range grp.items {
		item.DirectiveNode.RemoveSelf()
	}

	if grp.backlog != nil {
		if grp.hasCancel {
			grp.backlog.RemoveSelf()
		} else {
			props := logseqext.BlockProperties(grp.backlog)
			props.Remove(logseqext.PropertyCancelled)

			if props.FirstChild() == nil {
				props.RemoveSelf()
			}
		}
	}

	return true
}

func applyDirectiveGroup(
	graph *logseq.Graph,
	logseqAPI logseqapi.LogseqAPI,
	items []*blockDirective,
	currentTime func() time.Time,
) error {
	if len(items) == 0 {
		return nil
	}

	block, transaction, err := logseqapi.FindBlockOnDisk(graph, logseqAPI, items[0].UUID)
	if err != nil {
		return fmt.Errorf("finding block on disk: %w", err)
	}

	for _, item := range items {
		applyErr := applyDirectiveToBlock(block, item, currentTime)
		if applyErr != nil {
			return applyErr
		}
	}

	saveErr := transaction.Save()
	if saveErr != nil {
		return fmt.Errorf("failed to save task page: %w", saveErr)
	}

	return nil
}

func applyDirectiveToBlock(
	block *content.Block,
	directive *blockDirective,
	currentTime func() time.Time,
) error {
	switch directive.Kind {
	case directiveCancel:
		cancelErr := logseqext.SetTaskCanceled(block, currentTime())
		if cancelErr != nil {
			return fmt.Errorf("failed to set task canceled: %w", cancelErr)
		}

	case directiveWaiting:
		waitErr := logseqext.SetTaskWaiting(block)
		if waitErr != nil {
			return fmt.Errorf("failed to set task waiting: %w", waitErr)
		}

	case directiveTodo:
		todoErr := logseqext.SetTaskTodo(block)
		if todoErr != nil {
			return fmt.Errorf("failed to set task todo: %w", todoErr)
		}

	case directivePriority:
		prioErr := logseqext.SetPriority(block, directive.Priority)
		if prioErr != nil {
			return fmt.Errorf("failed to set priority: %w", prioErr)
		}
	}

	return nil
}

func kindName(kind directiveKind) string {
	switch kind {
	case directiveCancel:
		return content.TaskStringCanceled
	case directiveWaiting:
		return content.TaskStringWaiting
	case directiveTodo:
		return content.TaskStringTodo
	case directivePriority:
		return "priority"
	}

	return "unknown"
}
