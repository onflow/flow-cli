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
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext(t *testing.T) *mcpContext {
	t.Helper()
	lsp, err := NewLSPWrapper(false)
	require.NoError(t, err)
	return &mcpContext{lsp: lsp}
}

func TestTool_CadenceCheck_Valid(t *testing.T) {
	t.Parallel()
	mctx := newTestContext(t)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"code": `access(all) fun hello(): String { return "hello" }`,
	}

	result, err := mctx.cadenceCheck(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent := result.Content[0].(mcplib.TextContent)
	assert.Contains(t, textContent.Text, "No errors found")
}

func TestTool_CadenceCheck_Invalid(t *testing.T) {
	t.Parallel()
	mctx := newTestContext(t)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"code": `access(all) fun hello(): String { return 42 }`,
	}

	result, err := mctx.cadenceCheck(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent := result.Content[0].(mcplib.TextContent)
	assert.Contains(t, textContent.Text, "Error")
}

func TestTool_CadenceCheck_MissingCode(t *testing.T) {
	t.Parallel()
	mctx := newTestContext(t)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := mctx.cadenceCheck(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestTool_CadenceSymbols(t *testing.T) {
	t.Parallel()
	mctx := newTestContext(t)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"code": `
access(all) contract MyContract {
	access(all) fun greet(): String {
		return "hi"
	}
}
`,
	}

	result, err := mctx.cadenceSymbols(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent := result.Content[0].(mcplib.TextContent)
	assert.Contains(t, textContent.Text, "MyContract")
}
