# Flow MCP Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `flow mcp` command that starts an MCP server over stdio, exposing 9 tools for Cadence development (LSP + on-chain query + code review).

**Architecture:** In-process LSP via `cadence-tools/languageserver`, on-chain queries via `flowkit` gRPC gateways, code review via regex rules. All wrapped in an MCP server using `mcp-go` with stdio transport.

**Tech Stack:** Go, mcp-go, cadence-tools/languageserver, flowkit/v2

**Spec:** `docs/superpowers/specs/2026-03-25-flow-mcp-server-design.md`

---

## File Structure

```
internal/mcp/
  mcp.go          - Cobra command, MCP server creation, tool registration
  mcp_test.go     - End-to-end MCP tool call tests
  lsp.go          - LSPWrapper: in-process server.Server lifecycle, diagnostic capture
  lsp_test.go     - LSP wrapper unit tests
  audit.go        - Code review rules (cadence_code_review)
  audit_test.go   - Code review rule tests
  tools.go        - All 9 tool handler implementations
  tools_test.go   - Tool handler tests

Modified:
  cmd/flow/main.go  - Register mcp.Cmd
  go.mod / go.sum   - Add mcp-go dependency
```

---

### Task 1: Add mcp-go dependency and scaffold the command

**Files:**
- Create: `internal/mcp/mcp.go`
- Modify: `cmd/flow/main.go`
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add mcp-go dependency**

Run:
```bash
go get github.com/mark3labs/mcp-go@latest
```

- [ ] **Step 2: Create the MCP command with help text**

Create `internal/mcp/mcp.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/afero"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
)

var Cmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the Cadence MCP server",
	Long: `Start a Model Context Protocol (MCP) server for Cadence smart contract development.

The server provides tools for checking Cadence code, inspecting types,
querying on-chain contracts, executing scripts, and reviewing code for
common issues.

Claude Code:
  claude mcp add cadence-mcp -- flow mcp

Cursor / Claude Desktop (add to settings JSON):
  {
    "mcpServers": {
      "cadence-mcp": {
        "command": "flow",
        "args": ["mcp"]
      }
    }
  }

Available tools:
  cadence_check          Check Cadence code for syntax and type errors
  cadence_hover          Get type info for a symbol at a position
  cadence_definition     Find where a symbol is defined
  cadence_symbols        List all symbols in Cadence code
  cadence_completion     Get completions at a position
  get_contract_source    Fetch on-chain contract manifest
  get_contract_code      Fetch contract source code from an address
  cadence_code_review    Review Cadence code for common issues
  cadence_execute_script Execute a read-only Cadence script on-chain`,
	Run: runMCP,
}

func runMCP(cmd *cobra.Command, args []string) {
	// Try to load flow.json for custom network configs
	loader := &afero.Afero{Fs: afero.NewOsFs()}
	state, _ := flowkit.Load(config.DefaultPaths(), loader)

	s := mcpserver.NewMCPServer("cadence-mcp", "1.0.0")

	// TODO: register tools in subsequent tasks

	if err := mcpserver.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

// resolveNetwork returns a config.Network for the given network name.
// Uses flow.json config if available, otherwise falls back to defaults.
func resolveNetwork(state *flowkit.State, network string) (*config.Network, error) {
	if network == "" {
		network = "mainnet"
	}

	if state != nil {
		net, err := state.Networks().ByName(network)
		if err == nil {
			return net, nil
		}
	}

	net, err := config.DefaultNetworks.ByName(network)
	if err != nil {
		return nil, fmt.Errorf("unknown network %q", network)
	}
	return net, nil
}

// createGateway creates a gRPC gateway for the given network.
func createGateway(state *flowkit.State, network string) (gateway.Gateway, error) {
	net, err := resolveNetwork(state, network)
	if err != nil {
		return nil, err
	}
	return gateway.NewGrpcGateway(*net)
}
```

- [ ] **Step 3: Register the command in main.go**

In `cmd/flow/main.go`, add the import and registration:

Add import:
```go
"github.com/onflow/flow-cli/internal/mcp"
```

Add after the `cmd.AddCommand(schedule.Cmd)` line:
```go
cmd.AddCommand(mcp.Cmd)
```

- [ ] **Step 4: Verify it builds**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go build ./cmd/flow/...
```
Expected: builds with no errors.

- [ ] **Step 5: Verify help text**

Run:
```bash
go run ./cmd/flow mcp --help
```
Expected: prints the long description with installation instructions and tool list.

- [ ] **Step 6: Commit**

```bash
git add internal/mcp/mcp.go cmd/flow/main.go go.mod go.sum
git commit -m "Add flow mcp command scaffold with mcp-go dependency"
```

---

### Task 2: LSP wrapper — in-process server with diagnostic capture

**Files:**
- Create: `internal/mcp/lsp.go`
- Create: `internal/mcp/lsp_test.go`

- [ ] **Step 1: Write the failing test for LSP wrapper initialization and diagnostics**

Create `internal/mcp/lsp_test.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLSPWrapper_Check_ValidCode(t *testing.T) {
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)

	diags, err := lsp.Check("access(all) fun main() {}", "")
	require.NoError(t, err)
	assert.Empty(t, diags, "valid code should produce no diagnostics")
}

