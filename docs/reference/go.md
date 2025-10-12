# Go API Reference

This section contains the API documentation for the Go implementation of Logseq Doctor.

**[📚 Browse Local Go Documentation](../go-api/index.html)** | **[View on pkg.go.dev](https://pkg.go.dev/github.com/andreoliwa/logseq-doctor)**

!!! tip "Local Documentation"
The local Go documentation is generated with [doc2go](https://github.com/abhinav/doc2go) and provides detailed API reference for all packages, **including internal packages**. Click the link above to browse the interactive documentation.

## Package Overview

### `cmd` - Command-line Interface

The `cmd` package contains the CLI implementation using Cobra:

- **root.go** - Root command and CLI entry point
- **md.go** - Markdown processing commands
- **tidy_up.go** - Tidy up commands
- **content.go** - Content manipulation commands
- **backlog.go** - Backlog management commands

### `internal` - Internal Packages

The `internal` package contains implementation details not meant for external use:

- **api.go** - Logseq API client implementation
- **content.go** - Content parsing and manipulation
- **md.go** - Markdown processing utilities
- **tasks.go** - Task management
- **nodes.go** - AST node handling
- **finder.go** - File and content discovery

### `pkg/set` - Public Utilities

The `pkg/set` package provides a generic set implementation:

- Generic set data structure using Go generics
- Type-safe operations (Add, Remove, Contains)
- Sorted value retrieval

## Installation

```bash
go get github.com/andreoliwa/lsd
```

Or install the binary:

```bash
# Using Homebrew
brew install andreoliwa/formulae/logseq-doctor

# Or download from releases
gh release download -R andreoliwa/logseq-doctor
```

## Usage

```go
import (
    "github.com/andreoliwa/lsd/pkg/set"
    "github.com/andreoliwa/lsd/internal"
)

// Use the set package
s := set.NewSet[string]()
s.Add("item")
```

Or use the CLI:

```bash
lsd --help
```
