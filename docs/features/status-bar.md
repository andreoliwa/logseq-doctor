# Status Bar

`lqd-statusbar` is a macOS status bar app that shows how many tasks are currently `DOING` in your Logseq graph. It sits in the menu bar, polls your graph directory every 2 seconds, and lets you jump straight to the `DOING` page in Logseq from the menu.

---

## Prerequisites

- **macOS only** - `lqd-statusbar` is a Darwin-only binary (build constraint `//go:build darwin`).
- **`rg` (ripgrep) on PATH** - the app counts `DOING` tasks by running `rg` against your graph directory. Install with `brew install ripgrep`.
- **`LOGSEQ_GRAPH_PATH` set** - the app exits immediately on startup if this variable is missing.

---

## Installation

`lqd-statusbar` is built and installed separately from the main `lqd` binary:

```bash
make build-statusbar    # build only
make install-statusbar  # build and install to ~/go/bin/lqd-statusbar
```

It is also distributed alongside `lqd` in the darwin GitHub Releases (amd64 and arm64).

---

## Usage

Export the required environment variables, then launch the binary:

```bash
export LOGSEQ_GRAPH_PATH="/path/to/your/graph"
lqd-statusbar
```

The app has no interactive terminal output - it runs entirely in the macOS menu bar.

### Environment variables

| Variable            | Required | Default | Description                                  |
| ------------------- | -------- | ------- | -------------------------------------------- |
| `LOGSEQ_GRAPH_PATH` | Yes      | -       | Path to your Logseq graph root directory     |
| `LQD_SERVE_PORT`    | No       | `8091`  | Port used to build the "Open Backlog UI" URL |

### Starting with lqd dashboard

If you have `lqd-statusbar` installed, you can start it automatically alongside the dashboard:

```bash
lqd dashboard --status
```

The status bar process is tied to the dashboard - it is killed automatically when the dashboard exits (Ctrl+C or normal return). If `lqd-statusbar` is not on PATH, the dashboard starts normally with a warning to stderr.

---

## Status bar icons

| Icon   | Meaning                                  |
| ------ | ---------------------------------------- |
| `🛑`   | No `DOING` tasks found (or `rg` errored) |
| `🟢 N` | N tasks currently marked `DOING`         |

The count updates every 2 seconds. If `rg` fails for any reason (missing binary, unreadable path), the icon falls back to `🛑` silently.

---

## Menu items

| Item                     | Action                                                                                    |
| ------------------------ | ----------------------------------------------------------------------------------------- |
| **Open DOING in Logseq** | Opens `logseq://graph/<name>?page=DOING` in the Logseq app via `open`                     |
| **Open Backlog UI**      | Opens `http://localhost:<port>` in the default browser (`LQD_SERVE_PORT`, default `8091`) |
| **Quit**                 | Stops polling and exits the status bar app                                                |

The graph name in the deep-link URL is always the last path component of `LOGSEQ_GRAPH_PATH`.

---

## Notes

- **macOS only** - the binary will not compile or run on Linux or Windows.
- **Requires `rg`** - if ripgrep is not on PATH, all polls return 0 and the icon stays `🛑` with no error shown.
- **Poll interval** - fixed at 2 seconds; there is no configuration option to change it.
- **No authentication** - reads `.md` files directly from disk; does not connect to the Logseq HTTP API or PocketBase.
