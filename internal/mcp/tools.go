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

	"sort"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/onflow/cadence"
	flow "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/arguments"
)

// mcpContext holds shared dependencies for all MCP tool handlers.
type mcpContext struct {
	lsp   *LSPWrapper
	state *flowkit.State // may be nil
}

// resolveCode extracts the required "code" parameter from the request.
func resolveCode(req mcplib.CallToolRequest) (string, error) {
	return req.RequireString("code")
}

// parseAddress parses a Flow address string and validates it is not empty.
func parseAddress(address string) (flow.Address, error) {
	addr := flow.HexToAddress(address)
	if addr == flow.EmptyAddress {
		return flow.EmptyAddress, fmt.Errorf("invalid Flow address: %q", address)
	}
	return addr, nil
}

// registerTools registers all MCP tools on the given server.
func registerTools(s *mcpserver.MCPServer, mctx *mcpContext) {
	// LSP tools — only register if the LSP wrapper is available.
	if mctx.lsp != nil {
		s.AddTool(
			mcplib.NewTool("cadence_check",
				mcplib.WithDescription("Check Cadence code for syntax and type errors."),
				mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence source code to check")),
			),
			mctx.cadenceCheck,
		)

		s.AddTool(
			mcplib.NewTool("cadence_hover",
				mcplib.WithDescription("Get type information for a symbol at a position in Cadence code."),
				mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence source code")),

				mcplib.WithNumber("line", mcplib.Required(), mcplib.Description("0-based line number")),
				mcplib.WithNumber("character", mcplib.Required(), mcplib.Description("0-based column number")),
			),
			mctx.cadenceHover,
		)

		s.AddTool(
			mcplib.NewTool("cadence_definition",
				mcplib.WithDescription("Find where a symbol is defined in Cadence code."),
				mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence source code")),

				mcplib.WithNumber("line", mcplib.Required(), mcplib.Description("0-based line number")),
				mcplib.WithNumber("character", mcplib.Required(), mcplib.Description("0-based column number")),
			),
			mctx.cadenceDefinition,
		)

		s.AddTool(
			mcplib.NewTool("cadence_symbols",
				mcplib.WithDescription("List all symbols in Cadence code."),
				mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence source code")),
			),
			mctx.cadenceSymbols,
		)

		s.AddTool(
			mcplib.NewTool("cadence_completion",
				mcplib.WithDescription("Get completion suggestions at a position in Cadence code."),
				mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence source code")),

				mcplib.WithNumber("line", mcplib.Required(), mcplib.Description("0-based line number")),
				mcplib.WithNumber("character", mcplib.Required(), mcplib.Description("0-based column number")),
			),
			mctx.cadenceCompletion,
		)
	}

	// Audit / network tools — always registered.
	s.AddTool(
		mcplib.NewTool("get_contract_source",
			mcplib.WithDescription("Fetch on-chain contract manifest (names and sizes) for a Flow account"),
			mcplib.WithString("address", mcplib.Required(), mcplib.Description("Flow account address (hex, with or without 0x prefix)")),
			mcplib.WithString("network", mcplib.Description("Flow network to query"), mcplib.Enum("mainnet", "testnet", "emulator")),
		),
		mctx.getContractSource,
	)

	s.AddTool(
		mcplib.NewTool("get_contract_code",
			mcplib.WithDescription("Fetch contract source code from a Flow account"),
			mcplib.WithString("address", mcplib.Required(), mcplib.Description("Flow account address (hex, with or without 0x prefix)")),
			mcplib.WithString("contract_name", mcplib.Description("Specific contract name to retrieve; omit for all contracts")),
			mcplib.WithString("network", mcplib.Description("Flow network to query"), mcplib.Enum("mainnet", "testnet", "emulator")),
		),
		mctx.getContractCode,
	)

	s.AddTool(
		mcplib.NewTool("cadence_code_review",
			mcplib.WithDescription("Review Cadence code for common issues and anti-patterns."),
			mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence source code to review")),
		),
		mctx.cadenceCodeReview,
	)

	s.AddTool(
		mcplib.NewTool("cadence_execute_script",
			mcplib.WithDescription("Execute a read-only Cadence script on-chain."),
			mcplib.WithString("code", mcplib.Required(), mcplib.Description("Cadence script source code")),

			mcplib.WithString("network", mcplib.Description("Flow network to execute against"), mcplib.Enum("mainnet", "testnet", "emulator")),
			mcplib.WithString("arguments", mcplib.Description("JSON array of arguments as strings, e.g. [\"String:hello\", \"UFix64:1.0\"]")),
		),
		mctx.cadenceExecuteScript,
	)
}

// ---------------------------------------------------------------------------
// LSP tool handlers
// ---------------------------------------------------------------------------

func (m *mcpContext) cadenceCheck(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	diags, err := m.lsp.Check(code)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("LSP check failed: %v", err)), nil
	}
	return mcplib.NewToolResultText(formatDiagnostics(diags)), nil
}

func (m *mcpContext) cadenceHover(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	line, err := req.RequireInt("line")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	character, err := req.RequireInt("character")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	result, err := m.lsp.Hover(code, line, character)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("LSP hover failed: %v", err)), nil
	}
	return mcplib.NewToolResultText(formatHover(result)), nil
}

