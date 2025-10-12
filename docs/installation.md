# Installation

Logseq Doctor provides both Python and Go implementations. You can install either or both depending on your needs.

## Python Executable

The recommended way is to install `logseq-doctor` globally with [pipx](https://github.com/pypa/pipx):

```bash
pipx install logseq-doctor
```

You can also install the development version with:

```bash
pipx install git+https://github.com/andreoliwa/logseq-doctor
```

You will then have the `lsdpy` command available globally in your system.

### Alternative: pip

If you prefer to use pip:

```bash
pip install logseq-doctor
```

!!! warning
Installing with pip may conflict with other Python packages in your system. We recommend using pipx instead.

## Go Binary Executable

### macOS and Linux (Homebrew)

The recommended way for macOS and Linux is to install with Homebrew:

```bash
brew install andreoliwa/formulae/logseq-doctor
```

### Manual Installation

You can install manually using Go:

```bash
go install github.com/andreoliwa/logseq-doctor@latest
```

Confirm if it's in your path:

```bash
which lsd
# or
ls -l $(go env GOPATH)/bin/lsd
```

!!! tip
Make sure your `GOPATH/bin` directory is in your system's PATH. You can add it to your shell profile:

    ```bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

## Build from Source

To build and install from the source (both Python and Go executables), clone the repository and run:

```bash
git clone https://github.com/andreoliwa/logseq-doctor.git
cd logseq-doctor
make install
```

This will:

1. Set up a Python virtual environment
2. Install Python dependencies
3. Build the Go binary
4. Install both executables

## Verify Installation

### Python CLI

```bash
lsdpy --help
```

You should see the help message with available commands.

### Go CLI

```bash
lsd --help
```

You should see the help message with available commands.

## Next Steps

Once installed, check out the [Usage Guide](usage.md) to learn how to use Logseq Doctor.
