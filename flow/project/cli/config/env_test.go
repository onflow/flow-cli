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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PrivateKeyEnv(t *testing.T) {
	os.Setenv("TEST", "123")

	test := `{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		},
		"contracts": {},
		"networks": {
			"emulator": {
				"host": "127.0.0.1:3569",
				"chain": "flow-emulator"
			}
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "${env:TEST}",
				"chain": "flow-emulator"
			}
		},
		"deploys": {}
	}`

	result := ReplaceEnv(test)
	assert.JSONEq(t, `{
			"emulators": {
				"default": {
					"port": 3569,
					"serviceAccount": "emulator-account"
				}
			},
			"contracts": {},
			"networks": {
				"emulator": {
					"host": "127.0.0.1:3569",
					"chain": "flow-emulator"
				}
			},
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"keys": "123",
					"chain": "flow-emulator"
				}
			},
			"deploys": {}
		}`, result)
}

func Test_PrivateKeyEnvMultipleAccounts(t *testing.T) {
	os.Setenv("TEST", "123")
	os.Setenv("TEST2", "333")

	test := `{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		},
		"contracts": {},
		"networks": {
			"emulator": {
				"host": "127.0.0.1:3569",
				"chain": "flow-emulator"
			}
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "${env:TEST}",
				"chain": "flow-emulator"
			},
			"emulator-account-2": {
				"address": "f8d6e0586b0a20c7",
				"keys": "${env:TEST2}",
				"chain": "flow-emulator"
			}
		},
		"deploys": {}
	}`

	result := ReplaceEnv(test)
	assert.JSONEq(t, `{
			"emulators": {
				"default": {
					"port": 3569,
					"serviceAccount": "emulator-account"
				}
			},
			"contracts": {},
			"networks": {
				"emulator": {
					"host": "127.0.0.1:3569",
					"chain": "flow-emulator"
				}
			},
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"keys": "123",
					"chain": "flow-emulator"
				}, 
				"emulator-account-2": {
					"address": "f8d6e0586b0a20c7",
					"keys": "333",
					"chain": "flow-emulator"
				}
			},
			"deploys": {}
		}`, result)
}
