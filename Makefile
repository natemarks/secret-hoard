.DEFAULT_GOAL := help

# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
DEFAULT_BRANCH := main
THIS_FILE := $(lastword $(MAKEFILE_LIST))
PKG := github.com/natemarks/secret-hoard
COMMIT := $(shell git rev-parse HEAD)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
CDIR = $(shell pwd)
EXECUTABLES := sh-pull sh-push sh-generate sh-contents
GOOS := linux
GOARCH := amd64

CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
DEFAULT_BRANCH := main

help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

${EXECUTABLES}:
	@for o in $(GOOS); do \
	  for a in $(GOARCH); do \
        echo "$(COMMIT)/$${o}/$${a}" ; \
        mkdir -p build/$(COMMIT)/$${o}/$${a} ; \
        echo "COMMIT: $(COMMIT)" >> build/$(COMMIT)/$${o}/$${a}/version.txt ; \
        env GOOS=$${o} GOARCH=$${a} \
        go build  -v -o build/$(COMMIT)/$${o}/$${a}/$@ \
				-ldflags="-X github.com/natemarks/secret-hoard/version.Version=${COMMIT}" ${PKG}/cmd/$@; \
	  done \
    done ; \

build: git-status ${EXECUTABLES}
	rm -f build/current
	ln -s $(CDIR)/build/$(COMMIT) $(CDIR)/build/current

check-release: ## Check version and description of latest GitHub release
	@echo "Checking latest release..."
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "Error: gh (GitHub CLI) is not installed"; \
		echo "Install it from: https://cli.github.com/"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "=== Latest Release ==="; \
	gh release view --json tagName,name,body,createdAt,author,url | \
		jq -r '"Version: \(.tagName)\nName: \(.name)\nAuthor: \(.author.login)\nCreated: \(.createdAt)\nURL: \(.url)\n\nDescription:\n\(.body)"' || \
		echo "No releases found or error accessing GitHub"

semver-release: ## Create a semver release tarball with install script
	@read -p "Enter semver version (e.g., v1.0.0): " VERSION; \
	if [ -z "$$VERSION" ]; then \
		echo "Error: Version cannot be empty"; \
		exit 1; \
	fi; \
	echo "Creating release $$VERSION..."; \
	$(MAKE) git-status; \
	$(MAKE) build; \
	mkdir -p release/$$VERSION; \
	for o in $(GOOS); do \
	  for a in $(GOARCH); do \
		RELEASE_DIR="release/$$VERSION/tmp_$${o}_$${a}"; \
		mkdir -p $$RELEASE_DIR; \
		cp build/$(COMMIT)/$${o}/$${a}/sh-* $$RELEASE_DIR/; \
		echo '#!/bin/bash' > $$RELEASE_DIR/install.sh; \
		echo 'set -e' >> $$RELEASE_DIR/install.sh; \
		echo '' >> $$RELEASE_DIR/install.sh; \
		echo '# Install secret-hoard binaries to $$HOME/bin' >> $$RELEASE_DIR/install.sh; \
		echo 'INSTALL_DIR="$$HOME/bin"' >> $$RELEASE_DIR/install.sh; \
		echo '' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "Installing secret-hoard to $$INSTALL_DIR..."' >> $$RELEASE_DIR/install.sh; \
		echo '' >> $$RELEASE_DIR/install.sh; \
		echo '# Create directory if it does not exist' >> $$RELEASE_DIR/install.sh; \
		echo 'mkdir -p "$$INSTALL_DIR"' >> $$RELEASE_DIR/install.sh; \
		echo '' >> $$RELEASE_DIR/install.sh; \
		echo '# Get the directory where this script is located' >> $$RELEASE_DIR/install.sh; \
		echo 'SCRIPT_DIR="$$(cd "$$(dirname "$$0")" && pwd)"' >> $$RELEASE_DIR/install.sh; \
		echo '' >> $$RELEASE_DIR/install.sh; \
		echo '# Copy binaries and make them executable' >> $$RELEASE_DIR/install.sh; \
		echo 'for binary in sh-pull sh-push sh-generate sh-contents; do' >> $$RELEASE_DIR/install.sh; \
		echo '  if [ -f "$$SCRIPT_DIR/$$binary" ]; then' >> $$RELEASE_DIR/install.sh; \
		echo '    echo "  Installing $$binary..."' >> $$RELEASE_DIR/install.sh; \
		echo '    cp "$$SCRIPT_DIR/$$binary" "$$INSTALL_DIR/"' >> $$RELEASE_DIR/install.sh; \
		echo '    chmod +x "$$INSTALL_DIR/$$binary"' >> $$RELEASE_DIR/install.sh; \
		echo '  else' >> $$RELEASE_DIR/install.sh; \
		echo '    echo "  Warning: $$binary not found"' >> $$RELEASE_DIR/install.sh; \
		echo '  fi' >> $$RELEASE_DIR/install.sh; \
		echo 'done' >> $$RELEASE_DIR/install.sh; \
		echo '' >> $$RELEASE_DIR/install.sh; \
		echo 'echo ""' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "Installation complete!"' >> $$RELEASE_DIR/install.sh; \
		echo 'echo ""' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "Installed binaries:"' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "  - $$INSTALL_DIR/sh-pull"' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "  - $$INSTALL_DIR/sh-push"' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "  - $$INSTALL_DIR/sh-generate"' >> $$RELEASE_DIR/install.sh; \
		echo 'echo "  - $$INSTALL_DIR/sh-contents"' >> $$RELEASE_DIR/install.sh; \
		echo 'echo ""' >> $$RELEASE_DIR/install.sh; \
		echo 'if echo "$$PATH" | grep -q "$$INSTALL_DIR"; then' >> $$RELEASE_DIR/install.sh; \
		echo '  echo "$$INSTALL_DIR is in your PATH"' >> $$RELEASE_DIR/install.sh; \
		echo 'else' >> $$RELEASE_DIR/install.sh; \
		echo '  echo "NOTE: Add $$INSTALL_DIR to your PATH by adding this to your ~/.bashrc or ~/.zshrc:"' >> $$RELEASE_DIR/install.sh; \
		echo '  echo \"  export PATH=\$$HOME/bin:\$$PATH\"' >> $$RELEASE_DIR/install.sh; \
		echo 'fi' >> $$RELEASE_DIR/install.sh; \
		chmod +x $$RELEASE_DIR/install.sh; \
		tar -C $$RELEASE_DIR -czf release/$$VERSION/secret-hoard_$$VERSION\_$${o}_$${a}.tar.gz .; \
		rm -rf $$RELEASE_DIR; \
		echo "Created release/$$VERSION/secret-hoard_$$VERSION\_$${o}_$${a}.tar.gz"; \
	  done; \
	done; \
	echo ""; \
	echo "Release $$VERSION created successfully!"; \
	echo "Release files:"; \
	ls -lh release/$$VERSION/; \
	echo ""; \
	if command -v gh >/dev/null 2>&1; then \
		read -p "Create GitHub release? (y/N): " CREATE_GH; \
		if [ "$$CREATE_GH" = "y" ] || [ "$$CREATE_GH" = "Y" ]; then \
			echo ""; \
			read -p "Enter release title (default: $$VERSION): " TITLE; \
			if [ -z "$$TITLE" ]; then \
				TITLE="$$VERSION"; \
			fi; \
			echo ""; \
			echo "Enter release notes (Ctrl-D when done):"; \
			NOTES=$$(cat); \
			if [ -z "$$NOTES" ]; then \
				NOTES="Release $$VERSION"; \
			fi; \
			echo ""; \
			echo "Creating GitHub release..."; \
			gh release create "$$VERSION" \
				release/$$VERSION/*.tar.gz \
				--title "$$TITLE" \
				--notes "$$NOTES"; \
			echo "GitHub release created!"; \
			echo "View at: $$(gh release view $$VERSION --json url -q .url)"; \
		fi; \
	else \
		echo "Note: Install gh CLI to create GitHub releases automatically"; \
		echo "  https://cli.github.com/"; \
	fi

