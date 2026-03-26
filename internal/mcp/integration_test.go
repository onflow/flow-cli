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

	mcplib "github.com/mark3labs/mcp-go/mcp"
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
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"address": "0x1654653399040a61",
		"network": "mainnet",
	}

	result, err := mctx.getContractSource(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	text := result.Content[0].(mcplib.TextContent).Text
	assert.Contains(t, text, "FungibleToken")
}

func TestIntegration_GetContractCode(t *testing.T) {
	skipIfNoNetwork(t)

	mctx := &mcpContext{state: nil}
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"address":       "0x1654653399040a61",
		"contract_name": "FungibleToken",
		"network":       "mainnet",
	}

	result, err := mctx.getContractCode(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	text := result.Content[0].(mcplib.TextContent).Text
	assert.Contains(t, text, "FungibleToken")
}

func TestIntegration_ExecuteScript(t *testing.T) {
	skipIfNoNetwork(t)

	mctx := &mcpContext{state: nil}
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"code":    `access(all) fun main(): Int { return 42 }`,
		"network": "mainnet",
	}

	result, err := mctx.cadenceExecuteScript(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	text := result.Content[0].(mcplib.TextContent).Text
	assert.Contains(t, text, "42")
}
