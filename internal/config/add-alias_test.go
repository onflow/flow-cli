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

package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_AddAlias(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)

		// Setup: Add a contract to the state first
		contract := config.Contract{
			Name:     "MyContract",
			Location: "contracts/MyContract.cdc",
		}
		state.Contracts().AddOrUpdate(contract)

		// Set flags
		addAliasFlags.Contract = "MyContract"
		addAliasFlags.Network = "testnet"
		addAliasFlags.Address = "0x1234567890abcdef"

		// Call the function
		result, err := addAlias(
			[]string{},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		// Verify no errors
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "Alias for contract MyContract on network testnet added")

		// Verify the state was modified correctly
		updatedContract, err := state.Contracts().ByName("MyContract")
		require.NoError(t, err)

		// Verify the alias was added for the specified network
		alias := updatedContract.Aliases.ByNetwork("testnet")
		require.NotNil(t, alias)
		assert.Equal(t, "1234567890abcdef", alias.Address.String())

		// Reset flags
		addAliasFlags = flagsAddAlias{}
	})

	t.Run("Success with multiple aliases", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)

		// Get the emulator service account address
		serviceAcc, err := state.EmulatorServiceAccount()
		require.NoError(t, err)

		// Setup: Add a contract with an existing alias
		contract := config.Contract{
			Name:     "MultiContract",
			Location: "contracts/MultiContract.cdc",
			Aliases: config.Aliases{{
				Network: "emulator",
				Address: serviceAcc.Address,
			}},
		}
		state.Contracts().AddOrUpdate(contract)

		// Add testnet alias
		addAliasFlags.Contract = "MultiContract"
		addAliasFlags.Network = "testnet"
		addAliasFlags.Address = "0xabcdef1234567890"

		result, err := addAlias(
			[]string{},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify both aliases exist
		updatedContract, err := state.Contracts().ByName("MultiContract")
		require.NoError(t, err)

		emulatorAlias := updatedContract.Aliases.ByNetwork("emulator")
		require.NotNil(t, emulatorAlias)
		assert.Equal(t, serviceAcc.Address.String(), emulatorAlias.Address.String())

		testnetAlias := updatedContract.Aliases.ByNetwork("testnet")
		require.NotNil(t, testnetAlias)
		assert.Equal(t, "abcdef1234567890", testnetAlias.Address.String())

		// Reset flags
		addAliasFlags = flagsAddAlias{}
	})

	t.Run("Fail contract not found", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)

		addAliasFlags.Contract = "NonExistentContract"
		addAliasFlags.Network = "testnet"
		addAliasFlags.Address = "0x1234567890abcdef"

		result, err := addAlias(
			[]string{},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "contract NonExistentContract not found in configuration")

		// Reset flags
		addAliasFlags = flagsAddAlias{}
	})

	t.Run("Verify flow.json is modified correctly", func(t *testing.T) {
		srv, state, rw := util.TestMocks(t)

		// Setup: Add a contract to the state
		contract := config.Contract{
			Name:     "TestContract",
			Location: "contracts/TestContract.cdc",
		}
		state.Contracts().AddOrUpdate(contract)

		// Set flags
		addAliasFlags.Contract = "TestContract"
		addAliasFlags.Network = "mainnet"
		addAliasFlags.Address = "0xabcdef1234567890"

		// Call the function
		result, err := addAlias(
			[]string{},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		assert.NotNil(t, result)

		// Read the flow.json file
		flowJSON, err := rw.ReadFile("flow.json")
		require.NoError(t, err)

		// Unmarshal and verify the JSON structure
		var flowConfig map[string]interface{}
		err = json.Unmarshal(flowJSON, &flowConfig)
		require.NoError(t, err)

		// Verify contracts section exists
		contracts, ok := flowConfig["contracts"].(map[string]interface{})
		require.True(t, ok, "contracts section should exist in flow.json")

		// Verify TestContract exists
		testContract, ok := contracts["TestContract"].(map[string]interface{})
		require.True(t, ok, "TestContract should exist in flow.json")

		// Verify aliases section exists in the contract
		aliases, ok := testContract["aliases"].(map[string]interface{})
		require.True(t, ok, "aliases section should exist in TestContract")

		// Verify mainnet alias exists with correct address (stored without 0x prefix)
		mainnetAlias, ok := aliases["mainnet"].(string)
		require.True(t, ok, "mainnet alias should exist")
		assert.Equal(t, "abcdef1234567890", mainnetAlias)

		// Reset flags
		addAliasFlags = flagsAddAlias{}
	})
}

func Test_FlagsToAliasData(t *testing.T) {
	t.Run("Success with all flags", func(t *testing.T) {
		flags := flagsAddAlias{
			Contract: "TestContract",
			Network:  "testnet",
			Address:  "0x1234567890abcdef",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		require.NoError(t, err)
		assert.True(t, flagsProvided)
		assert.Equal(t, "TestContract", data.Contract)
		assert.Equal(t, "testnet", data.Network)
		assert.Equal(t, "0x1234567890abcdef", data.Address)
	})

	t.Run("No flags provided", func(t *testing.T) {
		flags := flagsAddAlias{}

		data, flagsProvided, err := flagsToAliasData(flags)

		require.NoError(t, err)
		assert.False(t, flagsProvided)
		assert.Nil(t, data)
	})

	t.Run("Fail missing contract name", func(t *testing.T) {
		flags := flagsAddAlias{
			Network: "testnet",
			Address: "0x1234567890abcdef",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		assert.Nil(t, data)
		assert.True(t, flagsProvided)
		assert.EqualError(t, err, "contract name must be provided")
	})

	t.Run("Fail missing network", func(t *testing.T) {
		flags := flagsAddAlias{
			Contract: "TestContract",
			Address:  "0x1234567890abcdef",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		assert.Nil(t, data)
		assert.True(t, flagsProvided)
		assert.EqualError(t, err, "network name must be provided")
	})

	t.Run("Fail missing address", func(t *testing.T) {
		flags := flagsAddAlias{
			Contract: "TestContract",
			Network:  "testnet",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		assert.Nil(t, data)
		assert.True(t, flagsProvided)
		assert.EqualError(t, err, "address must be provided")
	})

	t.Run("Fail invalid address", func(t *testing.T) {
		flags := flagsAddAlias{
			Contract: "TestContract",
			Network:  "testnet",
			Address:  "invalid-address",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		assert.Nil(t, data)
		assert.True(t, flagsProvided)
		assert.EqualError(t, err, "invalid address")
	})

	t.Run("Fail empty address", func(t *testing.T) {
		flags := flagsAddAlias{
			Contract: "TestContract",
			Network:  "testnet",
			Address:  "0x0000000000000000",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		assert.Nil(t, data)
		assert.True(t, flagsProvided)
		assert.EqualError(t, err, "invalid address")
	})

	t.Run("Success with address without 0x prefix", func(t *testing.T) {
		flags := flagsAddAlias{
			Contract: "TestContract",
			Network:  "testnet",
			Address:  "1234567890abcdef",
		}

		data, flagsProvided, err := flagsToAliasData(flags)

		require.NoError(t, err)
		assert.True(t, flagsProvided)
		assert.Equal(t, "1234567890abcdef", data.Address)
	})
}
