package cli

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
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
				"account-1": ["NonFungibleToken"]
			}
		},
		"accounts": {
			"account-1": {
				"address": "service"
			}
		}
	}`)

	config := new(config.Config)
	json.Unmarshal(b, &config)

	p, _ := newProject(config)
	contracts := p.GetContractsByNetwork("emulator")

	assert.Equal(t, 1, len(contracts))
	assert.Equal(t, "NonFungibleToken", contracts[0].Name)
	assert.Equal(t, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc", contracts[0].Source)
	assert.Equal(t, p.conf.Accounts.GetByName("account-1").Address, contracts[0].Target)
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
				"account-3": ["KittyItems", "KittyItemsMarket"],
				"account-4": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
			}
		},

		"accounts": {
			"account-2": {
				"address": "0x2c1162386b0a245f",
				"keys": "22232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			},
			"account-3": {
				"address": "service",
				"keys": "service"
			},
			"account-4": {
				"address": "0xf8d6e0586b0a20c1",
				"keys": "4442967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		}
	}`)

	config := new(config.Config)
	json.Unmarshal(b, &config)

	p, _ := newProject(config)
	contracts := p.GetContractsByNetwork("emulator")

	assert.Equal(t, 7, len(contracts))

	//sort names so tests are deterministic
	contractNames := funk.Map(contracts, func(c Contract) string {
		return c.Name
	}).([]string)
	sort.Strings(contractNames)

	assert.Equal(t, "FungibleToken", contractNames[0])
	assert.Equal(t, "Kibble", contractNames[1])
	assert.Equal(t, "KittyItems", contractNames[2])
	assert.Equal(t, "KittyItems", contractNames[3])
	assert.Equal(t, "KittyItemsMarket", contractNames[4])
	assert.Equal(t, "KittyItemsMarket", contractNames[5])
	assert.Equal(t, "NonFungibleToken", contractNames[6])

	assert.Equal(t, "../cadence/kittyItems/contracts/KittyItems.cdc", contracts[0].Source)
	assert.Equal(t, "../cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts[1].Source)
	assert.Equal(t, "../hungry-kitties/cadence/contracts/FungibleToken.cdc", contracts[2].Source)
	assert.Equal(t, "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc", contracts[3].Source)
	assert.Equal(t, "../cadence/kibble/contracts/Kibble.cdc", contracts[4].Source)
	assert.Equal(t, "../cadence/kittyItems/contracts/KittyItems.cdc", contracts[5].Source)
	assert.Equal(t, "../cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts[6].Source)

	assert.Equal(t, "f8d6e0586b0a20c7", contracts[0].Target.String())
	assert.Equal(t, "f8d6e0586b0a20c7", contracts[1].Target.String())
	assert.Equal(t, "f8d6e0586b0a20c1", contracts[2].Target.String())
	assert.Equal(t, "f8d6e0586b0a20c1", contracts[3].Target.String())
	assert.Equal(t, "f8d6e0586b0a20c1", contracts[4].Target.String())
	assert.Equal(t, "f8d6e0586b0a20c1", contracts[5].Target.String())
	assert.Equal(t, "f8d6e0586b0a20c1", contracts[6].Target.String())
}
