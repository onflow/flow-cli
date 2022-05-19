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

package flowkit

import (
	"fmt"
	"sort"
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/config/json"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

var af = afero.Afero{Fs: afero.NewMemMapFs()}
var composer = config.NewLoader(af)

func keys() []crypto.PrivateKey {
	var keys []crypto.PrivateKey
	key, _ := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
	keys = append(keys, key)
	key, _ = crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "388e3fbdc654b765942610679bb3a66b74212149ab9482187067ee116d9a8118")
	keys = append(keys, key)
	key, _ = crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "27bbeba308e501f8485ddaab77e285c0bc0d611096a79b4f0b4ccc927c6dbf04")
	keys = append(keys, key)

	return keys
}

func generateComplexProject() State {
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
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[0],
			},
		}, {
			Name:    "account-2",
			Address: flow.HexToAddress("2c1162386b0a245f"),
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[1],
			},
		}, {
			Name:    "account-4",
			Address: flow.HexToAddress("f8d6e0586b0a20c1"),
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[2],
			},
		}, {
			Name:    "emulator-account-2",
			Address: flow.HexToAddress("2c1162386b0a245f"),
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[1],
			},
		}},
		Networks: config.Networks{{
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}
func generateSimpleProject() State {
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
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[0],
			},
		}},
		Networks: config.Networks{{
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}

	composer.AddConfigParser(json.NewParser())
	p, err := newProject(&config, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateAliasesProject() State {
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
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[0],
			},
		}},
		Networks: config.Networks{{
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateAliasesComplexProject() State {
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
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[0],
			},
		}, {
			Name:    "testnet-account",
			Address: flow.HexToAddress("1e82856bf20e2aa6"),
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[1],
			},
		}},
		Networks: config.Networks{{
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}, {
			Name: "testnet",
			Host: "127.0.0.1.3569",
		}},
	}

	p, err := newProject(&config, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func Test_GetContractsByNameSimple(t *testing.T) {
	p := generateSimpleProject()

	contracts, _ := p.DeploymentContractsByNetwork("emulator")
	account, err := p.conf.Accounts.ByName("emulator-account")
	assert.NoError(t, err)
	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
	assert.Equal(t, contracts[0].Source, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
	assert.Equal(t, account.Address, contracts[0].AccountAddress)
}

func Test_EmulatorConfigSimple(t *testing.T) {
	p := generateSimpleProject()
	emulatorServiceAccount, _ := p.EmulatorServiceAccount()

	assert.Equal(t, emulatorServiceAccount.name, "emulator-account")
	assert.Equal(t, emulatorServiceAccount.key.ToConfig().PrivateKey, keys()[0])
	assert.Equal(t, flow.ServiceAddress("flow-emulator"), emulatorServiceAccount.Address())
}

func Test_AccountByAddressSimple(t *testing.T) {
	p := generateSimpleProject()
	acc, _ := p.Accounts().ByAddress(flow.ServiceAddress("flow-emulator"))

	assert.Equal(t, acc.name, "emulator-account")
}

func Test_AccountByNameSimple(t *testing.T) {
	p := generateSimpleProject()
	acc, _ := p.Accounts().ByName("emulator-account")

	assert.Equal(t, flow.ServiceAddress("flow-emulator"), acc.Address())
	assert.Equal(t, acc.key.ToConfig().PrivateKey, keys()[0])
}

func Test_HostSimple(t *testing.T) {
	p := generateSimpleProject()
	network, err := p.Networks().ByName("emulator")

	assert.NoError(t, err)
	assert.Equal(t, network.Host, "127.0.0.1.3569")
}

func Test_GetContractsByNameComplex(t *testing.T) {
	p := generateComplexProject()

	contracts, _ := p.DeploymentContractsByNetwork("emulator")

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
		return c.AccountAddress.String()
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
	assert.Equal(t, emulatorServiceAccount.key.ToConfig().PrivateKey, keys()[0])
	assert.Equal(t, emulatorServiceAccount.Address(), flow.ServiceAddress("flow-emulator"))
}
func Test_AccountByNameWithDuplicateAddress(t *testing.T) {
	p := generateComplexProject()
	acc1, err := p.Accounts().ByName("emulator-account")

	assert.NoError(t, err)
	acc2, err := p.Accounts().ByName("emulator-account-2")
	assert.NoError(t, err)

	assert.Equal(t, acc1.name, "emulator-account")
	assert.Equal(t, acc2.name, "emulator-account-2")
}
func Test_AccountByNameComplex(t *testing.T) {
	p := generateComplexProject()
	acc, _ := p.Accounts().ByName("account-2")

	assert.Equal(t, acc.Address().String(), "2c1162386b0a245f")
	assert.Equal(t, acc.key.ToConfig().PrivateKey, keys()[1])
}

func Test_HostComplex(t *testing.T) {
	p := generateComplexProject()
	network, err := p.Networks().ByName("emulator")

	assert.NoError(t, err)

	assert.Equal(t, network.Host, "127.0.0.1.3569")
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
	contracts, _ := p.DeploymentContractsByNetwork("emulator")

	assert.Len(t, aliases, 1)
	assert.Equal(t, aliases["../hungry-kitties/cadence/contracts/FungibleToken.cdc"], "ee82856bf20e2aa6")
	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
}

func Test_GetAliasesComplex(t *testing.T) {
	p := generateAliasesComplexProject()

	aEmulator := p.AliasesForNetwork("emulator")
	cEmulator, _ := p.DeploymentContractsByNetwork("emulator")

	aTestnet := p.AliasesForNetwork("testnet")
	cTestnet, _ := p.DeploymentContractsByNetwork("testnet")

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

func Test_ChangingState(t *testing.T) {
	p := generateSimpleProject()

	em, err := p.EmulatorServiceAccount()
	assert.NoError(t, err)

	em.SetName("foo")
	em.SetAddress(flow.HexToAddress("0x1"))

	pk, _ := crypto.GeneratePrivateKey(
		crypto.ECDSA_P256,
		[]byte("seedseedseedseedseedseedseedseedseedseedseedseed"),
	)
	key := NewHexAccountKeyFromPrivateKey(em.Key().Index(), em.Key().HashAlgo(), pk)
	em.SetKey(key)

	foo, err := p.Accounts().ByName("foo")
	assert.NoError(t, err)
	assert.NotNil(t, foo)
	assert.Equal(t, foo.Name(), "foo")
	assert.Equal(t, foo.Address(), flow.HexToAddress("0x1"))

	pkey, err := foo.Key().PrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, (*pkey).String(), pk.String())

	bar, err := p.Accounts().ByName("foo")
	assert.NoError(t, err)
	bar.SetName("zoo")
	zoo, err := p.Accounts().ByName("zoo")
	assert.NotNil(t, zoo)
	assert.NoError(t, err)

	a := Account{}
	a.SetName("bobo")
	p.Accounts().AddOrUpdate(&a)
	bobo, err := p.Accounts().ByName("bobo")
	assert.NotNil(t, bobo)
	assert.NoError(t, err)

	zoo2, _ := p.Accounts().ByName("zoo")
	zoo2.SetName("emulator-account")
	assert.Equal(t, "emulator-account", zoo2.name)

	pk, _ = crypto.GeneratePrivateKey(
		crypto.ECDSA_P256,
		[]byte("seedseedseedseedseedseedseedseedseed123seedseedseed123"),
	)

	p.SetEmulatorKey(pk)
	em, _ = p.EmulatorServiceAccount()
	pkey, err = em.Key().PrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, (*pkey).String(), pk.String())
}

func Test_LoadState(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", b, 0644)
	assert.NoError(t, err)

	paths := []string{"flow.json"}
	state, err := Load(paths, af)
	assert.NoError(t, err)

	acc, err := state.Accounts().ByName("emulator-account")
	assert.NoError(t, err)
	assert.Equal(t, acc.Name(), "emulator-account")
}

func Test_LoadStateMultiple(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff0347a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			},
			"charlie": {
				"address": "0xf3fcd2c1a78f5eee",
				"key": "9463ceedf08627108ea0b394c96b18446d1370e7332c91ce332aba1594096ba0"
			},
			"bob": {
				"address": "179b6b1cb6755e31",
				"key": "748d21a762aa192976f4d264afe26379b8a63f5d1343773f813a37d4262b9f52"
			}
		}
	}`)

	a := []byte(`{
		"accounts": {
			"alice": {
				"address": "179b6b1cb6755e31",
				"key": "728d21a7622a29d976f4d264afe26379b8a63f5d1343773f813a37d4262b9f52"
			},
			"charlie": {
				"address": "0xe03daebed8ca0615",
				"key": "9463ceedf04427208ea0b394c96b18446d1370e7332c91ce332aba1594096ba0"
			}
		}
	}`)

	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", b, 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(af.Fs, "a.json", a, 0644)
	assert.NoError(t, err)

	paths := []string{"flow.json", "a.json"}
	state, err := Load(paths, af)
	assert.NoError(t, err)

	acc, err := state.Accounts().ByName("emulator-account")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address().String(), "f8d6e0586b0a20c7")

	acc, err = state.Accounts().ByName("charlie")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address().String(), "e03daebed8ca0615")

	acc, err = state.Accounts().ByName("bob")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address().String(), "179b6b1cb6755e31")

	acc, err = state.Accounts().ByName("alice")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address().String(), "179b6b1cb6755e31")
}

func Test_Saving(t *testing.T) {
	s := generateSimpleProject()

	err := s.SaveEdited([]string{"a.json", "b.json"})
	assert.EqualError(t, err, "specifying multiple paths is not supported when updating configuration")

	err = s.SaveEdited([]string{config.GlobalPath(), config.DefaultPath})
	assert.EqualError(t, err, "default configuration not found, please initialize it first or specify another configuration file")

	err = s.SaveEdited([]string{"a.json"})
	assert.NoError(t, err)

	_ = afero.WriteFile(af.Fs, config.DefaultPath, []byte(`{
		"networks": {
			"foo": "localhost:3000"
		}
	}`), 0644)

	err = s.SaveEdited([]string{config.GlobalPath(), config.DefaultPath})
	assert.NoError(t, err)
}
