# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build the binary (requires CGO for BLS crypto)
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" GO111MODULE=on go build -o ./cmd/flow/flow ./cmd/flow

# Or use Make
make binary

# Run directly without building
go run cmd/flow/main.go [command]
```

## Testing

```bash
# Run all tests
make test
# Equivalent: CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" GO111MODULE=on go test -coverprofile=coverage.txt ./...

# Run a single test package
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/accounts/...

# Run a specific test
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/accounts/... -run TestFunctionName

# Skip network-dependent tests (e.g. in sandboxed environments)
SKIP_NETWORK_TESTS=1 make test
```

## Linting

```bash
make lint         # Run golangci-lint
make fix-lint     # Auto-fix lint issues
make check-headers  # Verify Apache license headers on all Go files
go generate ./... # Regenerate generated code (required before lint)
```

## Architecture

The CLI is a [Cobra](https://github.com/spf13/cobra)-based application with three main layers:

### Entry Point
`cmd/flow/main.go` — wires all subcommands into the root `flow` command.

### Command Framework (`internal/command/`)
The `command.Command` struct wraps a `cobra.Command` with two execution modes:
- `Run` — for commands that don't need a loaded `flow.json` state
- `RunS` — for commands that require an initialized project state (`*flowkit.State`)

`Command.AddToParent()` handles all shared boilerplate: loading `flow.json`, resolving network/host, creating the gRPC gateway, initializing `flowkit.Services`, version checking, analytics, and error formatting. **All new commands should use this pattern.**

Every command's run function returns a `command.Result` interface with three output methods: `String()` (human-readable), `Oneliner()` (grep-friendly inline), and `JSON()` (structured). The framework handles `--output`, `--filter`, and `--save` flags automatically.

### Command Packages (`internal/`)
Each feature area is its own package with a top-level `Cmd *cobra.Command` that aggregates subcommands. Pattern:
- `accounts.Cmd` (`internal/accounts/`) — registered in `main.go` via `cmd.AddCommand(accounts.Cmd)`
- Subcommands (e.g., `get.go`, `create.go`) define a package-level `var getCommand = &command.Command{...}` and register via `init()` or the parent's `init()`

Key packages:
- `internal/super/` — high-level "super commands": `flow init`, `flow dev`, `flow generate`, `flow flix`
- `internal/super/generator/` — code generation engine for Cadence contracts, scripts, transactions, and tests
- `internal/dependencymanager/` — `flow deps` commands for managing on-chain contract dependencies
- `internal/config/` — `flow config` subcommands for managing `flow.json`
- `internal/emulator/` — wraps the Flow emulator

### flowkit Dependency
The CLI delegates all blockchain interactions to the `github.com/onflow/flowkit/v2` module (external). The `flowkit.Services` interface is the primary abstraction for network calls. The local `flowkit/` directory is a historical artifact (migrated to the external module) and contains only a README and schema.

### Global Flags
Defined in `internal/command/global_flags.go`, applied to every command: `--network`, `--host`, `--log`, `--output`, `--filter`, `--save`, `--config-path`, `--yes`, `--skip-version-check`.

### Configuration
`flow.json` is the project config file. `flowkit.Load()` reads it. The `internal/config/` commands modify it. `state.Networks()`, `state.Accounts()`, etc. provide typed access.

## CLI Design Conventions
- Commands follow `noun verb` pattern (`flow accounts get`)
- Prefer flags over positional args; use args only for the primary required value
- `--output json` must always work for machine-readable output
- Errors go to stderr; normal output to stdout
- Progress indicators for long-running operations via `logger.StartProgress()` / `logger.StopProgress()`
- Long-running commands support `--yes` to skip confirmation prompts

## License Headers
All Go source files must have the Apache 2.0 license header. Run `make check-headers` to verify.
