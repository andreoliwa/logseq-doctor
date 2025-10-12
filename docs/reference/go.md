# Go API Reference

This section contains the API documentation for the Go implementation of Logseq Doctor.

**[View on pkg.go.dev](https://pkg.go.dev/github.com/andreoliwa/logseq-doctor)**

!!! tip "Documentation"
The documentation below is automatically generated from the Go source code using [gomarkdoc](https://github.com/princjef/gomarkdoc) and includes detailed API reference for all packages, **including internal packages**.

## Packages

Browse the documentation for each package:

### Command-line Interface

- **[cmd](go/cmd/README.md)** - CLI implementation using Cobra

### Public Packages

- **[pkg/set](go/pkg/set/README.md)** - Generic set data structure

### Internal Packages

- **[internal](go/internal/README.md)** - Internal implementation details
- **[internal/backlog](go/internal/backlog/README.md)** - Backlog management
- **[internal/testutils](go/internal/testutils/README.md)** - Testing utilities

## Installation

```bash
go get github.com/andreoliwa/logseq-doctor
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
    "github.com/andreoliwa/logseq-doctor/pkg/set"
)

// Use the set package
s := set.NewSet[string]()
s.Add("item")
```

Or use the CLI:

```bash
lsd --help
```
