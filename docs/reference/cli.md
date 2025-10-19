# CLI Reference

This page provides detailed reference documentation for all CLI commands in Logseq Doctor.

## Commands

### `backlog`

Aggregate tasks from multiple pages into a unified backlog.

**Usage:**

```bash
lsd backlog [partial page names]
```

**Description:**

This command aggregates tasks from one or more pages into a unified backlog page. It reads configuration from a "backlog" page in your graph, where each line with page references or tags defines a separate backlog.

**Features:**

- **Automatic task aggregation**: Queries tasks from specified pages using the Logseq API
- **Smart categorization**: Organizes tasks into sections (Focus, Overdue, New tasks, Scheduled)
- **Concurrent processing**: Processes multiple pages in parallel for better performance
- **Obsolete task removal**: Automatically removes tasks that no longer exist in source pages
- **Pin support**: Tasks marked with 📌 emoji are preserved and not removed
- **Overdue detection**: Identifies and highlights tasks past their deadline or scheduled date
- **Focus page**: Aggregates focus tasks from all backlogs into a central focus page

**Configuration:**

Create a page named "backlog" with lines containing page references or tags. The first page reference determines the backlog page name, and all referenced pages/tags are used as input sources.

**Example:**

```bash
# Process all backlogs
lsd backlog

# Process only backlogs matching "work"
lsd backlog work

# Process multiple specific backlogs
lsd backlog computer house
```

**Environment Variables:**

