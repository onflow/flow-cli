package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigAccountKeysSimple(t *testing.T) {
	b := []byte(`{
		"test": {
			"address": "service",
			"chain": "flow-emulator",
			"keys": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, "f8d6e0586b0a20c7", accounts.GetByName("test").Address.String())
	assert.Equal(t, "flow-emulator", accounts.GetByName("test").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("test").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("test").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("test").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("test").Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("test").Keys[0].Context["privateKey"])
}

func Test_ConfigAccountKeysAdvanced(t *testing.T) {
	b := []byte(`{
		"test": {
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
		}
	}`)

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()
	account := accounts.GetByName("test")

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
		"test": {
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
		}
	}`)

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()
	account := accounts.GetByName("test")

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

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, "emulator-account", accounts.GetByName("emulator-account").Name)
	assert.Equal(t, "0000000000000000", accounts.GetByName("emulator-account").Address.String())
	assert.Equal(t, "flow-emulator", accounts.GetByName("emulator-account").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("emulator-account").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("emulator-account").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("emulator-account").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("emulator-account").Keys[0].SigAlgo.String())
	assert.Equal(t, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("emulator-account").Keys[0].Context["privateKey"])

	assert.Equal(t, "2c1162386b0a245f", accounts.GetByName("testnet-account").Address.String())
	assert.Equal(t, "testnet", accounts.GetByName("testnet-account").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("testnet-account").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("testnet-account").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("testnet-account").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("testnet-account").Keys[0].SigAlgo.String())
	assert.Equal(t, "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("testnet-account").Keys[0].Context["privateKey"])
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

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, "f8d6e0586b0a20c7", accounts.GetByName("emulator-account").Address.String())
	assert.Equal(t, "flow-emulator", accounts.GetByName("emulator-account").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("emulator-account").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("emulator-account").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("emulator-account").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("emulator-account").Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("emulator-account").Keys[0].Context["privateKey"])

	assert.Equal(t, "1c1162386b0a245f", accounts.GetByName("testnet-account").Address.String())
	assert.Equal(t, "testnet", accounts.GetByName("testnet-account").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("testnet-account").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("testnet-account").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("testnet-account").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("testnet-account").Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("testnet-account").Keys[0].Context["privateKey"])
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

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, "f8d6e0586b0a20c7", accounts.GetByName("emulator-account").Address.String())
	assert.Equal(t, "flow-emulator", accounts.GetByName("emulator-account").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("emulator-account").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("emulator-account").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("emulator-account").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("emulator-account").Keys[0].SigAlgo.String())
	assert.Equal(t, "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("emulator-account").Keys[0].Context["privateKey"])

	assert.Equal(t, "3c1162386b0a245f", accounts.GetByName("testnet-account").Address.String())
	assert.Equal(t, "testnet", accounts.GetByName("testnet-account").ChainID.String())
	assert.Equal(t, 1, len(accounts.GetByName("testnet-account").Keys))
	assert.Equal(t, "SHA3_256", accounts.GetByName("testnet-account").Keys[0].HashAlgo.String())
	assert.Equal(t, 0, accounts.GetByName("testnet-account").Keys[0].Index)
	assert.Equal(t, "ECDSA_P256", accounts.GetByName("testnet-account").Keys[0].SigAlgo.String())
	assert.Equal(t, "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47", accounts.GetByName("testnet-account").Keys[0].Context["privateKey"])
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

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, "0000000000000000", accounts.GetByName("emulator-account").Address.String())
	assert.Equal(t, "emulator-account", accounts.GetByName("emulator-account").Name)
}

func Test_TransformDefaultAccountToJSON(t *testing.T) {
	b := []byte(`{"emulator-account":{"address":"f8d6e0586b0a20c7","chain":"flow-emulator","keys":[{"type":"hex","index":0,"signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256","context":{"privateKey":"1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}}]},"testnet-account":{"address":"3c1162386b0a245f","keys":"2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47","chain":"testnet"}}`)

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	j := jsonAccounts.transformToJSON(accounts)
	x, _ := json.Marshal(j)

	// our output format is shorted - improve test
	assert.NotEqual(t, string(b), string(x))
}

func Test_TransformAccountToJSON(t *testing.T) {
	b := []byte(`{"emulator-account":{"address":"f8d6e0586b0a20c7","chain":"flow-emulator","keys":[{"type":"hex","index":1,"signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256","context":{"privateKey":"1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}}]},"testnet-account":{"address":"3c1162386b0a245f","keys":"2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47","chain":"testnet"}}`)

	var jsonAccounts jsonAccounts
	json.Unmarshal(b, &jsonAccounts)

	accounts := jsonAccounts.transformToConfig()

	j := jsonAccounts.transformToJSON(accounts)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
