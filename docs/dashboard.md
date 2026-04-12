# Dashboard Guide

The Logseq Doctor dashboard is a lightweight web UI for browsing, filtering, reordering, and deep-linking tasks across all your backlogs. It talks directly to PocketBase — no separate sync layer required.

---

## Prerequisites

Before starting the dashboard you need:

1. **PocketBase running** — the dashboard manages PocketBase as a subprocess, but PocketBase itself must be installed and on your `PATH`. Download from [pocketbase.io](https://pocketbase.io).

2. **At least one sync completed** — run `lqd sync` to populate PocketBase with your tasks:

   ```bash
   lqd sync
   ```

   On a fresh PocketBase instance, run `lqd sync --init` the first time to create the collection schema.

3. **Environment variables set:**

   | Variable              | Description                                           |
   | --------------------- | ----------------------------------------------------- |
   | `POCKETBASE_URL`      | PocketBase base URL (default `http://127.0.0.1:8090`) |
   | `POCKETBASE_USERNAME` | PocketBase admin email                                |
   | `POCKETBASE_PASSWORD` | PocketBase admin password                             |
   | `LOGSEQ_GRAPH_PATH`   | Path to your Logseq graph root                        |

---

## Starting the Dashboard

```bash
lqd dashboard
# or the short alias:
lqd dash
```

This command:

1. Starts PocketBase as a managed subprocess (stdout/stderr appear in the same terminal).
2. Waits up to 5 seconds for PocketBase to be ready.
3. Starts the backlog UI at `http://localhost:8091`.
4. On macOS, opens the browser automatically.

To use a different port:

```bash
lqd dash --port 9000
# or via env var:
LQD_SERVE_PORT=9000 lqd dash
```

Press `Ctrl-C` to stop both the UI server and PocketBase together.

---

## The Filter Bar

The filter bar is sticky — it stays visible as you scroll through the task list. It has two rows.

### Row 1: Scope and search

| Control                   | What it does                                                                              |
| ------------------------- | ----------------------------------------------------------------------------------------- |
| **Quick filter** dropdown | Pre-configured filter presets (see below)                                                 |
| **Backlog** dropdown      | Narrow to a single backlog or view all                                                    |
| **Search** field          | Free-text filter on task name (space-separated terms are AND; `term1 OR term2` works too) |
| **Tags** field            | Multi-select tag filter with autocomplete chips (AND/OR toggle on the right)              |

### Row 2: Status and date type

**Status checkboxes** (left side): `TODO`, `DOING`, `WAITING` are checked by default. Uncheck any to hide those tasks.

**Date type checkboxes** (right side):

| Checkbox    | When checked, shows...                                  |
| ----------- | ------------------------------------------------------- |
| Overdue     | Tasks past their scheduled or deadline date             |
| Scheduled   | Tasks with a `scheduled::` date that are not overdue    |
| Deadline    | Tasks with a `deadline::` date that are not overdue     |
| No due date | Tasks with neither scheduled nor deadline               |
| Journal     | Tasks that originated from a journal (daily note) page  |
| Page        | Tasks that originated from a regular (non-journal) page |

> **Overdue rescue:** Overdue tasks always appear in the Overdue column even if the Scheduled, Deadline, Journal, or Page checkboxes are unchecked. Unchecking Overdue explicitly hides them.

**Clear filters** resets all controls to their defaults.

### Quick filters

| Option                       | Effect                                                                               |
| ---------------------------- | ------------------------------------------------------------------------------------ |
| _(none)_                     | No preset applied — individual controls in full effect                               |
| **Overdue**                  | Resets all filters, then shows only overdue tasks                                    |
| **Waiting**                  | Resets all filters, then shows only WAITING tasks                                    |
| **Waiting without due date** | Resets all filters, then shows only WAITING tasks with no scheduled or deadline date |

Selecting a quick filter resets all other filters first, then applies its own subset. Changing any individual control after selecting a quick filter clears the quick filter selection automatically.

---

## Ranked Table

Tasks with a rank assigned appear in the top table, grouped by backlog and ordered by rank.

### Reading the table

- **Position** (the `#` column): sequential 1, 2, 3… within each backlog — derived from sort order, not the stored sparse rank value.
- **Backlog** column: backlog name (links to the Logseq page) plus the ⤵️ action on hover.
- **Overdue / Scheduled / Deadline / Journal / Sort date / Groomed** columns: dates associated with the task.

### Reordering tasks

Drag the ⠿ handle on the left to reorder tasks within a backlog. The rank is written to PocketBase immediately (one PATCH request in the common case). Run `lqd backlog` to propagate the new order back to your Logseq `.md` files.

### Moving tasks to unranked

Hover a ranked row to reveal the ⤵️ link in the Backlog column. Clicking it moves that task **and all tasks ranked below it** in the same backlog to the unranked table. A confirmation dialog shows how many tasks will be affected.

---

## Unranked Table

Tasks with no rank appear in the bottom table. They belong to a backlog but have not been manually ordered.

### Sorting

The sort bar above the unranked table lets you build up to 4 sort criteria. Each criterion is a `(field, ASC/DESC)` pair. Criteria can be reordered by dragging. Available fields:

`journal` · `name` · `tags` · `status` · `sort_date` · `groomed` · `backlog_name`

Default sort: `journal ASC` (oldest task first).

---

## URL Bookmarking

All filter and sort settings are written to the URL as query parameters. Copy the URL to bookmark a view or share a deep link.

| Parameter    | Stores                                                           |
| ------------ | ---------------------------------------------------------------- |
| `q`          | Search text                                                      |
| `backlog`    | Selected backlog name                                            |
| `quick`      | Quick filter value (`overdue`, `waiting`, `waiting-no-due-date`) |
| `tags`       | Comma-separated selected tags                                    |
| `tagMode`    | `AND` or `OR`                                                    |
| `status`     | Comma-separated checked status values                            |
| `dateHidden` | Comma-separated **unchecked** date type values                   |
| `sort`       | Sort criteria, e.g. `journal:ASC,name:DESC`                      |

The URL updates 400 ms after you stop interacting with any filter control. The browser title also updates to reflect active filters so bookmarks are self-describing.

---

## Deep Links into Logseq

Each task name in both tables has a 🔗 icon on hover. Clicking it opens the task's block directly in Logseq using the `logseq://` protocol.

---

## Workflow

A typical daily workflow:

```bash
# 1. Sync latest tasks from Logseq files into PocketBase
lqd sync

# 2. Open the dashboard
lqd dash

# 3. Browse, filter, and reorder tasks in the UI

# 4. Propagate rank changes back to Logseq files
lqd backlog
```

---

## Troubleshooting

**"PocketBase not found"** — ensure `pocketbase` is on your `PATH`. Check with `which pocketbase`.

**Dashboard shows no tasks** — run `lqd sync` first. If the collection doesn't exist yet, run `lqd sync --init`.

**Rank changes disappear after `lqd sync`** — this is expected behavior: `lqd sync` never overwrites ranks that were set via the UI. If ranks are being reset, check that you are not running `lqd sync --init`, which drops and recreates the collection.

**Browser doesn't open automatically** — navigate manually to `http://localhost:8091`. The auto-open only runs on macOS.
