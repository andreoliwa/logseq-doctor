ACTIVATE_VENV = source ~/.pyenv/versions/logseq-doctor/bin/activate

help: # Display this help
	@cat Makefile | egrep '^[a-z0-9 ./-]*:.*#' | sed -E -e 's/:.+# */@ /g' -e 's/ .+@/@/g' | sort | awk -F@ '{printf "\033[1;34m%-15s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: build-go # Build the Rust crate and Python package
	$(ACTIVATE_VENV) && maturin build
.PHONY: build

build-go: # Build the Golang executable
	go mod tidy
	go build
.PHONY: build-go

develop: build-go # Install the crate as module in the current virtualenv, rehash pyenv to put CLI scripts in PATH
	$(ACTIVATE_VENV) && maturin develop
.PHONY: develop

rehash: # Rehashing is needed (once) to make the [project.scripts] section of pyproject.toml available in the PATH
	pyenv rehash
.PHONY: rehash

print-config: # Print the configuration used by maturin
	PYO3_PRINT_CONFIG=1 $(ACTIVATE_VENV) && maturin develop
.PHONY: print-config

install: # Create the virtualenv and setup the local development environment
	-rm .python-version
	@echo $$(basename $$(pwd))
	-pyenv virtualenv $$(basename $$(pwd))
	pyenv local $$(basename $$(pwd))
# TODO: keep the list of dev packages in a single place; this was copied from tox.ini
	$(MAKE) deps
# Can't activate virtualenv from Makefile · Issue #372 · pyenv/pyenv-virtualenv
# https://github.com/pyenv/pyenv-virtualenv/issues/372
	$(MAKE) develop
	@echo "Run 'make smoke' to check if the development environment is working"
.PHONY: install

pipx-install: # Install the package with pipx in editable mode. Do this when you want to use "lsd" outside of the development environment
	-pipx install -e --force .
.PHONY: pipx-install

pipx-uninstall: # Uninstall only the pipx virtualenv. Use this when developing, so the local venv "lsd" is available instead of the pipx one
	-pipx uninstall logseq-doctor
	$(MAKE) rehash
.PHONY: pipx-uninstall

deps: # Install the development dependencies
	$(ACTIVATE_VENV) && python -m pip install -U pip pytest pytest-cov pytest-datadir responses pytest-env pytest-watch pytest-testmon
	$(MAKE) freeze
.PHONY: deps

freeze: # Show the installed packages
	$(ACTIVATE_VENV) && python -m pip freeze
.PHONY: freeze

uninstall: pipx-uninstall # Remove the virtualenv
	-rm .python-version
	-pyenv uninstall -f $$(basename $$(pwd))
.PHONY: uninstall

example: develop # Run a simple example of Python code calling Rust code
	python -c "from logseq_doctor import rust_ext; print(rust_ext.remove_consecutive_spaces('    - abc   123     def  \n'))"
.PHONY: example

run: develop rehash # Run the CLI with a Python click script as the entry point
	lsd --help
.PHONY: run

test: # Run tests on both Python and Rust
	cargo test
	tox -e py311
.PHONY: test

watch: # Run tests and watch for changes
	$(ACTIVATE_VENV) && ptw --runner "pytest --testmon"
.PHONY: watch

pytest: # Run tests with pytest
	$(ACTIVATE_VENV) && pytest --cov --cov-report=term-missing -vv tests
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

smoke: run example test # Run simple tests to make sure the package is working
.PHONY: smoke

clippy: develop # Run clippy on the Rust code
	cargo clippy
.PHONY: clippy
