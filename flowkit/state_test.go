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
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/onflow/flow-cli/flowkit/accounts"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thoas/go-funk"

	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/config/json"
	"github.com/onflow/flow-cli/flowkit/project"
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
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:     "NonFungibleToken",
			Location: "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		}, {
			Name:     "FungibleToken",
			Location: "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
		}, {
			Name:     "Kibble",
			Location: "./cadence/kibble/contracts/Kibble.cdc",
		}, {
			Name:     "KittyItems",
			Location: "./cadence/kittyItems/contracts/KittyItems.cdc",
		}, {
			Name:     "KittyItemsMarket",
			Location: "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			Aliases: config.Aliases{{
				Network: "testnet",
				Address: flow.HexToAddress("0x123123123"),
			}},
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

	p, err := newProject(&cfg, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateSimpleProject() State {
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:     "NonFungibleToken",
			Location: "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
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
	p, err := newProject(&cfg, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateAliasesProject() State {
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:     "NonFungibleToken",
			Location: "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		}, {
			Name:     "FungibleToken",
			Location: "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			Aliases: config.Aliases{{
				Network: "emulator",
				Address: flow.HexToAddress("ee82856bf20e2aa6"),
			}},
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

	p, err := newProject(&cfg, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func generateAliasesComplexProject() State {
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           9000,
			ServiceAccount: "emulator-account",
		}},
		Contracts: config.Contracts{{
			Name:     "NonFungibleToken",
			Location: "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		}, {
			Name:     "FungibleToken",
			Location: "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			Aliases: config.Aliases{{
				Network: "emulator",
				Address: flow.HexToAddress("ee82856bf20e2aa6"),
			}},
		}, {
			Name:     "Kibble",
			Location: "../hungry-kitties/cadence/contracts/Kibble.cdc",
			Aliases: config.Aliases{{
				Network: "testnet",
				Address: flow.HexToAddress("ee82856bf20e2aa6"),
			}},
		}, {
			Name:     "Kibble",
			Location: "../hungry-kitties/cadence/contracts/Kibble.cdc",
			Aliases: config.Aliases{{
				Network: "emulator",
				Address: flow.HexToAddress("ee82856bf20e2aa6"),
			}},
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

	p, err := newProject(&cfg, composer, af)
	if err != nil {
		fmt.Println(err)
	}

	return *p
}

func Test_GetContractsByNameSimple(t *testing.T) {
	p := generateSimpleProject()
	path := filepath.FromSlash("../hungry-kitties/cadence/contracts/NonFungibleToken.cdc")
	af.WriteFile(path, []byte("pub contract{}"), os.ModePerm)

	contracts, err := p.DeploymentContractsByNetwork(config.EmulatorNetwork)
	require.NoError(t, err)
	account, err := p.conf.Accounts.ByName("emulator-account")
	require.NoError(t, err)
	require.Len(t, contracts, 1)
	assert.Equal(t, "NonFungibleToken", contracts[0].Name)
	assert.Equal(t, path, contracts[0].Location())
	assert.Equal(t, account.Address, contracts[0].AccountAddress)
}

func Test_EmulatorConfigSimple(t *testing.T) {
	p := generateSimpleProject()
	emulatorServiceAccount, _ := p.EmulatorServiceAccount()

	assert.Equal(t, "emulator-account", emulatorServiceAccount.Name)
	assert.Equal(t, emulatorServiceAccount.Key.ToConfig().PrivateKey, keys()[0])
	assert.Equal(t, flow.ServiceAddress("flow-emulator"), emulatorServiceAccount.Address)
}

func Test_AccountByAddressSimple(t *testing.T) {
	p := generateSimpleProject()
	acc, _ := p.Accounts().ByAddress(flow.ServiceAddress("flow-emulator"))

	assert.Equal(t, "emulator-account", acc.Name)
}

func Test_AccountByNameSimple(t *testing.T) {
	p := generateSimpleProject()
	acc, _ := p.Accounts().ByName("emulator-account")

	assert.Equal(t, flow.ServiceAddress("flow-emulator"), acc.Address)
	assert.Equal(t, acc.Key.ToConfig().PrivateKey, keys()[0])
}

func Test_HostSimple(t *testing.T) {
	p := generateSimpleProject()
	network, err := p.Networks().ByName("emulator")

	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1.3569", network.Host)
}

func Test_GetContractsByNameComplex(t *testing.T) {
	p := generateComplexProject()

	for _, c := range p.conf.Contracts {
		_ = af.WriteFile(c.Location, []byte("pub contract{}"), os.ModePerm)
	}

	contracts, err := p.DeploymentContractsByNetwork(config.EmulatorNetwork)
	require.NoError(t, err)
	require.Equal(t, 7, len(contracts))

	//sort contracts by name so tests are deterministic
	sort.Slice(contracts, func(i, j int) bool {
		return contracts[i].Name < contracts[j].Name
	})

	contractNames := funk.Map(contracts, func(c *project.Contract) string {
		return c.Name
	}).([]string)

	sources := funk.Map(contracts, func(c *project.Contract) string {
		return c.Location()
	}).([]string)

	targets := funk.Map(contracts, func(c *project.Contract) string {
		return c.AccountAddress.String()
	}).([]string)

	assert.Equal(t, contractNames[0], "FungibleToken")
	assert.Equal(t, contractNames[1], "Kibble")
	assert.Equal(t, contractNames[2], "KittyItems")
	assert.Equal(t, contractNames[3], "KittyItems")
	assert.Equal(t, contractNames[4], "KittyItemsMarket")
	assert.Equal(t, contractNames[5], "KittyItemsMarket")
	assert.Equal(t, contractNames[6], "NonFungibleToken")

	assert.Equal(t, sources[0], filepath.FromSlash("../hungry-kitties/cadence/contracts/FungibleToken.cdc"))
	assert.Equal(t, sources[1], filepath.FromSlash("cadence/kibble/contracts/Kibble.cdc"))
	assert.Equal(t, sources[2], filepath.FromSlash("cadence/kittyItems/contracts/KittyItems.cdc"))
	assert.Equal(t, sources[3], filepath.FromSlash("cadence/kittyItems/contracts/KittyItems.cdc"))
	assert.Equal(t, sources[4], filepath.FromSlash("cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"))
	assert.Equal(t, sources[5], filepath.FromSlash("cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"))
	assert.Equal(t, sources[6], filepath.FromSlash("../hungry-kitties/cadence/contracts/NonFungibleToken.cdc"))

	assert.Equal(t, targets[0], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[1], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[2], "f8d6e0586b0a20c7")
	assert.Equal(t, targets[3], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[4], "f8d6e0586b0a20c7")
	assert.Equal(t, targets[5], "f8d6e0586b0a20c1")
	assert.Equal(t, targets[6], "f8d6e0586b0a20c1")
}

func Test_EmulatorConfigComplex(t *testing.T) {
	p := generateComplexProject()
	emulatorServiceAccount, _ := p.EmulatorServiceAccount()

	assert.Equal(t, emulatorServiceAccount.Name, "emulator-account")
	assert.Equal(t, emulatorServiceAccount.Key.ToConfig().PrivateKey, keys()[0])
	assert.Equal(t, emulatorServiceAccount.Address, flow.ServiceAddress("flow-emulator"))
}

func Test_AccountByNameWithDuplicateAddress(t *testing.T) {
	p := generateComplexProject()
	acc1, err := p.Accounts().ByName("emulator-account")

	assert.NoError(t, err)
	acc2, err := p.Accounts().ByName("emulator-account-2")
	assert.NoError(t, err)

	assert.Equal(t, acc1.Name, "emulator-account")
	assert.Equal(t, acc2.Name, "emulator-account-2")
}

func Test_AccountByNameComplex(t *testing.T) {
	p := generateComplexProject()
	acc, _ := p.Accounts().ByName("account-2")

	assert.Equal(t, acc.Address.String(), "2c1162386b0a245f")
	assert.Equal(t, acc.Key.ToConfig().PrivateKey, keys()[1])
}

func Test_HostComplex(t *testing.T) {
	p := generateComplexProject()
	network, err := p.Networks().ByName("emulator")

	assert.NoError(t, err)

	assert.Equal(t, network.Host, "127.0.0.1.3569")
}

func Test_GetAliases(t *testing.T) {
	p := generateAliasesProject()

	aliases := p.AliasesForNetwork(config.EmulatorNetwork)
	contracts, _ := p.DeploymentContractsByNetwork(config.EmulatorNetwork)

	assert.Len(t, aliases, 2)
	assert.Equal(t, aliases[filepath.FromSlash("../hungry-kitties/cadence/contracts/FungibleToken.cdc")], "ee82856bf20e2aa6")
	assert.Len(t, contracts, 1)
	assert.Equal(t, contracts[0].Name, "NonFungibleToken")
}

func Test_GetAliasesComplex(t *testing.T) {
	p := generateAliasesComplexProject()

	aEmulator := p.AliasesForNetwork(config.EmulatorNetwork)
	cEmulator, _ := p.DeploymentContractsByNetwork(config.EmulatorNetwork)

	aTestnet := p.AliasesForNetwork(config.TestnetNetwork)
	cTestnet, _ := p.DeploymentContractsByNetwork(config.TestnetNetwork)

	assert.Len(t, cEmulator, 1)
	assert.Equal(t, cEmulator[0].Name, "NonFungibleToken")

	assert.Len(t, aEmulator, 4)
	assert.Equal(t, aEmulator[filepath.FromSlash("../hungry-kitties/cadence/contracts/FungibleToken.cdc")], "ee82856bf20e2aa6")
	assert.Equal(t, aEmulator[filepath.FromSlash("../hungry-kitties/cadence/contracts/Kibble.cdc")], "ee82856bf20e2aa6")

	assert.Len(t, aTestnet, 2)
	assert.Equal(t, aTestnet[filepath.FromSlash("../hungry-kitties/cadence/contracts/Kibble.cdc")], "ee82856bf20e2aa6")

	assert.Len(t, cTestnet, 2)
	assert.Equal(t, cTestnet[0].Name, "NonFungibleToken")
	assert.Equal(t, cTestnet[1].Name, "FungibleToken")
}

func Test_ChangingState(t *testing.T) {
	p := generateSimpleProject()

	em, err := p.EmulatorServiceAccount()
	assert.NoError(t, err)

	em.Name = "foo"
	em.Address = flow.HexToAddress("0x1")

	pk, _ := crypto.GeneratePrivateKey(
		crypto.ECDSA_P256,
		[]byte("seedseedseedseedseedseedseedseedseedseedseedseed"),
	)
	key := accounts.NewHexKeyFromPrivateKey(em.Key.Index(), em.Key.HashAlgo(), pk)
	em.Key = key

	foo, err := p.Accounts().ByName("foo")
	assert.NoError(t, err)
	assert.NotNil(t, foo)
	assert.Equal(t, foo.Name, "foo")
	assert.Equal(t, foo.Address, flow.HexToAddress("0x1"))

	pkey, err := foo.Key.PrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, (*pkey).String(), pk.String())

	bar, err := p.Accounts().ByName("foo")
	assert.NoError(t, err)
	bar.Name = "zoo"
	zoo, err := p.Accounts().ByName("zoo")
	assert.NotNil(t, zoo)
	assert.NoError(t, err)

	a := accounts.Account{}
	a.Name = "bobo"
	p.Accounts().AddOrUpdate(&a)
	bobo, err := p.Accounts().ByName("bobo")
	assert.NotNil(t, bobo)
	assert.NoError(t, err)

	zoo2, _ := p.Accounts().ByName("zoo")
	zoo2.Name = "emulator-account"
	assert.Equal(t, "emulator-account", zoo2.Name)

	pk, _ = crypto.GeneratePrivateKey(
		crypto.ECDSA_P256,
		[]byte("seedseedseedseedseedseedseedseedseed123seedseedseed123"),
	)

	p.SetEmulatorKey(pk)
	em, _ = p.EmulatorServiceAccount()
	pkey, err = em.Key.PrivateKey()
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
	assert.Equal(t, acc.Name, "emulator-account")
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
	assert.Equal(t, acc.Address.String(), "f8d6e0586b0a20c7")

	acc, err = state.Accounts().ByName("charlie")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address.String(), "e03daebed8ca0615")

	acc, err = state.Accounts().ByName("bob")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address.String(), "179b6b1cb6755e31")

	acc, err = state.Accounts().ByName("alice")
	assert.NoError(t, err)
	assert.Equal(t, acc.Address.String(), "179b6b1cb6755e31")
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

