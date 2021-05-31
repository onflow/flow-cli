/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
)

func Test_ConfigAccountKeysSimple(t *testing.T) {
	b := []byte(`{
		"test": {
			"address": "service",
			"key": "0x1fae488ce86422698f1c13468b137d62de488e7e978d7090396f7883a60abdcf"
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("test")
	key := account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x1fae488ce86422698f1c13468b137d62de488e7e978d7090396f7883a60abdcf")
}

func Test_ConfigAccountKeysAdvancedHex(t *testing.T) {
	b := []byte(`{
		"test": {
			"address": "service",
			"key": {
				"type": "hex",
				"index": 1,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA2_256",
				"privateKey": "271cec6bb5221d12713759188166bdfa00079db5789c36b54dcf1d794d8d8cdf"
			}
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("test")
	key := account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA2_256")
	assert.Equal(t, key.Index, 1)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x271cec6bb5221d12713759188166bdfa00079db5789c36b54dcf1d794d8d8cdf")
	assert.Equal(t, key.ResourceID, "")
}

func Test_ConfigAccountKeysAdvancedKMS(t *testing.T) {
	b := []byte(`{
		"test": {
			"address": "service",
			"key": {
				"type": "kms",
				"index": 1,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"resourceID": "projects/flow/locations/us/keyRings/foo/bar/cryptoKeyVersions/1"
			}
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("test")
	key := account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 1)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.ResourceID, "projects/flow/locations/us/keyRings/foo/bar/cryptoKeyVersions/1")
	assert.Nil(t, key.PrivateKey)
}

func Test_ConfigAccountOldFormats(t *testing.T) {
	b := []byte(`{
		"old-format-1": {
			"address": "service",
			"key": {
				"type": "hex",
				"index": 0,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA2_256",
				"context": {
					"privateKey": "f988fd7a959d96d0e36ca13a240bbfc4a78098cc56cfd1fa6c918080c8a0f55c"
				}
			}
		},
		"old-format-2": {
			"address": "service",
			"keys": [{
				"type": "hex",
				"index": 1,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "271cec6bb5221d12713759188166bdfa00079db5789c36b54dcf1d794d8d8cdf"
				}
			}]
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("old-format-1")
	key := account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA2_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0xf988fd7a959d96d0e36ca13a240bbfc4a78098cc56cfd1fa6c918080c8a0f55c")

	account = accounts.GetByName("old-format-2")
	key = account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 1)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x271cec6bb5221d12713759188166bdfa00079db5789c36b54dcf1d794d8d8cdf")
}

func Test_ConfigMultipleAccountsSimple(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service-1",
			"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		},
		"testnet-account": {
			"address": "0x2c1162386b0a245f",
			"key": "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("emulator-account")
	key := account.Key

	assert.Equal(t, account.Name, "emulator-account")
	assert.Equal(t, account.Address.String(), "0000000000000000")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0xdd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	account = accounts.GetByName("testnet-account")
	key = account.Key

	assert.Equal(t, account.Address.String(), "2c1162386b0a245f")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_ConfigMultipleAccountsAdvanced(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service",
			"key": {
				"type": "hex",
				"index": 0,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
				}
			}
		},
		"testnet-account": {
			"address": "1c1162386b0a245f",
			"key": {
				"type": "0x18d6e0586b0a20c7",
				"index": 0,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
				}
			}
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("emulator-account")
	key := account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	account = accounts.GetByName("testnet-account")
	key = account.Key

	assert.Equal(t, account.Address.String(), "1c1162386b0a245f")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_ConfigMixedAccounts(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service",
			"key": {
				"type": "hex",
				"index": 0,
				"signatureAlgorithm": "ECDSA_P256",
				"hashAlgorithm": "SHA3_256",
				"context": {
					"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
				}
			}
		},
		"testnet-account": {
			"address": "3c1162386b0a245f",
			"key": "2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	account := accounts.GetByName("emulator-account")
	key := account.Key

	assert.Equal(t, account.Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	account = accounts.GetByName("testnet-account")
	key = account.Key

	assert.Equal(t, account.Address.String(), "3c1162386b0a245f")
	assert.Equal(t, key.HashAlgo.String(), "SHA3_256")
	assert.Equal(t, key.Index, 0)
	assert.Equal(t, key.SigAlgo.String(), "ECDSA_P256")
	assert.Equal(t, key.PrivateKey.String(), "0x2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}

func Test_ConfigAccountsMap(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service-1",
			"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		},
		"testnet-account": {
			"address": "3c1162386b0a245f",
			"key": "1232967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	assert.Equal(t, accounts.GetByName("emulator-account").Address.String(), "0000000000000000")
	assert.Equal(t, accounts.GetByName("emulator-account").Name, "emulator-account")
}

func Test_TransformDefaultAccountToJSON(t *testing.T) {
	b := []byte(`{"emulator-account":{"address":"f8d6e0586b0a20c7","key":{"type":"hex","index":0,"signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256","privateKey":"1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}},"testnet-account":{"address":"3c1162386b0a245f","key":"2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	j := transformAccountsToJSON(accounts)
	x, _ := json.Marshal(j)

	// our output format is shorted - improve test
	assert.NotEqual(t, string(b), string(x))
}

func Test_TransformAccountToJSON(t *testing.T) {
	b := []byte(`{"emulator-account":{"address":"f8d6e0586b0a20c7","key":{"type":"hex","index":1,"signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256","privateKey":"1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}},"testnet-account":{"address":"3c1162386b0a245f","key":"2272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"}}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	accounts, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	j := transformAccountsToJSON(accounts)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}

func Test_SupportForOldFormatWithMultipleKeys(t *testing.T) {
	b := []byte(`{
		"emulator-account": {
			"address": "service-1",
			"chain": "flow-emulator",
			"keys": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
		},
		"testnet-account": {
			"address": "3c1162386b0a245f",
			"chain": "testnet",
			"keys": [
				{
					"type": "hex",
					"index": 0,
					"signatureAlgorithm": "ECDSA_P256",
					"hashAlgorithm": "SHA3_256",
					"context": {
						"privateKey": "1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
					}
				},{
					"type": "hex",
					"index": 0,
					"signatureAlgorithm": "ECDSA_P256",
					"hashAlgorithm": "SHA3_256",
					"context": {
						"privateKey": "2332967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b44"
					}
				}
			]
		}
	}`)

	var jsonAccounts jsonAccounts
	err := json.Unmarshal(b, &jsonAccounts)
	assert.NoError(t, err)

	conf, err := jsonAccounts.transformToConfig()
	assert.NoError(t, err)

	key := conf.GetByName("testnet-account").Key
	assert.Equal(t, key.PrivateKey.String(), "0x1272967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")

	key = conf.GetByName("emulator-account").Key
	assert.Equal(t, key.PrivateKey.String(), "0xdd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
}
