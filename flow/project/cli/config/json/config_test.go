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
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			}
		},
		"deploys": {}
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "test1-flow.json", b, 0644)

	require.NoError(t, err)

	conf, loadErr := Load("test1-flow.json", mockFS)

	require.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "emulator-account", conf.Accounts[0].Name)
	assert.Equal(t, "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Keys[0].Context["privateKey"])
	assert.Equal(t, "127.0.0.1:3569", conf.Networks.GetByName("emulator").Host)
}

func Test_SimpleJSONConfigEnvAccount(t *testing.T) {
	b := []byte(`{
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
				"keys": "${env:EMULATOR__KEY}",
				"chain": "flow-emulator"
			}
		},
		"deploys": {}
	}`)

	os.Setenv("EMULATOR__KEY", "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7")

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)

	require.NoError(t, err)

	conf, loadErr := Load("test2-flow.json", mockFS)

	require.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Keys[0].Context["privateKey"])
}
