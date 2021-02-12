/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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
	json.Unmarshal(b, &jsonContracts)
	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItemsMarket.cdc", contracts.GetByName("KittyItemsMarket").Source)
}

func Test_ConfigContractsComplex(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
      "testnet": "0x123123123",
      "emulator": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"
    }
  }`)

	var jsonContracts jsonContracts
	json.Unmarshal(b, &jsonContracts)
	contracts := jsonContracts.transformToConfig()

	assert.Equal(t, len(contracts), 3)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts.GetByNameAndNetwork("KittyItemsMarket", "emulator").Source)
	assert.Equal(t, "0x123123123", contracts.GetByNameAndNetwork("KittyItemsMarket", "testnet").Source)

	assert.Equal(t, 2, len(contracts.GetByNetwork("testnet")))
	assert.Equal(t, 2, len(contracts.GetByNetwork("emulator")))

	//REF: sort result since it is not arranged
	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByNetwork("testnet")[0].Source)
	assert.Equal(t, "0x123123123", contracts.GetByNetwork("testnet")[1].Source)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByNetwork("emulator")[0].Source)
	assert.Equal(t, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts.GetByNetwork("emulator")[1].Source)
}

func Test_TransformContractToJSON(t *testing.T) {
	b := []byte(`{"KittyItems":"./cadence/kittyItems/contracts/KittyItems.cdc","KittyItemsMarket":{"emulator":"./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc","testnet":"0x123123123"}}`)

	var jsonContracts jsonContracts
	json.Unmarshal(b, &jsonContracts)
	contracts := jsonContracts.transformToConfig()

	j := jsonContracts.transformToJSON(contracts)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
