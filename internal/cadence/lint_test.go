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

package cadence

import (
	"testing"

	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/tools/analysis"
	flowsdk "github.com/onflow/flow-go-sdk"
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

		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "NoError.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("lints file with import", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "foo/WithImports.cdc")
		require.NoError(t, err)

		// Should not have results for imported file, only for the file being linted
		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "foo/WithImports.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("lints multiple files", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "NoError.cdc", "foo/WithImports.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
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
			},
			results,
		)
	})

	t.Run("lints file with warning", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "LintWarning.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
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
			},
			results,
		)
	})

	t.Run("lints file with error", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "LintError.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
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
								SecondaryMessage: "not found in this scope; check for typos or declare it",
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
			},
			results,
		)
	})

	t.Run("linter resolves imports from flowkit state", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "WithFlowkitImport.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "WithFlowkitImport.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("resolves stdlib imports contracts", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsContract.cdc")
		require.NoError(t, err)

		// Expects an error because getAuthAccount is only available in scripts
		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath: "StdlibImportsContract.cdc",
						Diagnostics: []analysis.Diagnostic{
							{
								Category:         "semantic-error",
								Message:          "cannot find variable in this scope: `getAuthAccount`",
								SecondaryMessage: "not found in this scope; check for typos or declare it",
								Location:         common.StringLocation("StdlibImportsContract.cdc"),
								Range: ast.Range{
									StartPos: ast.Position{Line: 6, Column: 13, Offset: 99},
									EndPos:   ast.Position{Line: 6, Column: 26, Offset: 112},
								},
							},
						},
					},
				},
				exitCode: 1,
			},
			results,
		)
	})

	t.Run("resolves stdlib imports transactions", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsTransaction.cdc")
		require.NoError(t, err)

		// Expects an error because getAuthAccount is only available in scripts
		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath: "StdlibImportsTransaction.cdc",
						Diagnostics: []analysis.Diagnostic{
							{
								Category:         "semantic-error",
								Message:          "cannot find variable in this scope: `getAuthAccount`",
								SecondaryMessage: "not found in this scope; check for typos or declare it",
								Location:         common.StringLocation("StdlibImportsTransaction.cdc"),
								Range: ast.Range{
									StartPos: ast.Position{Line: 6, Column: 13, Offset: 98},
									EndPos:   ast.Position{Line: 6, Column: 26, Offset: 111},
								},
							},
						},
					},
				},
				exitCode: 1,
			},
			results,
		)
	})

	t.Run("resolves stdlib imports scripts", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsScript.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "StdlibImportsScript.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("resolves stdlib imports Crypto", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "StdlibImportsCrypto.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "StdlibImportsCrypto.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("resolves nested imports when contract imported by name", func(t *testing.T) {
		t.Parallel()

		state := setupMockState(t)

		results, err := lintFiles(state, "TransactionImportingContractWithNestedImports.cdc")
		require.NoError(t, err)

		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "TransactionImportingContractWithNestedImports.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("allows access(account) when contracts on same account", func(t *testing.T) {
		t.Parallel()

		state := setupMockStateWithAccountAccess(t)

		results, err := lintFiles(state, "ContractA.cdc")
		require.NoError(t, err)

		// Should have no errors since ContractA and ContractB are on same account
		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "ContractA.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("denies access(account) when contracts on different accounts", func(t *testing.T) {
		t.Parallel()

		state := setupMockStateWithAccountAccess(t)

		results, err := lintFiles(state, "ContractC.cdc")
		require.NoError(t, err)

		// Should have error since ContractC and ContractB are on different accounts
		require.Len(t, results.Results, 1)
		require.Len(t, results.Results[0].Diagnostics, 1)
		require.Equal(t, "semantic-error", results.Results[0].Diagnostics[0].Category)
		require.Contains(t, results.Results[0].Diagnostics[0].Message, "access denied")
		require.Equal(t, 1, results.exitCode)
	})

	t.Run("allows access(account) when dependencies on same account (peak-money repro)", func(t *testing.T) {
		t.Parallel()

		state := setupMockStateWithDependencies(t)

		results, err := lintFiles(state, "imports/testaddr/DepA.cdc")
		require.NoError(t, err)

		// Should have no errors since DepA and DepB are dependencies on same address
		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "imports/testaddr/DepA.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
	})

	t.Run("allows access(account) when dependencies have Source but no Aliases", func(t *testing.T) {
		t.Parallel()

		state := setupMockStateWithSourceOnly(t)

		// Verify that AddDependencyAsContract automatically adds Source to Aliases
		sourceAContract, _ := state.Contracts().ByName("SourceA")
		require.NotNil(t, sourceAContract, "SourceA contract should exist")

		// Check if the alias was automatically added from Source
		alias := sourceAContract.Aliases.ByNetwork("testnet")
		require.NotNil(t, alias, "Alias should be automatically created from Source")
		require.Equal(t, "dfc20aee650fcbdf", alias.Address.String(), "Alias address should match Source address")

		results, err := lintFiles(state, "imports/testaddr/SourceA.cdc")
		require.NoError(t, err)

		// Should have no errors since SourceA and SourceB have same Source.Address (converted to Aliases)
		require.Equal(t,
			&lintResult{
				Results: []fileResult{
					{
						FilePath:    "imports/testaddr/SourceA.cdc",
						Diagnostics: []analysis.Diagnostic{},
					},
				},
				exitCode: 0,
			},
			results,
		)
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
	import Test
	import BlockchainHelpers
	access(all) fun main(): Void {
		let foo = getAuthAccount<&Account>(0x01)
		log(RLP.getType())
	}`), 0644)
	_ = afero.WriteFile(mockFs, "StdlibImportsCrypto.cdc", []byte(`
	import Crypto

	access(all) contract CryptoImportTest {
		access(all) fun test(): Void {
			let _ = Crypto.hash([1, 2, 3], algorithm: HashAlgorithm.SHA3_256)
		}
	}
	`), 0644)

	// Regression test files for nested import bug
	_ = afero.WriteFile(mockFs, "Helper.cdc", []byte(`
	access(all) contract Helper {
		access(all) let name: String
		
		init() {
			self.name = "Helper"
		}
		
		access(all) fun greet(): String {
			return "Hello from ".concat(self.name)
		}
	}
	`), 0644)

	_ = afero.WriteFile(mockFs, "ContractWithNestedImports.cdc", []byte(`
	import Helper from "./Helper.cdc"
	
	access(all) contract ContractWithNestedImports {
		access(all) fun test(): String {
			return Helper.greet()
		}
		init() {}
	}
	`), 0644)

	_ = afero.WriteFile(mockFs, "TransactionImportingContractWithNestedImports.cdc", []byte(`
	import ContractWithNestedImports from "ContractWithNestedImports"
	
	transaction() {
		prepare(signer: auth(Storage) &Account) {
			log(ContractWithNestedImports.test())
		}
	}
	`), 0644)

	rw := afero.Afero{Fs: mockFs}
	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	// Mock flowkit contracts
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "NoError",
		Location: "NoError.cdc",
	})
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "Helper",
		Location: "Helper.cdc",
	})
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "ContractWithNestedImports",
		Location: "ContractWithNestedImports.cdc",
	})

	return state
}

func setupMockStateWithAccountAccess(t *testing.T) *flowkit.State {
	// Mock file system
	mockFs := afero.NewMemMapFs()

	// ContractB has an access(account) function
	_ = afero.WriteFile(mockFs, "ContractB.cdc", []byte(`
	access(all) contract ContractB {
		access(account) fun accountOnlyFunction() {
			log("This requires account access")
		}
		init() {}
	}
	`), 0644)

	// ContractA imports and calls ContractB's account function - should work (same account)
	_ = afero.WriteFile(mockFs, "ContractA.cdc", []byte(`
	import ContractB from "ContractB"
	
	access(all) contract ContractA {
		access(all) fun callB() {
			ContractB.accountOnlyFunction()
		}
		init() {}
	}
	`), 0644)

	// ContractC imports and calls ContractB's account function - should fail (different account)
	_ = afero.WriteFile(mockFs, "ContractC.cdc", []byte(`
	import ContractB from "ContractB"
	
	access(all) contract ContractC {
		access(all) fun callB() {
			ContractB.accountOnlyFunction()
		}
		init() {}
	}
	`), 0644)

	rw := afero.Afero{Fs: mockFs}
	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	// Configure contracts with deployments
	// ContractA and ContractB are on the same account (0x01)
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "ContractA",
		Location: "ContractA.cdc",
		Aliases: config.Aliases{
			{
				Network: "testnet",
				Address: flowsdk.HexToAddress("0000000000000001"),
			},
		},
	})

	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "ContractB",
		Location: "ContractB.cdc",
		Aliases: config.Aliases{
			{
				Network: "testnet",
				Address: flowsdk.HexToAddress("0000000000000001"),
			},
		},
	})

	// ContractC is on a different account (0x02)
	state.Contracts().AddOrUpdate(config.Contract{
		Name:     "ContractC",
		Location: "ContractC.cdc",
		Aliases: config.Aliases{
			{
				Network: "testnet",
				Address: flowsdk.HexToAddress("0000000000000002"),
			},
		},
	})

	// Add network
	state.Networks().AddOrUpdate(config.Network{
		Name: "testnet",
		Host: "access.testnet.nodes.onflow.org:9000",
	})

	return state
}

func setupMockStateWithDependencies(t *testing.T) *flowkit.State {
	// Reproduce peak-money structure: dependencies with aliases, not contracts
	mockFs := afero.NewMemMapFs()

	// DepB has an access(account) function (like FlowEVMBridgeCustomAssociations)
	_ = afero.WriteFile(mockFs, "imports/testaddr/DepB.cdc", []byte(`
	access(all) contract DepB {
		access(account) fun pauseConfig(forType: Type) {
			log("This requires account access")
		}
		init() {}
	}
	`), 0644)

	// DepA imports and calls DepB's account function (like FlowEVMBridgeConfig)
	_ = afero.WriteFile(mockFs, "imports/testaddr/DepA.cdc", []byte(`
	import DepB from "DepB"
	
	access(all) contract DepA {
		access(all) fun callDepB(forType: Type) {
			DepB.pauseConfig(forType: forType)
		}
		init() {}
	}
	`), 0644)

	rw := afero.Afero{Fs: mockFs}
	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	// Add network first
	state.Networks().AddOrUpdate(config.Network{
		Name: "testnet",
		Host: "access.testnet.nodes.onflow.org:9000",
	})

	// Add as DEPENDENCIES (not contracts) with same aliases - this is the key difference from peak-money
	state.Dependencies().AddOrUpdate(config.Dependency{
		Name: "DepA",
		Source: config.Source{
			NetworkName:  "testnet",
			Address:      flowsdk.HexToAddress("dfc20aee650fcbdf"),
			ContractName: "DepA",
		},
		Aliases: config.Aliases{
			{
				Network: "testnet",
				Address: flowsdk.HexToAddress("dfc20aee650fcbdf"),
			},
			{
				Network: "emulator",
				Address: flowsdk.HexToAddress("f8d6e0586b0a20c7"),
			},
		},
	})

	state.Dependencies().AddOrUpdate(config.Dependency{
		Name: "DepB",
		Source: config.Source{
			NetworkName:  "testnet",
			Address:      flowsdk.HexToAddress("dfc20aee650fcbdf"),
			ContractName: "DepB",
		},
		Aliases: config.Aliases{
			{
				Network: "testnet",
				Address: flowsdk.HexToAddress("dfc20aee650fcbdf"),
			},
			{
				Network: "emulator",
				Address: flowsdk.HexToAddress("f8d6e0586b0a20c7"),
			},
		},
	})

	// Dependencies should also be added as contracts for import resolution
	// This is what happens when you run `flow dependencies install`
	state.Contracts().AddDependencyAsContract(
		*state.Dependencies().ByName("DepA"),
		"testnet",
	)
	state.Contracts().AddDependencyAsContract(
		*state.Dependencies().ByName("DepB"),
		"testnet",
	)

	// Set the Location field so imports can resolve the files
	depAContract, _ := state.Contracts().ByName("DepA")
	if depAContract != nil {
		depAContract.Location = "imports/testaddr/DepA.cdc"
	}
	depBContract, _ := state.Contracts().ByName("DepB")
	if depBContract != nil {
		depBContract.Location = "imports/testaddr/DepB.cdc"
	}

	return state
}

func setupMockStateWithSourceOnly(t *testing.T) *flowkit.State {
	// Test dependencies with ONLY Source (no Aliases) to see if we need to check Source
	mockFs := afero.NewMemMapFs()

	// SourceB has an access(account) function
	_ = afero.WriteFile(mockFs, "imports/testaddr/SourceB.cdc", []byte(`
	access(all) contract SourceB {
		access(account) fun sourceOnlyFunction() {
			log("This requires account access")
		}
		init() {}
	}
	`), 0644)

	// SourceA imports and calls SourceB's account function
	_ = afero.WriteFile(mockFs, "imports/testaddr/SourceA.cdc", []byte(`
	import SourceB from "SourceB"
	
	access(all) contract SourceA {
		access(all) fun callSourceB() {
			SourceB.sourceOnlyFunction()
		}
		init() {}
	}
	`), 0644)

	rw := afero.Afero{Fs: mockFs}
	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	// Add network
	state.Networks().AddOrUpdate(config.Network{
		Name: "testnet",
		Host: "access.testnet.nodes.onflow.org:9000",
	})

	// Add dependencies with ONLY Source, NO Aliases
	state.Dependencies().AddOrUpdate(config.Dependency{
		Name: "SourceA",
		Source: config.Source{
			NetworkName:  "testnet",
			Address:      flowsdk.HexToAddress("dfc20aee650fcbdf"),
			ContractName: "SourceA",
		},
		// NO Aliases!
	})

	state.Dependencies().AddOrUpdate(config.Dependency{
		Name: "SourceB",
		Source: config.Source{
			NetworkName:  "testnet",
			Address:      flowsdk.HexToAddress("dfc20aee650fcbdf"),
			ContractName: "SourceB",
		},
		// NO Aliases!
	})

	// Add as contracts for import resolution
	state.Contracts().AddDependencyAsContract(
		*state.Dependencies().ByName("SourceA"),
		"testnet",
	)
	state.Contracts().AddDependencyAsContract(
		*state.Dependencies().ByName("SourceB"),
		"testnet",
	)

	// Set the Location field so imports can resolve
	sourceAContract, _ := state.Contracts().ByName("SourceA")
	if sourceAContract != nil {
		sourceAContract.Location = "imports/testaddr/SourceA.cdc"
	}
	sourceBContract, _ := state.Contracts().ByName("SourceB")
	if sourceBContract != nil {
		sourceBContract.Location = "imports/testaddr/SourceB.cdc"
	}

	return state
}