func TestLSPWrapper_Check_InvalidCode(t *testing.T) {
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)

	diags, err := lsp.Check("access(all) fun main() { let x: Int = \"hello\" }", "")
	require.NoError(t, err)
	assert.NotEmpty(t, diags, "type mismatch should produce diagnostics")
}

func TestLSPWrapper_Check_SyntaxError(t *testing.T) {
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)

	diags, err := lsp.Check("this is not valid cadence {{{", "")
	require.NoError(t, err)
	assert.NotEmpty(t, diags, "syntax error should produce diagnostics")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestLSPWrapper -v
```
Expected: FAIL — `NewLSPWrapper` undefined.

- [ ] **Step 3: Implement the LSP wrapper**

Create `internal/mcp/lsp.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"fmt"
	"strings"
	"sync"

	"github.com/onflow/cadence-tools/languageserver/integration"
	"github.com/onflow/cadence-tools/languageserver/protocol"
	"github.com/onflow/cadence-tools/languageserver/server"
)

const scratchURI = protocol.DocumentURI("file:///mcp/scratch.cdc")

// LSPWrapper manages an in-process cadence-tools language server.
// All operations are serialized — the LSP server is single-threaded.
type LSPWrapper struct {
	server      *server.Server
	conn        *diagConn
	mu          sync.Mutex
	docVersion  int32
	docOpen     bool
}

// NewLSPWrapper creates a new LSP wrapper with an in-process language server.
// enableFlowClient enables on-chain import resolution (requires network access).
func NewLSPWrapper(enableFlowClient bool) (*LSPWrapper, error) {
	s, err := server.NewServer()
	if err != nil {
		return nil, fmt.Errorf("creating LSP server: %w", err)
	}

	_, err = integration.NewFlowIntegration(s, enableFlowClient)
	if err != nil {
		return nil, fmt.Errorf("initializing Flow integration: %w", err)
	}

	conn := &diagConn{}

	// Initialize the server (required before any LSP operations)
	_, err = s.Initialize(conn, &protocol.InitializeParams{
		InitializationOptions: map[string]interface{}{
			"accessCheckMode": "strict",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initializing LSP: %w", err)
	}

	return &LSPWrapper{
		server: s,
		conn:   conn,
	}, nil
}

// updateDocument opens or updates the scratch document with the given code.
// Must be called with w.mu held.
func (w *LSPWrapper) updateDocument(code string) {
	w.docVersion++
	if !w.docOpen {
		w.server.DidOpenTextDocument(w.conn, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        scratchURI,
				LanguageID: "cadence",
				Version:    w.docVersion,
				Text:       code,
			},
		})
		w.docOpen = true
	} else {
		w.server.DidChangeTextDocument(w.conn, &protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: scratchURI},
				Version:                w.docVersion,
			},
			ContentChanges: []protocol.TextDocumentContentChangeEvent{
				{Text: code},
			},
		})
	}
}

// Check analyzes Cadence code and returns diagnostics.
func (w *LSPWrapper) Check(code string, network string) ([]protocol.Diagnostic, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()
	w.updateDocument(code)
	return w.conn.getDiagnostics(), nil
}

// Hover returns type information at the given position.
func (w *LSPWrapper) Hover(code string, line, character int, network string) (*protocol.Hover, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()
	w.updateDocument(code)
	return w.server.Hover(w.conn, &protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
		Position:     protocol.Position{Line: uint32(line), Character: uint32(character)},
	})
}

// Definition returns the definition location of a symbol at the given position.
func (w *LSPWrapper) Definition(code string, line, character int, network string) (*protocol.Location, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()
	w.updateDocument(code)
	return w.server.Definition(w.conn, &protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
		Position:     protocol.Position{Line: uint32(line), Character: uint32(character)},
	})
}

// Symbols returns all document symbols in the code.
func (w *LSPWrapper) Symbols(code string, network string) ([]*protocol.DocumentSymbol, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()
	w.updateDocument(code)
	return w.server.DocumentSymbol(w.conn, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
	})
}

// Completion returns completion items at the given position.
func (w *LSPWrapper) Completion(code string, line, character int, network string) ([]*protocol.CompletionItem, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()
	w.updateDocument(code)
	return w.server.Completion(w.conn, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
			Position:     protocol.Position{Line: uint32(line), Character: uint32(character)},
		},
	})
}

// diagConn implements protocol.Conn to capture diagnostics pushed by the LSP.
type diagConn struct {
	mu          sync.Mutex
	diagnostics []protocol.Diagnostic
}

func (c *diagConn) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.diagnostics = nil
}

func (c *diagConn) getDiagnostics() []protocol.Diagnostic {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]protocol.Diagnostic, len(c.diagnostics))
	copy(result, c.diagnostics)
	return result
}

