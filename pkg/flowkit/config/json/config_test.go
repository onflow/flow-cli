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
package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SimpleJSONConfig(t *testing.T) {
	b := []byte(`{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		},
		"contracts": {},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"deployments": {}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "emulator-account", conf.Accounts[0].Name)
	assert.Equal(t, "0x11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Key.PrivateKey.String())
	network, err := conf.Networks.ByName("emulator")
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:3569", network.Host)
}

func Test_NonExistingContractForDeployment(t *testing.T) {
	b := []byte(`{
		"contracts": {},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"deployments": {
			"emulator": {
				"emulator-account": ["FungibleToken"]
			}
		}
	}`)

	parser := NewParser()
	config, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = config.Validate()
	assert.Equal(t, err.Error(), "deployment contains nonexisting contract FungibleToken")
}

func Test_NonExistingAccountForDeployment(t *testing.T) {
	b := []byte(`{
		"contracts": {
			"FungibleToken": "./test.cdc"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"deployments": {
			"emulator": {
				"test-1": ["FungibleToken"]
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, err.Error(), "deployment contains nonexisting account test-1")
}

func Test_NonExistingNetworkForDeployment(t *testing.T) {
	b := []byte(`{
		"contracts": {
			"FungibleToken": "./test.cdc"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"networks": {},
		"deployments": {
			"foo": {
				"test-1": ["FungibleToken"]
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, err.Error(), "deployment contains nonexisting network foo")
}

func Test_NonExistingAccountForEmulator(t *testing.T) {
	b := []byte(`{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, err.Error(), "emulator default contains nonexisting service account emulator-account")
}
