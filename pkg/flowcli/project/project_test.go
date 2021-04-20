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

package project

import (
	"fmt"
	"sort"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
)

var composer = config.NewLoader(afero.NewOsFs())

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
		Deployments: config.Deployments{{
			Network: "emulator",
			Account: "emulator-account",
			Contracts: []config.ContractDeployment{
				{Name: "KittyItems", Args: nil},
				{Name: "KittyItemsMarket", Args: nil},
			},
		}, {
			Network: "emulator",
			Account: "account-4",
			Contracts: []config.ContractDeployment{
				{Name: "FungibleToken", Args: nil},
				{Name: "NonFungibleToken", Args: nil},
				{Name: "Kibble", Args: nil},
				{Name: "KittyItems", Args: nil},
				{Name: "KittyItemsMarket", Args: nil},
			},
		}, {
			Network: "testnet",
			Account: "account-2",
			Contracts: []config.ContractDeployment{
				{Name: "FungibleToken", Args: nil},
				{Name: "NonFungibleToken", Args: nil},
				{Name: "Kibble", Args: nil},
				{Name: "KittyItems", Args: nil},
			},
		}},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
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
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer)
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
		Deployments: config.Deployments{{
			Network: "emulator",
			Account: "emulator-account",
			Contracts: []config.ContractDeployment{
				{Name: "NonFungibleToken", Args: nil},
			},
		}},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
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
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer)
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
		Deployments: config.Deployments{{
			Network: "emulator",
			Account: "emulator-account",
			Contracts: []config.ContractDeployment{
				{Name: "NonFungibleToken", Args: nil},
			},
		}},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
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
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer)
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
		Deployments: config.Deployments{{
			Network: "emulator",
			Account: "emulator-account",
			Contracts: []config.ContractDeployment{
				{Name: "NonFungibleToken", Args: nil},
			},
		}, {
			Network: "testnet",
			Account: "testnet-account",
			Contracts: []config.ContractDeployment{
				{Name: "NonFungibleToken", Args: nil},
				{Name: "FungibleToken", Args: nil},
			},
		}},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
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
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}, {
			Name: "testnet",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer)
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

	contracts, _ := p.ContractsByNetwork("emulator")

	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
	assert.Equal(t, contracts[0].Source, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
	assert.Equal(t, p.conf.Accounts.GetByName("emulator-account").Address, contracts[0].Target)
}

