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
	var b = []byte(`{
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
	var b = []byte(`{
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
	var b = []byte(`{
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
	var b = []byte(`{
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

	var accounts map[string]Account
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "service-1", accounts["emulator-account"].Address)
	assert.Equal(t, "flow-emulator", accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "0x123123", accounts["testnet-account"].Address)
	assert.Equal(t, "testnet", accounts["testnet-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts["testnet-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts["testnet-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts["testnet-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts["testnet-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts["testnet-account"].Keys[0].Context["privateKey"])
}

func Test_ConfigMultipleAccountsAdvanced(t *testing.T) {
	var b = []byte(`{
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

	var accounts map[string]Account
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "service", accounts["emulator-account"].Address)
	assert.Equal(t, "flow-emulator", accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "0x123", accounts["testnet-account"].Address)
	assert.Equal(t, "testnet", accounts["testnet-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts["testnet-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts["testnet-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts["testnet-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts["testnet-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts["testnet-account"].Keys[0].Context["privateKey"])
}

func Test_ConfigMixedAccounts(t *testing.T) {
	var b = []byte(`{
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

	var accounts map[string]Account
	json.Unmarshal(b, &accounts)

	assert.Equal(t, "service", accounts["emulator-account"].Address)
	assert.Equal(t, "flow-emulator", accounts["emulator-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts["emulator-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts["emulator-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts["emulator-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts["emulator-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts["emulator-account"].Keys[0].Context["privateKey"])

	assert.Equal(t, "0x123", accounts["testnet-account"].Address)
	assert.Equal(t, "testnet", accounts["testnet-account"].ChainID.String())
	assert.Equal(t, 1, len(accounts["testnet-account"].Keys))
	assert.Equal(t, "SHA3_256", accounts["testnet-account"].Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts["testnet-account"].Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts["testnet-account"].Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts["testnet-account"].Keys[0].Context["privateKey"])
}

/* ================================================================
Config Tests
================================================================ */
