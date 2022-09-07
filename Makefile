DIR := ${CURDIR}

############################################################################
# OS/ARCH detection
############################################################################
os1=$(shell uname -s)
os2=
ifeq ($(os1),Darwin)
os1=darwin
os2=osx
else ifeq ($(os1),Linux)
os1=linux
os2=linux
else
$(error unsupported OS: $(os1))
endif

arch1=$(shell uname -m)
ifeq ($(arch1),x86_64)
arch2=amd64
else ifeq ($(arch1),aarch64)
arch2=arm64
else
$(error unsupported ARCH: $(arch1))
endif

############################################################################
# Vars
############################################################################

build_dir := $(DIR)/.build/$(os1)-$(arch1)

golangci_lint_version = v1.49.0
golangci_lint_dir = $(build_dir)/golangci_lint/$(golangci_lint_version)
golangci_lint_bin = $(golangci_lint_dir)/golangci-lint
golangci_lint_cache = $(golangci_lint_dir)/cache

# There may be more than one tag. Only use one that starts with 'v' followed by
# a number, e.g., v0.9.3.
git_tag = $(shell git tag --points-at HEAD | grep '^v[0-9]*')
git_commit = $(shell git rev-parse --short=7 HEAD)
git_dirty = $(if $(shell git status -s),true,)

# The ldflags are only influenced by the GIT_* variables passed in as Makefile
# arguments. These are normally only passed by the Dockerfile.
go_ldflags := -s -w
ifneq ($(GIT_TAG),)
go_ldflags += -X github.com/spiffe/spiffe-csi/internal/version.gitTag=$(GIT_TAG)
endif
ifneq ($(GIT_COMMIT),)
go_ldflags += -X github.com/spiffe/spiffe-csi/internal/version.gitCommit=$(GIT_COMMIT)
endif
ifneq ($(GIT_DIRTY),)
go_ldflags += -X github.com/spiffe/spiffe-csi/internal/version.gitDirty=$(GIT_DIRTY)
endif
go_ldflags := '${go_ldflags}'

.PHONY: default
default: docker-build

.PHONY: docker-build
docker-build:
	docker build \
		--build-arg GIT_TAG=$(git_tag:v%=%) \
		--build-arg GIT_COMMIT=$(git_commit) \
		--build-arg GIT_DIRTY=$(git_dirty) \
		--target spiffe-csi-driver \
		-t ghcr.io/spiffe/spiffe-csi-driver:devel \
		.

.PHONY: build
build: | bin
	CGO_ENABLED=0 go build -ldflags ${go_ldflags} -o bin/spiffe-csi-driver ./cmd/spiffe-csi-driver

bin:
	mkdir bin

lint: $(golangci_lint_bin)
	@GOLANGCI_LINT_CACHE="$(golangci_lint_cache)" $(golangci_lint_bin) run ./...

$(golangci_lint_bin):
	@echo "Installing golangci-lint $(golangci_lint_version)..."
	@rm -rf $(dir $(golangci_lint_dir))
	@mkdir -p $(golangci_lint_dir)
	@mkdir -p $(golangci_lint_cache)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(golangci_lint_dir) $(golangci_lint_version)