This command uses the following [environment variables](#environment-variables): `LOGSEQ_GRAPH_PATH`, `LOGSEQ_HOST_URL`, and `LOGSEQ_API_TOKEN`.

### `completion`

Generate shell completion scripts.

**Usage (Go CLI):**

```bash
lsd completion [bash|zsh|fish|powershell]
```

**Description:**

Generate autocompletion scripts for your shell. This enables tab completion for commands and options.

**Examples:**

**Bash:**

```bash
lsd completion bash > /etc/bash_completion.d/lsd
```

**Zsh:**

```bash
lsd completion zsh > "${fpath[1]}/_lsd"
```

**Fish:**

```bash
lsd completion fish > ~/.config/fish/completions/lsd.fish
```

**PowerShell:**

```bash
lsd completion powershell > lsd.ps1
```

**Usage (Python CLI):**

Install shell completion for easier command-line usage:

**Bash:**

```bash
lsdpy --install-completion bash
```

**Zsh:**

```bash
lsdpy --install-completion zsh
```

**Fish:**

```bash
lsdpy --install-completion fish
```

**PowerShell:**

```bash
lsdpy --install-completion powershell
```

### `content`

Append raw Markdown content to your Logseq graph.

**Usage:**

```bash
lsd content [OPTIONS]
```

**Description:**

This command allows you to programmatically add content to your Logseq pages. It's useful for automation and scripting workflows where you want to append content to your knowledge base.

**Example:**

```bash
echo "- New note" | lsd content
```

### `md`

Add Markdown content to Logseq using the DOM.

**Usage:**

```bash
lsd md [OPTIONS]
```

**Description:**

This command reads Markdown content from stdin and adds it to your Logseq graph by parsing and manipulating the document object model (DOM). Unlike the `content` command which appends raw text, `md` parses the Markdown structure and properly integrates it into Logseq's outline format.

**Features:**

- **Markdown parsing**: Parses input as Markdown and converts it to Logseq's block structure
- **Parent block support**: Can add content as a child of a specific block using `--parent`
- **Journal targeting**: Supports adding to specific journal dates with `--journal`
- **Multi-line content**: Handles complex Markdown including tasks with properties and logbooks
- **Smart placement**: Adds to journal page by default, or under a parent block if specified

**Flags:**

```
-j, --journal YYYY-MM-DD    Journal date (default: today)
-p, --parent TEXT           Partial text of a block to use as parent
```

**Examples:**

```bash
# Add a simple note to today's journal
echo "New task" | lsd md

# Add content under a specific parent block
echo "Child task" | lsd md --parent "Project A"

# Add to a specific journal date
echo "Meeting notes" | lsd md --journal 2024-12-25

# Add a task with properties and logbook
echo "DOING Some task
collapsed:: true
:LOGBOOK:
CLOCK: [2025-08-27 Wed 21:12:50]
:END:" | lsd md --parent "work"
```

### `outline`

Convert flat Markdown files to Logseq's outline format.

**Usage:**

```bash
lsdpy -g /path/to/graph outline [OPTIONS] INPUT_FILE
```

**Arguments:**

- `INPUT_FILE`: Path to the Markdown file to convert

**Description:**

This command reads a flat Markdown file and converts it to Logseq's indented outline structure. It processes headings and content to create a hierarchical outline that works well with Logseq's outliner interface.

**Example:**

```bash
lsdpy -g ~/logseq/my-graph outline notes.md
```

### `task`

Manage tasks in your Logseq graph.

**Usage:**

```bash
lsd task [subcommand]
```

**Description:**

Parent command for task management operations. Use subcommands to add, list, or modify tasks in your Logseq graph.

**Subcommands:**

#### `task add`

Add a new task to Logseq or update an existing one.

**Usage:**

```bash
lsd task add [task description] [OPTIONS]
```

**Description:**

Adds a new TODO task to your Logseq graph. By default, tasks are added to today's journal page. You can specify a different target using flags.

**Features:**

- **Key-based updates**: Use `--key` to search for and update existing tasks (case-insensitive)
- **Flexible targeting**: Add to journal pages, regular pages, or under specific blocks
- **Smart search**: When using `--key`, searches within the specified page/journal or block scope
- **Preserve structure**: When updating tasks, preserves children, properties, and logbook entries

**Flags:**

```
-j, --journal YYYY-MM-DD    Journal date (default: today)
-p, --page NAME             Page name to add the task to
-b, --block TEXT            Partial text of a block to use as parent
-k, --key TEXT              Unique key to search for existing task
```

**Examples:**

```bash
# Add a task to today's journal
lsd task add "Review pull request"

# Add a task to a specific page
lsd task add "Call client" --page "Work"

# Add a task to a specific journal date
lsd task add "Buy groceries" --journal 2024-12-25

# Update an existing task by key (preserves children and properties)
lsd task add "Water plants in living room" --key "water plants"

# Add a task under a specific block
lsd task add "Meeting notes" --block "Project A"

# Update a task within a specific block scope
lsd task add "Updated task name" --block "Project A" --key "task"
```

### `tasks`

List all tasks in your Logseq graph.

**Usage:**

```bash
lsdpy -g /path/to/graph tasks [OPTIONS]
```

**Description:**

This command scans your Logseq graph and lists all tasks found in your pages. It helps you get an overview of your TODO items across all your notes.

**Example:**

```bash
lsdpy -g ~/logseq/my-graph tasks
```

### `tidy-up`

Clean up and standardize your Markdown files.

**Usage:**

```bash
lsd tidy-up [OPTIONS] [FILES...]
```

**Description:**

This command helps ensure your Markdown files follow consistent formatting rules. It performs various cleanup operations:

- Removes double spaces (except in tables)
- Removes empty bullets
- Removes unnecessary brackets from tags
- Standardizes formatting

**Example:**

```bash
lsd tidy-up /path/to/markdown/files/*.md
```

## Global Flags

**Go CLI (`lsd`):**

```
-h, --help
```

Show help information for any command.

**Python CLI (`lsdpy`):**

```
-g, --graph DIRECTORY
```

Path to your Logseq graph directory. Can also be set via the `LOGSEQ_GRAPH_PATH` environment variable.

**Required:** Yes (for most commands)

## Environment Variables

### `LOGSEQ_GRAPH_PATH`

Path to your Logseq graph directory. Used by the Python CLI (`lsdpy`) as the default value for the `-g` flag, and by the `backlog` command to locate your graph.

**Example:**

```bash
export LOGSEQ_GRAPH_PATH=~/logseq/my-graph
lsdpy tasks  # No need to specify -g flag
```

### `LOGSEQ_HOST_URL`

Logseq API host URL. Used by the `backlog` command to connect to the Logseq API.

**Default:** `http://localhost:12315`

**Example:**

```bash
export LOGSEQ_HOST_URL=http://localhost:12315
lsd backlog
```

### `LOGSEQ_API_TOKEN`

Logseq API authentication token. Required by the `backlog` command to authenticate with the Logseq API.

**Example:**

```bash
export LOGSEQ_API_TOKEN=your-api-token-here
lsd backlog
```

## Exit Code

Both CLIs use standard exit codes:

- `0`: Success
- `1`: General error
- `2`: Command-line usage error

## Tips

!!! tip "Combining Commands"
You can combine multiple commands using shell pipes and scripting:

    ```bash
    # Example: Tidy up all Markdown files in a directory
    find ~/notes -name "*.md" -exec lsd tidy-up {} \;
    ```

!!! tip "Automation"
Use the Go CLI (`lsd`) in scripts and automation workflows for better performance:

    ```bash
    #!/bin/bash
    # Daily note automation
    echo "- $(date): Daily standup notes" | lsd content
    ```

!!! tip "Shell Aliases"
Create shell aliases for frequently used commands:

    ```bash
    # Add to ~/.bashrc or ~/.zshrc
    alias lsd-tasks='lsdpy -g ~/logseq/my-graph tasks'
    alias lsd-tidy='lsd tidy-up'
    ```
