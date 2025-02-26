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
else ifeq ($(arch1),arm64)
arch2=arm64
else
$(error unsupported ARCH: $(arch1))
endif

############################################################################
# Vars
############################################################################

BINARIES := spiffe-csi-driver

PLATFORMS ?= linux/amd64,linux/arm64

build_dir := $(DIR)/.build/$(os1)-$(arch1)

golangci_lint_version = v1.64.5
golangci_lint_dir = $(build_dir)/golangci_lint/$(golangci_lint_version)
golangci_lint_bin = $(golangci_lint_dir)/golangci-lint
golangci_lint_cache = $(golangci_lint_dir)/cache

.PHONY: FORCE
FORCE: ;

.PHONY: default
default: docker-build

.PHONY: container-builder
container-builder:
	docker buildx create --platform $(PLATFORMS) --name container-builder --node container-builder0 --use

.PHONY: docker-build
docker-build: $(addsuffix -image.tar,$(BINARIES))

spiffe-csi-driver-image.tar: Dockerfile FORCE | container-builder
	docker buildx build \
		--platform $(PLATFORMS) \
		--target spiffe-csi-driver \
		-o type=oci,dest=$@ \
		.

.PHONY: build
build: $(addprefix bin/,$(BINARIES))

bin/%: cmd/% FORCE
	CGO_ENABLED=0 go build -o $@ ./$<

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: $(golangci_lint_bin)
	@GOLANGCI_LINT_CACHE="$(golangci_lint_cache)" $(golangci_lint_bin) run ./...

$(golangci_lint_bin):
	@echo "Installing golangci-lint $(golangci_lint_version)..."
	@rm -rf $(dir $(golangci_lint_dir))
	@mkdir -p $(golangci_lint_dir)
	@mkdir -p $(golangci_lint_cache)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(golangci_lint_dir) $(golangci_lint_version)

.PHONY: load-images
load-images: $(addsuffix -image.tar,$(BINARIES))
	./.github/workflows/scripts/load-oci-archives.sh