func (c *diagConn) Notify(method string, params any) error {
	if method == "textDocument/publishDiagnostics" {
		if p, ok := params.(*protocol.PublishDiagnosticsParams); ok {
			c.mu.Lock()
			c.diagnostics = append(c.diagnostics, p.Diagnostics...)
			c.mu.Unlock()
		}
	}
	return nil
}

func (c *diagConn) ShowMessage(params *protocol.ShowMessageParams) {}

func (c *diagConn) ShowMessageRequest(params *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}

func (c *diagConn) LogMessage(params *protocol.LogMessageParams) {}

func (c *diagConn) PublishDiagnostics(params *protocol.PublishDiagnosticsParams) error {
	c.mu.Lock()
	c.diagnostics = append(c.diagnostics, params.Diagnostics...)
	c.mu.Unlock()
	return nil
}

func (c *diagConn) RegisterCapability(params *protocol.RegistrationParams) error {
	return nil
}

// formatDiagnostics formats LSP diagnostics into a human-readable string.
func formatDiagnostics(diagnostics []protocol.Diagnostic) string {
	if len(diagnostics) == 0 {
		return "No errors found."
	}

	severityLabels := map[protocol.DiagnosticSeverity]string{
		protocol.SeverityError:       "error",
		protocol.SeverityWarning:     "warning",
		protocol.SeverityInformation: "info",
		protocol.SeverityHint:        "hint",
	}

	var b strings.Builder
	for _, d := range diagnostics {
		label := severityLabels[d.Severity]
		if label == "" {
			label = "error"
		}
		fmt.Fprintf(&b, "[%s] line %d:%d: %s\n",
			label,
			d.Range.Start.Line+1,
			d.Range.Start.Character+1,
			d.Message,
		)
	}
	return b.String()
}

// formatHover formats a hover result into readable text.
func formatHover(result *protocol.Hover) string {
	if result == nil {
		return "No information available."
	}
	return result.Contents.Value
}

// formatSymbols formats document symbols into readable text.
func formatSymbols(symbols []*protocol.DocumentSymbol, indent int) string {
	if len(symbols) == 0 {
		return "No symbols found."
	}

	var b strings.Builder
	prefix := strings.Repeat("  ", indent)
	for _, sym := range symbols {
		detail := ""
		if sym.Detail != "" {
			detail = " — " + sym.Detail
		}
		fmt.Fprintf(&b, "%s%s %s%s\n", prefix, symbolKindName(sym.Kind), sym.Name, detail)
		if len(sym.Children) > 0 {
			b.WriteString(formatSymbols(sym.Children, indent+1))
		}
	}
	return b.String()
}

func symbolKindName(kind protocol.SymbolKind) string {
	names := map[protocol.SymbolKind]string{
		1: "File", 2: "Module", 3: "Namespace", 4: "Package",
		5: "Class", 6: "Method", 7: "Property", 8: "Field",
		9: "Constructor", 10: "Enum", 11: "Interface", 12: "Function",
		13: "Variable", 14: "Constant", 15: "String", 16: "Number",
		17: "Boolean", 18: "Array", 19: "Object", 20: "Key",
		21: "Null", 22: "EnumMember", 23: "Struct", 24: "Event",
		25: "Operator", 26: "TypeParameter",
	}
	if name, ok := names[kind]; ok {
		return name
	}
	return fmt.Sprintf("kind(%d)", kind)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestLSPWrapper -v
```
Expected: all 3 tests PASS.

- [ ] **Step 5: Add tests for hover and symbols**

Append to `internal/mcp/lsp_test.go`:

```go
func TestLSPWrapper_Hover(t *testing.T) {
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)

	// Hover over "Int" on line 0, character ~30
	code := "access(all) fun main(): Int { return 42 }"
	result, err := lsp.Hover(code, 0, 24, "")
	require.NoError(t, err)
	assert.NotNil(t, result, "should get hover info for Int type")
}

func TestLSPWrapper_Symbols(t *testing.T) {
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)

	code := `access(all) contract Foo {
		access(all) resource Bar {}
		access(all) fun baz() {}
	}`
	symbols, err := lsp.Symbols(code, "")
	require.NoError(t, err)
	assert.NotEmpty(t, symbols, "should find symbols in contract")
}

func TestLSPWrapper_Completion(t *testing.T) {
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)

	// Get completions at empty position — should return at least some items
	code := "access(all) fun main() {\n  \n}"
	items, err := lsp.Completion(code, 1, 2, "")
	require.NoError(t, err)
	assert.NotEmpty(t, items, "should get completion items")
}
```

- [ ] **Step 6: Run all LSP tests**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestLSPWrapper -v
```
Expected: all 6 tests PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/mcp/lsp.go internal/mcp/lsp_test.go
git commit -m "Add LSP wrapper with in-process language server"
```

---

### Task 3: Code review rules (cadence_code_review)

**Files:**
- Create: `internal/mcp/audit.go`
- Create: `internal/mcp/audit_test.go`

- [ ] **Step 1: Write failing tests for code review rules**

Create `internal/mcp/audit_test.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeReview_CleanCode(t *testing.T) {
	code := `access(contract) let balance: UFix64
access(contract) fun transfer() {}`
	result := codeReview(code)
	assert.Empty(t, result.Findings, "clean code should have no findings")
}

