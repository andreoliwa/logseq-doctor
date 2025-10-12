# Python API Reference

This section contains the auto-generated API documentation for the Python implementation of Logseq Doctor.

The Python package provides:

- **API Client** (`logseq_doctor.api`): Logseq API client for interacting with your graph
- **CLI** (`logseq_doctor.cli`): Command-line interface implementation
- **Constants** (`logseq_doctor.constants`): Shared constants and configuration

## Quick Links

- [logseq_doctor.api](python/api.md) - API client for Logseq
- [logseq_doctor.cli](python/cli.md) - CLI implementation
- [logseq_doctor.constants](python/constants.md) - Constants and configuration

## Installation

```bash
pip install logseq-doctor
```

Or with `uv`:

```bash
uv add logseq-doctor
```

## Usage

```python
from logseq_doctor.api import Block, LogseqClient

# Use the API client
client = LogseqClient(path="/path/to/graph")
```

Or use the CLI:

```bash
lsdpy --help
```