release: git-status build
	mkdir -p release/$(COMMIT)
	@for o in $(GOOS); do \
	  for a in $(GOARCH); do \
        tar -C ./build/$(COMMIT)/$${o}/$${a} -czvf release/$(COMMIT)/secret-hoard_$(COMMIT)_$${o}_$${a}.tar.gz . ; \
	  done \
    done ; \

shellcheck: ## use black to format python files
	( \
       git ls-files 'scripts/*.sh' |  xargs shellcheck --format=gcc; \
    )

test:
	@go test -v ${PKG_LIST}
#	@go test -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

goimports: ## check imports
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

lint:  ##  run golint
	go install golang.org/x/lint/golint@latest
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

fmt: ## run gofmt
	@go fmt ${PKG_LIST}

gocyclo: ## run cyclomatic complexity check
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 25 .

godeadcode: ## unreachable code check
	go install golang.org/x/tools/cmd/deadcode@latest
	deadcode -test github.com/natemarks/secret-hoard/...

govulncheck: # run cyclomatic complexity check
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
unittest: ## run unit tests (no external dependencies)
	@go test -v ./secretlogic/...

unittest-update: ## update golden files for unit tests
	@go test -v ./secretlogic/... -update

unittest-coverage: ## run unit tests with coverage report
	@go test -cover ./secretlogic/...

static: unittest goimports fmt vet lint gocyclo godeadcode govulncheck test ## run all static checks including unit tests

clean:
	-@rm ${OUT} ${OUT}-v*


git-status: ## require status is clean so we can use undo_edits to put things back
	@status=$$(git status --porcelain); \
	if [ ! -z "$${status}" ]; \
	then \
		echo "Error - working directory is dirty. Commit those changes!"; \
		exit 1; \
	fi

download: ## download biometric aware ssl cert tarball
	bash scripts/download.sh

upload: ## upload biometric aware ssl tarball
	bash scripts/upload.sh

.PHONY: build release semver-release check-release static vet lint fmt gocyclo goimports test ${EXECUTABLES}
