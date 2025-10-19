# Usage

Logseq Doctor provides two CLI tools: `lqdpy` (Python) and `lqd` (Go). This guide covers both.

!!! info
The Go CLI (`lqd`) is the recommended version for new features. The Python CLI (`lqdpy`) is maintained for backward compatibility but new features will only be added to the Go version.

## Python CLI (`lqdpy`)

### Overview

Type `lqdpy` without arguments to check the current commands and options:

```bash
lqdpy --help
```

Output:

```
Usage: lqdpy [OPTIONS] COMMAND [ARGS]...

Logseq Doctor: heal your flat old Markdown files before importing them.

Options:
  -g, --graph DIRECTORY           Logseq graph  [env var: LOGSEQ_GRAPH_PATH; required]
  --install-completion [bash|zsh|fish|powershell|pwsh]
                                  Install completion for the specified shell.
  --show-completion [bash|zsh|fish|powershell|pwsh]
                                  Show completion for the specified shell, to
                                  copy it or customize the installation.
  --help                          Show this message and exit.

Commands:
  outline  Convert flat Markdown to outline.
  tasks    List tasks in Logseq.
```

### Setting the Logseq Graph Path

Most commands require you to specify your Logseq graph directory. You can do this in two ways:

1. **Using the `-g` flag:**

```bash
lqdpy -g /path/to/your/logseq/graph outline input.md
```

2. **Using the environment variable:**

```bash
export LOGSEQ_GRAPH_PATH=/path/to/your/logseq/graph
lqdpy outline input.md
```

### Commands

#### `outline` - Convert Flat Markdown to Outline

Convert flat Markdown files to Logseq's outline format:

```bash
lqdpy -g /path/to/graph outline input.md
```

This command reads a flat Markdown file and converts it to Logseq's indented outline structure.

#### `tasks` - List Tasks in Logseq

List all tasks in your Logseq graph:

```bash
lqdpy -g /path/to/graph tasks
```

This will display all tasks found in your Logseq pages.

## Go CLI (`lqd`)

### Overview

Type `lqd` without arguments to check the current commands and options:

```bash
lqd --help
```

Output:

```
Logseq Doctor heals your Markdown files for Logseq.

Convert flat Markdown to Logseq outline, clean up Markdown,
prevent invalid content, and more stuff to come.

"lqdpy" is the CLI tool originally written in Python; "lqd" is the Go version.
The intention is to slowly convert everything to Go.

Usage:
  lqd [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  content     Append raw Markdown content to Logseq
  help        Help about any command
  tidy-up     Tidy up your Markdown files.

Flags:
  -h, --help   help for lqd

Use "lqd [command] --help" for more information about a command.
```

### Commands

#### `content` - Append Raw Markdown Content

Append raw Markdown content to your Logseq graph:

```bash
lqd content --help
```

This command allows you to add content to your Logseq pages programmatically.

#### `tidy-up` - Tidy Up Markdown Files

Clean up and standardize your Markdown files:

```bash
lqd tidy-up --help
```

This command helps ensure your Markdown files follow consistent formatting rules.

#### `completion` - Shell Completion

Generate autocompletion scripts for your shell:

```bash
# For bash
lqd completion bash > /etc/bash_completion.d/lqd

# For zsh
lqd completion zsh > "${fpath[1]}/_lqd"

# For fish
lqd completion fish > ~/.config/fish/completions/lqd.fish
```

## Examples

### Converting a Flat Markdown File

If you have a flat Markdown file like this:

```markdown
# My Notes

Some content here.

## Section 1

More content.

### Subsection

Details.
```

You can convert it to Logseq's outline format:

```bash
lqdpy -g /path/to/graph outline my-notes.md
```

### Tidying Up Markdown Files

To clean up and standardize your Markdown files:

```bash
lqd tidy-up /path/to/your/markdown/files
```

## Tips

!!! tip "Environment Variables"
Set `LOGSEQ_GRAPH_PATH` in your shell profile to avoid typing it every time:

    ```bash
    # Add to ~/.bashrc, ~/.zshrc, etc.
    export LOGSEQ_GRAPH_PATH=/path/to/your/logseq/graph
    ```

!!! tip "Shell Completion"
Enable shell completion for a better CLI experience. See the `completion` command above.

## Troubleshooting

### Command Not Found

If you get a "command not found" error:

1. Make sure the tool is installed (see [Installation](installation.md))
2. Check that the installation directory is in your PATH
3. Try running with the full path to the executable

### Permission Denied

If you get a "permission denied" error:

1. Make sure the executable has execute permissions: `chmod +x /path/to/lqd`
2. Check that you have write permissions to your Logseq graph directory

## Next Steps

- Explore the [CLI Reference](reference/cli.md) for detailed command documentation
- Check out [Contributing](contributing.md) if you want to help improve Logseq Doctor
