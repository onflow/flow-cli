# Set OS-related environment variables
ifeq ($(OS),Windows_NT)
	SHELL := cmd.exe
	GOPATH ?= $(USERPROFILE)\go
	PATH := $(PATH);$(GOPATH)\bin
	MKDIR_GOPATH := if not exist "$(GOPATH)" md "$(GOPATH)"
	SET_GOPATH := go env -w GOPATH="$(GOPATH)"
	SET_GO111MODULE := go env -w GO111MODULE=on
	SET_CGO_DISABLED := go env -w CGO_ENABLED=0
else
	SHELL := /bin/bash
	GOPATH ?= $(HOME)/go
	PATH := $(PATH):$(GOPATH)/bin
	MKDIR_GOPATH := mkdir -p "$(GOPATH)"
	SET_GOPATH := export GOPATH="$(GOPATH)"
	SET_GO111MODULE := export GO111MODULE=on
	SET_CGO_DISABLED := export CGO_ENABLED=0
endif

# The short Git commit hash
SHORT_COMMIT := $(shell git rev-parse --short HEAD)
# The Git commit hash
COMMIT := $(shell git rev-parse HEAD)
# The tag of the current commit, otherwise empty
VERSION := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
# Name of the cover profile
COVER_PROFILE := coverage.txt
# Disable go sum database lookup for private repos
GOPRIVATE := github.com/dapperlabs/*

MIXPANEL_PROJECT_TOKEN := 3fae49de272be1ceb8cf34119f747073
ACCOUNT_TOKEN := lilico:sF60s3wughJBmNh2

BINARY ?= ./cmd/flow/flow

.PHONY: binary
binary: $(BINARY)

.PHONY: install-tools
install-tools:
	$(MKDIR_GOPATH) && \
	$(SET_GOPATH) && \
	$(SET_GO111MODULE) && \
	go install github.com/axw/gocov/gocov@latest && \
	go install github.com/matm/gocov-html/cmd/gocov-html@latest && \
	go install github.com/sanderhahn/gozip/cmd/gozip@latest && \
	go install github.com/vektra/mockery/v2@latest

.PHONY: test
test:
	$(SET_GO111MODULE) && \
	$(SET_CGO_DISABLED) && \
	go test -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./... && \
	cd flowkit && \
	go test -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./...

.PHONY: test-e2e-emulator
test-e2e-emulator:
	flow -f tests/flow.json emulator start

.PHONY: coverage
coverage:
ifeq ($(COVER), true)
	# file has to be called index.html
	gocov convert $(COVER_PROFILE) > cover.json
	./cover-summary.sh
	gocov-html cover.json > index.html
	# coverage.zip will automatically be picked up by teamcity
	gozip -c coverage.zip index.html
endif

.PHONY: ci
ci: install-tools generate test coverage

$(BINARY):
	GO111MODULE=on go build \
		-trimpath \
		-ldflags \
		"-X github.com/onflow/flow-cli/build.commit=$(COMMIT) -X github.com/onflow/flow-cli/build.semver=$(VERSION) -X github.com/onflow/flow-cli/flowkit/util.MIXPANEL_PROJECT_TOKEN=${MIXPANEL_PROJECT_TOKEN} -X github.com/onflow/flow-cli/internal/accounts.accountToken=${ACCOUNT_TOKEN}"\
		-o $(BINARY) ./cmd/flow

.PHONY: versioned-binaries
versioned-binaries: generate
	$(MAKE) OS=linux ARCH=amd64 ARCHNAME=x86_64 versioned-binary
	$(MAKE) OS=linux ARCH=arm64 versioned-binary
	$(MAKE) OS=darwin ARCH=amd64 ARCHNAME=x86_64 versioned-binary
	$(MAKE) OS=darwin ARCH=arm64 versioned-binary
	$(MAKE) OS=windows ARCH=amd64 ARCHNAME=x86_64 versioned-binary

.PHONY: versioned-binary
versioned-binary:
	GOOS=$(OS) GOARCH=$(ARCH) $(MAKE) BINARY=./cmd/flow/flow-$(or ${ARCHNAME},${ARCHNAME},${ARCH})-$(OS)-$(VERSION) binary

.PHONY: publish
publish:
	gsutil -m cp cmd/flow/flow-*-$(VERSION) gs://flow-cli

.PHONY: clean
clean:
	rm ./cmd/flow/flow*

.PHONY: install-linter
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.47.2

.PHONY: lint
lint: generate
	golangci-lint run -v ./...

.PHONY: fix-lint
fix-lint:
	golangci-lint run -v --fix ./...

.PHONY: check-headers
check-headers:
	@./check-headers.sh

.PHONY: check-schema
check-schema:
	cd flowkit && go run ./cmd/flow-schema/flow-schema.go --verify=true ./schema.json 

.PHONY: check-tidy
check-tidy:
	go mod tidy
	cd flowkit && go mod tidy

.PHONY: generate-schema
generate-schema:
	echo %PATH% && dir $(GOPATH) && cd flowkit && go run ./cmd/flow-schema/flow-schema.go ./schema.json

.PHONY: generate
generate: install-tools
	echo %PATH% && dir $(GOPATH) && \
	cd flowkit && \
 	go generate ./...