func TestCodeReview_OverlyPermissiveAccess(t *testing.T) {
	code := `access(all) var balance: UFix64`
	result := codeReview(code)
	assert.NotEmpty(t, result.Findings)
	assert.Equal(t, "overly-permissive-access", result.Findings[0].Rule)
	assert.Equal(t, "warning", string(result.Findings[0].Severity))
}

func TestCodeReview_DeprecatedPub(t *testing.T) {
	code := `pub fun doSomething() {}`
	result := codeReview(code)
	found := false
	for _, f := range result.Findings {
		if f.Rule == "deprecated-pub" {
			found = true
		}
	}
	assert.True(t, found, "should detect deprecated pub keyword")
}

func TestCodeReview_ForceUnwrap(t *testing.T) {
	code := `access(all) fun main() { let x = optional! }`
	result := codeReview(code)
	found := false
	for _, f := range result.Findings {
		if f.Rule == "unsafe-force-unwrap" {
			found = true
		}
	}
	assert.True(t, found, "should detect force-unwrap")
}

func TestCodeReview_HardcodedAddress(t *testing.T) {
	code := `let addr = 0xf233dcee88fe0abe`
	result := codeReview(code)
	found := false
	for _, f := range result.Findings {
		if f.Rule == "hardcoded-address" {
			found = true
		}
	}
	assert.True(t, found, "should detect hardcoded address")
}

func TestCodeReview_AddressImportNotFlagged(t *testing.T) {
	code := `import FungibleToken from 0xf233dcee88fe0abe`
	result := codeReview(code)
	for _, f := range result.Findings {
		assert.NotEqual(t, "hardcoded-address", f.Rule,
			"address imports should not be flagged as hardcoded addresses")
	}
}

func TestCodeReview_FormatResult(t *testing.T) {
	result := ReviewResult{
		Findings: []Finding{
			{Rule: "test", Severity: "warning", Line: 1, Message: "test message"},
		},
		Summary: map[string]int{"warning": 1},
	}
	text := formatReviewResult(result)
	assert.Contains(t, text, "1 issue(s)")
	assert.Contains(t, text, "[WARNING]")
	assert.Contains(t, text, "test message")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestCodeReview -v
```
Expected: FAIL — `codeReview` undefined.

- [ ] **Step 3: Implement code review rules**

Create `internal/mcp/audit.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"fmt"
	"regexp"
	"strings"
)

// Severity represents the severity level of a finding.
type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityNote    Severity = "note"
	SeverityInfo    Severity = "info"
)

// Finding represents a single code review finding.
type Finding struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Line     int      `json:"line"`
	Message  string   `json:"message"`
}

// ReviewResult contains all findings from a code review.
type ReviewResult struct {
	Findings []Finding      `json:"findings"`
	Summary  map[string]int `json:"summary"`
}

type rule struct {
	id       string
	severity Severity
	pattern  *regexp.Regexp
	message  string // static message, or empty if messageFunc is set
	msgFunc  func([]string) string
	perLine  bool // true = match per line (default), false = full text
}

var rules = []rule{
	{
		id:       "overly-permissive-access",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`access\(all\)\s+(var|let)\s+`),
		message:  "State field with access(all) — consider restricting access with entitlements",
		perLine:  true,
	},
	{
		id:       "overly-permissive-function",
		severity: SeverityNote,
		pattern:  regexp.MustCompile(`access\(all\)\s+fun\s+(\w+)`),
		msgFunc: func(m []string) string {
			name := "unknown"
			if len(m) > 1 {
				name = m[1]
			}
			return fmt.Sprintf("Function '%s' has access(all) — review if public access is intended", name)
		},
		perLine: true,
	},
	{
		id:       "deprecated-pub",
		severity: SeverityInfo,
		pattern:  regexp.MustCompile(`\bpub\s+(var|let|fun|resource|struct|event|contract|enum)\b`),
		message:  "`pub` is deprecated in Cadence 1.0 — use `access(all)` or a more restrictive access modifier",
		perLine:  true,
	},
	{
		id:       "unsafe-force-unwrap",
		severity: SeverityNote,
		pattern:  regexp.MustCompile(`[)\w]\s*!`),
		message:  "Force-unwrap (!) used — consider nil-coalescing (??) or optional binding for safer handling",
		perLine:  true,
	},
	{
		id:       "auth-account-exposure",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`\bAuthAccount\b`),
		message:  "AuthAccount reference found — passing AuthAccount gives full account access, use capabilities instead",
		perLine:  true,
	},
	{
		id:       "auth-reference-exposure",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`\bauth\s*\(.*?\)\s*&Account\b`),
		message:  "auth(…) &Account reference found — this grants broad account access, prefer scoped capabilities",
		perLine:  true,
	},
	{
		id:       "hardcoded-address",
		severity: SeverityInfo,
		pattern:  regexp.MustCompile(`0x[0-9a-fA-F]{8,16}\b`),
		message:  "Hardcoded address detected — consider using named address imports for portability",
		perLine:  true,
	},
	{
		id:       "unguarded-capability",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`\.publish\s*\(`),
		message:  "Capability published — verify that proper entitlements guard this capability",
		perLine:  true,
	},
	{
		id:       "resource-loss-destroy",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`destroy\s*\(`),
		message:  "Explicit destroy call — ensure the resource is intentionally being destroyed and not lost",
		perLine:  true,
	},
}

