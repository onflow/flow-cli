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

/* ================================================================
Project Tests
================================================================ */
func Test_GetContractsByNameSimple(t *testing.T) {
	p := generateSimpleProject()

	contracts := p.GetContractsByNetwork("emulator")

	assert.Equal(t, 1, len(contracts))
	assert.Equal(t, "NonFungibleToken", contracts[0].Name)
	assert.Equal(t, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc", contracts[0].Source)
	assert.Equal(t, p.conf.Accounts.GetByName("emulator-account").Address, contracts[0].Target)
}

func Test_EmulatorConfigSimple(t *testing.T) {
	p := generateSimpleProject()
	emulatorServiceAccount := p.EmulatorServiceAccount()

	assert.Equal(t, "emulator-account", emulatorServiceAccount.Name)
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", emulatorServiceAccount.Keys[0].Context["privateKey"])
	assert.Equal(t, flow.ServiceAddress("flow-emulator"), emulatorServiceAccount.Address)
}

func Test_AccountByAddressSimple(t *testing.T) {
	p := generateSimpleProject()
	acc := p.GetAccountByAddress(flow.ServiceAddress("flow-emulator").String())

	assert.Equal(t, "emulator-account", acc.name)
}

func Test_AccountByNameSimple(t *testing.T) {
	p := generateSimpleProject()
	acc := p.GetAccountByName("emulator-account")

	assert.Equal(t, flow.ServiceAddress("flow-emulator").String(), acc.Address().String())
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", acc.DefaultKey().ToConfig().Context["privateKey"])
}

func Test_HostSimple(t *testing.T) {
	p := generateSimpleProject()
	host := p.Host("emulator")

	assert.Equal(t, "127.0.0.1.3569", host)
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

	assert.Equal(t, "FungibleToken", contractNames[0])
	assert.Equal(t, "Kibble", contractNames[1])
	assert.Equal(t, "KittyItems", contractNames[2])
	assert.Equal(t, "KittyItems", contractNames[3])
	assert.Equal(t, "KittyItemsMarket", contractNames[4])
	assert.Equal(t, "KittyItemsMarket", contractNames[5])
	assert.Equal(t, "NonFungibleToken", contractNames[6])

	assert.Equal(t, "../hungry-kitties/cadence/contracts/FungibleToken.cdc", sources[0])
	assert.Equal(t, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc", sources[1])
	assert.Equal(t, "cadence/kibble/contracts/Kibble.cdc", sources[2])
	assert.Equal(t, "cadence/kittyItems/contracts/KittyItems.cdc", sources[3])
	assert.Equal(t, "cadence/kittyItems/contracts/KittyItems.cdc", sources[4])
	assert.Equal(t, "cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", sources[5])
	assert.Equal(t, "cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", sources[6])

	assert.Equal(t, "f8d6e0586b0a20c1", targets[0])
	assert.Equal(t, "f8d6e0586b0a20c1", targets[1])
	assert.Equal(t, "f8d6e0586b0a20c1", targets[2])
	assert.Equal(t, "f8d6e0586b0a20c1", targets[3])
	assert.Equal(t, "f8d6e0586b0a20c1", targets[4])
	assert.Equal(t, "f8d6e0586b0a20c7", targets[5])
	assert.Equal(t, "f8d6e0586b0a20c7", targets[6])
}

func Test_EmulatorConfigComplex(t *testing.T) {
	p := generateComplexProject()
	emulatorServiceAccount := p.EmulatorServiceAccount()

	assert.Equal(t, "emulator-account", emulatorServiceAccount.Name)
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", emulatorServiceAccount.Keys[0].Context["privateKey"])
	assert.Equal(t, flow.ServiceAddress("flow-emulator"), emulatorServiceAccount.Address)
}

func Test_AccountByAddressComplex(t *testing.T) {
	p := generateComplexProject()
	acc1 := p.GetAccountByAddress("f8d6e0586b0a20c1")
	acc2 := p.GetAccountByAddress("0x2c1162386b0a245f")

	assert.Equal(t, "account-4", acc1.name)
	assert.Equal(t, "account-2", acc2.name)
}

func Test_AccountByNameComplex(t *testing.T) {
	p := generateComplexProject()
	acc := p.GetAccountByName("account-2")

	assert.Equal(t, "2c1162386b0a245f", acc.Address().String())
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", acc.DefaultKey().ToConfig().Context["privateKey"])
}

func Test_HostComplex(t *testing.T) {
	p := generateComplexProject()
	host := p.Host("emulator")

	assert.Equal(t, "127.0.0.1.3569", host)
}

func Test_ContractConflictComplex(t *testing.T) {
	p := generateComplexProject()
	exists := p.ContractConflictExists("emulator")
	notexists := p.ContractConflictExists("testnet")

	assert.True(t, exists)
	assert.False(t, notexists)

}
