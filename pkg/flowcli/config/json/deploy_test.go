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
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"

	"github.com/stretchr/testify/assert"
)

func cleanSpecialChars(code []byte) string {
	space := regexp.MustCompile(`\s+`)
	return strings.ReplaceAll(space.ReplaceAllString(string(code), " "), " ", "")
}

func Test_ConfigDeploymentsSimple(t *testing.T) {
	b := []byte(`{
		"testnet": {
			"account-1": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"]
		},
		"emulator": {
			"account-2": ["KittyItems", "KittyItemsMarket"],
			"account-3": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
		}
	}`)

	var jsonDeployments jsonDeployments
	err := json.Unmarshal(b, &jsonDeployments)
	assert.NoError(t, err)

	deployments := jsonDeployments.transformToConfig()

	const account1Name = "account-1"
	const account2Name = "account-2"
	const account3Name = "account-3"

	assert.Equal(t, 1, len(deployments.GetByNetwork("testnet")))
	assert.Equal(t, 2, len(deployments.GetByNetwork("emulator")))

	account1Deployment := deployments.GetByAccountAndNetwork(account1Name, "testnet")
	account2Deployment := deployments.GetByAccountAndNetwork(account2Name, "emulator")
	account3Deployment := deployments.GetByAccountAndNetwork(account3Name, "emulator")

	require.Len(t, account1Deployment, 1)
	require.Len(t, account2Deployment, 1)
	require.Len(t, account3Deployment, 1)

	assert.Equal(t, account1Name, account1Deployment[0].Account)
	assert.Equal(t, account2Name, account2Deployment[0].Account)
	assert.Equal(t, account3Name, account3Deployment[0].Account)

	assert.Len(t, account1Deployment[0].Contracts, 4)

	for i, name := range []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"} {
		assert.Equal(t, account1Deployment[0].Contracts[i].Name, name)
	}

	for i, name := range []string{"KittyItems", "KittyItemsMarket"} {
		assert.Equal(t, account2Deployment[0].Contracts[i].Name, name)
	}

	for i, name := range []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"} {
		assert.Equal(t, account3Deployment[0].Contracts[i].Name, name)
	}

}

func Test_TransformDeployToJSON(t *testing.T) {
	b := []byte(`{
		"emulator":{
			"account-3":["KittyItems", {
					"name": "Kibble",
					"args": [{ "type": "String", "value": "Hello World" }]
			}],
			"account-4":["FungibleToken","NonFungibleToken","Kibble","KittyItems","KittyItemsMarket"]
		},
		"testnet":{
			"account-2":["FungibleToken","NonFungibleToken","Kibble","KittyItems"]
		}
	}`)

	var jsonDeployments jsonDeployments
	err := json.Unmarshal(b, &jsonDeployments)
	assert.NoError(t, err)

	deployments := jsonDeployments.transformToConfig()

	j := transformDeploymentsToJSON(deployments)
	x, _ := json.Marshal(j)

	assert.Equal(t, cleanSpecialChars(b), cleanSpecialChars(x))
}

func Test_DeploymentAdvanced(t *testing.T) {
	b := []byte(`{
		"emulator": {
			"alice": [
				{
					"name": "Kibble",
					"args": [{ "type": "String", "value": "Hello World" }]
				},
				"KittyItemsMarket"
			]
		}
	}`)

	var jsonDeployments jsonDeployments
	err := json.Unmarshal(b, &jsonDeployments)
	assert.NoError(t, err)

	deployments := jsonDeployments.transformToConfig()

	alice := deployments.GetByAccountAndNetwork("alice", "emulator")
	assert.Len(t, alice, 1)
	assert.Len(t, alice[0].Contracts, 2)
	assert.Equal(t, alice[0].Contracts[0].Name, "Kibble")
	assert.Equal(t, alice[0].Contracts[0].Args[0].Value, cadence.String("Hello World"))
	assert.Equal(t, alice[0].Contracts[1].Name, "KittyItemsMarket")
	assert.Len(t, alice[0].Contracts[1].Args, 0)
}
