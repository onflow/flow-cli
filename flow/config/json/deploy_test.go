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
	"github.com/stretchr/testify/require"
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
	err := json.Unmarshal(b, &jsonDeploys)
	require.NoError(t, err)

	deploys := jsonDeploys.transformToConfig()

	assert.Equal(t, deploys.GetByAccountAndNetwork("account-2", "testnet")[0].Account, "account-2")
	assert.Equal(t, deploys.GetByAccountAndNetwork("account-2", "testnet")[0].Contracts, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"})

	assert.Len(t, deploys.GetByNetwork("emulator"), 2)
	assert.Equal(t, deploys.GetByAccountAndNetwork("account-3", "emulator")[0].Account, "account-3")
	assert.Equal(t, deploys.GetByAccountAndNetwork("account-4", "emulator")[0].Account, "account-4")
	assert.Equal(t, deploys.GetByAccountAndNetwork("account-3", "emulator")[0].Contracts, []string{"KittyItems", "KittyItemsMarket"})
	assert.Equal(t, deploys.GetByAccountAndNetwork("account-4", "emulator")[0].Contracts, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"})
}

func Test_TransformDeployToJSON(t *testing.T) {
	b := []byte(`{"emulator":{"account-3":["KittyItems","KittyItemsMarket"],"account-4":["FungibleToken","NonFungibleToken","Kibble","KittyItems","KittyItemsMarket"]},"testnet":{"account-2":["FungibleToken","NonFungibleToken","Kibble","KittyItems"]}}`)

	var jsonDeploys jsonDeploys
	err := json.Unmarshal(b, &jsonDeploys)
	require.NoError(t, err)

	deploys := jsonDeploys.transformToConfig()

	j := transformDeploysToJSON(deploys)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
