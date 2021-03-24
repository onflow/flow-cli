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

	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItemsMarket.cdc", contracts.GetByName("KittyItemsMarket").Source)
}

func Test_ConfigContractsComplex(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
			"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			"aliases": {
      	"testnet": "0x123123123"
      }
    }
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, len(contracts), 2)

	assert.Equal(t, contracts.GetByName("KittyItems").Source, "./cadence/kittyItems/contracts/KittyItems.cdc")
	assert.Equal(t, contracts.GetByNameAndNetwork("KittyItemsMarket", "emulator").Source, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")
	assert.Equal(t, contracts.GetByNameAndNetwork("KittyItemsMarket", "testnet").Source, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")

	assert.Equal(t, contracts.GetByNameAndNetwork("KittyItems", "emulator").Alias, "")
	assert.Equal(t, contracts.GetByNameAndNetwork("KittyItems", "testnet").Alias, "")

	assert.Equal(t, contracts.GetByNameAndNetwork("KittyItemsMarket", "testnet").Alias, "0x123123123")
	assert.Equal(t, contracts.GetByNameAndNetwork("KittyItemsMarket", "emulator").Alias, "")
}

func Test_ConfigContractsAliases(t *testing.T) {
	b := []byte(`{
		"NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		"Kibble": {
			"source": "../hungry-kitties/cadence/contracts/Kibble.cdc",
			"aliases": {
				"emulator": "ee82856bf20e2aa6",
				"testnet": "1e82856bf20e2aa6"
			}
		},
    "FungibleToken": {
			"source": "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			"aliases": {
				"emulator": "2e82856bf20e2aa6"
			}
		}
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, contracts.GetByName("FungibleToken").Network, "emulator")
	assert.Equal(t, contracts.GetByName("FungibleToken").Alias, "2e82856bf20e2aa6")
	assert.Equal(t, contracts.GetByName("FungibleToken").Source, "../hungry-kitties/cadence/contracts/FungibleToken.cdc")
	assert.Equal(t, contracts.GetByNameAndNetwork("FungibleToken", "emulator").Alias, "2e82856bf20e2aa6")
	assert.Equal(t, contracts.GetByNameAndNetwork("FungibleToken", "testnet").Alias, "")
	assert.Equal(t, contracts.GetByNameAndNetwork("FungibleToken", "testnet").Network, "testnet")
	assert.Equal(t, contracts.GetByNameAndNetwork("FungibleToken", "testnet").Source, "../hungry-kitties/cadence/contracts/FungibleToken.cdc")
	assert.Equal(t, contracts.GetByNameAndNetwork("FungibleToken", "emulator").Source, "../hungry-kitties/cadence/contracts/FungibleToken.cdc")

	assert.Equal(t, contracts.GetByNameAndNetwork("Kibble", "testnet").Network, "testnet")
	assert.Equal(t, contracts.GetByNameAndNetwork("Kibble", "testnet").Alias, "1e82856bf20e2aa6")
	assert.Equal(t, contracts.GetByNameAndNetwork("Kibble", "emulator").Alias, "ee82856bf20e2aa6")
	assert.Equal(t, contracts.GetByNameAndNetwork("Kibble", "testnet").Source, "../hungry-kitties/cadence/contracts/Kibble.cdc")

	assert.Equal(t, contracts.GetByNameAndNetwork("NonFungibleToken", "testnet").Network, "testnet")
	assert.Equal(t, contracts.GetByNameAndNetwork("NonFungibleToken", "testnet").Alias, "")
	assert.Equal(t, contracts.GetByNameAndNetwork("NonFungibleToken", "testnet").Source, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
}

func Test_TransformContractToJSON(t *testing.T) {
	b := []byte(`{
		"KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
		"KittyItemsMarket": {
			"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			"aliases": {
				"testnet":"0x123123123"
			}
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts := jsonContracts.transformToConfig()

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
				"testnet":"0x123123123"
			}
		},
		"Kibble": {
			"source": "./cadence/kittyItems/contracts/KittyItems.cdc",
			"aliases": {
				"testnet": "0x22222222",
				"emulator": "0x123123123"
			}
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts := jsonContracts.transformToConfig()

	j := transformContractsToJSON(contracts)
	x, _ := json.Marshal(j)

	assert.JSONEq(t, string(b), string(x))
}