// ensures that default emulator values are in config when no emulator is defined in flow.json
func Test_DefaultEmulatorNotPresentInConfig(t *testing.T) {
	configJson := []byte(`{
		"contracts": {},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           3569,
			ServiceAccount: "emulator-account",
		}},
		Contracts:   config.Contracts{},
		Deployments: config.Deployments{},
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
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, err := Load(paths, af)
	assert.Equal(t, state.conf, &cfg)
	assert.NoError(t, err)
}

// ensures that default emulator values are not in config when no emulator is defined in flow.json
func Test_DefaultEmulatorWithoutEmulatorAccountInConfig(t *testing.T) {
	configJson := []byte(`{
		"contracts": {},
		"accounts": {
			"testnet-account": {
      			"address": "1e82856bf20e2aa6",
				"key": "388e3fbdc654b765942610679bb3a66b74212149ab9482187067ee116d9a8118"
    		}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, _ := Load(paths, af)
	assert.Equal(t, config.Emulators{}, state.conf.Emulators)
}

// ensures that default emulator values are in config when emulator is defined in flow.json
func Test_DefaultEmulatorWithEmulatorAccountInConfig(t *testing.T) {
	configJson := []byte(`{
		"contracts": {},
		"accounts": {
			"testnet-account": {
      			"address": "1e82856bf20e2aa6",
				"key": "388e3fbdc654b765942610679bb3a66b74212149ab9482187067ee116d9a8118"
    		},
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, _ := Load(paths, af)
	assert.Len(t, state.conf.Emulators, 1)
	assert.Equal(t, state.conf.Emulators, config.DefaultEmulators)
}

// backward compatibility test to ensure that default emulator values are still observed in flow.json
func Test_DefaultEmulatorPresentInConfig(t *testing.T) {
	configJson := []byte(`{
		"contracts": {},
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           3569,
			ServiceAccount: "emulator-account",
		}},
		Contracts:   config.Contracts{},
		Deployments: config.Deployments{},
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
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, err := Load(paths, af)
	assert.Equal(t, 1, len(state.conf.Emulators))
	assert.Equal(t, state.conf, &cfg)
	assert.NoError(t, err)
}

// ensures that custom emulator values are still observed in flow.json
func Test_CustomEmulatorValuesInConfig(t *testing.T) {
	configJson := []byte(`{
		"contracts": {},
		"emulators": {
			"custom-emulator": {
				"port": 2000,
				"serviceAccount": "emulator-account"
			}
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "custom-emulator",
			Port:           2000,
			ServiceAccount: "emulator-account",
		}},
		Contracts:   config.Contracts{},
		Deployments: config.Deployments{},
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
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, err := Load(paths, af)
	assert.Equal(t, "custom-emulator", state.conf.Emulators[0].Name)
	assert.Equal(t, 1, len(state.conf.Emulators))
	assert.Equal(t, state.conf, &cfg)
	assert.NoError(t, err)
}
