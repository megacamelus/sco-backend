
PROJECT_NAME ?= sco-backend
PROJECT_VERSION ?= 0.0.1

CONTAINER_REGISTRY ?= quay.io
CONTAINER_REGISTRY_ORG ?= sco1237896
CONTAINER_IMAGE_VERSION ?= $(PROJECT_VERSION)
CONTAINER_IMAGE ?= $(CONTAINER_REGISTRY)/$(CONTAINER_REGISTRY_ORG)/$(PROJECT_NAME):$(CONTAINER_IMAGE_VERSION)

LINT_GOGC ?= 10
LINT_DEADLINE ?= 10m

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCALBIN := $(PROJECT_PATH)/bin

## Tool Versions
KIND_VERSION ?= v0.20.0
LINTER_VERSION ?= v1.52.2
GOVULNCHECK_VERSION ?= latest

## Tool Binaries
LINTER ?= $(LOCALBIN)/golangci-lint
GOIMPORT ?= $(LOCALBIN)/goimports
KIND ?= $(LOCALBIN)/kind
GOVULNCHECK ?= $(LOCALBIN)/govulncheck

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

ifndef ignore-not-found
  ignore-not-found = false
endif

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-\/]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Development

.PHONY: fmt
fmt: goimport ## Run go fmt, gomiport against code.
	$(GOIMPORT) -l -w .
	go fmt ./...


.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...


.PHONY: run
run: serve


.PHONY: serve
serve: ## Run the server
	go run -ldflags="$(GOLDFLAGS)" cmd/main.go serve


.PHONY: test
test: fmt vet
	go test -ldflags="$(GOLDFLAGS)" -v ./pkg/...


.PHONY: test/e2e
test/e2e: fmt vet
	go test -ldflags="$(GOLDFLAGS)" -v ./e2e-test/...


.PHONY: test/apply-crds
test/apply-crds:
	kubectl apply -k ./config/manifests/e2e/ --server-side --force-conflicts


.PHONY: check/all ## single target for checking everything
check/all: deps fmt vet check build test test/e2e docker/build


##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -ldflags="$(GOLDFLAGS)" -o bin/sco cmd/main.go


.PHONY: deps
deps:  ## Tidy up deps.
	go mod tidy


.PHONY: check
check: check/lint check/vuln


.PHONY: check/lint
check/lint: golangci-lint
	@$(LINTER) run \
		--config .golangci.yml \
		--out-format tab \
		--skip-dirs etc \
		--deadline $(LINT_DEADLINE) \
		--verbose


.PHONY: check/lint/fix
check/lint/fix: golangci-lint
	@$(LINTER) run \
		--config .golangci.yml \
		--out-format tab \
		--skip-dirs etc \
		--deadline $(LINT_DEADLINE) \
		--fix

.PHONY: check/vuln
check/vuln: govulncheck
	@echo "run govulncheck"
	@$(GOVULNCHECK) ./...

.PHONY: docker/build
docker/build:
	$(CONTAINER_TOOL) build -t $(CONTAINER_IMAGE) .


.PHONY: docker/push
docker/push:
	$(CONTAINER_TOOL) push $(CONTAINER_IMAGE)


.PHONY: docker/push/kind
docker/push/kind: docker/build
	kind load docker-image $(CONTAINER_IMAGE)


##@ Build Dependencies

## Location to install dependencies to
$(LOCALBIN):
	mkdir -p $(LOCALBIN)


## Tool Binaries
.PHONY: golangci-lint
golangci-lint: $(LINTER)
$(LINTER): $(LOCALBIN)
	@test -s $(LOCALBIN)/golangci-lint || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINTER_VERSION)


.PHONY: goimport
goimport: $(GOIMPORT)
$(GOIMPORT): $(LOCALBIN)
	@test -s $(LOCALBIN)/goimport || \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@latest


.PHONY: kind
kind: $(KIND)
$(KIND): $(LOCALBIN)
	@test -s $(LOCALBIN)/kind || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind@$(KIND_VERSION)

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK)
$(GOVULNCHECK): $(LOCALBIN)
	@test -s $(GOVULNCHECK) || \
	GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
