/*
 * Flow CLI
 *
 * Copyright 2022 Dapper Labs, Inc.
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

package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/stdlib"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/tests"

	"github.com/onflow/flow-cli/internal/util"
)

func TestExecutingTests(t *testing.T) {
	t.Parallel()

	aliases := config.Aliases{{
		Network: "testing",
		Address: flowsdk.HexToAddress("0x0000000000000007"),
	}}

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		_, state, _ := util.TestMocks(t)

		script := tests.TestScriptSimple
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		result, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, result.Results, 1)
		assert.NoError(t, result.Results[script.Filename][0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()

		_, state, _ := util.TestMocks(t)

		script := tests.TestScriptSimpleFailing
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		result, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, result.Results, 1)

		err = result.Results[script.Filename][0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()

		_, state, _ := util.TestMocks(t)

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
			Aliases:  aliases,
		}
		state.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		result, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, result.Results, 1)
		assert.NoError(t, result.Results[script.Filename][0].Error)
	})

	t.Run("with relative imports", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)
		readerWriter := state.ReaderWriter()
		_ = readerWriter.WriteFile(
			"../contracts/contractHello.cdc",
			tests.ContractHelloString.Source,
			os.ModeTemporary,
		)
		_ = readerWriter.WriteFile(
			"../contracts/FooContract.cdc",
			tests.ContractFooCoverage.Source,
			os.ModeTemporary,
		)

		contractHello := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
			Aliases:  aliases,
		}
		state.Contracts().AddOrUpdate(contractHello)
		contractFoo := config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
			Aliases:  aliases,
		}
		state.Contracts().AddOrUpdate(contractFoo)

		// Execute script
		script := tests.TestScriptWithRelativeImports
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		result, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, result.Results, 1)
		assert.NoError(t, result.Results[script.Filename][0].Error)
	})

	t.Run("with helper script import", func(t *testing.T) {
		t.Parallel()

		_, state, _ := util.TestMocks(t)

		// Execute script
		script := tests.TestScriptWithHelperImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		result, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, result.Results, 1)
		assert.NoError(t, result.Results[script.Filename][0].Error)
	})

	t.Run("with missing contract in config", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		// Execute script
		script := tests.TestScriptWithMissingContract
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		_, err := testCode(testFiles, state, flagsTests{})

		require.Error(t, err)
		assert.ErrorContains(
			t,
			err,
			"cannot find contract with location 'ApprovalVoting' in configuration",
		)
	})

	t.Run("with missing testing alias in config", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
			Aliases: config.Aliases{{
				Network: "emulator",
				Address: flowsdk.HexToAddress("0x0000000000000007"),
			}},
		}
		state.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		_, err := testCode(testFiles, state, flagsTests{})

		require.Error(t, err)
		assert.ErrorContains(
			t,
			err,
			"could not find the address of contract: Hello",
		)
	})

	t.Run("without testing alias for common contracts", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
			Aliases:  aliases,
		}
		state.Contracts().AddOrUpdate(c)
		// fungibleToken has no `testing` alias, but it is not
		// actually deployed/used, so there is no errror.
		fungibleToken := config.Contract{
			Name:     "FungibleToken",
			Location: "cadence/contracts/FungibleToken.cdc",
		}
		state.Contracts().AddOrUpdate(fungibleToken)

		// Execute script
		script := tests.TestScriptWithImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		_, err := testCode(testFiles, state, flagsTests{})

		assert.NoError(t, err)
	})

	t.Run("with file read", func(t *testing.T) {
		t.Parallel()

		_, state, rw := util.TestMocks(t)

		_ = rw.WriteFile(
			tests.SomeFile.Filename,
			tests.SomeFile.Source,
			os.ModeTemporary,
		)

		// Execute script
		script := tests.TestScriptWithFileRead
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		result, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, result.Results, 1)
		assert.NoError(t, result.Results[script.Filename][0].Error)
	})

	t.Run("with code coverage", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
			Aliases:  aliases,
		})

		// Execute script
		script := tests.TestScriptWithCoverage
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Cover: true,
		}
		result, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, result.Results[script.Filename], 2)
		for _, result := range result.Results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		coverageReport := result.CoverageReport
		location := common.AddressLocation{
			Name:    "FooContract",
			Address: common.Address{0, 0, 0, 0, 0, 0, 0, 7},
		}
		coverage := coverageReport.Coverage[location]

		assert.Equal(t, []int{}, coverage.MissedLines())
		assert.Equal(t, 15, coverage.Statements)
		assert.Equal(t, "100.0%", coverage.Percentage())
		assert.EqualValues(
			t,
			map[int]int{
				6: 1, 14: 1, 18: 10, 19: 1, 20: 9, 21: 1, 22: 8, 23: 1,
				24: 7, 25: 1, 26: 6, 27: 1, 30: 5, 31: 4, 34: 1,
			},
			coverage.LineHits,
		)

		assert.True(t, coverageReport.TotalLocations() > 1)
		assert.ElementsMatch(
			t,
			[]string{
				"s.7465737400000000000000000000000000000000000000000000000000000000",
				"I.Crypto",
				"I.Test",
				"A.0000000000000001.NodeVersionBeacon",
				"A.0000000000000001.FlowServiceAccount",
				"A.0000000000000002.FungibleToken",
				"A.0000000000000001.FlowClusterQC",
				"A.0000000000000001.FlowDKG",
				"A.0000000000000002.FungibleTokenMetadataViews",
				"A.0000000000000001.FlowIDTableStaking",
				"A.0000000000000001.LockedTokens",
				"A.0000000000000001.ExampleNFT",
				"A.0000000000000001.FlowStakingCollection",
				"A.0000000000000001.StakingProxy",
				"A.0000000000000003.FlowToken",
				"A.0000000000000001.FlowEpoch",
				"A.0000000000000001.FlowStorageFees",
				"A.0000000000000004.FlowFees",
				"A.0000000000000001.MetadataViews",
				"A.0000000000000001.NonFungibleToken",
				"A.0000000000000001.ViewResolver",
				"A.0000000000000001.RandomBeaconHistory",
				"A.0000000000000001.EVM",
				"A.0000000000000002.FungibleTokenSwitchboard",
				"I.BlockchainHelpers",
				"A.0000000000000001.Burner",
			},
			coverageReport.ExcludedLocationIDs(),
		)

		expected := "Coverage: 93.8% of statements"

		assert.Equal(
			t,
			expected,
			coverageReport.String(),
		)
		assert.Contains(
			t,
			result.String(),
			expected,
		)

		lcovReport, _ := coverageReport.MarshalLCOV()
		assert.Contains(t, string(lcovReport), "TN:\nSF:FooContract.cdc\n")

		jsonReport, _ := coverageReport.MarshalJSON()
		assert.Contains(t, string(jsonReport), `{"coverage":{"FooContract.cdc":{`)
	})

	t.Run("with code coverage for contracts only", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
			Aliases:  aliases,
		})

		// Execute script
		script := tests.TestScriptWithCoverage
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Cover:     true,
			CoverCode: contractsCoverCode,
		}
		result, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, result.Results[script.Filename], 2)
		for _, result := range result.Results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		coverageReport := result.CoverageReport
		location := common.AddressLocation{
			Name:    "FooContract",
			Address: common.Address{0, 0, 0, 0, 0, 0, 0, 7},
		}
		coverage := coverageReport.Coverage[location]

		assert.Equal(t, []int{}, coverage.MissedLines())
		assert.Equal(t, 15, coverage.Statements)
		assert.Equal(t, "100.0%", coverage.Percentage())
		assert.EqualValues(
			t,
			map[int]int{
				6: 1, 14: 1, 18: 10, 19: 1, 20: 9, 21: 1, 22: 8, 23: 1,
				24: 7, 25: 1, 26: 6, 27: 1, 30: 5, 31: 4, 34: 1,
			},
			coverage.LineHits,
		)

		assert.Equal(t, 1, coverageReport.TotalLocations())
		assert.ElementsMatch(
			t,
			[]string{
				"s.7465737400000000000000000000000000000000000000000000000000000000",
				"I.Crypto",
				"I.Test",
				"A.0000000000000001.NodeVersionBeacon",
				"A.0000000000000001.FlowServiceAccount",
				"A.0000000000000002.FungibleToken",
				"A.0000000000000001.FlowClusterQC",
				"A.0000000000000001.FlowDKG",
				"A.0000000000000002.FungibleTokenMetadataViews",
				"A.0000000000000001.FlowIDTableStaking",
				"A.0000000000000001.LockedTokens",
				"A.0000000000000001.ExampleNFT",
				"A.0000000000000001.FlowStakingCollection",
				"A.0000000000000001.StakingProxy",
				"A.0000000000000003.FlowToken",
				"A.0000000000000001.FlowEpoch",
				"A.0000000000000001.FlowStorageFees",
				"A.0000000000000004.FlowFees",
				"A.0000000000000001.MetadataViews",
				"A.0000000000000001.NonFungibleToken",
				"A.0000000000000001.ViewResolver",
				"A.0000000000000001.RandomBeaconHistory",
				"A.0000000000000001.EVM",
				"A.0000000000000002.FungibleTokenSwitchboard",
				"I.BlockchainHelpers",
				"A.0000000000000001.Burner",
			},
			coverageReport.ExcludedLocationIDs(),
		)
		assert.Equal(
			t,
			"Coverage: 100.0% of statements",
			coverageReport.String(),
		)
		assert.Contains(
			t,
			result.String(),
			"Coverage: 100.0% of statements",
		)

		lcovReport, _ := coverageReport.MarshalLCOV()
		assert.Contains(t, string(lcovReport), "TN:\nSF:FooContract.cdc\n")

		jsonReport, _ := coverageReport.MarshalJSON()
		assert.Contains(t, string(jsonReport), `{"coverage":{"FooContract.cdc":{`)
	})

	t.Run("with random test case execution", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
			Aliases:  aliases,
		})

		// Execute script
		script := tests.TestScriptWithCoverage
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Random: true,
		}
		result, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, result.Results[script.Filename], 2)
		for _, result := range result.Results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		assert.Contains(
			t,
			result.String(),
			fmt.Sprintf("Seed: %d", result.RandomSeed),
		)
	})

	t.Run("with input seed for test case execution", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
			Aliases:  aliases,
		})

		// Execute script
		script := tests.TestScriptWithCoverage
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Seed: 1521,
		}
		result, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, result.Results[script.Filename], 2)
		for _, result := range result.Results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		assert.Contains(
			t,
			result.String(),
			fmt.Sprintf("Seed: %d", flags.Seed),
		)

		// Note that `testGetIntegerTrait` is the first test case,
		// but it gets executed after `testAddSpecialNumber`, due
		// to random test execution.
		expected := `Test results: "testScriptWithCoverage.cdc"
- PASS: testAddSpecialNumber
- PASS: testGetIntegerTrait
Seed: 1521
`
		assert.Equal(
			t,
			expected,
			result.Oneliner(),
		)
	})

	t.Run("with JSON output", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
			Aliases:  aliases,
		})

		// Execute script
		script := tests.TestScriptWithCoverage
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Seed:      1521,
			Cover:     true,
			CoverCode: contractsCoverCode,
		}
		result, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, result.Results[script.Filename], 2)
		for _, result := range result.Results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		output, err := json.Marshal(result.JSON())
		require.NoError(t, err)

		expected := `
        {
            "meta": {
                "coverage": "100.0%",
                "seed": "1521"
            },
            "testScriptWithCoverage.cdc": {
                "testGetIntegerTrait": "PASS",
                "testAddSpecialNumber": "PASS"
            }
        }
		`

		assert.JSONEq(
			t,
			expected,
			string(output),
		)
	})

	t.Run("run specific test case by name", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		// Execute script
		script := tests.TestScriptSimple
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Name: "testSimple",
		}

		result, err := testCode(testFiles, state, flags)

		assert.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.NoError(t, result.Results[script.Filename][0].Error)

		expected := "Test results: \"./testScriptSimple.cdc\"\n- PASS: testSimple\n"
		assert.Equal(
			t,
			expected,
			result.Oneliner(),
		)
	})

	t.Run("run specific test case by name multiple files", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		scriptPassing := tests.TestScriptSimple
		scriptFailing := tests.TestScriptSimpleFailing

		// Execute script
		testFiles := map[string][]byte{
			scriptPassing.Filename: scriptPassing.Source,
			scriptFailing.Filename: scriptFailing.Source,
		}
		flags := flagsTests{
			Name: "testSimple",
		}

		result, err := testCode(testFiles, state, flags)

		assert.NoError(t, err)
		assert.Len(t, result.Results, 2)
		assert.NoError(t, result.Results[scriptPassing.Filename][0].Error)
		assert.Error(t, result.Results[scriptFailing.Filename][0].Error)
		assert.ErrorAs(t, result.Results[scriptFailing.Filename][0].Error, &stdlib.AssertionError{})

		assert.Contains(
			t,
			result.Oneliner(),
			"Test results: \"./testScriptSimple.cdc\"\n- PASS: testSimple",
		)
		assert.Contains(
			t,
			result.Oneliner(),
			"Test results: \"./testScriptSimpleFailing.cdc\"\n- FAIL: "+
				"testSimple\n\t\tExecution failed:\n\t\t\terror: assertion failed\n"+
				"\t\t\t --> ./testScriptSimpleFailing.cdc:5:12",
		)
	})

	t.Run("run specific test case by name will do nothing if not found", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		// Execute script
		script := tests.TestScriptSimple
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Name: "doesNotExist",
		}

		result, err := testCode(testFiles, state, flags)

		assert.NoError(t, err)
		assert.Len(t, result.Results, 0)
		assert.Equal(
			t,
			"No tests found",
			result.Oneliner(),
		)
	})
}