// isAddressImportLine returns true if the line is an import-from-address statement.
var addressImportPattern = regexp.MustCompile(`^\s*import\s+\w[\w, ]*\s+from\s+0x`)

func isAddressImportLine(line string) bool {
	return addressImportPattern.MatchString(line)
}

// codeReview runs static analysis rules against Cadence source code.
func codeReview(code string) ReviewResult {
	var findings []Finding
	lines := strings.Split(code, "\n")

	for _, r := range rules {
		if !r.perLine {
			continue // multi-line rules handled below
		}
		for i, line := range lines {
			// Skip address import lines for hardcoded-address rule
			if r.id == "hardcoded-address" && isAddressImportLine(line) {
				continue
			}

			match := r.pattern.FindStringSubmatch(line)
			if match == nil {
				continue
			}

			msg := r.message
			if r.msgFunc != nil {
				msg = r.msgFunc(match)
			}

			findings = append(findings, Finding{
				Rule:     r.id,
				Severity: r.severity,
				Line:     i + 1,
				Message:  msg,
			})
		}
	}

	summary := map[string]int{}
	for _, f := range findings {
		summary[string(f.Severity)]++
	}

	return ReviewResult{
		Findings: findings,
		Summary:  summary,
	}
}

// formatReviewResult formats a ReviewResult into a human-readable string.
func formatReviewResult(result ReviewResult) string {
	var b strings.Builder

	total := len(result.Findings)
	fmt.Fprintf(&b, "## Code Review Results\n")
	fmt.Fprintf(&b, "Found %d issue(s): %d warning, %d note, %d info\n\n",
		total,
		result.Summary["warning"],
		result.Summary["note"],
		result.Summary["info"],
	)

	if total == 0 {
		b.WriteString("No issues detected.\n")
		return b.String()
	}

	for _, f := range result.Findings {
		fmt.Fprintf(&b, "- [%s] Line %d: (%s) %s\n",
			strings.ToUpper(string(f.Severity)),
			f.Line,
			f.Rule,
			f.Message,
		)
	}

	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestCodeReview -v
```
Expected: all 7 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/mcp/audit.go internal/mcp/audit_test.go
git commit -m "Add code review rules for cadence_code_review tool"
```

---

### Task 4: Tool handler implementations

**Files:**
- Create: `internal/mcp/tools.go`
- Modify: `internal/mcp/mcp.go` (wire tools into server)

- [ ] **Step 1: Implement all 9 tool handlers**

Create `internal/mcp/tools.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/arguments"
)

// mcpContext holds shared state for all tool handlers.
type mcpContext struct {
	lsp   *LSPWrapper
	state *flowkit.State // may be nil if no flow.json
}

// --- LSP Tool Handlers ---

func (m *mcpContext) cadenceCheck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	diags, err := m.lsp.Check(code, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LSP error: %v", err)), nil
	}
	return mcp.NewToolResultText(formatDiagnostics(diags)), nil
}

func (m *mcpContext) cadenceHover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	line, err := req.RequireInt("line")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	character, err := req.RequireInt("character")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	result, err := m.lsp.Hover(code, line, character, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LSP error: %v", err)), nil
	}
	return mcp.NewToolResultText(formatHover(result)), nil
}

func (m *mcpContext) cadenceDefinition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	line, err := req.RequireInt("line")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	character, err := req.RequireInt("character")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	loc, err := m.lsp.Definition(code, line, character, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LSP error: %v", err)), nil
	}
	if loc == nil {
		return mcp.NewToolResultText("No definition found."), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Definition: %s at line %d:%d",
		loc.URI,
		loc.Range.Start.Line+1,
		loc.Range.Start.Character+1,
	)), nil
}

func (m *mcpContext) cadenceSymbols(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	symbols, err := m.lsp.Symbols(code, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LSP error: %v", err)), nil
	}
	return mcp.NewToolResultText(formatSymbols(symbols, 0)), nil
}

func (m *mcpContext) cadenceCompletion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	line, err := req.RequireInt("line")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	character, err := req.RequireInt("character")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	items, err := m.lsp.Completion(code, line, character, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LSP error: %v", err)), nil
	}
	if len(items) == 0 {
		return mcp.NewToolResultText("No completions available."), nil
	}

	var b strings.Builder
	for _, item := range items {
		detail := ""
		if item.Detail != "" {
			detail = " — " + item.Detail
		}
		fmt.Fprintf(&b, "%s%s\n", item.Label, detail)
	}
	return mcp.NewToolResultText(b.String()), nil
}

// --- Audit Tool Handlers ---

func (m *mcpContext) getContractSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	address, err := req.RequireString("address")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	gw, err := createGateway(m.state, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gateway error: %v", err)), nil
	}

	addr := flow.HexToAddress(address)
	account, err := gw.GetAccount(ctx, addr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error fetching account: %v", err)), nil
	}

	type contractEntry struct {
		Name    string   `json:"name"`
		Size    int      `json:"size"`
		Imports []string `json:"imports,omitempty"`
	}

	var entries []contractEntry
	for name, code := range account.Contracts {
		entries = append(entries, contractEntry{
			Name: name,
			Size: len(code),
		})
	}

	result, err := json.MarshalIndent(map[string]any{
		"address":   address,
		"network":   network,
		"contracts": entries,
	}, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("JSON error: %v", err)), nil
	}
	return mcp.NewToolResultText(string(result)), nil
}

func (m *mcpContext) getContractCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	address, err := req.RequireString("address")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	contractName, _ := req.GetString("contract_name", "")
	network, _ := req.GetString("network", "mainnet")

	gw, err := createGateway(m.state, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gateway error: %v", err)), nil
	}

	addr := flow.HexToAddress(address)
	account, err := gw.GetAccount(ctx, addr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error fetching account: %v", err)), nil
	}

	var parts []string
	for name, code := range account.Contracts {
		if contractName != "" && name != contractName {
			continue
		}
		parts = append(parts, fmt.Sprintf("// === %s (%s) ===\n\n%s", name, address, string(code)))
	}

	if len(parts) == 0 {
		if contractName != "" {
			names := make([]string, 0, len(account.Contracts))
			for name := range account.Contracts {
				names = append(names, name)
			}
			return mcp.NewToolResultError(fmt.Sprintf(
				"Contract '%s' not found on %s. Available: %s",
				contractName, address, strings.Join(names, ", "),
			)), nil
		}
		return mcp.NewToolResultText("No contracts found on this address."), nil
	}

	return mcp.NewToolResultText(strings.Join(parts, "\n\n")), nil
}

func (m *mcpContext) cadenceCodeReview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")

	// Run static analysis rules
	result := codeReview(code)
	text := formatReviewResult(result)

	// Also run LSP type check if available
	if m.lsp != nil {
		diags, err := m.lsp.Check(code, network)
		if err == nil && len(diags) > 0 {
			text += "\n## Type Check (LSP)\n" + formatDiagnostics(diags)
		}
	}

	return mcp.NewToolResultText(text), nil
}

func (m *mcpContext) cadenceExecuteScript(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	network, _ := req.GetString("network", "mainnet")
	argsJSON, _ := req.GetString("args", "[]")

	gw, err := createGateway(m.state, network)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gateway error: %v", err)), nil
	}

	// Parse script arguments
	var argStrings []string
	if err := json.Unmarshal([]byte(argsJSON), &argStrings); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid args format: %v", err)), nil
	}

	cadenceArgs, err := arguments.ParseWithoutType(argStrings)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Argument parse error: %v", err)), nil
	}

	value, err := gw.ExecuteScript(ctx, flowkit.Script{
		Code: []byte(code),
		Args: cadenceArgs,
	}, flowkit.LatestScriptQuery)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Script execution failed:\n%v", err)), nil
	}

	return mcp.NewToolResultText(value.String()), nil
}
```

- [ ] **Step 2: Wire tools into the MCP server**

Update `internal/mcp/mcp.go` — replace the `runMCP` function:

```go
func runMCP(cmd *cobra.Command, args []string) {
	// Try to load flow.json for custom network configs
	loader := &afero.Afero{Fs: afero.NewOsFs()}
	state, _ := flowkit.Load(config.DefaultPaths(), loader)

	// Initialize LSP wrapper (disable flow client for now — network queries
	// use gateways directly, and the LSP's flow client would prompt for
	// flow.json interactively which doesn't work over MCP stdio)
	lsp, err := NewLSPWrapper(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: LSP initialization failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "LSP tools will be unavailable.\n")
	}

	mctx := &mcpContext{
		lsp:   lsp,
		state: state,
	}

	s := mcpserver.NewMCPServer("cadence-mcp", "1.0.0")
	registerTools(s, mctx)

	if err := mcpserver.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func registerTools(s *mcpserver.MCPServer, mctx *mcpContext) {
	networkParam := mcp.WithString("network",
		mcp.Description("Flow network: mainnet, testnet, or emulator (default: mainnet)"),
		mcp.Enum("mainnet", "testnet", "emulator"),
	)

	// --- LSP Tools ---

	if mctx.lsp != nil {
		s.AddTool(mcp.NewTool("cadence_check",
			mcp.WithDescription("Check Cadence smart contract code for syntax and type errors. Returns diagnostics."),
			mcp.WithString("code", mcp.Required(), mcp.Description("Cadence source code to check")),
			mcp.WithString("filename", mcp.Description("Virtual filename (default: check.cdc)")),
			networkParam,
		), mctx.cadenceCheck)

		s.AddTool(mcp.NewTool("cadence_hover",
			mcp.WithDescription("Get type information and documentation for a symbol at a given position in Cadence code."),
			mcp.WithString("code", mcp.Required(), mcp.Description("Cadence source code")),
			mcp.WithNumber("line", mcp.Required(), mcp.Description("0-based line number")),
			mcp.WithNumber("character", mcp.Required(), mcp.Description("0-based column number")),
			mcp.WithString("filename", mcp.Description("Virtual filename")),
			networkParam,
		), mctx.cadenceHover)

		s.AddTool(mcp.NewTool("cadence_definition",
			mcp.WithDescription("Find the definition location of a symbol at a given position in Cadence code."),
			mcp.WithString("code", mcp.Required(), mcp.Description("Cadence source code")),
			mcp.WithNumber("line", mcp.Required(), mcp.Description("0-based line number")),
			mcp.WithNumber("character", mcp.Required(), mcp.Description("0-based column number")),
			mcp.WithString("filename", mcp.Description("Virtual filename")),
			networkParam,
		), mctx.cadenceDefinition)

		s.AddTool(mcp.NewTool("cadence_symbols",
			mcp.WithDescription("List all symbols (contracts, resources, functions, events, etc.) in Cadence code."),
			mcp.WithString("code", mcp.Required(), mcp.Description("Cadence source code")),
			mcp.WithString("filename", mcp.Description("Virtual filename")),
			networkParam,
		), mctx.cadenceSymbols)

		s.AddTool(mcp.NewTool("cadence_completion",
			mcp.WithDescription("Get code completions at a position in Cadence code. Returns available members, methods, and keywords."),
			mcp.WithString("code", mcp.Required(), mcp.Description("Cadence source code")),
			mcp.WithNumber("line", mcp.Required(), mcp.Description("0-based line number")),
			mcp.WithNumber("character", mcp.Required(), mcp.Description("0-based column number")),
			mcp.WithString("filename", mcp.Description("Virtual filename")),
			networkParam,
		), mctx.cadenceCompletion)
	}

	// --- Audit Tools ---

	s.AddTool(mcp.NewTool("get_contract_source",
		mcp.WithDescription("Fetch on-chain contract manifest from a Flow address: lists all contracts with names and sizes."),
		mcp.WithString("address", mcp.Required(), mcp.Description("Flow address (0x...)")),
		networkParam,
	), mctx.getContractSource)

	s.AddTool(mcp.NewTool("get_contract_code",
		mcp.WithDescription("Fetch the source code of contracts from a Flow address."),
		mcp.WithString("address", mcp.Required(), mcp.Description("Flow address (0x...)")),
		mcp.WithString("contract_name", mcp.Description("Name of specific contract to fetch. If omitted, returns all.")),
		networkParam,
	), mctx.getContractCode)

	s.AddTool(mcp.NewTool("cadence_code_review",
		mcp.WithDescription("Review Cadence code for common issues and best practices. Uses pattern matching to flag potential problems — not a substitute for a proper audit."),
		mcp.WithString("code", mcp.Required(), mcp.Description("Cadence source code to review")),
		networkParam,
	), mctx.cadenceCodeReview)

	s.AddTool(mcp.NewTool("cadence_execute_script",
		mcp.WithDescription("Execute a read-only Cadence script on the Flow network. Scripts can query on-chain state. Cannot modify state."),
		mcp.WithString("code", mcp.Required(), mcp.Description("Cadence script code (must have `access(all) fun main()` entry point)")),
		mcp.WithString("args", mcp.Description(`Script arguments as JSON array of strings in "Type:Value" format, e.g. ["Address:0x1654653399040a61", "UFix64:10.0"]`)),
		networkParam,
	), mctx.cadenceExecuteScript)
}
```

- [ ] **Step 3: Verify it builds**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go build ./cmd/flow/...
```
Expected: builds with no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/mcp/tools.go internal/mcp/mcp.go
git commit -m "Add tool handler implementations and wire into MCP server"
```

---

### Task 5: Tool handler tests

**Files:**
- Create: `internal/mcp/tools_test.go`

- [ ] **Step 1: Write tool handler tests**

Create `internal/mcp/tools_test.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext(t *testing.T) *mcpContext {
	t.Helper()
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)
	return &mcpContext{lsp: lsp}
}

func makeRequest(args map[string]any) mcp.CallToolRequest {
	raw, _ := json.Marshal(args)
	var params struct {
		Arguments map[string]any `json:"arguments"`
	}
	params.Arguments = args
	rawParams, _ := json.Marshal(params)
	var req mcp.CallToolRequest
	json.Unmarshal(rawParams, &req)
	// mcp-go uses req.Params.Arguments
	req.Params.Arguments = args
	return req
}

func TestTool_CadenceCheck_Valid(t *testing.T) {
	mctx := newTestContext(t)
	req := makeRequest(map[string]any{
		"code": "access(all) fun main() {}",
	})

	result, err := mctx.cadenceCheck(context.Background(), req)
	require.NoError(t, err)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "No errors found")
}

func TestTool_CadenceCheck_Invalid(t *testing.T) {
	mctx := newTestContext(t)
	req := makeRequest(map[string]any{
		"code": "access(all) fun main() { let x: Int = \"bad\" }",
	})

	result, err := mctx.cadenceCheck(context.Background(), req)
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "error")
}

func TestTool_CadenceCheck_MissingCode(t *testing.T) {
	mctx := newTestContext(t)
	req := makeRequest(map[string]any{})

	result, err := mctx.cadenceCheck(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestTool_CadenceSymbols(t *testing.T) {
	mctx := newTestContext(t)
	req := makeRequest(map[string]any{
		"code": `access(all) contract Foo {
			access(all) fun bar() {}
		}`,
	})

	result, err := mctx.cadenceSymbols(context.Background(), req)
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "Foo")
}

func TestTool_CadenceCodeReview(t *testing.T) {
	mctx := newTestContext(t)
	req := makeRequest(map[string]any{
		"code": "access(all) var balance: UFix64",
	})

	result, err := mctx.cadenceCodeReview(context.Background(), req)
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "overly-permissive-access")
}
```

- [ ] **Step 2: Run tests**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestTool -v
```
Expected: all tests PASS.

Note: The `makeRequest` helper may need adjustment based on the exact mcp-go `CallToolRequest` struct. Check the struct definition after adding the dependency and adjust accordingly.

- [ ] **Step 3: Commit**

```bash
git add internal/mcp/tools_test.go
git commit -m "Add tool handler tests"
```

---

### Task 6: End-to-end verification and license headers

**Files:**
- Modify: all files in `internal/mcp/`

- [ ] **Step 1: Verify license headers on all files**

Run:
```bash
make check-headers
```
Expected: PASS. All Go files in `internal/mcp/` already have the Apache 2.0 header from the code above. If any are flagged, add the header.

- [ ] **Step 2: Run linter**

Run:
```bash
make lint
```
Expected: PASS with no new lint issues.

- [ ] **Step 3: Run full test suite**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -v
```
Expected: all tests PASS.

- [ ] **Step 4: Manual smoke test**

Run the MCP server interactively to verify it starts and responds:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | go run ./cmd/flow mcp
```

Expected: JSON response with server capabilities including the tool list.

- [ ] **Step 5: Commit any fixes**

If any fixes were needed:
```bash
git add internal/mcp/
git commit -m "Fix lint and license header issues"
```

---

### Task 7: Integration tests for network tools (optional, behind flag)

**Files:**
- Create: `internal/mcp/integration_test.go`

These tests hit the real network and are skipped unless `SKIP_NETWORK_TESTS` is unset.

- [ ] **Step 1: Write integration tests**

Create `internal/mcp/integration_test.go`:

```go
/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoNetwork(t *testing.T) {
	t.Helper()
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("Skipping network test (SKIP_NETWORK_TESTS is set)")
	}
}

