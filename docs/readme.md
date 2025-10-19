# Logseq Doctor

<!-- --8<-- [start:badges] -->

[![Documentation](https://img.shields.io/badge/docs-mkdocs-blue)](https://andreoliwa.github.io/logseq-doctor/)
[![Go Build Status](https://github.com/andreoliwa/logseq-doctor/actions/workflows/go.yaml/badge.svg)](https://github.com/andreoliwa/logseq-doctor/actions)
[![Tox Build Status](https://github.com/andreoliwa/logseq-doctor/actions/workflows/tox.yaml/badge.svg)](https://github.com/andreoliwa/logseq-doctor/actions)
[![Coverage Status](https://codecov.io/gh/andreoliwa/logseq-doctor/branch/master/graphs/badge.svg?branch=master)](https://codecov.io/github/andreoliwa/logseq-doctor)
[![PyPI Package](https://img.shields.io/pypi/v/logseq-doctor.svg)](https://pypi.org/project/logseq-doctor)

<!-- --8<-- [end:badges] -->

<!-- --8<-- [start:tagline] -->

**Logseq Doctor: heal your flat old Markdown files before importing them to Logseq.**

<!-- --8<-- [end:tagline] -->

📚 **[Read the full documentation](https://andreoliwa.github.io/logseq-doctor/)**

<!-- --8<-- [start:status] -->

> [!NOTE]
> This project is still alpha, so it's very rough on the edges
> (documentation and feature-wise).
>
> At the moment, it has both a Python and Go CLI.
>
> The long-term plan is to convert it to Go and slowly remove Python.
> New features will be added to the Go CLI only.

<!-- --8<-- [end:status] -->

<!-- --8<-- [start:description] -->

## What is Logseq Doctor?

Logseq Doctor is a command-line tool with commands to manipulate your [Logseq](https://logseq.com/) Markdown files. It provides utilities to:

- Convert flat Markdown to Logseq's outline format
- Append content to pages and journals
- Create task backlogs that are easily viewed and prioritized in the mobile app
- Manage tasks in Logseq
- Clean up and tidy Markdown files
- Prevent invalid content to be committed
- And more stuff to come...
<!-- --8<-- [end:description] -->

<!-- --8<-- [start:features] -->

## Features

- **Backlog Management** (`backlog`): Aggregate tasks from multiple pages into unified backlogs with smart categorization, overdue detection, and focus page generation
- **Content Management** (`content`): Append raw Markdown content to Logseq pages or journals
- **Markdown Integration** (`md`): Parse and add Markdown content using DOM manipulation with support for parent blocks and journal targeting
- **Task Management** (`task add`): Add new tasks or update existing ones with key-based search, preserving children and properties
- **Tidy Up** (`tidy-up`): Clean up and standardize your Markdown files
- **Fast Performance**: Written in Go for speed and efficiency
- **Outline Conversion** (`outline`): Convert flat Markdown files to Logseq's outline format
- **Task Listing** (`tasks`): List and manage tasks in your Logseq graph
<!-- --8<-- [end:features] -->

<!-- --8<-- [start:installation] -->

## Installation

### Go binary executable

The recommended way for macOS and Linux is to install with Homebrew:

    brew install andreoliwa/formulae/logseq-doctor

Or you can install manually:

    go install github.com/andreoliwa/logseq-doctor@latest

Confirm if it's in your path:

    which lqd
    # or
    ls -l $(go env GOPATH)/bin/lqd

### Python executable

The recommended way is to install `logseq-doctor` globally with
[pipx](https://github.com/pypa/pipx):

    pipx install logseq-doctor

You can also install the development version with:

    pipx install git+https://github.com/andreoliwa/logseq-doctor

You will then have the `lqdpy` command available globally in your system.

### Build from source

To build and install from the source (both Python and Go executables), clone the repository and run:

    make install

<!-- --8<-- [end:installation] -->

<!-- --8<-- [start:quickstart] -->

## Quick start

Type `lqd` (the Go executable) without arguments to check the current commands and options:

    Logseq Doctor heals your Markdown files for Logseq.

    Convert flat Markdown to Logseq outline, clean up Markdown,
    prevent invalid content, and more stuff to come.

    "lqdpy" is the CLI tool originally written in Python; "lqd" is the Go version.
    The intention is to slowly convert everything to Go.

    Usage:
    lqd [command]

    Available Commands:
    backlog     Aggregate tasks from multiple pages into a unified backlog
    completion  Generate the autocompletion script for the specified shell
    content     Append raw Markdown content to Logseq
    help        Help about any command
    md          Add Markdown content to Logseq using the DOM
    task        Manage tasks in Logseq (subcommands: add)
    tidy-up     Tidy up your Markdown files

    Flags:
    -h, --help   help for lqd

    Use "lqd [command] --help" for more information about a command.

Type `lqdpy` without arguments to check the current commands and options:

    Usage: lqdpy [OPTIONS] COMMAND [ARGS]...

    Logseq Doctor: heal your flat old Markdown files before importing them.

    Options:
    -g, --graph DIRECTORY           Logseq graph  [env var: LOGSEQ_GRAPH_PATH;
    required]
    --install-completion [bash|zsh|fish|powershell|pwsh]
    Install completion for the specified shell.
    --show-completion [bash|zsh|fish|powershell|pwsh]
    Show completion for the specified shell, to
    copy it or customize the installation.
    --help                          Show this message and exit.

    Commands:
    outline  Convert flat Markdown to outline.
    tasks    List tasks in Logseq.

<!-- --8<-- [end:quickstart] -->

<!-- --8<-- [start:development] -->

## Development

To set up local development:

    make setup

Run this to see help on all available targets:

    make

To run all the tests run:

    tox

Note, to combine the coverage data from all the tox environments run:

| OS      |                                     |
| ------- | ----------------------------------- |
| Windows | set PYTEST_ADDOPTS=--cov-append tox |
| Other   | PYTEST_ADDOPTS=--cov-append tox     |

<!-- --8<-- [end:development] -->
