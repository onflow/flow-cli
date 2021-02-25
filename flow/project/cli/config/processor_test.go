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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	result := PreProcess(test)
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

	result := PreProcess(test)
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

func Test_PrivateConfigFileAccounts(t *testing.T) {
	b := `{
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
				"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
				"chain": "flow-emulator"
			},
			"admin-account": "${file:flow.admin.json}"
		},
		"deploys": {}
	}`

	f := []byte(`{
		"address": "f669cb8d41ce0c74",
		"keys": "17a616e230d38c04fb887dd83283a45f9a3082579db512c96eb84c5c562ac054",
		"chain": "flow-emulator"  
	}`)

	err := ioutil.WriteFile("flow.admin.json", f, os.ModePerm)
	require.NoError(t, err)

	result := PreProcess(b)
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
					"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
					"chain": "flow-emulator"
				}, 
				"admin-account": {
					"address": "f669cb8d41ce0c74",
					"keys": "17a616e230d38c04fb887dd83283a45f9a3082579db512c96eb84c5c562ac054",
					"chain": "flow-emulator" 
				}
			},
			"deploys": {}
		}`, result)

	os.Remove("flow.admin.json")
}

func Test_PrivateConfigFileAndEnvAccounts(t *testing.T) {
	b := `{
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
				"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
				"chain": "flow-emulator"
			},
			"admin-account": "${env:ADMIN_FILE}"
		},
		"deploys": {}
	}`

	f := []byte(`{
		"address": "f669cb8d41ce0c74",
		"keys": "17a616e230d38c04fb887dd83283a45f9a3082579db512c96eb84c5c562ac054",
		"chain": "flow-emulator"  
	}`)

	// REF: don't create files but mock it
	os.Setenv("ADMIN_FILE", "${file:test.flow.json}")
	err := ioutil.WriteFile("test.flow.json", f, os.ModePerm)
	defer os.Remove("test.flow.json")
	require.NoError(t, err)

	result := PreProcess(b)
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
					"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
					"chain": "flow-emulator"
				}, 
				"admin-account": {
					"address": "f669cb8d41ce0c74",
					"keys": "17a616e230d38c04fb887dd83283a45f9a3082579db512c96eb84c5c562ac054",
					"chain": "flow-emulator" 
				}
			},
			"deploys": {}
		}`, result)
}
