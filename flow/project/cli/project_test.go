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

package cli

import (
	"fmt"
	"sort"
	"testing"

	"github.com/onflow/flow-cli/flow/project/cli/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

func generateComplexProject() Project {
	config := config.Config{
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
		Deploys: config.Deploys{{
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

	p, err := newProject(&config)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateSimpleProject() Project {
	config := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:    "NonFungibleToken",
			Source:  "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
			Network: "emulator",
		}},
		Deploys: config.Deploys{{
			Network:   "emulator",
			Account:   "emulator-account",
			Contracts: []string{"NonFungibleToken"},
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
		}},
		Networks: config.Networks{{
			Name:    "emulator",
			Host:    "127.0.0.1.3569",
			ChainID: flow.Emulator,
		}},
	}

	p, err := newProject(&config)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateAliasesProject() Project {
	config := config.Config{
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
			Alias:   "ee82856bf20e2aa6",
		}},
		Deploys: config.Deploys{{
			Network:   "emulator",
			Account:   "emulator-account",
			Contracts: []string{"NonFungibleToken"},
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
		}},
		Networks: config.Networks{{
			Name:    "emulator",
			Host:    "127.0.0.1.3569",
			ChainID: flow.Emulator,
		}},
	}

	p, err := newProject(&config)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateAliasesComplexProject() Project {
	config := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:   "NonFungibleToken",
			Source: "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		}, {
			Name:    "FungibleToken",
			Source:  "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			Network: "emulator",
			Alias:   "ee82856bf20e2aa6",
		}, {
			Name:    "Kibble",
			Source:  "../hungry-kitties/cadence/contracts/Kibble.cdc",
			Network: "testnet",
			Alias:   "ee82856bf20e2aa6",
		}, {
			Name:    "Kibble",
			Source:  "../hungry-kitties/cadence/contracts/Kibble.cdc",
			Network: "emulator",
			Alias:   "ee82856bf20e2aa6",
		}},
		Deploys: config.Deploys{{
			Network:   "emulator",
			Account:   "emulator-account",
			Contracts: []string{"NonFungibleToken"},
		}, {
			Network:   "testnet",
			Account:   "testnet-account",
			Contracts: []string{"NonFungibleToken", "FungibleToken"},
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
			Name:    "testnet-account",
			Address: flow.HexToAddress("1e82856bf20e2aa6"),
			ChainID: flow.Testnet,
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
		}, {
			Name:    "testnet",
			Host:    "127.0.0.1.3569",
			ChainID: flow.Testnet,
		}},
	}

	p, err := newProject(&config)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

