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

func TestAliases_Add(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("0xabcdef"))

	alias := aliases.ByNetwork("testnet")
	assert.NotNil(t, alias)
}

func TestAliases_Add_Duplicate(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("0xabcdef"))
	aliases.Add("testnet", flow.HexToAddress("0x123456"))

	assert.Len(t, aliases, 1)
}

func TestContracts_AddOrUpdate_Add(t *testing.T) {
	contracts := Contracts{}
	contracts.AddOrUpdate(Contract{Name: "mycontract", Location: "path/to/contract.cdc"})

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("mycontract")
	assert.NoError(t, err)
	assert.Equal(t, "path/to/contract.cdc", contract.Location)
}

func TestContracts_AddOrUpdate_Update(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	contracts.AddOrUpdate(Contract{Name: "mycontract", Location: "new/path/to/contract.cdc"})

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("mycontract")
	assert.NoError(t, err)
	assert.Equal(t, "new/path/to/contract.cdc", contract.Location)
}

func TestContracts_Remove(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	err := contracts.Remove("mycontract")
	assert.NoError(t, err)
	assert.Len(t, contracts, 0)
}

func TestContracts_Remove_NotFound(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract1", Location: "path/to/contract.cdc"},
		Contract{Name: "mycontract2", Location: "path/to/contract.cdc"},
		Contract{Name: "mycontract3", Location: "path/to/contract.cdc"},
	}
	err := contracts.Remove("mycontract2")
	assert.NoError(t, err)
	assert.Len(t, contracts, 2)
	_, err = contracts.ByName("mycontract1")
	assert.NoError(t, err)
	_, err = contracts.ByName("mycontract3")
	assert.NoError(t, err)
	_, err = contracts.ByName("mycontract2")
	assert.EqualError(t, err, "contract mycontract2 does not exist")
}

func TestContracts_AddDependencyAsContract(t *testing.T) {
	contracts := Contracts{}
	contracts.AddDependencyAsContract(Dependency{
		Name: "testcontract",
		RemoteSource: RemoteSource{
			NetworkName:  "testnet",
			Address:      flow.HexToAddress("0x0000000000abcdef"),
			ContractName: "TestContract",
		},
	}, "testnet")

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("testcontract")
	assert.NoError(t, err)
	assert.Equal(t, "imports/0000000000abcdef/TestContract.cdc", contract.Location)
	assert.Len(t, contract.Aliases, 1)
}
