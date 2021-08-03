/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigContractsSimple(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": "./cadence/kittyItems/contracts/KittyItemsMarket.cdc"
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	contract, err := contracts.ByName("KittyItems")
	assert.NoError(t, err)

	marketContract, err := contracts.ByName("KittyItemsMarket")
	assert.NoError(t, err)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contract.Source)
	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItemsMarket.cdc", marketContract.Source)
}

func Test_ConfigContractsComplex(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
		"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
		"aliases": {
			"testnet": "f8d6e0586b0a20c7"
		}
    }
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	assert.Equal(t, len(contracts), 2)

	contract, err := contracts.ByName("KittyItems")
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, contract.Source, "./cadence/kittyItems/contracts/KittyItems.cdc")
	assert.Equal(t, contracts.ByNameAndNetwork("KittyItemsMarket", "emulator").Source, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")
	assert.Equal(t, contracts.ByNameAndNetwork("KittyItemsMarket", "testnet").Source, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")

	assert.Equal(t, contracts.ByNameAndNetwork("KittyItems", "emulator").Alias, "")
	assert.Equal(t, contracts.ByNameAndNetwork("KittyItems", "testnet").Alias, "")

	assert.Equal(t, contracts.ByNameAndNetwork("KittyItemsMarket", "testnet").Alias, "f8d6e0586b0a20c7")
	assert.Equal(t, contracts.ByNameAndNetwork("KittyItemsMarket", "emulator").Alias, "")
}

func Test_ConfigContractsAliases(t *testing.T) {
	b := []byte(`{
		"NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		"Kibble": {
			"source": "../hungry-kitties/cadence/contracts/Kibble.cdc",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7",
				"testnet": "ead892083b3e2c6c"
			}
		},
		"FungibleToken": {
			"source": "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			"aliases": {
				"emulator": "e5a8b7f23e8b548f"
			}
		}
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	fungibleToken, err := contracts.ByName("FungibleToken")
	assert.NoError(t, err)

	assert.Equal(t, fungibleToken.Network, "emulator")
	assert.Equal(t, fungibleToken.Alias, "e5a8b7f23e8b548f")
	assert.Equal(t, fungibleToken.Source, "../hungry-kitties/cadence/contracts/FungibleToken.cdc")
	assert.Equal(t, contracts.ByNameAndNetwork("FungibleToken", "emulator").Alias, "e5a8b7f23e8b548f")
	assert.Equal(t, contracts.ByNameAndNetwork("FungibleToken", "testnet").Alias, "")
	assert.Equal(t, contracts.ByNameAndNetwork("FungibleToken", "testnet").Network, "testnet")
	assert.Equal(t, contracts.ByNameAndNetwork("FungibleToken", "testnet").Source, "../hungry-kitties/cadence/contracts/FungibleToken.cdc")
	assert.Equal(t, contracts.ByNameAndNetwork("FungibleToken", "emulator").Source, "../hungry-kitties/cadence/contracts/FungibleToken.cdc")

	assert.Equal(t, contracts.ByNameAndNetwork("Kibble", "testnet").Network, "testnet")
	assert.Equal(t, contracts.ByNameAndNetwork("Kibble", "testnet").Alias, "ead892083b3e2c6c")
	assert.Equal(t, contracts.ByNameAndNetwork("Kibble", "emulator").Alias, "f8d6e0586b0a20c7")
	assert.Equal(t, contracts.ByNameAndNetwork("Kibble", "testnet").Source, "../hungry-kitties/cadence/contracts/Kibble.cdc")

	assert.Equal(t, contracts.ByNameAndNetwork("NonFungibleToken", "testnet").Network, "testnet")
	assert.Equal(t, contracts.ByNameAndNetwork("NonFungibleToken", "testnet").Alias, "")
	assert.Equal(t, contracts.ByNameAndNetwork("NonFungibleToken", "testnet").Source, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
}

func Test_TransformContractToJSON(t *testing.T) {
	b := []byte(`{
		"KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
		"KittyItemsMarket": {
			"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			"aliases": {
				"testnet":"e5a8b7f23e8b548f"
			}
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	j := transformContractsToJSON(contracts)
	x, _ := json.Marshal(j)

	assert.JSONEq(t, string(b), string(x))
}

func Test_TransformComplexContractToJSON(t *testing.T) {
	b := []byte(`{
		"KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
		"KittyItemsMarket": {
			"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			"aliases": {
				"testnet":"e5a8b7f23e8b548f"
			}
		},
		"Kibble": {
			"source": "./cadence/kittyItems/contracts/KittyItems.cdc",
			"aliases": {
				"testnet": "e5a8b7f23e8b548f",
				"emulator": "f8d6e0586b0a20c7"
			}
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	j := transformContractsToJSON(contracts)
	x, _ := json.Marshal(j)

	assert.JSONEq(t, string(b), string(x))
}
