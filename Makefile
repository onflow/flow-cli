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
# Ensure go bin path is in path (Especially for CI)
PATH := $(PATH):$(GOPATH)/bin
# OS
UNAME := $(shell uname)

GOPATH ?= $(HOME)/go

BINARY ?= ./cmd/flow/flow

.PHONY: binary
binary: $(BINARY)

.PHONY: install-tools
install-tools:
	cd ${GOPATH}; \
	mkdir -p ${GOPATH}; \
	GO111MODULE=on go install github.com/axw/gocov/gocov@latest; \
	GO111MODULE=on go install github.com/matm/gocov-html@latest; \
	GO111MODULE=on go install github.com/sanderhahn/gozip/cmd/gozip@latest;

.PHONY: test
test:
	GO111MODULE=on go test $$(go list ./... | grep -v /e2e) -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,)
	cd pkg/flowkit; GO111MODULE=on go test -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./...

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

.PHONY: test-e2e
test-e2e:
	GO111MODULE=on go build -o ./e2e/flow ./cmd/flow/main.go 
	cd ./e2e; go test

.PHONY: ci
ci: install-tools test coverage

.PHONY: install
install:
	GO111MODULE=on go install \
		-trimpath \
		-ldflags \
		"-X github.com/onflow/flow-cli/build.commit=$(COMMIT) -X github.com/onflow/flow-cli/build.semver=$(VERSION) -X github.com/onflow/flow-cli/pkg/flowkit/util.MIXPANEL_PROJECT_TOKEN=${MIXPANEL_PROJECT_TOKEN}" \
		./cmd/flow

$(BINARY):
	GO111MODULE=on go build \
		-trimpath \
		-ldflags \
		"-X github.com/onflow/flow-cli/build.commit=$(COMMIT) -X github.com/onflow/flow-cli/build.semver=$(VERSION) -X github.com/onflow/flow-cli/pkg/flowkit/util.MIXPANEL_PROJECT_TOKEN=${MIXPANEL_PROJECT_TOKEN}"\
		-o $(BINARY) ./cmd/flow

.PHONY: versioned-binaries
versioned-binaries:
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
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.47.2

.PHONY: lint
lint:
	golangci-lint run -v ./...

.PHONY: fix-lint
fix-lint:
	golangci-lint run -v --fix ./...

.PHONY: check-headers
check-headers:
	@./check-headers.sh

.PHONY: check-tidy
check-tidy:
	go mod tidy
	cd pkg/flowkit; go mod tidy