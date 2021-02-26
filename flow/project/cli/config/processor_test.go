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

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_PrivateKeyEnv(t *testing.T) {
	os.Setenv("TEST", "123")

	test := `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "${env:TEST}",
				"chain": "flow-emulator"
			}
		}
	}`

	preprocessor := NewPreprocessor(new(afero.MemMapFs))
	result := preprocessor.Run(test)

	assert.JSONEq(t, `{
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"keys": "123",
					"chain": "flow-emulator"
				}
			}
		}`, result)
}

func Test_PrivateKeyEnvMultipleAccounts(t *testing.T) {
	os.Setenv("TEST", "123")
	os.Setenv("TEST2", "333")

	test := `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "${env:TEST}",
				"chain": "flow-emulator"
			}
		}
	}`

	preprocessor := NewPreprocessor(new(afero.MemMapFs))
	result := preprocessor.Run(test)

	assert.JSONEq(t, `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "123",
				"chain": "flow-emulator"
			}
		}
	}`, result)
}

func Test_PrivateConfigFileAccounts(t *testing.T) {
	b := `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
				"chain": "flow-emulator"
			},
			"admin-account": "${file:test.flow.json}"
		}
	}`

	f := []byte(`{
		"address": "f669cb8d41ce0c74",
		"keys": "17a616e230d38c04fb887dd83283a45f9a3082579db512c96eb84c5c562ac054",
		"chain": "flow-emulator"  
	}`)

	mockFS := afero.NewMemMapFs()
	afero.WriteFile(mockFS, "test.flow.json", f, 0644)

	preprocessor := NewPreprocessor(mockFS)
	result := preprocessor.Run(string(b))

	assert.JSONEq(t, `{
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
			}
		}`, result)
}

func Test_PrivateConfigFileAndEnvAccounts(t *testing.T) {
	b := `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
				"chain": "flow-emulator"
			},
			"admin-account": "${env:ADMIN_FILE}"
		}
	}`

	f := []byte(`{
		"address": "f669cb8d41ce0c74",
		"keys": "17a616e230d38c04fb887dd83283a45f9a3082579db512c96eb84c5c562ac054",
		"chain": "flow-emulator"  
	}`)

	os.Setenv("ADMIN_FILE", "${file:test.flow.json}")

	mockFS := afero.NewMemMapFs()
	afero.WriteFile(mockFS, "test.flow.json", f, 0644)

	preprocessor := NewPreprocessor(mockFS)
	result := preprocessor.Run(string(b))

	assert.JSONEq(t, `{
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
			}
		}`, result)
}
