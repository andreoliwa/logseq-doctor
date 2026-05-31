# Usage

Logseq Doctor provides a single Go CLI tool: `lqd`.

## Commands

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

#### `outline` - Convert Flat Markdown to Logseq Outline

Convert flat Markdown files to Logseq's bullet outline format:

```bash
# From stdin
printf '# Header\n\nParagraph.\n' | lqd outline

# Single file to stdout
lqd outline notes.md

# Multiple files to stdout
lqd outline a.md b.md

# Edit in place
lqd outline --in-place notes.md

# Move converted file to a directory (fails if destination already exists)
lqd outline --move-to ~/logseq/pages notes.md

# Preserve blank lines as empty bullet lines
lqd outline --keep-breaks notes.md
```

Ordered lists are converted to Logseq's native ordered list format using the
`logseq.order-list-type:: number` block property. The conversion is idempotent:
running it twice produces the same result.

#### `task ls` - List Tasks

List tasks from your Logseq graph via the HTTP API:

```bash
# All active tasks
lqd task ls

# Filter by tag or page
lqd task ls work

# Include completed tasks
lqd task ls --done --canceled
lqd task ls --completed   # shorthand for both

# JSON output for scripting
lqd task ls --json

# Print the Datalog query before results
lqd task ls -v work
```

Requires `LOGSEQ_HOST_URL` and `LOGSEQ_API_TOKEN` environment variables.

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
lqd outline my-notes.md
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
