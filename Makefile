# The short Git commit hash
SHORT_COMMIT := $(shell git rev-parse --short HEAD)
# The Git commit hash
COMMIT := $(shell git rev-parse HEAD)
# The tag of the current commit, otherwise empty
VERSION := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
# Name of the cover profile
COVER_PROFILE := cover.out
# Disable go sum database lookup for private repos
GOPRIVATE := github.com/dapperlabs/*
# Ensure go bin path is in path (Especially for CI)
PATH := $(PATH):$(GOPATH)/bin
# OS
UNAME := $(shell uname)

BINARY ?= ./cmd/flow/flow

.PHONY: install-tools
install-tools:
	cd ${GOPATH}; \
	GO111MODULE=on go get github.com/axw/gocov/gocov; \
	GO111MODULE=on go get github.com/matm/gocov-html; \
	GO111MODULE=on go get github.com/sanderhahn/gozip/cmd/gozip;

.PHONY: test
test:
	GO111MODULE=on go test -mod vendor -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./...

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
ci: install-tools test coverage

.PHONY: binary
binary: $(BINARY)

$(BINARY):
	GO111MODULE=on go build \
		-mod vendor \
		-trimpath \
		-ldflags \
		"-X github.com/dapperlabs/flow-cli/build.commit=$(COMMIT) -X github.com/dapperlabs/flow-cli/build.semver=$(VERSION)" \
		-o $(BINARY) ./cmd/flow

.PHONY: versioned-binaries
versioned-binaries:
	$(MAKE) OS=linux versioned-binary
	$(MAKE) OS=darwin versioned-binary
	$(MAKE) OS=windows versioned-binary

.PHONY: versioned-binary
versioned-binary:
	GOOS=$(OS) GOARCH=amd64 $(MAKE) BINARY=./cmd/flow/flow-x86_64-$(OS)-$(VERSION) binary

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor
