# AGENTS.md

Guidance for AI coding agents working in this repo. Every claim below is sourced from the
Makefile, `go.mod`, `ci.yml`, `CONTRIBUTING.md`, or verified source files — do not add
unverified commands or paths.

## Overview

`flow-cli` is the official command-line tool for the Flow blockchain: deploy contracts, run
transactions/scripts, manage accounts/keys, and run a bundled emulator. Go 1.25.1 module
(`github.com/onflow/flow-cli`), built on [Cobra](https://github.com/spf13/cobra). All
blockchain logic is delegated to the external `github.com/onflow/flowkit/v2` module. Entry
point is `cmd/flow/main.go`. License: Apache-2.0.

## Build and Test Commands

CGO is required (BLS crypto). `go build` / `go test` need these env vars set:
`CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11"`.

- `make binary` — build `./cmd/flow/flow`; ldflags inject version, commit, and analytics tokens
- `make test` — `go test -coverprofile=coverage.txt ./...` with CGO flags set
- `make ci` — `generate test coverage` (this is what GitHub Actions runs)
- `make coverage` — emits `index.html` and `cover-summary.txt`, only when `COVER=true`
- `make lint` — `golangci-lint run -v ./...`; depends on `make generate`
- `make fix-lint` — golangci-lint with `--fix`
- `make generate` — `go generate ./...`; run before `lint`, `ci`, or any test touching generated code
- `make check-headers` — `./check-headers.sh`, verifies Apache-2.0 header on every `.go` file
- `make check-tidy` — `go mod tidy` (CI runs this; fails if `go.mod`/`go.sum` drift)
- `make clean` — removes binaries under `cmd/flow/`
- `make versioned-binaries` — cross-compiles linux/darwin/windows × amd64/arm64
- `make publish` — uploads versioned binaries to `gs://flow-cli` via `gsutil`
- `make release` — runs `ghcr.io/goreleaser/goreleaser-cross:v1.25.0` in Docker
- `make test-e2e-emulator` — `flow -f tests/flow.json emulator start`
- `SKIP_NETWORK_TESTS=1 make test` — skip tests that reach Flow mainnet/testnet (CONTRIBUTING.md)
- `nix develop` — enter dev shell from `flake.nix`; then `go run cmd/flow/main.go`

## Architecture

Cobra CLI. `cmd/flow/main.go` wires every subcommand into the root `flow` command and defines
eight command groups (super, resources, interactions, tools, project, security, manager, schedule).

**`internal/command/`** — shared framework. `command.Command` wraps `cobra.Command` with two
run modes: `Run` (no project state) and `RunS` (requires `*flowkit.State` loaded from
`flow.json`). `AddToParent()` handles loading `flow.json`, gateway/network resolution,
`flowkit.Services` init, version check, analytics, and error formatting. Global flags
(`internal/command/global_flags.go`): `--network`, `--host`, `--log`, `--output`, `--filter`,
`--save`, `--config-path`, `--yes`, `--skip-version-check`. Every `Result` must implement
`String()`, `Oneliner()`, and `JSON()`.

**`internal/super/`** — super commands (`flow init`, `flow dev`, `flow generate`, `flow flix`).
Scaffolding engine under `internal/super/generator/` with `templates/` and `fixtures/`.

**Feature packages** (`internal/<name>/`) — one per top-level command; each exports a
`Cmd *cobra.Command` (or `Command`) registered in `main.go`:
`accounts`, `blocks`, `cadence`, `collections`, `config`, `dependencymanager`, `emulator`,
`events`, `evm`, `keys`, `mcp`, `project`, `quick` (`flow deploy`, `flow run`), `schedule`
(transaction scheduler: `setup`/`get`/`list`/`cancel`/`parse`), `scripts`, `settings`,
`signatures`, `snapshot`, `status`, `test`, `tools` (`dev-wallet`, `flowser`), `transactions`,
`version`. Support: `internal/util/`, `internal/prompt/`.

**`build/build.go`** — version/commit variables injected via `-ldflags` at build time.
**`common/branding/`** — styling/ASCII constants.
**`flowkit/`** (top-level) — **historical artifact**; contains only `README.md` and
`schema.json`. All Go code moved to the external `github.com/onflow/flowkit/v2`.
**`docs/`** — hand-maintained Markdown reference pages, one per command, published to
developers.flow.com.
**`testing/better/`** — shared test helpers.

## Conventions and Gotchas

- **`make generate` before `make lint` and CI workflows.** `lint` declares `generate` as a
  prerequisite; `ci` runs `generate test coverage` in that order.
- **CGO is not optional.** Plain `go build ./...` / `go test ./...` without the CGO env vars
  above will fail on the BLS crypto dependency (`__BLST_PORTABLE__`).
- **Register new commands via `command.Command.AddToParent(cmd)`** (not raw `cmd.AddCommand`)
  so shared boilerplate — `flow.json` load, gateway init, error formatting — runs. See
  `cmd/flow/main.go` for both registration styles.
- **Command naming is `noun verb`** (`flow accounts get`, not `flow get-accounts`) — see
  "CLI Guidelines" in `CONTRIBUTING.md`.
- **Prefer flags over positional args.** Use an arg only for the single primary required value.
- **`--output json` must always work.** Every `Result` implements `JSON()`; never gate
  machine-readable output behind a subcommand.
- **stdout for normal output, stderr for errors.** No stack traces on error; `--log debug`
  is the escape hatch.
- **Every `.go` file needs the Apache-2.0 header.** `check-headers.sh` greps for
  `Licensed under the Apache License` or `Code generated (from|by)` and fails CI otherwise.
- **goimports `local-prefixes: github.com/onflow/flow-cli`** (`.golangci.yml`) — internal
  imports group separately from third-party.
- **Linters enabled:** `errcheck`, `govet`, `ineffassign`, `misspell`, plus `goimports`
  formatter. CI pins `golangci-lint v2.4.0` (`.github/workflows/ci.yml`).
- **`SKIP_NETWORK_TESTS=1`** skips tests that reach mainnet/testnet nodes — use in Nix or
  egress-restricted CI (CONTRIBUTING.md "Skipping Network-Dependent Tests").
- **`syscall.Exit` in `cmd/flow/main.go` is intentional** — works around a gRPC cleanup
  regression that appeared in Go 1.23.1 (inline comment in `main.go`).
- **`version.txt` is deprecated** for CLI versions after v1.18.0 (CONTRIBUTING.md
  "Releasing"). The semver is derived from the git tag via `-ldflags` into `build.semver`.
- **Analytics tokens (`MIXPANEL_PROJECT_TOKEN`, `ACCOUNT_TOKEN`) are baked in at build time**
  via ldflags in the Makefile — rebuild, don't patch the binary.

## Files Not to Modify

- `go.sum` — regenerate via `go mod tidy` / `make check-tidy`, never hand-edit.
- `flake.lock` — update via `nix flake update`.
- `flowkit/` top-level directory — legacy stub; real code lives in `github.com/onflow/flowkit/v2`.
- `version.txt` — deprecated post v1.18.0; leave it.
- `cli-banner.svg`, `cli.gif` — release artifacts.
