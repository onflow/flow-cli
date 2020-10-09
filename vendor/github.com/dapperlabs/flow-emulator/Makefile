# The short Git commit hash
SHORT_COMMIT := $(shell git rev-parse --short HEAD)
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

# Enable docker build kit
export DOCKER_BUILDKIT := 1

.PHONY: install-tools
install-tools:
	cd ${GOPATH}; \
	GO111MODULE=on go get github.com/golang/mock/mockgen@v1.3.1; \
	GO111MODULE=on go get github.com/axw/gocov/gocov; \
	GO111MODULE=on go get github.com/matm/gocov-html; \
	GO111MODULE=on go get github.com/sanderhahn/gozip/cmd/gozip;

.PHONY: test
test:
	GO111MODULE=on go test -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./...

.PHONY: run
run:
	GO111MODULE=on go run ./cmd/emulator

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

.PHONY: generate
generate: generate-mocks

.PHONY: generate-mocks
generate-mocks:
	GO111MODULE=on mockgen -destination=server/backend/mocks/emulator.go -package=mocks github.com/dapperlabs/flow-emulator/server/backend Emulator
	GO111MODULE=on mockgen -destination=storage/mocks/store.go -package=mocks github.com/dapperlabs/flow-emulator/storage Store

.PHONY: ci
ci: install-tools generate test coverage

.PHONY: docker-build-emulator
docker-build:
	docker build --ssh default -f cmd/emulator/Dockerfile -t gcr.io/dl-flow/emulator:latest -t "gcr.io/dl-flow/emulator:$(SHORT_COMMIT)" .
ifneq (${VERSION},)
	docker tag gcr.io/dl-flow/emulator:latest gcr.io/dl-flow/emulator:${VERSION}
endif

.PHONY: docker-push-emulator
docker-push:
	docker push gcr.io/dl-flow/emulator:latest
	docker push "gcr.io/dl-flow/emulator:$(SHORT_COMMIT)"
ifneq (${VERSION},)
	docker push "gcr.io/dl-flow/emulator:${VERSION}"
endif
