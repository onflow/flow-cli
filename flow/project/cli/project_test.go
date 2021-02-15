package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/onflow/flow-cli/flow/project/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

/* ================================================================
Project Tests
================================================================ */
func Test_GetContractsByNameSimple(t *testing.T) {
	b := []byte(`{
		"contracts": {
			"NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc"
		},
		"deploy": { 
			"emulator": {
				"emulator-service": ["NonFungibleToken"]
			}
		},
		"accounts": {
			"emulator-service": {
				"address": "service",
				"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		}
	}`)

	config := new(config.Config)
	json.Unmarshal(b, &config)

	p, err := newProject(config)
	if err != nil {
		fmt.Println(err)
	}

	contracts := p.GetContractsByNetwork("emulator")

	assert.Equal(t, 1, len(contracts))
	assert.Equal(t, "NonFungibleToken", contracts[0].Name)
	assert.Equal(t, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc", contracts[0].Source)
	assert.Equal(t, p.conf.Accounts.GetByName("emulator-service").Address, contracts[0].Target)
}

func Test_GetContractsByNameComplex(t *testing.T) {
	b := []byte(`{
		"contracts": {
			"NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
			"FungibleToken": "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			"Kibble": "./cadence/kibble/contracts/Kibble.cdc",
			"KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
			"KittyItemsMarket": {
				"testnet": "0x123123123",
				"emulator": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"
			}
		},

		"deploy": {
			"testnet": {
				"account-2": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"]
			}, 
			"emulator": {
				"emulator-service": ["KittyItems", "KittyItemsMarket"],
				"account-4": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
			}
		},

		"accounts": {
			"account-2": {
				"address": "0x2c1162386b0a245f",
				"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			},
			"emulator-service": {
				"address": "service",
				"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			},
			"account-4": {
				"address": "0xf8d6e0586b0a20c1",
				"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		}
	}`)

	config := new(config.Config)
	json.Unmarshal(b, &config)

	p, err := newProject(config)
	if err != nil {
		fmt.Println(err)
	}

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
