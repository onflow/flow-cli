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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeReview_CleanCode(t *testing.T) {
	t.Parallel()
	code := `
access(all) contract MyContract {
    access(self) var counter: Int

    init() {
        self.counter = 0
    }

    access(all) fun getCounter(): Int {
        return self.counter
    }
}
`
	result := codeReview(code)
	// overly-permissive-function will fire for "getCounter", so we only check
	// that no warnings or hard-error-class findings exist from the truly
	// problematic rules.
	for _, f := range result.Findings {
		assert.NotEqual(t, "overly-permissive-access", f.Rule)
		assert.NotEqual(t, "deprecated-pub", f.Rule)
		assert.NotEqual(t, "unsafe-force-unwrap", f.Rule)
		assert.NotEqual(t, "auth-account-exposure", f.Rule)
		assert.NotEqual(t, "hardcoded-address", f.Rule)
	}
}

func TestCodeReview_OverlyPermissiveAccess(t *testing.T) {
	t.Parallel()
	code := `
access(all) contract MyContract {
    access(all) var balance: UFix64
    access(all) let name: String
}
`
	result := codeReview(code)

	var found []Finding
	for _, f := range result.Findings {
		if f.Rule == "overly-permissive-access" {
			found = append(found, f)
		}
	}
	require.Len(t, found, 2)
	assert.Equal(t, SeverityWarning, found[0].Severity)
	assert.Contains(t, found[0].Message, "access(all)")
}

func TestCodeReview_DeprecatedPub(t *testing.T) {
	t.Parallel()
	code := `
pub fun greet(): String {
    return "hello"
}
`
	result := codeReview(code)

	var found []Finding
	for _, f := range result.Findings {
		if f.Rule == "deprecated-pub" {
			found = append(found, f)
		}
	}
	require.Len(t, found, 1)
	assert.Equal(t, SeverityInfo, found[0].Severity)
	assert.Equal(t, 2, found[0].Line)
	assert.Contains(t, found[0].Message, "pub")
}

func TestCodeReview_ForceUnwrap(t *testing.T) {
	t.Parallel()
	code := `
let value = someOptional!
let other = foo()!
`
	result := codeReview(code)

	var found []Finding
	for _, f := range result.Findings {
		if f.Rule == "unsafe-force-unwrap" {
			found = append(found, f)
		}
	}
	require.GreaterOrEqual(t, len(found), 1)
	assert.Equal(t, SeverityNote, found[0].Severity)
	assert.Contains(t, found[0].Message, "Force-unwrap")
}

func TestCodeReview_HardcodedAddress(t *testing.T) {
	t.Parallel()
	code := `
let addr: Address = 0x1234567890abcdef
`
	result := codeReview(code)

	var found []Finding
	for _, f := range result.Findings {
		if f.Rule == "hardcoded-address" {
			found = append(found, f)
		}
	}
	require.Len(t, found, 1)
	assert.Equal(t, SeverityInfo, found[0].Severity)
	assert.Contains(t, found[0].Message, "Hardcoded address")
}

func TestCodeReview_AddressImportNotFlagged(t *testing.T) {
	t.Parallel()
	code := `
import FungibleToken from 0xf233dcee88fe0abe
import NonFungibleToken from 0x1d7e57aa55817448
`
	result := codeReview(code)

	for _, f := range result.Findings {
		assert.NotEqual(t, "hardcoded-address", f.Rule,
			"import-from-address lines should not trigger hardcoded-address rule")
	}
}

func TestCodeReview_FormatResult(t *testing.T) {
	t.Parallel()
	result := ReviewResult{
		Findings: []Finding{
			{Rule: "overly-permissive-access", Severity: SeverityWarning, Line: 3, Message: "State field with access(all) — consider restricting access with entitlements"},
			{Rule: "deprecated-pub", Severity: SeverityInfo, Line: 7, Message: "`pub` is deprecated in Cadence 1.0 — use `access(all)` or a more restrictive access modifier"},
		},
		Summary: map[string]int{
			string(SeverityWarning): 1,
			string(SeverityNote):    0,
			string(SeverityInfo):    1,
		},
	}

	output := formatReviewResult(result)

	assert.Contains(t, output, "[warning]")
	assert.Contains(t, output, "line 3")
	assert.Contains(t, output, "overly-permissive-access")
	assert.Contains(t, output, "[info]")
	assert.Contains(t, output, "line 7")
	assert.Contains(t, output, "deprecated-pub")
	assert.Contains(t, output, "Summary:")
	assert.Contains(t, output, "1 warning(s)")
	assert.Contains(t, output, "1 info(s)")

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.GreaterOrEqual(t, len(lines), 3)
}
