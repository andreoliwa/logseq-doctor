# Logseq Doctor

[![Documentation Status](https://img.shields.io/badge/docs-mkdocs-blue)](https://andreoliwa.github.io/logseq-doctor/)
[![Go Build Status](https://github.com/andreoliwa/logseq-doctor/actions/workflows/go.yaml/badge.svg)](https://github.com/andreoliwa/logseq-doctor/actions)
[![Tox Build Status](https://github.com/andreoliwa/logseq-doctor/actions/workflows/tox.yaml/badge.svg)](https://github.com/andreoliwa/logseq-doctor/actions)
[![Coverage Status](https://codecov.io/gh/andreoliwa/logseq-doctor/branch/master/graphs/badge.svg?branch=master)](https://codecov.io/github/andreoliwa/logseq-doctor)
[![PyPI Package](https://img.shields.io/pypi/v/logseq-doctor.svg)](https://pypi.org/project/logseq-doctor)

**Logseq Doctor: heal your flat old Markdown files before importing them to Logseq.**

!!! note "Project Status"
This project is still in alpha, so it's very rough on the edges (documentation and feature-wise).

    At the moment, it has both a Python and Go CLI.

    The long-term plan is to convert it to Go and slowly remove Python.
    New features will be added to the Go CLI only.

## What is Logseq Doctor?

Logseq Doctor is a command-line tool that helps you prepare your Markdown files for import into [Logseq](https://logseq.com/). It provides utilities to:

- Convert flat Markdown to Logseq's outline format
- Clean up and tidy Markdown files
- Prevent invalid content
- Manage tasks in Logseq
- Append content to your Logseq graph

## Features

### Python CLI (`lsdpy`)

- **Outline Conversion**: Convert flat Markdown files to Logseq's outline format
- **Task Management**: List and manage tasks in your Logseq graph

### Go CLI (`lsd`)

- **Content Management**: Append raw Markdown content to Logseq
- **Tidy Up**: Clean up and standardize your Markdown files
- **Fast Performance**: Written in Go for speed and efficiency

## Quick Links

- [Installation Guide](installation.md) - Get started with Logseq Doctor
- [Usage Guide](usage.md) - Learn how to use the CLI tools
- [Contributing](contributing.md) - Help improve Logseq Doctor
- [GitHub Repository](https://github.com/andreoliwa/logseq-doctor)

## Project Goals

The primary goal of Logseq Doctor is to make it easier to migrate existing Markdown content into Logseq. Whether you're coming from another note-taking system or have a collection of flat Markdown files, Logseq Doctor helps ensure your content is properly formatted for Logseq's outliner-based structure.

## License

This project is licensed under the MIT License.
