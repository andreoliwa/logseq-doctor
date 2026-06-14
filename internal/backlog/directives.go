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

// detectDirective inspects the BlockRef's preceding sibling in its parent Paragraph.
// Returns a populated blockDirective if a recognized directive node is found, nil otherwise.
func detectDirective(blockRef *content.BlockRef) *blockDirective {
	prev := blockRef.PreviousSibling()
	if prev == nil {
		return nil
	}

	// blockRef lives inside a Paragraph inside a Block.
	var backlogBlock *content.Block

	if para, ok := blockRef.Parent().(*content.Paragraph); ok {
		if block, ok := para.Parent().(*content.Block); ok {
			backlogBlock = block
		}
	}

	switch node := prev.(type) {
	case *content.TaskMarker:
		switch node.Status { //nolint:exhaustive // only CANCELED/WAITING are valid directives; others are ignored
		case content.TaskStatusCanceled, content.TaskStatusCancelled:
			return &blockDirective{
				UUID:          blockRef.ID,
				Kind:          directiveCancel,
				Priority:      content.PriorityNone,
				DirectiveNode: node,
				BacklogBlock:  backlogBlock,
			}
		case content.TaskStatusWaiting, content.TaskStatusWait:
			return &blockDirective{
				UUID:          blockRef.ID,
				Kind:          directiveWaiting,
				Priority:      content.PriorityNone,
				DirectiveNode: node,
				BacklogBlock:  backlogBlock,
			}
		}
	case *content.Priority:
		return &blockDirective{
			UUID:          blockRef.ID,
			Kind:          directivePriority,
			Priority:      node.Priority,
			DirectiveNode: node,
			BacklogBlock:  backlogBlock,
		}
	}

	return nil
}

// applyDirectives processes all collected directives: modifies the real task block on disk,
// then strips the directive node from the backlog page.
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
	applied := false

	for i := range directives {
		directive := &directives[i]

		err := applySingleDirective(graph, logseqAPI, directive, currentTime)
		if err != nil {
			color.Yellow("[backlog] WARNING: directive %s on block %s: %v",
				kindName(directive.Kind), directive.UUID, err)

			continue
		}

		// Strip the directive node from the backlog page AST.
		// The caller's transaction.Save() will persist this change.
		directive.DirectiveNode.RemoveSelf()

		if directive.BacklogBlock != nil {
			if directive.Kind == directiveCancel {
				// Cancel directives remove the task from the backlog immediately —
				// a canceled task no longer belongs in any backlog section.
				directive.BacklogBlock.RemoveSelf()
			} else {
				// Remove any stale properties left on the backlog ref block (e.g. cancelled:: added manually).
				// If the Properties node becomes empty after removal, remove it entirely to avoid a blank line.
				props := logseqext.BlockProperties(directive.BacklogBlock)
				props.Remove(logseqext.PropertyCancelled)

				if props.FirstChild() == nil {
					props.RemoveSelf()
				}
			}
		}

		applied = true
	}

	return applied
}

func applySingleDirective(
	graph *logseq.Graph,
	logseqAPI logseqapi.LogseqAPI,
	directive *blockDirective,
	currentTime func() time.Time,
) error {
	block, transaction, err := logseqapi.FindBlockOnDisk(graph, logseqAPI, directive.UUID)
	if err != nil {
		return fmt.Errorf("finding block on disk: %w", err)
	}

	applyErr := applyDirectiveToBlock(block, directive, currentTime)
	if applyErr != nil {
		return applyErr
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
	case directivePriority:
		return "priority"
	}

	return "unknown"
}
