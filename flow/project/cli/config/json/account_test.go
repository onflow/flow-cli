/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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
package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, accounts.GetByName("test").Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, accounts.GetByName("test").ChainID.String(), "flow-emulator")
	assert.Len(t, accounts.GetByName("test").Keys, 1)
	assert.Equal(t, accounts.GetByName("test").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("test").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("test").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("test").Keys[0].Context["privateKey"], "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()
	account := accounts.GetByName("test")

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, account.ChainID.String(), "flow-emulator")
	assert.Len(t, account.Keys, 1)
	assert.Equal(t, account.Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, account.Keys[0].Index, 0)
	assert.Equal(t, account.Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, account.Keys[0].Context["privateKey"], "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()
	account := accounts.GetByName("test")

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, account.ChainID.String(), "flow-emulator")
	assert.Len(t, account.Keys, 2)

	assert.Equal(t, account.Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, account.Keys[0].Index, 0)
	assert.Equal(t, account.Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, account.Keys[0].Context["privateKey"], "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	assert.Equal(t, account.Keys[1].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, account.Keys[1].Index, 1)
	assert.Equal(t, account.Keys[1].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, account.Keys[1].Context["privateKey"], "2372967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, accounts.GetByName("emulator-account").Name, "emulator-account")
	assert.Equal(t, accounts.GetByName("emulator-account").Address.String(), "0000000000000000")
	assert.Equal(t, accounts.GetByName("emulator-account").ChainID.String(), "flow-emulator")
	assert.Len(t, accounts.GetByName("emulator-account").Keys, 1)
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].Context["privateKey"], "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	assert.Equal(t, accounts.GetByName("testnet-account").Address.String(), "2c1162386b0a245f")
	assert.Equal(t, accounts.GetByName("testnet-account").ChainID.String(), "testnet")
	assert.Len(t, accounts.GetByName("testnet-account").Keys, 1)
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].Context["privateKey"], "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, accounts.GetByName("emulator-account").Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, accounts.GetByName("emulator-account").ChainID.String(), "flow-emulator")
	assert.Len(t, accounts.GetByName("emulator-account").Keys, 1)
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].Context["privateKey"], "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	assert.Equal(t, accounts.GetByName("testnet-account").Address.String(), "1c1162386b0a245f")
	assert.Equal(t, accounts.GetByName("testnet-account").ChainID.String(), "testnet")
	assert.Len(t, accounts.GetByName("testnet-account").Keys, 1)
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].Context["privateKey"], "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, accounts.GetByName("emulator-account").Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, accounts.GetByName("emulator-account").ChainID.String(), "flow-emulator")
	assert.Len(t, accounts.GetByName("emulator-account").Keys, 1)
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("emulator-account").Keys[0].Context["privateKey"], "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	assert.Equal(t, accounts.GetByName("testnet-account").Address.String(), "3c1162386b0a245f")
	assert.Equal(t, accounts.GetByName("testnet-account").ChainID.String(), "testnet")
	assert.Len(t, accounts.GetByName("testnet-account").Keys, 1)
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].HashAlgo.String(), "SHA3_256")
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].Index, 0)
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, accounts.GetByName("testnet-account").Keys[0].Context["privateKey"], "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
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
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	assert.Equal(t, accounts.GetByName("emulator-account").Address.String(), "0000000000000000")
	assert.Equal(t, accounts.GetByName("emulator-account").Name, "emulator-account")
}

func Test_TransformDefaultAccountToJSON(t *testing.T) {
	b := []byte(`{"emulator-account":{"address":"f8d6e0586b0a20c7","chain":"flow-emulator","keys":[{"type":"hex","index":0,"signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256","context":{"privateKey":"1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}}]},"testnet-account":{"address":"3c1162386b0a245f","keys":"2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47","chain":"testnet"}}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	j := transformAccountsToJSON(accounts)
	x, _ := json.Marshal(j)

	// our output format is shorted - improve test
	assert.NotEqual(t, string(b), string(x))
}

func Test_TransformAccountToJSON(t *testing.T) {
	b := []byte(`{"emulator-account":{"address":"f8d6e0586b0a20c7","chain":"flow-emulator","keys":[{"type":"hex","index":1,"signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256","context":{"privateKey":"1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}}]},"testnet-account":{"address":"3c1162386b0a245f","keys":"2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47","chain":"testnet"}}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	require.NoError(t, err)

	accounts := jsonAccounts.transformToConfig()

	j := transformAccountsToJSON(accounts)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}