func (m *mcpContext) cadenceDefinition(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	line, err := req.RequireInt("line")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	character, err := req.RequireInt("character")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	loc, err := m.lsp.Definition(code, line, character)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("LSP definition failed: %v", err)), nil
	}
	if loc == nil {
		return mcplib.NewToolResultText("No definition found."), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("%s line %d:%d",
		loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)), nil
}

func (m *mcpContext) cadenceSymbols(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	symbols, err := m.lsp.Symbols(code)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("LSP symbols failed: %v", err)), nil
	}
	return mcplib.NewToolResultText(formatSymbols(symbols, 0)), nil
}

func (m *mcpContext) cadenceCompletion(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	line, err := req.RequireInt("line")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	character, err := req.RequireInt("character")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	items, err := m.lsp.Completion(code, line, character)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("LSP completion failed: %v", err)), nil
	}

	var b strings.Builder
	for _, item := range items {
		b.WriteString(item.Label)
		if item.Detail != "" {
			fmt.Fprintf(&b, " — %s", item.Detail)
		}
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return mcplib.NewToolResultText("No completions available."), nil
	}
	return mcplib.NewToolResultText(b.String()), nil
}

// ---------------------------------------------------------------------------
// Audit / network tool handlers
// ---------------------------------------------------------------------------

func (m *mcpContext) getContractSource(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	address, err := req.RequireString("address")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	network := req.GetString("network", "mainnet")

	gw, err := createGateway(m.state, network)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to create gateway: %v", err)), nil
	}

	addr, err := parseAddress(address)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	account, err := gw.GetAccount(ctx, addr)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to get account: %v", err)), nil
	}

	type contractInfo struct {
		Name string `json:"name"`
		Size int    `json:"size"`
	}

	contracts := make([]contractInfo, 0, len(account.Contracts))
	for name, code := range account.Contracts {
		contracts = append(contracts, contractInfo{Name: name, Size: len(code)})
	}
	sort.Slice(contracts, func(i, j int) bool {
		return contracts[i].Name < contracts[j].Name
	})

	result := struct {
		Address   string         `json:"address"`
		Contracts []contractInfo `json:"contracts"`
	}{
		Address:   addr.String(),
		Contracts: contracts,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcplib.NewToolResultText(string(data)), nil
}

func (m *mcpContext) getContractCode(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	address, err := req.RequireString("address")
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	contractName := req.GetString("contract_name", "")
	network := req.GetString("network", "mainnet")

	gw, err := createGateway(m.state, network)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to create gateway: %v", err)), nil
	}

	addr, err := parseAddress(address)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	account, err := gw.GetAccount(ctx, addr)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to get account: %v", err)), nil
	}

	if contractName != "" {
		code, ok := account.Contracts[contractName]
		if !ok {
			return mcplib.NewToolResultError(fmt.Sprintf("contract %q not found on account %s", contractName, addr.String())), nil
		}
		return mcplib.NewToolResultText(string(code)), nil
	}

	// Return all contracts.
	var b strings.Builder
	names := make([]string, 0, len(account.Contracts))
	for name := range account.Contracts {
		names = append(names, name)
	}
	sort.Strings(names)

	for i, name := range names {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "// === %s ===\n%s", name, string(account.Contracts[name]))
	}
	if b.Len() == 0 {
		return mcplib.NewToolResultText("No contracts found on this account."), nil
	}
	return mcplib.NewToolResultText(b.String()), nil
}

func (m *mcpContext) cadenceCodeReview(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}

	result := codeReview(code)
	text := formatReviewResult(result)

	// If LSP is available, also run a check and append diagnostics.
	if m.lsp != nil {
		diags, lspErr := m.lsp.Check(code)
		if lspErr == nil && len(diags) > 0 {
			text += "\nLSP diagnostics:\n" + formatDiagnostics(diags)
		}
	}

	return mcplib.NewToolResultText(text), nil
}

func (m *mcpContext) cadenceExecuteScript(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	code, err := resolveCode(req)
	if err != nil {
		return mcplib.NewToolResultError(err.Error()), nil
	}
	network := req.GetString("network", "mainnet")
	argsJSON := req.GetString("arguments", "")

	gw, err := createGateway(m.state, network)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("failed to create gateway: %v", err)), nil
	}

	var cadenceArgs []cadence.Value
	if argsJSON != "" {
		var argStrings []string
		if jsonErr := json.Unmarshal([]byte(argsJSON), &argStrings); jsonErr != nil {
			return mcplib.NewToolResultError(fmt.Sprintf("failed to parse arguments JSON: %v", jsonErr)), nil
		}
		parsed, parseErr := arguments.ParseWithoutType(argStrings, []byte(code), "")
		if parseErr != nil {
			return mcplib.NewToolResultError(fmt.Sprintf("failed to parse arguments: %v", parseErr)), nil
		}
		cadenceArgs = parsed
	}

	val, err := gw.ExecuteScript(ctx, []byte(code), cadenceArgs)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("script execution failed: %v", err)), nil
	}

	return mcplib.NewToolResultText(val.String()), nil
}

