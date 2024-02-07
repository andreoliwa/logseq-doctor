## v0.3.0 (2024-02-07)

### Feat

- **journal**: option to prepend content
- **journal**: choose the date (with natural language)
- **journal**: output content to stdout
- **journal**: option to convert to outline (#17)
- **journal**: pipe content from stdin (#17)
- **journal**: append content to the current journal page (#17)

## v0.2.1 (2024-02-04)

### Fix

- automated release with "make release" from local computer

## v0.2.0 (2024-02-04)

### Feat

- "logseq" crate for reusable functions
- remove double spaces
- remove Python 3.7
- list tasks, add Kanban board (#78)
- tidy up files by removing empty bullets (#63)
- rich CLI with Typer

### Fix

- remove Python 3.8 support
- preserve line break at the end
- **deps**: update dependency mistletoe to v1.3.0
- remove --format kanban
- **deps**: update dependency mistletoe to v1.2.1
- **deps**: update dependency mistletoe to v1.2.0
- **deps**: update dependency click to v8.1.7
- Typer needs some imports, UP006 ignores target-version=py38
- don't print "Done" on success
- **deps**: update dependency click to v8.1.6
- **deps**: update dependency click to v8.1.5
- **deps**: update dependency click to v8.1.4
- **deps**: update dependency mistletoe to v1.1.0
- **deps**: update dependency requests to v2.31.0
- **deps**: update dependency requests to v2.30.0
- **deps**: update dependency typer to v0.9.0
- **deps**: update dependency typer to v0.8.0
- add Ruff and adjust code
- **deps**: update dependency mistletoe to v1
- **deps**: upgrade shelligham and pre-commit hooks, fix tox
- **tidy-up**: display file name that was fixed
- handle thematic breaks and setext headers
- handle nested lists with single or multiple levels (#47)

### Refactor

- rename Python module

## v0.1.1 (2022-08-21)

### Fix

- ImportError on lsd --help

## v0.1.0 (2022-03-26)

### Feat

- convert headers and flat paragraphs to an outline
