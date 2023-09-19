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
	"os"
	"testing"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/onflow/flow-cli/internal/util"
)

func TestExecutingTests(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		script := tests.TestScriptSimple
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		results, _, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[script.Filename][0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		script := tests.TestScriptSimpleFailing
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		results, _, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, results, 1)

		err = results[script.Filename][0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		results, _, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[script.Filename][0].Error)
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
		}
		state.Contracts().AddOrUpdate(contractHello)
		contractFoo := config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
		}
		state.Contracts().AddOrUpdate(contractFoo)

		// Execute script
		script := tests.TestScriptWithRelativeImports
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		results, _, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[script.Filename][0].Error)
	})

	t.Run("with helper script import", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		// Execute script
		script := tests.TestScriptWithHelperImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		results, _, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[script.Filename][0].Error)
	})

	t.Run("with missing contract location from config", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: "SomeHelloContract.cdc",
		}
		state.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		_, _, err := testCode(testFiles, state, flagsTests{})

		require.Error(t, err)
		assert.Error(
			t,
			err,
			"cannot find contract with location 'contractHello.cdc' in configuration",
		)
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
		results, _, err := testCode(testFiles, state, flagsTests{})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[script.Filename][0].Error)
	})

	t.Run("with code coverage", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
		})

		// Execute script
		script := tests.TestScriptWithCoverage
		testFiles := map[string][]byte{
			script.Filename: script.Source,
		}
		flags := flagsTests{
			Cover: true,
		}
		results, coverageReport, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, results[script.Filename], 3)
		for _, result := range results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		location := common.StringLocation("FooContract")
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
				"A.0000000000000003.FlowToken",
				"A.0000000000000001.FlowStorageFees",
				"A.0000000000000001.FlowDKG",
				"A.0000000000000001.ExampleNFT",
				"A.0000000000000001.FlowIDTableStaking",
				"A.0000000000000001.FlowClusterQC",
				"A.0000000000000001.NodeVersionBeacon",
				"A.0000000000000001.StakingProxy",
				"A.0000000000000004.FlowFees",
				"A.0000000000000002.FungibleToken",
				"A.0000000000000001.FlowStakingCollection",

				// TODO: enable
				//"A.0000000000000001.NFTStorefrontV2",
				//"A.0000000000000001.NFTStorefront",

				"A.0000000000000001.LockedTokens",
				"A.0000000000000001.FlowServiceAccount",
				"A.0000000000000001.FlowEpoch",
			},
			coverageReport.ExcludedLocationIDs(),
		)
		assert.Equal(
			t,
			"Coverage: 42.7% of statements",
			coverageReport.String(),
		)
	})

	t.Run("with code coverage for contracts only", func(t *testing.T) {
		t.Parallel()

		// Setup
		_, state, _ := util.TestMocks(t)

		state.Contracts().AddOrUpdate(config.Contract{
			Name:     tests.ContractFooCoverage.Name,
			Location: tests.ContractFooCoverage.Filename,
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
		results, coverageReport, err := testCode(testFiles, state, flags)

		require.NoError(t, err)
		require.Len(t, results[script.Filename], 3)
		for _, result := range results[script.Filename] {
			assert.NoError(t, result.Error)
		}

		location := common.StringLocation("FooContract")
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

		assert.Equal(t, 3, coverageReport.TotalLocations())
		assert.ElementsMatch(
			t,
			[]string{
				"s.7465737400000000000000000000000000000000000000000000000000000000",
				"I.Crypto",
				"I.Test",
				"A.0000000000000003.FlowToken",
				"A.0000000000000001.FlowStorageFees",
				"A.0000000000000001.FlowDKG",
				"A.0000000000000001.ExampleNFT",
				"A.0000000000000001.FlowIDTableStaking",
				"A.0000000000000001.FlowClusterQC",
				"A.0000000000000001.NodeVersionBeacon",
				"A.0000000000000001.StakingProxy",
				"A.0000000000000004.FlowFees",
				"A.0000000000000002.FungibleToken",
				"A.0000000000000001.FlowStakingCollection",

				// TODO: enable
				//"A.0000000000000001.NFTStorefrontV2",
				//"A.0000000000000001.NFTStorefront",

				"A.0000000000000001.LockedTokens",
				"A.0000000000000001.FlowServiceAccount",
				"A.0000000000000001.FlowEpoch",
			},
			coverageReport.ExcludedLocationIDs(),
		)
		assert.Equal(
			t,
			"Coverage: 22.9% of statements",
			coverageReport.String(),
		)
	})
}
