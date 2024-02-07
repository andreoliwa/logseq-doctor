help: # Display this help
	@cat Makefile | egrep '^[a-z0-9 ./-]*:.*#' | sed -E -e 's/:.+# */@ /g' -e 's/ .+@/@/g' | sort | awk -F@ '{printf "\033[1;34m%-15s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: # Build the Rust crate and Python package
	maturin build
.PHONY: build

develop: # Install the crate as module in the current virtualenv, rehash pyenv to put CLI scripts in PATH
	maturin develop
.PHONY: develop

rehash: # Rehashing is needed (once) to make the [project.scripts] section of pyproject.toml available in the PATH
	pyenv rehash
.PHONY: rehash

print-config: # Print the configuration used by maturin
	PYO3_PRINT_CONFIG=1 maturin develop
.PHONY: print-config

install: # Create the virtualenv and setup the local development environment
	@echo $$(basename $$(pwd))
	pyenv virtualenv $$(basename $$(pwd))
	pyenv local $$(basename $$(pwd))
# TODO: keep the list of dev packages in a single place; this was copied from tox.ini
	$(MAKE) deps
# Can't activate virtualenv from Makefile · Issue #372 · pyenv/pyenv-virtualenv
# https://github.com/pyenv/pyenv-virtualenv/issues/372
	@echo "Run 'pyenv activate' then `make smoke' to make sure the development environment is working"
.PHONY: install

deps: # Install the development dependencies
	source ~/.pyenv/versions/logseq-doctor/bin/activate && \
    		python -m pip install -U pip pytest pytest-cov pytest-datadir responses pytest-env pytest-watch pytest-testmon
	$(MAKE) freeze
.PHONY: deps

freeze: # Show the installed packages
	source ~/.pyenv/versions/logseq-doctor/bin/activate && python -m pip freeze
.PHONY: freeze

uninstall: # Remove the virtualenv
	-rm .python-version
	-pyenv uninstall $$(basename $$(pwd))
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

test-watch: # Run tests and watch for changes
	source .tox/py311/bin/activate && ptw --runner "pytest --testmon"
.PHONY: test-watch

pytest: # Run tests with pytest
	source ~/.pyenv/versions/logseq-doctor/bin/activate && pytest --cov --cov-report=term-missing -vv tests
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
