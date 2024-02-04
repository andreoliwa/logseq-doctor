help: # Display this help
	@cat Makefile | egrep '^[a-z0-9 ./-]*:.*#' | sed -E -e 's/:.+# */@ /g' -e 's/ .+@/@/g' | sort | awk -F@ '{printf "  \033[1;34m%-18s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: # Build the Rust crate and Python package
	maturin build
.PHONY: build

develop: # Install the crate as module in the current virtualenv, rehash pyenv to put CLI scripts in PATH
	maturin develop
	# Rehashing is needed (once) to make the [project.scripts] section of pyproject.toml available in the PATH
	pyenv rehash
	python -m pip freeze
.PHONY: develop

print-config: # Print the configuration used by maturin
	PYO3_PRINT_CONFIG=1 maturin develop
.PHONY: print-config

install: # Create the virtualenv
	@echo $$(basename $$(pwd))
	pyenv virtualenv $$(basename $$(pwd))
	pyenv local $$(basename $$(pwd))
# Can't activate virtualenv from Makefile · Issue #372 · pyenv/pyenv-virtualenv
# https://github.com/pyenv/pyenv-virtualenv/issues/372
	@echo "Run 'pyenv activate' before running maturin commands"
.PHONY: install

uninstall: # Remove the virtualenv
	-rm .python-version
	-pyenv uninstall $$(basename $$(pwd))
.PHONY: uninstall

example: develop # Run a simple example of Python code calling Rust code
	python -c "from logseq_doctor import _logseq_doctor as rust; print(rust.rust_remove_consecutive_spaces('    - abc   123     def  \n'))"
.PHONY: example

cli: develop # Run the CLI with a Python click script as the entry point
	lsd --help
.PHONY: cli

test: # Run tests on both Python and Rust
	cargo test
	tox -e py311
.PHONY: test

test-watch: # Run tests and watch for changes
	source .tox/py311/bin/activate && ptw --runner "pytest --testmon"
.PHONY: test-watch
