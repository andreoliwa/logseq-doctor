# Tidy up

Run `lqd tidy-up path/to/page.md` to check and clean Markdown files. Back up
your graph before bulk cleanup.

Alongside existing checks for forbidden references, running tasks, double spaces,
empty bullets and tag brackets, tidy-up normalizes leading times:

```markdown
- **9h00 Meeting**
  - TODO [#A] _10;05_ Follow-up
```

becomes:

```markdown
- 09:00 **Meeting**
  - TODO [#A] 10:05 Follow-up
```

A time has one or two hour digits, a separator (`:`, `h`, `H`, `.`, `,`, or
`;`), and two minute digits. Hours range from 0 to 23; minutes from 0 to 59.
It must end at whitespace or the end of first-line content. An optional task
marker and priority may precede it. Bold and italic times are supported.

Each block, including nested blocks, is handled independently. This rule does
not split continuation lines, sort blocks, change hierarchy, resolve references,
or normalize times inside links, code, properties or logbooks. It preserves
formatting on the description. Other independent cleanup rules still apply.

Changes use the existing per-file reporting and transaction save. A changed
file returns the existing nonzero check result; a second pass on already clean
content returns zero and does not save the file. Invalid or absent times alone
do not cause a rewrite.
