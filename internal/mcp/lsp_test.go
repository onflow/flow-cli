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

func newTestWrapper(t *testing.T) *LSPWrapper {
	t.Helper()
	w, err := NewLSPWrapper(false)
	require.NoError(t, err)
	require.NotNil(t, w)
	return w
}

func TestLSPWrapper_Check_ValidCode(t *testing.T) {
	w := newTestWrapper(t)

	code := `
		access(all) fun hello(): String {
			return "hello"
		}
	`
	diags, err := w.Check(code, "")
	require.NoError(t, err)
	assert.Empty(t, diags, "valid code should produce no diagnostics")
}

func TestLSPWrapper_Check_InvalidCode(t *testing.T) {
	w := newTestWrapper(t)

	// Type mismatch: returning Int from a String function
	code := `
		access(all) fun hello(): String {
			return 42
		}
	`
	diags, err := w.Check(code, "")
	require.NoError(t, err)
	assert.NotEmpty(t, diags, "type mismatch should produce diagnostics")
}

func TestLSPWrapper_Check_SyntaxError(t *testing.T) {
	w := newTestWrapper(t)

	code := `
		access(all) fun hello( {
	`
	diags, err := w.Check(code, "")
	require.NoError(t, err)
	assert.NotEmpty(t, diags, "syntax error should produce diagnostics")
}

func TestLSPWrapper_Hover(t *testing.T) {
	w := newTestWrapper(t)

	code := `
access(all) fun hello(): String {
	return "hello"
}
`
	// Hover over "String" return type — line 1 (0-based), find the position of "String"
	result, err := w.Hover(code, 1, 25, "")
	require.NoError(t, err)
	// Hover may or may not return a result depending on the position;
	// we just verify it doesn't error. If non-nil, it should have contents.
	if result != nil {
		assert.NotEmpty(t, result.Contents.Value)
	}
}

func TestLSPWrapper_Symbols(t *testing.T) {
	w := newTestWrapper(t)

	code := `
access(all) contract MyContract {
	access(all) fun greet(): String {
		return "hi"
	}
}
`
	symbols, err := w.Symbols(code, "")
	require.NoError(t, err)
	require.NotEmpty(t, symbols, "contract with members should have symbols")

	// The top-level symbol should be the contract
	assert.Equal(t, "MyContract", symbols[0].Name)
}

func TestLSPWrapper_Completion(t *testing.T) {
	w := newTestWrapper(t)

	// Inside a function body, the LSP should offer completions
	code := `
access(all) fun main() {
	let x: String = "hello"
	x.
}
`
	// Position right after "x." — line 3, character 3
	items, err := w.Completion(code, 3, 3, "")
	require.NoError(t, err)
	// String methods should appear as completions
	assert.NotEmpty(t, items, "should get completion items for String methods")
}
