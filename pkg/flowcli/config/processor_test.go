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

	test := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "$TEST"
			}
		}
	}`)

	result, _ := ProcessorRun(test)

	assert.JSONEq(t, `{
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"key": "123"
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
				"key": "$TEST"
			}
		}
	}`)

	result, _ := ProcessorRun(test)

	assert.JSONEq(t, `{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "123"
			}
		}
	}`, string(result))
}

func Test_PrivateConfigFileAccounts(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			},
			"admin-account": { "fromFile": "test.json" }
		}
	}`)

	preprocessor, accFromFile := ProcessorRun(b)

	assert.Equal(t, len(accFromFile), 1)

	assert.JSONEq(t, `{
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
			}
		}`, string(preprocessor))
}
