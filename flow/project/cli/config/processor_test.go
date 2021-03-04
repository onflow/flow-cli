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

package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_PrivateKeyEnv(t *testing.T) {
	os.Setenv("TEST", "123")

	test := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "$TEST",
				"chain": "flow-emulator"
			}
		}
	}`)

	preprocessor := NewPreprocessor(NewLoader(afero.NewMemMapFs()))
	result := preprocessor.Run(test)

	assert.JSONEq(t, `{
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"keys": "123",
					"chain": "flow-emulator"
				}
			}
		}`, string(result))
}

func Test_PrivateKeyEnvMultipleAccounts(t *testing.T) {
	os.Setenv("TEST", "123")
	os.Setenv("TEST2", "333")

	test := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "$TEST",
				"chain": "flow-emulator"
			}
		}
	}`)

	preprocessor := NewPreprocessor(NewLoader(afero.NewMemMapFs()))
	result := preprocessor.Run(test)

	assert.JSONEq(t, `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "123",
				"chain": "flow-emulator"
			}
		}
	}`, string(result))
}

func Test_PrivateConfigFileAccounts(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
				"chain": "flow-emulator"
			},
			"admin-account": { "fromFile": "test.json" },
			"admin-account": { "fromFile": "test.json" },

			"admin-account":{ 
				"fromFile": "test.json" 
			}
		}
	}`)

	preprocessor := NewPreprocessor(NewLoader(afero.NewMemMapFs()))
	result := preprocessor.Run(b)

	assert.JSONEq(t, `{
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"keys": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
					"chain": "flow-emulator"
				}
			}
		}`, string(result))
}
