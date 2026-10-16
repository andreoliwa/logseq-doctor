# Add and update tasks

```sh
lqd task add "Review pull request"
lqd task add "**9h00** Meeting" --page Work
lqd task add "10;05 Follow-up" --journal 2026-10-15 --parent "Project A"
lqd task add "09:30 Review pull request" --key "review pull request"
```

The default target is today's journal. `--page` selects a page, `--journal`
selects a journal date, and `--parent` selects a matching parent block. If the
parent is absent, the existing fallback is page-level insertion.

## Timed descriptions

Leading times accept one or two hour digits, a separator (`:`, `h`, `H`, `.`,
`,`, or `;`), and exactly two minute digits. Valid values are 00:00 through
23:59. The token must end at whitespace or the end of first-line content.
An optional priority and bold/italic formatting are supported.

New tasks normalize their time immediately: `**9h00 Meeting**` becomes
`TODO 09:00 **Meeting**`. Description formatting stays intact. No time is
generated from the current clock. Untimed new tasks still append.

Timed tasks go immediately before the first sibling with a strictly later
effective time, scanning in existing order. Notes and all task statuses count.
Equal times are skipped. Untimed siblings retain their relative order. If no
later anchor exists, the task appends. Existing disorder is not sorted.
Placement compares only page-level blocks or the selected parent's direct
children; it never changes the selected parent.

A ref-only sibling such as `((uuid))` uses its **first timed direct child** as
its effective time, skipping untimed children. Properties are allowed. This is
not the earliest clock value, and deeper descendants or the referenced source
do not supply a time. Ordinary parents and text mixed with a ref do not inherit
child times.

For example, a ref with direct children at `12:00` and `12:20` stays whole:
a new `12:10` task goes after the ref at the outer level, not between its
children. A new `12:00` task goes before a ref whose first timed child is
`12:06`. Anchors retain their original time spelling and formatting.

## Key updates

`--key` finds an existing matching task case-insensitively, including nested
tasks, and retains its existing transition back to TODO. A timed updated
description is normalized, then the **same block** moves among its actual
siblings using the rule above. Its ID, properties, children and parent stay
intact; existing status/logbook behavior still applies.

Changing `11:00 Review` to `09:30 Review` can move the task earlier. Changing it
to `13:00 Review` can move it later. Equal-time anchors are skipped, excluding
the updated block itself. Removing the time leaves the block in place. If the
key has no match, the command creates a new task in the selected scope.
