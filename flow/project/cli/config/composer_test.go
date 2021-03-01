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
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_JSONSimple(t *testing.T) {
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
				"keys": "123123123123123123123123123123",
				"chain": "flow-emulator"
			}
		},
		"deploys": {}
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)

	require.NoError(t, err)

	composer := NewComposer(mockFS)
	conf, loadErr := composer.Load([]string{"test2-flow.json"})

	require.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Keys[0].Context["privateKey"])
}

func Test_JSONEnv(t *testing.T) {
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
				"keys": "${env:EMULATOR_KEY}",
				"chain": "flow-emulator"
			}
		},
		"deploys": {}
	}`)

	os.Setenv("EMULATOR_KEY", "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7")

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)

	require.NoError(t, err)

	composer := NewComposer(mockFS)
	conf, loadErr := composer.Load([]string{"test2-flow.json"})

	require.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Keys[0].Context["privateKey"])
}
