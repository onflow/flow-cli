/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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

package cadence

import (
	"testing"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_Lint(t *testing.T) {
	t.Parallel()

	// Test this to make sure that lintResult exit codes are actually propagated to CLI result
	t.Run("results.exitCode exported via result.ExitCode()", func(t *testing.T) {
		t.Parallel()

		results := lintResult{
			exitCode: 999,
		}
		require.Equal(t, 999, results.ExitCode())
	})

	t.Run("lints file with no issues", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "NoError.cdc")
		require.NoError(t, err)

		require.Equal(t, &lintResult{
			Results: []fileResult{
				{
					FilePath:    "NoError.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		}, results)
	})

	t.Run("lints file with import", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "foo/WithImports.cdc")
		require.NoError(t, err)

		// Should not have results for imported file, only for the file being linted
		require.Equal(t, &lintResult{
			Results: []fileResult{
				{
					FilePath:    "foo/WithImports.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		}, results)
	})

	t.Run("lints multiple files", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "NoError.cdc", "foo/WithImports.cdc")
		require.NoError(t, err)

		require.Equal(t, &lintResult{
			Results: []fileResult{
				{
					FilePath:    "NoError.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
				{
					FilePath:    "foo/WithImports.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		}, results)
	})

	t.Run("lints file with warning", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "LintWarning.cdc")
		require.NoError(t, err)

		require.Equal(t, &lintResult{
			Results: []fileResult{
				{
					FilePath: "LintWarning.cdc",
					Diagnostics: []analysis.Diagnostic{
						{
							Category: "removal-hint",
							Message:  "unnecessary force operator",
							Location: common.StringLocation("LintWarning.cdc"),
							Range: ast.Range{
								StartPos: ast.Position{Line: 4, Column: 11, Offset: 59},
								EndPos:   ast.Position{Line: 4, Column: 12, Offset: 60},
							},
						},
					},
				},
			},
			exitCode: 0,
		}, results)
	})

	t.Run("lints file with error", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "LintError.cdc")
		require.NoError(t, err)

		require.Equal(t, &lintResult{
			Results: []fileResult{
				{
					FilePath: "LintError.cdc",
					Diagnostics: []analysis.Diagnostic{
						{
							Category: "removal-hint",
							Message:  "unnecessary force operator",
							Location: common.StringLocation("LintError.cdc"),
							Range: ast.Range{
								StartPos: ast.Position{Line: 4, Column: 11, Offset: 57},
								EndPos:   ast.Position{Line: 4, Column: 12, Offset: 58},
							},
						},
						{
							Category:         "semantic-error",
							Message:          "cannot find variable in this scope: `qqq`",
							SecondaryMessage: "not found in this scope",
							Location:         common.StringLocation("LintError.cdc"),
							Range: ast.Range{
								StartPos: ast.Position{Line: 5, Column: 3, Offset: 63},
								EndPos:   ast.Position{Line: 5, Column: 5, Offset: 65},
							},
						},
						{
							Location: common.StringLocation("LintError.cdc"),
							Category: "unused-result-hint",
							Message:  "unused result",
							Range: ast.Range{
								StartPos: ast.Position{Offset: 63, Line: 5, Column: 3},
								EndPos:   ast.Position{Offset: 65, Line: 5, Column: 5},
							},
						},
					},
				},
			},
			exitCode: 1,
		}, results)
	})

	t.Run("linter resolves imports from flowkit state", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "WithFlowkitImport.cdc")
		require.NoError(t, err)

		require.Equal(t, results, &lintResult{
			Results: []fileResult{
				{
					FilePath:    "WithFlowkitImport.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		})
	})

	t.Run("resolves stdlib imports contracts", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsContract.cdc")
		require.NoError(t, err)

		// Expects an error because getAuthAccount is only available in scripts
		require.Equal(t, results, &lintResult{
			Results: []fileResult{
				{
					FilePath: "StdlibImportsContract.cdc",
					Diagnostics: []analysis.Diagnostic{
						{
							Category:         "semantic-error",
							Message:          "cannot find variable in this scope: `getAuthAccount`",
							SecondaryMessage: "not found in this scope",
							Location:         common.StringLocation("StdlibImportsContract.cdc"),
							Range: ast.Range{
								StartPos: ast.Position{Line: 7, Column: 13, Offset: 114},
								EndPos:   ast.Position{Line: 7, Column: 26, Offset: 127},
							},
						},
					},
				},
			},
			exitCode: 1,
		})
	})

	t.Run("resolves stdlib imports transactions", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsTransaction.cdc")
		require.NoError(t, err)

		// Expects an error because getAuthAccount is only available in scripts
		require.Equal(t, results, &lintResult{
			Results: []fileResult{
				{
					FilePath: "StdlibImportsTransaction.cdc",
					Diagnostics: []analysis.Diagnostic{
						{
							Category:         "semantic-error",
							Message:          "cannot find variable in this scope: `getAuthAccount`",
							SecondaryMessage: "not found in this scope",
							Location:         common.StringLocation("StdlibImportsTransaction.cdc"),
							Range: ast.Range{
								StartPos: ast.Position{Line: 7, Column: 13, Offset: 113},
								EndPos:   ast.Position{Line: 7, Column: 26, Offset: 126},
							},
						},
					},
				},
			},
			exitCode: 1,
		})
	})

	t.Run("resolves stdlib imports scripts", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsScript.cdc")
		require.NoError(t, err)

		require.Equal(t, results, &lintResult{
			Results: []fileResult{
				{
					FilePath:    "StdlibImportsScript.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		})
	})
}

func setupMockState(t *testing.T) *flowkit.State {
	// Mock file system
	mockFs := afero.NewMemMapFs()
	_ = afero.WriteFile(mockFs, "NoError.cdc", []byte(`
	access(all) contract NoError {
		access(all) fun test() {}
		init() {}
	}
	`), 0644)
	_ = afero.WriteFile(mockFs, "foo/WithImports.cdc", []byte(`
	import "../NoError.cdc"
	access(all) contract WithImports {
		init() {}
	}
	`), 0644)
	_ = afero.WriteFile(mockFs, "WithFlowkitImport.cdc", []byte(`
	import "NoError"
	access(all) contract WithFlowkitImport {
		init() {
			let foo = NoError.getType()
		}
	}
	`), 0644)
	_ = afero.WriteFile(mockFs, "LintWarning.cdc", []byte(`
	access(all) contract LintWarning {
		init() {
			let x = 1!
		}
	}`), 0644)
	_ = afero.WriteFile(mockFs, "LintError.cdc", []byte(`
	access(all) contract LintError {
		init() {
			let x = 1!
			qqq
		}
	}`), 0644)
	_ = afero.WriteFile(mockFs, "CadenceV1Error.cdc", []byte(`
	access(all) contract CadenceV1Error {
		init() {
			let test: PublicAccount = self.account
		}
	}`), 0644)
	_ = afero.WriteFile(mockFs, "StdlibImportsContract.cdc", []byte(`
	import Crypto
	import Test
	import BlockchainHelpers
	access(all) contract WithImports{
		init() {
			let foo = getAuthAccount<&Account>(0x01)
			log(RLP.getType())
		}
	}
	`), 0644)
	_ = afero.WriteFile(mockFs, "StdlibImportsTransaction.cdc", []byte(`
	import Crypto
	import Test
	import BlockchainHelpers
	transaction {
		prepare(signer: &Account) {
			let foo = getAuthAccount<&Account>(0x01)
			log(RLP.getType())
		}
	}
	`), 0644)
	_ = afero.WriteFile(mockFs, "StdlibImportsScript.cdc", []byte(`
	import Crypto
	import Test
	import BlockchainHelpers
	access(all) fun main(): Void {
		let foo = getAuthAccount<&Account>(0x01)
		log(RLP.getType())
	}`), 0644)

	rw := afero.Afero{Fs: mockFs}
	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	// Mock flowkit contracts
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "NoError",
		Location: "NoError.cdc",
	})

	return state
}
