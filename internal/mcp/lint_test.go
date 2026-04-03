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

func TestLintCode_ValidCode(t *testing.T) {
	t.Parallel()

	code := `access(all) fun main(): String { return "hello" }`
	diags, err := lintCode(code)
	require.NoError(t, err)
	// Valid code should only have lint hints, no errors
	for _, d := range diags {
		assert.NotEqual(t, "error", d.Category)
		assert.NotEqual(t, "semantic-error", d.Category)
		assert.NotEqual(t, "syntax-error", d.Category)
	}
}

func TestLintCode_SyntaxError(t *testing.T) {
	t.Parallel()

	code := `access(all) fun main( {`
	diags, err := lintCode(code)
	require.NoError(t, err)
	assert.NotEmpty(t, diags, "syntax errors should produce diagnostics")
}

func TestLintCode_TypeError(t *testing.T) {
	t.Parallel()

	code := `access(all) fun main(): String { return 42 }`
	diags, err := lintCode(code)
	require.NoError(t, err)
	assert.NotEmpty(t, diags, "type errors should produce diagnostics")

	hasError := false
	for _, d := range diags {
		if d.Category == "error" || d.Category == "semantic-error" {
			hasError = true
		}
	}
	assert.True(t, hasError, "should have error-level diagnostics")
}

func TestLintCode_AnalyzersRun(t *testing.T) {
	t.Parallel()

	// Redundant cast should be detected by lint analyzers
	code := `
		access(all) fun main(): Int {
			let x: Int = 42
			return x as Int
		}
	`
	diags, err := lintCode(code)
	require.NoError(t, err)
	// Should have at least one lint diagnostic (redundant cast or similar)
	assert.NotEmpty(t, diags, "lint analyzers should produce diagnostics for redundant cast")
}

func TestFormatLintDiagnostics_Empty(t *testing.T) {
	t.Parallel()

	result := formatLintDiagnostics(nil)
	assert.Contains(t, result, "Lint passed")
}
