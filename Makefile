GO_TEST = go test -v ./... -race -covermode=atomic

help: # Display this help
	@cat Makefile | egrep '^[a-z0-9 ./-]*:.*#' | sed -E -e 's/:.+# */@ /g' -e 's/ .+@/@/g' | sort | awk -F@ '{printf "\033[1;34m%-15s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: build-go # Build the Rust crate and Python package
	uv build
.PHONY: build

build-go: # Build the Golang executable
	go mod tidy
	go build -o lsdg main.go
	mv lsdg `go env GOPATH`/bin/
	$(MAKE) list-go
.PHONY: build-go

clean: # Clean the build artifacts
	cargo clean
	-rm `go env GOPATH`/bin/logseq-doctor
	-rm `go env GOPATH`/bin/lsdg
	-rm -rf .pytest_cache .ruff_cache build/
	$(MAKE) list-go
.PHONY: clean

list-go: # List the installed Go packages
	ls -l `go env GOPATH`/bin/
.PHONY: list-go

rehash: # Rehashing is needed (once) to make the [project.scripts] section of pyproject.toml available in the PATH
	pyenv rehash
.PHONY: rehash

setup: # Set up the local development environment
	uv sync
# TODO: keep the list of dev packages in a single place; this was copied from tox.ini
	uv add --dev pytest pytest-cov pytest-datadir responses pytest-env pytest-watch pytest-testmon
	$(MAKE) setup-go
	@echo "Run 'make smoke' to check if the development environment is working"
.PHONY: setup

setup-go: # Set up Go dependencies (logseq-go from the last commit of the local repo + golangci-lint)
	# Needed for the pre-commit hook
	# https://github.com/golangci/golangci-lint#install-golangci-lint
	brew install golangci-lint

	LAST_COMMIT=$$(cd ../logseq-go; git log -1 --format=%h); \
		echo "LAST_COMMIT: $$LAST_COMMIT"; \
		go get -u github.com/andreoliwa/logseq-go@$$LAST_COMMIT
	go mod tidy
.PHONY: setup-go

install: build-go # Install the package with pipx in editable mode. Do this when you want to use "lsd" outside of the development environment
	-pipx install -e --force .
	$(MAKE) rehash
	$(MAKE) which
.PHONY: install

which: # Show the location of the installed executables
	which lsd
	lsd
	which lsdg
	lsdg
.PHONY: which

uninstall: clean uninstall-pipx # Remove both local and global (virtualenv and pipx)
	-rm -rf .python-version .tox .venv
.PHONY: uninstall

uninstall-pipx: # Uninstall pipx virtualenv. Use this when developing, so the local venv "lsd" is available instead of the pipx one
	-pipx uninstall logseq-doctor
	$(MAKE) rehash
.PHONY: .uninstall-pipx

test: test-go # Run tests on Python, Rust and Go
	cargo test
	tox -e py311
.PHONY: test

test-go: # Run Go tests
	$(GO_TEST)
.PHONY: test-go

test-go-coverage: # Run Go tests with coverage
	$(GO_TEST) -coverprofile=coverage-go.out
.PHONY: test-go-coverage

watch: # Run tests and watch for changes
	uv run ptw --runner "pytest --testmon"
.PHONY: watch

pytest: # Run tests with pytest
	uv run pytest --cov --cov-report=term-missing -vv tests
.PHONY: pytest

release: # Bump the version, create a tag, commit and push. This will trigger the PyPI release on GitHub Actions
	# https://commitizen-tools.github.io/commitizen/bump/#configuration
	# See also: cz bump --help
	cz bump --check-consistency
	# TODO: publish the Rust crate on GitHub Actions instead of locally
	cargo publish -p logseq --locked
.PHONY: release

.release-post-bump:
	git push --atomic origin master ${CZ_POST_CURRENT_TAG_VERSION}
	gh release create ${CZ_POST_CURRENT_TAG_VERSION} --notes-from-tag
	gh repo view --web
.PHONY: .release-post-bump

smoke: rehash test # Run simple tests to make sure the package is working
	uv run lsd --help
.PHONY: smoke

clippy: develop # Run clippy on the Rust code
	cargo clippy
.PHONY: clippy
