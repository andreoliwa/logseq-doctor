GO_BIN := $(shell command -v richgo >/dev/null 2>&1 && echo richgo || echo go)
GO_TEST = source ./.env.tests; $(GO_BIN) test ./... -race -covermode=atomic

help: # Display this help
	@cat Makefile | egrep '^[a-z0-9 ./-]*:.*#' | sed -E -e 's/:.+# */@ /g' -e 's/ .+@/@/g' | sort | awk -F@ '{printf "\033[1;34m%-15s\033[0m %s\n", $$1, $$2}'
.PHONY: help

upgrade: # Upgrade all dependencies
	go get -u -t ./...
	go mod tidy
.PHONY: upgrade

build: # Build the Go binary (and lqd-statusbar on macOS)
	go mod tidy
	go build ./cmd/lqd
	if [ "$$(uname)" = "Darwin" ]; then go build ./cmd/lqd-statusbar; fi
.PHONY: build

clean: # Clean the build artifacts
	-rm `go env GOPATH`/bin/logseq-doctor
	-rm `go env GOPATH`/bin/lqd
	-rm -rf .ruff_cache build/
	$(MAKE) list
.PHONY: clean

list: # List the installed Go packages
	ls -l `go env GOPATH`/bin/
.PHONY: list

install: build # Build and install all binaries to ~/go/bin
	mv lqd `go env GOPATH`/bin/
	if [ "$$(uname)" = "Darwin" ] && [ -f lqd-statusbar ]; then mv lqd-statusbar `go env GOPATH`/bin/; fi
	$(MAKE) list
.PHONY: install

setup: # Set up Go development dependencies
	go mod download
.PHONY: setup

which: # Run the main executable to confirm it is installed properly in the PATH
	-which lqd
	-lqd
.PHONY: which

test: # Run Go tests
	$(GO_TEST) $(opt)
.PHONY: test

test-cov: # Run Go tests with coverage
	$(GO_TEST) -coverprofile=coverage-go.out
.PHONY: test-cov

release: # Create a GitHub release for the Go package
	gh workflow run release.yaml
	# https://commitizen-tools.github.io/commitizen/bump/#configuration
	# See also: cz bump --help
.PHONY: release

.release-post-bump: # This is called in .cz.toml in post_bump_hooks
	git push --atomic origin master ${CZ_POST_CURRENT_TAG_VERSION}
	gh release create ${CZ_POST_CURRENT_TAG_VERSION} --notes-from-tag
	gh repo view --web
.PHONY: .release-post-bump

docs: # Generate Go API docs and build the MkDocs site
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
	gomarkdoc --output docs/reference/go.md ./...
	mkdocs build
.PHONY: docs
