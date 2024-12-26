# Overview

+---------+-----------------------------------------------------------+
| docs | [![Documentation                                          |
|         |  Status](https://readthedocs.org/projects/logseq-doctor/b |
| | adge/?style=flat)](https://logseq-doctor.readthedocs.io/) |
+---------+-----------------------------------------------------------+
| tests | | [![Git                                                  |
|         | Hub Actions Build Status](https://github.com/andreoliwa/l |
| | ogseq-doctor/actions/workflows/github-actions.yml/badge.s |
| | vg)](https://github.com/andreoliwa/logseq-doctor/actions) |
| | | [![Coverage Status](https://codecov.io/gh/andreoliw |
| | a/logseq-doctor/branch/master/graphs/badge.svg?branch=mas |
| | ter)](https://codecov.io/github/andreoliwa/logseq-doctor) |
+---------+-----------------------------------------------------------+
| package | | [![PyPI                                                 |
|         |  Package latest release](https://img.shields.io/pypi/v/lo |
| | gseq-doctor.svg)](https://pypi.org/project/logseq-doctor) |
| | [![PyPI Wheel](https://img.shields.io/pypi/wheel/lo |
| | gseq-doctor.svg)](https://pypi.org/project/logseq-doctor) |
| | [![Supp                                                 |
|         | orted versions](https://img.shields.io/pypi/pyversions/lo |
| | gseq-doctor.svg)](https://pypi.org/project/logseq-doctor) |
| | [![Supported imple                                      |
|         | mentations](https://img.shields.io/pypi/implementation/lo |
| | gseq-doctor.svg)](https://pypi.org/project/logseq-doctor) |
| | | [![Commits s                                            |
|         | ince latest release](https://img.shields.io/github/commit |
| | s-since/andreoliwa/logseq-doctor/v0.3.0.svg)](https://git |
| | hub.com/andreoliwa/logseq-doctor/compare/v0.3.0...master) |
+---------+-----------------------------------------------------------+

Logseq Doctor: heal your flat old Markdown files before importing them.

**NOTE:** This project is still alpha, so it\'s very rough on the edges
(documentation and feature-wise).

At the moment, it has a Python package shipped with a Rust module, plus
an external Go executable with recent additions.

The long-term plan is to convert it to Go and slowly remove Rust and
Python. New features will be added to the Go executable only.

## Installation

The recommended way is to install `logseq-doctor` globally with
[pipx](https://github.com/pypa/pipx):

    pipx install logseq-doctor

You can also install the development version with:

    pipx install git+https://github.com/andreoliwa/logseq-doctor

You will then have the `lsd` command available globally in your system.

If you want to use the `lsd tidy-up` command, (for now) you will need to
manually install the Go executable from the latest `master` branch:

    go install github.com/andreoliwa/logseq-doctor@f7c0f0f  # use the latest commit hash after the @

`lsdg` is the expected name for the Go executable, so you need to rename
it:

    mv $(go env GOPATH)/bin/logseq-doctor $(go env GOPATH)/bin/lsdg

Confirm if it\'s in your path:

    ls -l $(go env GOPATH)/bin/lsdg

## Quick start

Type `lsd` without arguments to check the current commands and options:

    Usage: lsd [OPTIONS] COMMAND [ARGS]...

    Logseq Doctor: heal your flat old Markdown files before importing them.

    ╭─ Options ────────────────────────────────────────────────────────────────────╮
    │ --install-completion          Install completion for the current shell.      │
    │ --show-completion             Show completion for the current shell, to copy │
    │                               it or customize the installation.              │
    │ --help                        Show this message and exit.                    │
    ╰──────────────────────────────────────────────────────────────────────────────╯
    ╭─ Commands ───────────────────────────────────────────────────────────────────╮
    │ outline  Convert flat Markdown to outline.                                   │
    │ tasks    List tasks in Logseq.                                               │
    │ tidy-up  Tidy up your Markdown files by removing empty bullets in any block. │
    ╰──────────────────────────────────────────────────────────────────────────────╯

## Development

To set up local development:

    make install-local

Run this to see help on all available targets:

    make

To run all the tests run:

    tox

Note, to combine the coverage data from all the tox environments run:

+------+---------------------------------------------------------------+
| Win | set PYTEST_ADDOPTS=--cov-append |
| dows | tox |
+------+---------------------------------------------------------------+
| O | PYTEST_ADDOPTS=--cov-append tox |
| ther | |
+------+---------------------------------------------------------------+