func TestIntegration_GetContractSource(t *testing.T) {
	skipIfNoNetwork(t)
	mctx := &mcpContext{state: nil}

	req := makeRequest(map[string]any{
		"address": "0x1654653399040a61",
		"network": "mainnet",
	})

	result, err := mctx.getContractSource(context.Background(), req)
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "FungibleToken")
}

func TestIntegration_GetContractCode(t *testing.T) {
	skipIfNoNetwork(t)
	mctx := &mcpContext{state: nil}

	req := makeRequest(map[string]any{
		"address":       "0x1654653399040a61",
		"contract_name": "FungibleToken",
		"network":       "mainnet",
	})

	result, err := mctx.getContractCode(context.Background(), req)
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "FungibleToken")
	assert.Contains(t, text, "access(all) contract interface")
}

func TestIntegration_ExecuteScript(t *testing.T) {
	skipIfNoNetwork(t)
	mctx := &mcpContext{state: nil}

	req := makeRequest(map[string]any{
		"code":    `access(all) fun main(): Int { return 42 }`,
		"network": "mainnet",
	})

	result, err := mctx.cadenceExecuteScript(context.Background(), req)
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "42")
}
```

- [ ] **Step 2: Run integration tests (if network available)**

Run:
```bash
CGO_ENABLED=1 CGO_CFLAGS="-O2 -D__BLST_PORTABLE__ -std=gnu11" go test ./internal/mcp/... -run TestIntegration -v
```
Expected: PASS (or skip if `SKIP_NETWORK_TESTS` is set).

- [ ] **Step 3: Commit**

```bash
git add internal/mcp/integration_test.go
git commit -m "Add integration tests for network tools"
```
