# Markdown insertion

`lqd md` reads stdin and inserts structured Markdown through a transaction. It
uses today's journal by default. Choose another journal with `--journal DATE`,
a named page with `--page`, or direct children with `--parent TEXT`. A missing
parent falls back to page roots.

```bash
printf -- '- **9h00** Call\n- 11:00 Review\n' | lqd md --journal 2026-10-16
echo '10h00 Meeting' | lqd md --page Work --parent Project
```

## Placement and input shape

An explicit dash-bulleted outline supplies independent root blocks. Each root is
placed in input order against the evolving sibling list, without an extra parent.
Its children, properties and continuation lines stay attached. An untimed heading
with timed children stays one group and appends.

Unbulleted input stays one block, including supporting paragraphs and child
bullets. For example, this places one call block, not three timeline entries:

```bash
printf '09:00 Call\n\nSupporting notes\n\n- Follow up\n' | lqd md --page Work
```

A recognized leading time places a new block before the first strictly later
sibling in stored order. Equal times keep arrival order; untimed blocks append.
Existing siblings are not sorted. Only the selected parent's direct children or
the selected page/journal's root blocks participate.

Notes and tasks of all supported statuses participate. A ref-only sibling uses
its first timed direct child, skipping untimed children, rather than its earliest
time or a deeper descendant. A ref with children at `12:00` and `12:20` stays
intact before an incoming root at `12:10`. Ordinary headings do not inherit child
times, and references are not resolved.

The command retains supplied time spelling, emphasis and task statuses. It does
not generate TODO markers or normalize times. Run `lqd tidy-up` for explicit time
normalization. Unlike `lqd content`, this path serializes structured Markdown and
does not promise raw byte preservation.

## Key updates

```bash
echo 'DOING **9h30** feature-123' | lqd md --page Work --key feature-123
```

`--key` finds an existing block case-insensitively, recursively within the selected
scope. It replaces visible content using the existing single-block update
contract, not batch insertion. Existing properties, IDs, children and logbooks
survive. Supplied task status follows normal Markdown content replacement.

A new leading time moves the same block among its actual siblings, excluding
itself as an anchor. A nested match stays under its actual parent, even without
`--parent`. Untimed updates stay in place. If the key is absent, normal creation
rules apply.
