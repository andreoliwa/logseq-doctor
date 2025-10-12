# CLI Reference

This page provides detailed reference documentation for all CLI commands in Logseq Doctor.

## Python CLI (`lsdpy`)

### Global Options

```
-g, --graph DIRECTORY
```

Path to your Logseq graph directory. Can also be set via the `LOGSEQ_GRAPH_PATH` environment variable.

**Required:** Yes (for most commands)

### Commands

#### `outline`

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

#### `tasks`

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

### Shell Completion

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

## Go CLI (`lsd`)

### Global Flags

```
-h, --help
```

Show help information for any command.

### Commands

#### `content`

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

**Get detailed help:**

```bash
lsd content --help
```

#### `tidy-up`

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

**Get detailed help:**

```bash
lsd tidy-up --help
```

#### `completion`

Generate shell completion scripts.

**Usage:**

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

#### `help`

Get help about any command.

**Usage:**

```bash
lsd help [COMMAND]
```

**Example:**

```bash
lsd help tidy-up
```

## Environment Variables

### `LOGSEQ_GRAPH_PATH`

Path to your Logseq graph directory. Used by the Python CLI (`lsdpy`) as the default value for the `-g` flag.

**Example:**

```bash
export LOGSEQ_GRAPH_PATH=~/logseq/my-graph
lsdpy tasks  # No need to specify -g flag
```

## Exit Codes

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
