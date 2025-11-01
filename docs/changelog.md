## [0.6.1](https://github.com/andreoliwa/logseq-doctor/compare/v0.6.0...v0.6.1) (2025-10-20)

### Bug Fixes

- installation and import paths ([ce0b743](https://github.com/andreoliwa/logseq-doctor/commit/ce0b743cf26d528a0d8c68357047c6960a4bc668))

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [0.5.0](https://github.com/andreoliwa/logseq-doctor/compare/v0.4.0...v0.5.0) (2025-02-13)

### Features

- **tasks:** JSON format ([#121](https://github.com/andreoliwa/logseq-doctor/issues/121)) ([07da841](https://github.com/andreoliwa/logseq-doctor/commit/07da841b4f64e5a72f9edb6f578f91dd47e53ed4))

# [0.4.0](https://github.com/andreoliwa/logseq-doctor/compare/v0.3.0...v0.4.0) (2025-02-11)

### Bug Fixes

- **deps:** update github.com/andreoliwa/logseq-go digest to 276dc3d ([62aa9ec](https://github.com/andreoliwa/logseq-doctor/commit/62aa9ec0ab9adfd92d5b99cc5223f88cd387b6e6))
- **deps:** update github.com/andreoliwa/logseq-go digest to 3b9f58b ([13e6274](https://github.com/andreoliwa/logseq-doctor/commit/13e6274ea8861e0c1267e90d91a09037f738ecac))
- **deps:** update module github.com/stretchr/testify to v1.10.0 ([f737ea4](https://github.com/andreoliwa/logseq-doctor/commit/f737ea4de9d4aa70d219cbf3da8389ed992568da))
- **deps:** update module gotest.tools/v3 to v3.5.2 ([7ee8efc](https://github.com/andreoliwa/logseq-doctor/commit/7ee8efc30b64081a54872a361129b968b513c874))
- **deps:** update rust crate anyhow to 1.0.80 ([3938baa](https://github.com/andreoliwa/logseq-doctor/commit/3938baa83d3968c4784aa2787a4dd6da36b90669))
- **deps:** update rust crate anyhow to 1.0.81 ([d5e9425](https://github.com/andreoliwa/logseq-doctor/commit/d5e9425467c225ac82e9107de63817bfbabd9f4e))
- **deps:** update rust crate anyhow to v1.0.95 ([a0bdcf7](https://github.com/andreoliwa/logseq-doctor/commit/a0bdcf705708f31edba418144c404e9a8acf23be))
- **deps:** update rust crate assert_fs to v1.1.2 ([b55d654](https://github.com/andreoliwa/logseq-doctor/commit/b55d65439b39b27afd1b67fa9b2c254d6905cab6))
- **deps:** update rust crate chrono to 0.4.34 ([8903b29](https://github.com/andreoliwa/logseq-doctor/commit/8903b294c690ecb1a21da999dad5f16f6089e3ef))
- **deps:** update rust crate chrono to 0.4.35 ([4001ea5](https://github.com/andreoliwa/logseq-doctor/commit/4001ea5e722d04300f94cc343d1d4c9a9d99d6a5))
- **deps:** update rust crate chrono to 0.4.37 ([8ff6a0b](https://github.com/andreoliwa/logseq-doctor/commit/8ff6a0b647efed6bbe0efc2362109f381dc28eae))
- **deps:** update rust crate pyo3 to 0.20.3 ([c0a566a](https://github.com/andreoliwa/logseq-doctor/commit/c0a566af231f249ed833efc7a8f1295f11f98b3e))
- **deps:** update rust crate pyo3 to 0.21.0 ([5a75518](https://github.com/andreoliwa/logseq-doctor/commit/5a755180b2a9d845a6ee11e1991aa9f43f86fc47))
- **deps:** update rust crate pyo3 to 0.21.1 ([4f9b6b5](https://github.com/andreoliwa/logseq-doctor/commit/4f9b6b5dcf7e4fa07464485705ef26e290a85ec2))
- **deps:** update rust crate regex to 1.10.4 ([fc033da](https://github.com/andreoliwa/logseq-doctor/commit/fc033da8061b6efdc29c8e2cb7c60c08fcc1c567))
- **deps:** update rust crate regex to v1.11.1 ([61ed776](https://github.com/andreoliwa/logseq-doctor/commit/61ed7764359adf711ab73abda595398efa649374))
- don't remove double spaces in tables ([253a1d0](https://github.com/andreoliwa/logseq-doctor/commit/253a1d07b72671227804245f366c08d2f366ec4e))
- make Journal properties public ([78b18a2](https://github.com/andreoliwa/logseq-doctor/commit/78b18a23a8835c8ae1eb00ce2f03acaba6bf4948))
- one transaction per file, to avoid touching unmodified files ([b66410b](https://github.com/andreoliwa/logseq-doctor/commit/b66410b3cbca0549cfbe72541e6bfdabbf2e48e4))
- **tasks:** use § as column separator, to parse with fzf ([ccd12e7](https://github.com/andreoliwa/logseq-doctor/commit/ccd12e74a3dd70edabdba5645a10fc5889362003))

### Features

- append raw Markdown from sdtin to a journal ([545b17b](https://github.com/andreoliwa/logseq-doctor/commit/545b17be5217e2d74a974ba4f7160215a16f440a))
- check for running tasks ([0bfef8f](https://github.com/andreoliwa/logseq-doctor/commit/0bfef8f3b0a56c5fc61929ebe39d19c56986e85c))
- check references to forbidden pages ([58f9f26](https://github.com/andreoliwa/logseq-doctor/commit/58f9f26d2a885ae26040d2f6d0617e5c4a215792))
- Python calling a Go executable with a simple function ([902e54d](https://github.com/andreoliwa/logseq-doctor/commit/902e54df5cf3e39043a58d6f90d2508df1afeb78))
- tidy-up command accepting Markdown files ([0f2e549](https://github.com/andreoliwa/logseq-doctor/commit/0f2e54919fb732010c64a9d30005fc5ee7c50ebd))
- **tidy-up:** check consecutive spaces in Go ([#115](https://github.com/andreoliwa/logseq-doctor/issues/115)) ([32a02d8](https://github.com/andreoliwa/logseq-doctor/commit/32a02d889e7a6945f24cdebba77b8993e33fa1c0))
- **tidy-up:** remove double spaces and save Markdown file ([212014a](https://github.com/andreoliwa/logseq-doctor/commit/212014af48f0585ca0e31aea0b75966efccdc432))
- **tidy-up:** remove empty bullets in Go, remove Python command ([a472b30](https://github.com/andreoliwa/logseq-doctor/commit/a472b30ab0fdeaa744866af69ab7dca14e340a2e))
- **tidy-up:** remove unnecessary brackets from tags ([d887e58](https://github.com/andreoliwa/logseq-doctor/commit/d887e58421132f941686966c3f9ac057f4232e14))
- **tidy-up:** remove unnecessary brackets from tags in Go ([f01fffd](https://github.com/andreoliwa/logseq-doctor/commit/f01fffd6ee75148a78d8b7c709537c3a841e9b5d))

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
