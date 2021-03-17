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
package config_test

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flow/config"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func generateComplexConfig() config.Config {
	return config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:    "NonFungibleToken",
			Source:  "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
			Network: "emulator",
		}, {
			Name:    "FungibleToken",
			Source:  "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			Network: "emulator",
		}, {
			Name:    "Kibble",
			Source:  "./cadence/kibble/contracts/Kibble.cdc",
			Network: "emulator",
		}, {
			Name:    "KittyItems",
			Source:  "./cadence/kittyItems/contracts/KittyItems.cdc",
			Network: "emulator",
		}, {
			Name:    "KittyItemsMarket",
			Source:  "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			Network: "emulator",
		}, {
			Name:    "KittyItemsMarket",
			Source:  "0x123123123",
			Network: "testnet",
		}},
		Deployments: config.Deployments{{
			Network:   "emulator",
			Account:   "emulator-account",
			Contracts: []string{"KittyItems", "KittyItemsMarket"},
		}, {
			Network:   "emulator",
			Account:   "account-4",
			Contracts: []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"},
		}, {
			Network:   "testnet",
			Account:   "account-2",
			Contracts: []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"},
		}},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
			ChainID: flow.Emulator,
			Keys: []config.AccountKey{{
				Type:     config.KeyTypeHex,
				Index:    0,
				SigAlgo:  crypto.ECDSA_P256,
				HashAlgo: crypto.SHA3_256,
				Context: map[string]string{
					"privateKey": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47",
				},
			}},
		}, {
			Name:    "account-2",
			Address: flow.HexToAddress("2c1162386b0a245f"),
			ChainID: flow.Emulator,
			Keys: []config.AccountKey{{
				Type:     config.KeyTypeHex,
				Index:    0,
				SigAlgo:  crypto.ECDSA_P256,
				HashAlgo: crypto.SHA3_256,
				Context: map[string]string{
					"privateKey": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47",
				},
			}},
		}, {
			Name:    "account-4",
			Address: flow.HexToAddress("f8d6e0586b0a20c1"),
			ChainID: flow.Emulator,
			Keys: []config.AccountKey{{
				Type:     config.KeyTypeHex,
				Index:    0,
				SigAlgo:  crypto.ECDSA_P256,
				HashAlgo: crypto.SHA3_256,
				Context: map[string]string{
					"privateKey": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47",
				},
			}},
		}},
		Networks: config.Networks{{
			Name:    "emulator",
			Host:    "127.0.0.1.3569",
			ChainID: flow.Emulator,
		}},
	}
}

func Test_GetContractsForNetworkComplex(t *testing.T) {
	conf := generateComplexConfig()
	kitty := conf.Contracts.GetByName("KittyItems")
	market := conf.Contracts.GetByName("KittyItemsMarket")

	assert.Equal(t, kitty.Name, "KittyItems")
	assert.Equal(t, market.Source, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")
}

func Test_GetContractsByNameAndNetworkComplex(t *testing.T) {
	conf := generateComplexConfig()
	market := conf.Contracts.GetByNameAndNetwork("KittyItemsMarket", "testnet")

	assert.Equal(t, market.Source, "0x123123123")
}

func Test_GetContractsByNetworkComplex(t *testing.T) {
	conf := generateComplexConfig()
	contracts := conf.Contracts.GetByNetwork("emulator")

	assert.Equal(t, 5, len(contracts))
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
	assert.Equal(t, contracts[1].Name, "FungibleToken")
	assert.Equal(t, contracts[2].Name, "Kibble")
	assert.Equal(t, contracts[3].Name, "KittyItems")
	assert.Equal(t, contracts[4].Name, "KittyItemsMarket")
}

func Test_GetAccountByNameComplex(t *testing.T) {
	conf := generateComplexConfig()
	acc := conf.Accounts.GetByName("account-4")

	assert.Equal(t, acc.Address.String(), "f8d6e0586b0a20c1")
}

func Test_GetAccountByAddressComplex(t *testing.T) {
	conf := generateComplexConfig()
	acc1 := conf.Accounts.GetByAddress("f8d6e0586b0a20c1")
	acc2 := conf.Accounts.GetByAddress("2c1162386b0a245f")

	assert.Equal(t, acc1.Name, "account-4")
	assert.Equal(t, acc2.Name, "account-2")
}

func Test_GetDeploymentsByNetworkComplex(t *testing.T) {
	conf := generateComplexConfig()
	deployments := conf.Deployments.GetByAccountAndNetwork("account-2", "testnet")

	assert.Equal(t, deployments[0].Contracts, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"})
}

func Test_GetNetworkByNameComplex(t *testing.T) {
	conf := generateComplexConfig()
	network := conf.Networks.GetByName("emulator")

	assert.Equal(t, network.Host, "127.0.0.1.3569")
}
