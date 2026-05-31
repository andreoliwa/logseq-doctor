# Installation

Logseq Doctor is a single Go binary, `lqd`. Install it with Homebrew or `go install`.

## Go Binary Executable

### macOS and Linux (Homebrew)

The recommended way for macOS and Linux is to install with Homebrew:

```bash
brew install andreoliwa/formulae/logseq-doctor
```

### Manual Installation

You can install manually using Go:

```bash
go install github.com/andreoliwa/logseq-doctor/cmd/lqd@latest
```

Confirm if it's in your path:

```bash
which lqd
# or
ls -l $(go env GOPATH)/bin/lqd
```

!!! tip
Make sure your `GOPATH/bin` directory is in your system's PATH. You can add it to your shell profile:

    ```bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

## Build from Source

To build and install the `lqd` binary from source, clone the repository and run:

```bash
git clone https://github.com/andreoliwa/logseq-doctor.git
cd logseq-doctor
make build-go
```

This builds the `lqd` binary and installs it into `$(go env GOPATH)/bin`.

## Verify Installation

### Go CLI

```bash
lqd --help
```

You should see the help message with available commands.

## Next Steps

Once installed, check out the [Usage Guide](usage.md) to learn how to use Logseq Doctor.
