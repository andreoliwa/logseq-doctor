# Raw journal content

Pipe Markdown into `lqd content --journal YYYY-MM-DD`. The default is today's journal.

```sh
printf '%s\n' '- **9h00** Call [[quick capture]]' '- 11:00 Review' |
  lqd content --journal 2026-10-16
```

Explicit root dash bullets form a batch. Each root is placed before the first
existing sibling with a strictly later time, scanning the current list in order.
Equal times retain arrival order; untimed roots append. Existing disorder is not
sorted. Children, properties and continuation lines remain attached to their root.
An untimed heading with timed children stays one group.

A ref-only sibling uses its first timed direct child's own time, skipping untimed
children. It remains one unit: a ref with children at 12:00 and 12:20 precedes a
new outer 12:10 note. References are not resolved and grandchildren are not used.

Existing and incoming source bytes, links, literal brackets, code and line endings
are retained. Only needed insertion separators are added. Time spelling, emphasis
and task status are not normalized; plain notes do not become TODOs. Use
[`tidy-up`](tidy-up.md) when you want explicit normalization.

Unoutlined text keeps raw append behavior. Empty input does nothing. The command
prepares and validates the result before replacing the journal, preserves existing
permissions, and reports failure without a partial journal update.

For example, a script can pipe meeting notes as timed root bullets, with agenda
items and follow-up notes as children. Each meeting enters the journal timeline
without a shared heading, and its supporting content stays attached.

If preserving indentation would turn a root into a child at an insertion boundary,
the command rejects the unsafe join and leaves the journal unchanged. It does not
reindent either source. Page-level preambles stay at page level.
