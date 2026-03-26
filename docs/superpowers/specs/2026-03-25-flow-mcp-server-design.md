# Flow MCP Server Design

## Overview

Add a `flow mcp` command to flow-cli that starts a Model Context Protocol (MCP)
server over stdio for Cadence smart contract development. The server exposes 9
tools across two categories: LSP tools (in-process language server) and
on-chain query / code review tools.

This replaces the need for a separate TypeScript MCP server (see
[cadence-lang.org PR #285](https://github.com/onflow/cadence-lang.org/pull/285))
by integrating directly into the Go CLI with no extra runtime dependencies.

## Tools

### LSP Tools (5)

These wrap the in-process `cadence-tools/languageserver` Server.

| Tool | Description | Parameters |
|---|---|---|
| `cadence_check` | Check Cadence code for syntax and type errors | `code`, `filename?`, `network?` |
| `cadence_hover` | Get type info and docs for a symbol at a position | `code`, `line`, `character`, `filename?`, `network?` |
| `cadence_definition` | Find definition location of a symbol | `code`, `line`, `character`, `filename?`, `network?` |
| `cadence_symbols` | List all symbols (contracts, resources, functions, events) | `code`, `filename?`, `network?` |
| `cadence_completion` | Get completions at a position | `code`, `line`, `character`, `filename?`, `network?` |

All LSP tools accept an optional `network` parameter (mainnet/testnet/emulator,
default mainnet) for resolving on-chain imports.

### Audit Tools (4)

These use flowkit gRPC gateways for on-chain data and pattern matching for code review.

| Tool | Description | Parameters |
|---|---|---|
| `get_contract_source` | Fetch on-chain contract manifest (names, sizes, imports, dependency graph) | `address`, `network?`, `recurse?` |
| `get_contract_code` | Fetch source code of contracts from an address | `address`, `contract_name?`, `network?` |
| `cadence_code_review` | Review Cadence code for common issues and best practices | `code`, `network?` |
| `cadence_execute_script` | Execute a read-only Cadence script on-chain | `code`, `network?`, `args?` |

## Architecture

```
flow mcp (stdio)
    |
    v
+-- MCP Server (mcp-go) ------------------------------------+
|                                                            |
|  LSP Tools --> LSPWrapper --> server.Server (cadence-tools) |
|                (doc lifecycle    (in-process)               |
|                 management)                                 |
|                                                            |
|  Audit Tools --> flowkit gateway --> Flow network           |
|                                                            |
+------------------------------------------------------------+
```

### Package Structure

```
internal/mcp/
  mcp.go      - Cobra command + MCP server setup, tool registration
  lsp.go      - LSP wrapper: server.Server lifecycle, diagnostic capture
  audit.go    - Code review rules (cadence_code_review)
  tools.go    - Tool handler implementations (all 9 tools)
```

Registered in `cmd/flow/main.go` alongside other top-level commands.

## Command

```go
var Cmd = &cobra.Command{
    Use:   "mcp",
    Short: "Start the Cadence MCP server",
}
```

Uses `Run` (not `RunS`) so it works without a `flow.json`. If `flow.json` is
found, its network configurations are used (allowing custom host overrides).
Otherwise, hardcoded defaults are used for mainnet/testnet/emulator.

The `--help` output includes installation instructions for Claude Code, Cursor,
and Claude Desktop, plus a summary of available tools.

## LSP Wrapper

### In-Process Server

The wrapper manages a `server.Server` instance from `cadence-tools/languageserver`.

```go
type LSPWrapper struct {
    server *server.Server
    mu     sync.Mutex
}
```

Created at startup with:
1. `server.NewServer()` to create the LSP server
2. `integration.NewFlowIntegration(server, true)` to enable on-chain import resolution

### Document Lifecycle

The LSP server stores documents in an in-memory map (`s.documents`), not on
disk. The `file:///` URI is purely virtual — no actual files are created.

Since the LSP server has no `DidCloseTextDocument` handler, opened documents
stay in the map forever. To avoid unbounded accumulation, we reuse a single
virtual URI (`file:///mcp/scratch.cdc`) as a scratch buffer:

1. First call: `DidOpenTextDocument` with the virtual URI and the code string
2. Every subsequent call: `DidChangeTextDocument` to replace the content
3. The LSP runs the type checker on the updated content
4. Call the LSP method (`Hover`, `Completion`, etc.) and return the result

Each MCP tool call is independent — it overwrites the scratch buffer with its
code, queries the LSP, and returns. Calls are serialized by the mutex so there
is no contention over the single URI.

### Diagnostic Capture

`DidOpenTextDocument` and `DidChangeTextDocument` trigger type checking, which
pushes diagnostics via `conn.Notify("textDocument/publishDiagnostics", ...)`.

A thin `protocol.Conn` adapter captures these:

```go
type diagConn struct {
    diagnostics []protocol.Diagnostic
}

func (c *diagConn) Notify(_ context.Context, method string, params any) error {
    if method == "textDocument/publishDiagnostics" {
        // extract and store diagnostics
    }
    return nil
}
```

The `cadence_check` tool returns these captured diagnostics. Other tools
(hover, completion, etc.) ignore them.

### Serialization

All LSP operations are serialized via `sync.Mutex`. The LSP server is
single-threaded by design — concurrent document updates would corrupt state.

## Network Configuration

### flow.json Detection

At startup:
1. Attempt `flowkit.Load()` to find and load `flow.json`
2. If found, use its network configurations (custom hosts, accounts, aliases)
3. If not found, proceed with defaults — the server still works

### Gateway Creation

```go
func (m *MCPServer) gatewayForNetwork(network string) (gateway.Gateway, error) {
    if m.state != nil {
        net, err := m.state.Networks().ByName(network)
        if err == nil {
            return gateway.NewGrpcGateway(net)
        }
    }
    return gateway.NewGrpcGateway(defaultNetworks[network])
}
```

Default network addresses:
- mainnet: `access.mainnet.nodes.onflow.org:9000`
- testnet: `access.devnet.nodes.onflow.org:9000`
- emulator: `127.0.0.1:3569`

## cadence_code_review Rules

Regex-based pattern matching for common Cadence issues and best practices.
These are heuristic checks — not a substitute for a proper security audit.
Ported from the TypeScript PR:

| Rule | Severity | Pattern |
|---|---|---|
| overly-permissive-access | warning | `access(all) var/let` on state fields |
| overly-permissive-function | note | `access(all) fun` |
| deprecated-pub | info | `pub` keyword (deprecated in Cadence 1.0) |
| unsafe-force-unwrap | note | Force-unwrap `!` |
| auth-account-exposure | warning | `AuthAccount` or `auth(...) &Account` |
| hardcoded-address | info | Hardcoded `0x...` not in imports |
| unguarded-capability | warning | `.publish(` calls |
| potential-reentrancy | note | `.borrow` followed by `self.` mutation |
| resource-loss-destroy | warning | `destroy()` calls |

When the LSP is available, `cadence_code_review` also runs a full type check
and merges those diagnostics into the output.

## Help Text

`flow mcp --help` includes:

- What the server does
- Installation for Claude Code: `claude mcp add cadence-mcp -- flow mcp`
- Configuration for Cursor / Claude Desktop (JSON snippet)
- List of all available tools with descriptions

## Testing

- **LSP wrapper:** Unit tests with real `server.Server` — check Cadence snippets
  for expected diagnostics, hover info, completions
- **cadence_code_review:** Unit tests against known-vulnerable and clean Cadence
  snippets, verify expected findings
- **Network tools:** Integration tests behind `SKIP_NETWORK_TESTS` for
  `get_contract_source`, `get_contract_code`, `cadence_execute_script`
- **MCP server:** End-to-end tool call tests with mock inputs

All tests in `internal/mcp/*_test.go`.

## Dependencies

- `github.com/mark3labs/mcp-go` — Go MCP SDK (stdio transport, tool registration)
- `github.com/onflow/cadence-tools/languageserver` — already in go.mod
- `github.com/onflow/flowkit/v2` — already in go.mod
