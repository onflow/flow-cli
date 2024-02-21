PACKAGE_NAME := github.com/onflow/flow-cli
GOLANG_CROSS_VERSION ?= v1.20.0

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
GOPATH ?= $(HOME)/go
PATH := $(PATH):$(GOPATH)/bin
# OS
UNAME := $(shell uname)

MIXPANEL_PROJECT_TOKEN := 3fae49de272be1ceb8cf34119f747073
ACCOUNT_TOKEN := lilico:sF60s3wughJBmNh2

BINARY ?= ./cmd/flow/flow

.PHONY: binary
binary: $(BINARY)

.PHONY: install-tools
install-tools:
	cd ${GOPATH}; \
	mkdir -p ${GOPATH}; \
	GO111MODULE=on go install github.com/axw/gocov/gocov@latest; \
	GO111MODULE=on go install github.com/matm/gocov-html/cmd/gocov-html@latest; \
	GO111MODULE=on go install github.com/sanderhahn/gozip/cmd/gozip@latest; \
	GO111MODULE=on go install github.com/vektra/mockery/v2@v2.38.0;

.PHONY: test
test:
	GO111MODULE=on go test -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./...

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
		"-X github.com/onflow/flow-cli/build.commit=$(COMMIT) -X github.com/onflow/flow-cli/build.semver=$(VERSION) -X github.com/onflow/flow-cli/internal/accounts.accountToken=${ACCOUNT_TOKEN}"\
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
install-linter:
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

.PHONY: check-tidy
check-tidy:
	go mod tidy

.PHONY: generate
generate: install-tools
	go generate ./...

.PHONY: release
release:
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean