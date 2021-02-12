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

func Test_ConfigDeploySimple(t *testing.T) {
	b := []byte(`{
		"testnet": {
			"account-2": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"]
		}, 
		"emulator": {
			"account-3": ["KittyItems", "KittyItemsMarket"],
			"account-4": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
		}
	}`)

	var jsonDeploys jsonDeploys
	json.Unmarshal(b, &jsonDeploys)
	deploys := jsonDeploys.transformToConfig()

	//TODO: fix test to be sorted since its not necessary correct order
	assert.Equal(t, "account-2", deploys.GetByNetwork("testnet")[0].Account)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"}, deploys.GetByNetwork("testnet")[0].Contracts)

	assert.Equal(t, 2, len(deploys.GetByNetwork("emulator")))
	assert.Equal(t, "account-3", deploys.GetByNetwork("emulator")[0].Account)
	assert.Equal(t, "account-4", deploys.GetByNetwork("emulator")[1].Account)
	assert.Equal(t, []string{"KittyItems", "KittyItemsMarket"}, deploys.GetByNetwork("emulator")[0].Contracts)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"}, deploys.GetByNetwork("emulator")[1].Contracts)
}

func Test_TransformDeployToJSON(t *testing.T) {
	b := []byte(`{"emulator":{"account-3":["KittyItems","KittyItemsMarket"],"account-4":["FungibleToken","NonFungibleToken","Kibble","KittyItems","KittyItemsMarket"]},"testnet":{"account-2":["FungibleToken","NonFungibleToken","Kibble","KittyItems"]}}`)

	var jsonDeploys jsonDeploys
	json.Unmarshal(b, &jsonDeploys)
	deploys := jsonDeploys.transformToConfig()

	j := jsonDeploys.transformToJSON(deploys)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
