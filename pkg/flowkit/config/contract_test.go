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

package config

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestAliases_ByNetwork(t *testing.T) {
	aliases := Aliases{
		{
			Network: "testnet",
			Address: flow.HexToAddress("01"),
		},
		{
			Network: "mainnet",
			Address: flow.HexToAddress("02"),
		},
	}

	assert.Equal(t, flow.HexToAddress("01"), aliases.ByNetwork("testnet").Address)
	assert.Equal(t, flow.HexToAddress("02"), aliases.ByNetwork("mainnet").Address)
	assert.Nil(t, aliases.ByNetwork("nonexistent"))
}

func TestAliases_Add(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("01"))

	assert.Equal(t, 1, len(aliases))
	assert.Equal(t, flow.HexToAddress("01"), aliases.ByNetwork("testnet").Address)

	aliases.Add("testnet", flow.HexToAddress("02")) // should not add a new entry
	assert.Equal(t, 1, len(aliases))
	assert.Equal(t, flow.HexToAddress("01"), aliases.ByNetwork("testnet").Address)
}

func TestContracts_IsAliased(t *testing.T) {
	contract := Contract{
		Name:     "TestContract",
		Location: "test.cdc",
		Aliases: Aliases{
			{
				Network: "testnet",
				Address: flow.HexToAddress("01"),
			},
		},
	}

	assert.True(t, contract.IsAliased())

	contract.Aliases = Aliases{}
	assert.False(t, contract.IsAliased())
}

func TestContracts_ByName(t *testing.T) {
	contracts := Contracts{
		{
			Name:     "TestContract",
			Location: "test.cdc",
		},
	}

	assert.NotNil(t, contracts.ByName("TestContract"))
	assert.Nil(t, contracts.ByName("NonExistentContract"))
}

func TestContracts_AddOrUpdate(t *testing.T) {
	contracts := Contracts{}
	contract := Contract{
		Name:     "TestContract",
		Location: "test.cdc",
	}

	assert.Equal(t, 0, len(contracts))

	contracts.AddOrUpdate(contract)
	assert.Equal(t, 1, len(contracts))
	assert.Equal(t, "test.cdc", contracts.ByName("TestContract").Location)

	contract.Location = "updated.cdc"
	contracts.AddOrUpdate(contract) // should update the existing contract
	assert.Equal(t, 1, len(contracts))
	assert.Equal(t, "updated.cdc", contracts.ByName("TestContract").Location)
}

func TestContracts_Remove(t *testing.T) {
	contracts := Contracts{
		{
			Name:     "TestContract",
			Location: "test.cdc",
		},
	}

	err := contracts.Remove("NonExistentContract")
	assert.NotNil(t, err)
	assert.Equal(t, 1, len(contracts))

	err = contracts.Remove("TestContract")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(contracts))
}
