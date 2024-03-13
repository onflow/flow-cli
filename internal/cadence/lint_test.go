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
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_Lint(t *testing.T) {
	state := setupMockState(t)

	t.Run("results.exitCode matches result.ExitCode()", func(t *testing.T) {
		results := lintResults{
			exitCode: 999,
		}
		require.Equal(t, 999, results.ExitCode())
	})

	t.Run("lints file with no issues", func(t *testing.T) {
		results, error := lintFiles(state, "NoError.cdc")
		require.NoError(t, error)

		require.Equal(t, &lintResults{
			Results: []lintResult{
				{
					FilePath:    "NoError.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		}, results)
	})

	t.Run("lints file with import", func(t *testing.T) {
		results, error := lintFiles(state, "foo/WithImports.cdc")
		require.NoError(t, error)

		// Should not have results for imported file, only for the file being linted
		require.Equal(t, &lintResults{
			Results: []lintResult{
				{
					FilePath:    "foo/WithImports.cdc",
					Diagnostics: []analysis.Diagnostic{},
				},
			},
			exitCode: 0,
		}, results)
	})

	t.Run("lints multiple files", func(t *testing.T) {
		results, error := lintFiles(state, "NoError.cdc", "foo/WithImports.cdc")
		require.NoError(t, error)

		require.Equal(t, &lintResults{
			Results: []lintResult{
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
		results, error := lintFiles(state, "LintWarning.cdc")
		require.NoError(t, error)

		require.Equal(t, &lintResults{
			Results: []lintResult{
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
		results, error := lintFiles(state, "LintError.cdc")
		require.NoError(t, error)

		require.Equal(t, &lintResults{
			Results: []lintResult{
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
					},
				},
			},
			exitCode: 1,
		}, results)
	})

	t.Run("linter resolves imports from flowkit state", func(t *testing.T) {
		results, error := lintFiles(state, "WithFlowkitImport.cdc")
		require.NoError(t, error)

		require.Equal(t, results, &lintResults{
			Results: []lintResult{
				{
					FilePath:    "WithFlowkitImport.cdc",
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
			log(NoError.getType())
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

	rw := afero.Afero{Fs: mockFs}
	state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	// Mock flowkit contracts
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "NoError",
		Location: "NoError.cdc",
	})

	return state
}
