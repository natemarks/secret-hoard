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
EXECUTABLES := sh-download sh-upload
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
static: goimports fmt vet lint gocyclo godeadcode govulncheck test

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

.PHONY: build release static vet lint fmt gocyclo goimports test