func Test_EmulatorConfigSimple(t *testing.T) {
	p := generateSimpleProject()
	emulatorServiceAccount, _ := p.EmulatorServiceAccount()

	assert.Equal(t, emulatorServiceAccount.name, "emulator-account")
	assert.Equal(t, emulatorServiceAccount.keys[0].ToConfig().Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
	assert.Equal(t, flow.ServiceAddress("flow-emulator").String(), emulatorServiceAccount.Address().String())
}

func Test_AccountByAddressSimple(t *testing.T) {
	p := generateSimpleProject()
	acc := p.AccountByAddress(flow.ServiceAddress("flow-emulator").String())

	assert.Equal(t, acc.name, "emulator-account")
}

func Test_AccountByNameSimple(t *testing.T) {
	p := generateSimpleProject()
	acc := p.AccountByName("emulator-account")

	assert.Equal(t, flow.ServiceAddress("flow-emulator").String(), acc.Address().String())
	assert.Equal(t, acc.DefaultKey().ToConfig().Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_HostSimple(t *testing.T) {
	p := generateSimpleProject()
	host := p.NetworkByName("emulator").Host

	assert.Equal(t, host, "127.0.0.1.3569")
}

func Test_GetContractsByNameComplex(t *testing.T) {
	p := generateComplexProject()

	contracts, _ := p.ContractsByNetwork("emulator")

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
	emulatorServiceAccount, _ := p.EmulatorServiceAccount()

	assert.Equal(t, emulatorServiceAccount.name, "emulator-account")
	assert.Equal(t, emulatorServiceAccount.keys[0].ToConfig().Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
	assert.Equal(t, emulatorServiceAccount.Address().String(), flow.ServiceAddress("flow-emulator").String())
}

func Test_AccountByAddressComplex(t *testing.T) {
	p := generateComplexProject()
	acc1 := p.AccountByAddress("f8d6e0586b0a20c1")
	acc2 := p.AccountByAddress("0x2c1162386b0a245f")

	assert.Equal(t, acc1.name, "account-4")
	assert.Equal(t, acc2.name, "account-2")
}

func Test_AccountByNameComplex(t *testing.T) {
	p := generateComplexProject()
	acc := p.AccountByName("account-2")

	assert.Equal(t, acc.Address().String(), "2c1162386b0a245f")
	assert.Equal(t, acc.DefaultKey().ToConfig().Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_HostComplex(t *testing.T) {
	p := generateComplexProject()
	host := p.NetworkByName("emulator").Host

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

	aliases := p.AliasesForNetwork("emulator")
	contracts, _ := p.ContractsByNetwork("emulator")

	assert.Len(t, aliases, 1)
	assert.Equal(t, aliases["../hungry-kitties/cadence/contracts/FungibleToken.cdc"], "ee82856bf20e2aa6")
	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
}

func Test_GetAliasesComplex(t *testing.T) {
	p := generateAliasesComplexProject()

	aEmulator := p.AliasesForNetwork("emulator")
	cEmulator, _ := p.ContractsByNetwork("emulator")

	aTestnet := p.AliasesForNetwork("testnet")
	cTestnet, _ := p.ContractsByNetwork("testnet")

	assert.Len(t, cEmulator, 1)
	assert.Equal(t, cEmulator[0].Name, "NonFungibleToken")

	assert.Len(t, aEmulator, 2)
	assert.Equal(t, aEmulator["../hungry-kitties/cadence/contracts/FungibleToken.cdc"], "ee82856bf20e2aa6")
	assert.Equal(t, aEmulator["../hungry-kitties/cadence/contracts/Kibble.cdc"], "ee82856bf20e2aa6")

	assert.Len(t, aTestnet, 1)
	assert.Equal(t, aTestnet["../hungry-kitties/cadence/contracts/Kibble.cdc"], "ee82856bf20e2aa6")

	assert.Len(t, cTestnet, 2)
	assert.Equal(t, cTestnet[0].Name, "NonFungibleToken")
	assert.Equal(t, cTestnet[1].Name, "FungibleToken")
}

func Test_SDKParsing(t *testing.T) {

	t.Run("Address Parsing", func(t *testing.T) {
		addr1 := flow.HexToAddress("0xf8d6e0586b0a20c7")
		addr2 := flow.HexToAddress("f8d6e0586b0a20c7")

		assert.True(t, addr1.IsValid(flow.Emulator))
		assert.True(t, addr2.IsValid(flow.Emulator))
		assert.Equal(t, addr1.String(), addr2.String())
	})
	/* TODO test this after it is implemented in sdk
	t.Run("Tx ID Parsing", func(t *testing.T) {
		txid := "09f24d9dcde4c4d63d2f790e42905427ba04e6b0d601a7ec790b663f7cf2d942"
		id1 := flow.HexToID(txid)
		id2 := flow.HexToID("0x" + txid)

		assert.Equal(t, id1.String(), id2.String())
	})

	t.Run("Public Key Hex Parsing", func(t *testing.T) {
		pubKey := "642fcceac4b0af1ea7b78c11d5ce2ed505bb41b13c9e3f57725246b75d828651d9387e0cd19c5ebb1a44d571ce58cc3a83f0d92a6d3a70a45fe359d5d25d15d7"
		k1, err := crypto.DecodePublicKeyHex(crypto.ECDSA_P256, pubKey)
		assert.NoError(t, err)

		k2, err := crypto.DecodePublicKeyHex(crypto.ECDSA_P256, "0x"+pubKey)
		assert.NoError(t, err)

		assert.Equal(t, k1.String(), k2.String())
	})
	*/
}