/* ================================================================
Project Tests
================================================================ */
func Test_GetContractsByNameSimple(t *testing.T) {
	p := generateSimpleProject()

	contracts := p.GetContractsByNetwork("emulator")

	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
	assert.Equal(t, contracts[0].Source, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
	assert.Equal(t, p.conf.Accounts.GetByName("emulator-account").Address, contracts[0].Target)
}

func Test_EmulatorConfigSimple(t *testing.T) {
	p := generateSimpleProject()
	emulatorServiceAccount := p.EmulatorServiceAccount()

	assert.Equal(t, emulatorServiceAccount.Name, "emulator-account")
	assert.Equal(t, emulatorServiceAccount.Keys[0].Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
	assert.Equal(t, flow.ServiceAddress("flow-emulator"), emulatorServiceAccount.Address)
}

func Test_AccountByAddressSimple(t *testing.T) {
	p := generateSimpleProject()
	acc := p.GetAccountByAddress(flow.ServiceAddress("flow-emulator").String())

	assert.Equal(t, acc.name, "emulator-account")
}

func Test_AccountByNameSimple(t *testing.T) {
	p := generateSimpleProject()
	acc := p.GetAccountByName("emulator-account")

	assert.Equal(t, flow.ServiceAddress("flow-emulator").String(), acc.Address().String())
	assert.Equal(t, acc.DefaultKey().ToConfig().Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_HostSimple(t *testing.T) {
	p := generateSimpleProject()
	host := p.Host("emulator")

	assert.Equal(t, host, "127.0.0.1.3569")
}

func Test_GetContractsByNameComplex(t *testing.T) {
	p := generateComplexProject()

	contracts := p.GetContractsByNetwork("emulator")

	assert.Equal(t, 7, len(contracts))

	//sort names so tests are deterministic
	contractNames := funk.Map(contracts, func(c Contract) string {
		return c.Name
	}).([]string)
	sort.Strings(contractNames)

	sources := funk.Map(contracts, func(c Contract) string {
		return c.Source
	}).([]string)
	sort.Strings(sources)

	targets := funk.Map(contracts, func(c Contract) string {
		return c.Target.String()
	}).([]string)
	sort.Strings(targets)

	assert.Equal(t, contractNames[0], "FungibleToken")
	assert.Equal(t, contractNames[1], "Kibble")
	assert.Equal(t, contractNames[2], "KittyItems")
	assert.Equal(t, contractNames[3], "KittyItems")
	assert.Equal(t, contractNames[4], "KittyItemsMarket")
	assert.Equal(t, contractNames[5], "KittyItemsMarket")
	assert.Equal(t, contractNames[6], "NonFungibleToken")

	assert.Equal(t, sources[0], "../hungry-kitties/cadence/contracts/FungibleToken.cdc")
	assert.Equal(t, sources[1], "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
	assert.Equal(t, sources[2], "cadence/kibble/contracts/Kibble.cdc")
	assert.Equal(t, sources[3], "cadence/kittyItems/contracts/KittyItems.cdc")
	assert.Equal(t, sources[4], "cadence/kittyItems/contracts/KittyItems.cdc")
	assert.Equal(t, sources[5], "cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")
	assert.Equal(t, sources[6], "cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc")

	assert.Equal(t, targets[0], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[1], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[2], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[3], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[4], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[5], "f8d6e0586b0a20c7")
	assert.Equal(t, targets[6], "f8d6e0586b0a20c7")
}

func Test_EmulatorConfigComplex(t *testing.T) {
	p := generateComplexProject()
	emulatorServiceAccount := p.EmulatorServiceAccount()

	assert.Equal(t, emulatorServiceAccount.Name, "emulator-account")
	assert.Equal(t, emulatorServiceAccount.Keys[0].Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
	assert.Equal(t, emulatorServiceAccount.Address, flow.ServiceAddress("flow-emulator"))
}

func Test_AccountByAddressComplex(t *testing.T) {
	p := generateComplexProject()
	acc1 := p.GetAccountByAddress("f8d6e0586b0a20c1")
	acc2 := p.GetAccountByAddress("0x2c1162386b0a245f")

	assert.Equal(t, acc1.name, "account-4")
	assert.Equal(t, acc2.name, "account-2")
}

func Test_AccountByNameComplex(t *testing.T) {
	p := generateComplexProject()
	acc := p.GetAccountByName("account-2")

	assert.Equal(t, acc.Address().String(), "2c1162386b0a245f")
	assert.Equal(t, acc.DefaultKey().ToConfig().Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_HostComplex(t *testing.T) {
	p := generateComplexProject()
	host := p.Host("emulator")

	assert.Equal(t, host, "127.0.0.1.3569")
}

func Test_ContractConflictComplex(t *testing.T) {
	p := generateComplexProject()
	exists := p.ContractConflictExists("emulator")
	notexists := p.ContractConflictExists("testnet")

	assert.True(t, exists)
	assert.False(t, notexists)

}

func Test_GetAliases(t *testing.T) {
	p := generateAliasesProject()

	aliases := p.GetAliases("emulator")
	contracts := p.GetContractsByNetwork("emulator")

	assert.Len(t, aliases, 1)
	assert.Equal(t, aliases["FungibleToken"], "ee82856bf20e2aa6")
	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
}

func Test_GetAliasesComplex(t *testing.T) {
	p := generateAliasesComplexProject()

	aEmulator := p.GetAliases("emulator")
	cEmulator := p.GetContractsByNetwork("emulator")

	aTestnet := p.GetAliases("testnet")
	cTestnet := p.GetContractsByNetwork("testnet")

	assert.Len(t, cEmulator, 1)
	assert.Equal(t, cEmulator[0].Name, "NonFungibleToken")

	assert.Len(t, aEmulator, 2)
	assert.Equal(t, aEmulator["FungibleToken"], "ee82856bf20e2aa6")
	assert.Equal(t, aEmulator["Kibble"], "ee82856bf20e2aa6")

	assert.Len(t, aTestnet, 1)
	assert.Equal(t, aTestnet["Kibble"], "ee82856bf20e2aa6")

	assert.Len(t, cTestnet, 2)
	assert.Equal(t, cTestnet[0].Name, "NonFungibleToken")
	assert.Equal(t, cTestnet[1].Name, "FungibleToken")
}
