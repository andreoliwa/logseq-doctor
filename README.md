# Overview

|         |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| docs    | [![Documentation Status](https://readthedocs.org/projects/logseq-doctor/badge/?style=flat)](https://logseq-doctor.readthedocs.io/)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| tests   | [![Go Build Status](https://github.com/andreoliwa/logseq-doctor/actions/workflows/go.yaml/badge.svg)](https://github.com/andreoliwa/logseq-doctor/actions) [![Tox Build Status](https://github.com/andreoliwa/logseq-doctor/actions/workflows/tox.yaml/badge.svg)](https://github.com/andreoliwa/logseq-doctor/actions) [![Coverage Status](https://codecov.io/gh/andreoliwa/logseq-doctor/branch/master/graphs/badge.svg?branch=master)](https://codecov.io/github/andreoliwa/logseq-doctor)                                                                                                                                                                                                 |
| package | [![PyPI Package latest release](https://img.shields.io/pypi/v/logseq-doctor.svg)](https://pypi.org/project/logseq-doctor) [![PyPI Wheel](https://img.shields.io/pypi/wheel/logseq-doctor.svg)](https://pypi.org/project/logseq-doctor) [![Supported versions](https://img.shields.io/pypi/pyversions/logseq-doctor.svg)](https://pypi.org/project/logseq-doctor) [![Supported implementations](https://img.shields.io/pypi/implementation/logseq-doctor.svg)](https://pypi.org/project/logseq-doctor) [![Commits since latest release](https://img.shields.io/github/commits-since/andreoliwa/logseq-doctor/v0.3.0.svg)](https://github.com/andreoliwa/logseq-doctor/compare/v0.3.0...master) |

Logseq Doctor: heal your flat old Markdown files before importing them.

> [!NOTE]
> This project is still alpha, so it\'s very rough on the edges
> (documentation and feature-wise).
>
> At the moment, it has both a Python and Go CLI.
>
> The long-term plan is to convert it to Go and slowly remove Python.
> New features will be added to the Go CLI only.

## Installation

### Python executable

The recommended way is to install `logseq-doctor` globally with
[pipx](https://github.com/pypa/pipx):

    pipx install logseq-doctor

You can also install the development version with:

    pipx install git+https://github.com/andreoliwa/logseq-doctor

You will then have the `lsdpy` command available globally in your system.

### Go binary executable

The recommended way for macOS and Linux is to install with Homebrew:

    brew install andreoliwa/formulae/logseq-doctor

Or you can install manually:

    go install github.com/andreoliwa/logseq-doctor@latest

Confirm if it\'s in your path:

    which lsd
    # or
    ls -l $(go env GOPATH)/bin/lsd

### Build from source

To build and install from the source (both Python and Go executables), clone the repository and run:

    make install

## Quick start

Type `lsdpy` without arguments to check the current commands and options:

    Usage: lsdpy [OPTIONS] COMMAND [ARGS]...

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

Type `lsd` (the Go executable) without arguments to check the current commands and options:

    Logseq Doctor (Go) heals your Markdown files for Logseq.

    Convert flat Markdown to Logseq outline, clean up Markdown,
    prevent invalid content, and more stuff to come.

    "lsdpy"" is the CLI tool originally written in Python; "lsd"" is the Go version.
    The intention is to slowly convert everything to Go.

    Usage:
    lsd [command]

    Available Commands:
    completion  Generate the autocompletion script for the specified shell
    content     Append raw Markdown content to Logseq
    help        Help about any command
    tidy-up     Tidy up your Markdown files.

    Flags:
    -h, --help   help for lsd

    Use "lsd [command] --help" for more information about a command.

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
