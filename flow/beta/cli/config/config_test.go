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

	assert.Equal(t, "service", account.Address)
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

	assert.Equal(t, "service", account.Address)
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
		"address": "service-1",
		"chain": "flow-emulator-1",
		"keys": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
	}`)

	var account Account
	json.Unmarshal(b, &account)

	assert.Equal(t, "service-1", account.Address)
	assert.Equal(t, "flow-emulator-1", account.ChainID.String())
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
			"address": "0x123123",
			"chain": "testnet",
			"keys": "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "emulator-account", accounts.Accounts["emulator-account"].Name)
	assert.Equal(t, "service-1", accounts.Accounts["emulator-account"].Address)
	assert.Equal(t, "flow-emulator", accounts.Accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "0x123123", accounts.Accounts["testnet-account"].Address)
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
			"address": "0x123",
			"chain": "testnet",
			"keys": [
				{
					"type": "hex",
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

	assert.Equal(t, "service", accounts.Accounts["emulator-account"].Address)
	assert.Equal(t, "flow-emulator", accounts.Accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "0x123", accounts.Accounts["testnet-account"].Address)
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
			"address": "0x123",
			"chain": "testnet",
			"keys": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "service", accounts.Accounts["emulator-account"].Address)
	assert.Equal(t, "flow-emulator", accounts.Accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts.Accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts.Accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.Accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.Accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.Accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "0x123", accounts.Accounts["testnet-account"].Address)
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
			"address": "0x123123",
			"chain": "testnet",
			"keys": "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var accounts AccountCollection
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "service-1", accounts.GetAccountByName("emulator-account").Address)
	assert.Equal(t, "emulator-account", accounts.GetAccountByName("emulator-account").Name)
}

/* ================================================================
Contracts Config Tests
================================================================ */

func Test_ConfigContractsSimple(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
      "testnet": "0x123123123",
      "emulator": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"
    }
  }`)

	var contracts map[string]string
	json.Unmarshal(b, &contracts)

	assert.Equal(t, "./cadence/kittyItems/contracts/KittyItems.cdc", contracts["KittyItems"])
	assert.Equal(t, `{"testnet": "0x123123123","emulator": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"}`, contracts["KittyItemsMarket"])

}

/* ================================================================
Deploy Config Tests
================================================================ */

func Test_ConfigDeploySimple(t *testing.T) {
	b := []byte(`{
		"deploy": {
			"testnet": {
				"account-2": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"]
			}, 
			"emulator": {
				"account-3": ["KittyItems", "KittyItemsMarket"],
				"account-4": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
			}
		}
	}`)

	config := new(Config)
	json.Unmarshal(b, &config)

	assert.Equal(t, "FungibleToken", config.Deploy["testnet"]["account-2"][0])

	networks := make([]string, 0)
	for k, _ := range config.Deploy {
		networks = append(networks, k)
	}

	assert.Equal(t, "testnet", networks[0])
	assert.Equal(t, "emulator", networks[1])

}
