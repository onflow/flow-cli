package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* ================================================================
Accounts Config Tests
================================================================ */
func Test_ConfigAccountKeysAdvanced(t *testing.T) {
	b := []byte(`{
		"address": "service",
		"chain": "flow-emulator",
		"keys": [
			{
				"type": "hex",
				"index": 0,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
				}
			}
		]
	}`)

	var account Account
	json.Unmarshal(b, &account)

	assert.Equal(t, "f8d6e0586b0a20c7", account.Address.String())
	assert.Equal(t, "flow-emulator", account.ChainID.String())
	assert.Equal(t, 1, len(account.Keys))
	assert.Equal(t, "SHA3_256", account.Keys[0].HashAlgo.String())
	assert.Equal(t, 0, account.Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", account.Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", account.Keys[0].Context["privateKey"])
}

func Test_ConfigAccountKeysAdvancedMultiple(t *testing.T) {
	b := []byte(`{
		"address": "service",
		"chain": "flow-emulator",
		"keys": [
			{
				"type": "hex",
				"index": 0,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
				}
			},
			{
				"type": "hex",
				"index": 1,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "2372967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
				}
			}
		]
	}`)

	var account Account
	json.Unmarshal(b, &account)

	assert.Equal(t, "f8d6e0586b0a20c7", account.Address.String())
	assert.Equal(t, "flow-emulator", account.ChainID.String())
	assert.Equal(t, 2, len(account.Keys))

	assert.Equal(t, "SHA3_256", account.Keys[0].HashAlgo.String())
	assert.Equal(t, 0, account.Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", account.Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", account.Keys[0].Context["privateKey"])

	assert.Equal(t, "SHA3_256", account.Keys[1].HashAlgo.String())
	assert.Equal(t, 1, account.Keys[1].Index)
	assert.Equal(t, "ECDSA_P256", account.Keys[1].SigAlgo.String())
	assert.Equal(t, "2372967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", account.Keys[1].Context["privateKey"])
}

func Test_ConfigAccountKeysSimple(t *testing.T) {
	b := []byte(`{
		"address": "service",
		"chain": "flow-emulator",
		"keys": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
	}`)

	var account Account
	json.Unmarshal(b, &account)

	assert.Equal(t, "f8d6e0586b0a20c7", account.Address.String())
	assert.Equal(t, "flow-emulator", account.ChainID.String())
	assert.Equal(t, 1, len(account.Keys))
	assert.Equal(t, "SHA3_256", account.Keys[0].HashAlgo.String())
	assert.Equal(t, 0, account.Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", account.Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", account.Keys[0].Context["privateKey"])
}

