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
	"errors"
	"fmt"
	"os"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

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
  cadence_lint           Run Cadence lint analyzers (AST-based)
  cadence_code_review    Review Cadence code for common issues
  cadence_execute_script Execute a read-only Cadence script on-chain`,
	Run: runMCP,
}

func runMCP(cmd *cobra.Command, args []string) {
	// Try to load flow.json for custom network configs
	loader := &afero.Afero{Fs: afero.NewOsFs()}
	state, err := flowkit.Load(config.DefaultPaths(), loader)
	if err != nil && !errors.Is(err, config.ErrDoesNotExist) {
		fmt.Fprintf(os.Stderr, "Warning: failed to load flow.json: %v\n", err)
	}

	// Initialize the LSP wrapper (without flow client for MCP use).
	var lsp *LSPWrapper
	if w, err := NewLSPWrapper(false); err == nil {
		lsp = w
	} else {
		fmt.Fprintf(os.Stderr, "Warning: LSP initialization failed, LSP tools will be unavailable: %v\n", err)
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
// Uses a secure gateway when the network has a configured key.
func createGateway(state *flowkit.State, network string) (gateway.Gateway, error) {
	net, err := resolveNetwork(state, network)
	if err != nil {
		return nil, err
	}
	if net.Key != "" {
		return gateway.NewSecureGrpcGateway(*net)
	}
	return gateway.NewGrpcGateway(*net)
}