func Test_ConfigMultipleAccountsSimple(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service-1",
			"chain": "flow-emulator",
			"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		},
		"testnet-account": {
			"address": "0x2c1162386b0a245f",
			"chain": "testnet",
			"keys": "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "emulator-account", accounts.Accounts["emulator-account"].Name)
	assert.Equal(t, "0000000000000000", accounts.Accounts["emulator-account"].Address.String())
	assert.Equal(t, "flow-emulator", accounts.Accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "2c1162386b0a245f", accounts.Accounts["testnet-account"].Address.String())
	assert.Equal(t, "testnet", accounts.Accounts["testnet-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["testnet-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["testnet-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["testnet-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["testnet-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["testnet-account"].Keys[0].Context["privateKey"])
}

func Test_ConfigMultipleAccountsAdvanced(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service",
			"chain": "flow-emulator",
			"keys": [
				{
					"type": "hex",
					"index": 0,
					"signatureAlgorithm": "ECDSA_P256",
					"hashAlgorithm": "SHA3_256",
					"context": {
						"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
					}
				}
			]
		},
		"testnet-account": {
			"address": "1c1162386b0a245f",
			"chain": "testnet",
			"keys": [
				{
					"type": "0x18d6e0586b0a20c7",
					"index": 0,
					"signatureAlgorithm": "ECDSA_P256",
					"hashAlgorithm": "SHA3_256",
					"context": {
						"privateKey": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
					}
				}
			]
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "f8d6e0586b0a20c7", accounts.Accounts["emulator-account"].Address.String())
	assert.Equal(t, "flow-emulator", accounts.Accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "1c1162386b0a245f", accounts.Accounts["testnet-account"].Address.String())
	assert.Equal(t, "testnet", accounts.Accounts["testnet-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["testnet-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["testnet-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["testnet-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["testnet-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["testnet-account"].Keys[0].Context["privateKey"])
}

func Test_ConfigMixedAccounts(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service",
			"chain": "flow-emulator",
			"keys": [
				{
					"type": "hex",
					"index": 0,
					"signatureAlgorithm": "ECDSA_P256",
					"hashAlgorithm": "SHA3_256",
					"context": {
						"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
					}
				}
			]
		},
		"testnet-account": {
			"address": "3c1162386b0a245f",
			"chain": "testnet",
			"keys": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "f8d6e0586b0a20c7", accounts.Accounts["emulator-account"].Address.String())
	assert.Equal(t, "flow-emulator", accounts.Accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "3c1162386b0a245f", accounts.Accounts["testnet-account"].Address.String())
	assert.Equal(t, "testnet", accounts.Accounts["testnet-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["testnet-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["testnet-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["testnet-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["testnet-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["testnet-account"].Keys[0].Context["privateKey"])
}

func Test_ConfigAccountsMap(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service-1",
			"chain": "flow-emulator",
			"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		},
		"testnet-account": {
			"address": "3c1162386b0a245f",
			"chain": "testnet",
			"keys": "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "0000000000000000", accounts.GetByName("emulator-account").Address.String())
	assert.Equal(t, "emulator-account", accounts.GetByName("emulator-account").Name)
}

/* ================================================================
Contracts Config Tests
================================================================ */

func Test_ConfigContractsSimple(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": "./cadence/kittyItems/contracts/KittyItemsMarket.cdc"
  }`)

	var contracts ContractCollection
	json.Unmarshal(b, &contracts)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItemsMarket.cdc", contracts.GetByName("KittyItemsMarket").Source)
}

func Test_ConfigContractsComplex(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
      "testnet": "0x123123123",
      "emulator": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"
    }
  }`)

	var contracts ContractCollection
	json.Unmarshal(b, &contracts)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByName("KittyItems").Source)
	assert.Equal(t, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts.GetByNameAndNetwork("KittyItemsMarket", "emulator").Source)
	assert.Equal(t, "0x123123123", contracts.GetByNameAndNetwork("KittyItemsMarket", "testnet").Source)

	assert.Equal(t, 2, len(contracts.GetByNetwork("testnet")))
	assert.Equal(t, 2, len(contracts.GetByNetwork("emulator")))

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByNetwork("testnet")[0].Source)
	assert.Equal(t, "0x123123123", contracts.GetByNetwork("testnet")[1].Source)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts.GetByNetwork("emulator")[0].Source)
	assert.Equal(t, "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc", contracts.GetByNetwork("emulator")[1].Source)
}

/* ================================================================
Deploy Config Tests
================================================================ */
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

	deploy := new(DeployCollection)
	json.Unmarshal(b, &deploy)

	assert.Equal(t, "account-2", deploy.GetByNetwork("testnet")[0].Account)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"}, deploy.GetByNetwork("testnet")[0].Contracts)

	assert.Equal(t, 2, len(deploy.GetByNetwork("emulator")))
	assert.Equal(t, "account-3", deploy.GetByNetwork("emulator")[0].Account)
	assert.Equal(t, "account-4", deploy.GetByNetwork("emulator")[1].Account)
	assert.Equal(t, []string{"KittyItems", "KittyItemsMarket"}, deploy.GetByNetwork("emulator")[0].Contracts)
	assert.Equal(t, []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"}, deploy.GetByNetwork("emulator")[1].Contracts)
}
