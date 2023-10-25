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
package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/onflow/flow-cli/flowkit/config"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/flowkit/config/json"
)

var mockFS = afero.NewMemMapFs()

var af = afero.Afero{Fs: mockFS}

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
			"emulator": "127.0.0.1:3569"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"deployments": {}
	}`)

	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(af)
	composer.AddConfigParser(json.NewParser())
	conf, loadErr := composer.Load([]string{"test2-flow.json"})

	assert.NoError(t, loadErr)
	assert.Len(t, conf.Accounts, 1)
	assert.Equal(t,
		"0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		conf.Accounts[0].Key.PrivateKey.String(),
	)
}

func Test_ErrorWhenMissingBothDefaultJsonFiles(t *testing.T) {
	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	_, loadErr := composer.Load(config.DefaultPaths())

	assert.Error(t, loadErr)
	assert.Contains(t, loadErr.Error(), "missing configuration")
}

func Test_AllowMissingLocalJson(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, config.GlobalPath(), b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load(config.DefaultPaths())
	assert.NoError(t, loadErr)

	acc, err := conf.Accounts.ByName("emulator-account")
	assert.NoError(t, err)

	assert.Len(t, conf.Accounts, 1)
	assert.Equal(t,
		"0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		acc.Key.PrivateKey.String(),
	)
}

func Test_PreferLocalJson(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"emulator-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, config.GlobalPath(), b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load(config.DefaultPaths())
	assert.NotNil(t, conf)
	assert.NoError(t, err)

	acc, err := conf.Accounts.ByName("emulator-account")
	assert.NoError(t, err)

	assert.NoError(t, loadErr)
	assert.Len(t, conf.Accounts, 1)
	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		acc.Key.PrivateKey.String(),
	)
}

func Test_MissingConfiguration(t *testing.T) {
	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, err := composer.Load([]string{"flow.json"})

	assert.Nil(t, conf)
	assert.EqualError(t, err, "missing configuration")
}

func Test_ConfigurationMalformedJSON(t *testing.T) {
	b := []byte(`{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account",
			}
		},
		"contracts": {},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"deployments": {}
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, err := composer.Load(config.DefaultPaths())
	assert.EqualError(t, err, "failed to preprocess config: failed to parse config JSON: invalid character '}' looking for beginning of object key string")
	assert.Nil(t, conf)
}

func Test_ConfigurationWrongFormat(t *testing.T) {
	b := []byte(`{
		"deployments": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, err := composer.Load(config.DefaultPaths())
	assert.EqualError(t, err, "configuration syntax error: json: cannot unmarshal string into Go struct field jsonConfig.deployments of type []json.deployment")
	assert.Nil(t, conf)
}

func Test_ComposeJSON(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"admin-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "flow-testnet.json", b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json", "flow-testnet.json"})

	assert.NoError(t, loadErr)
	assert.Len(t, conf.Accounts, 2)

	account, err := conf.Accounts.ByName("emulator-account")
	assert.NoError(t, err)

	adminAccount, err := conf.Accounts.ByName("admin-account")
	assert.NoError(t, err)

	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", account.Key.PrivateKey.String())
	assert.NotNil(t, adminAccount)
	assert.Equal(t, "0x3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", adminAccount.Key.PrivateKey.String())
}

func Test_ComposeCrossReference(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"test": {
				"address":"f1d6e0586b0a20c7",
				"key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"deployments": {
			"testnet": {
				"test": ["NFT"]
			}
		}
	}`)

	b2 := []byte(`{
		"networks": {
			"testnet": "access.devnet.nodes.onflow.org:9000"
		},
		"contracts": { "NFT": "./NFT.cdc" }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "b.json", b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json", "b.json"})

	assert.NoError(t, loadErr)
	account, err := conf.Accounts.ByName("test")
	assert.NoError(t, err)
	assert.NotNil(t, account)

	deployments := conf.Deployments.ByAccountAndNetwork(account.Name, "testnet")
	assert.NotNil(t, deployments)
}

func Test_ComposeJSONOverwrite(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"admin-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"admin-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "flow-testnet.json", b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json", "flow-testnet.json"})

	assert.NoError(t, loadErr)
	account, err := conf.Accounts.ByName("admin-account")
	assert.NoError(t, err)

	assert.Len(t, conf.Accounts, 1)
	assert.NotNil(t, account)
	assert.Equal(t, "0x3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", account.Key.PrivateKey.String())
}

func Test_JSONEnv(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "$EMULATOR_KEY"
			},
			"advanced": {
				"address": "f8d6e0586b0a20c7",
				"key": {
					"type": "hex",
					"index": 0,
					"signatureAlgorithm": "ECDSA_P256",
					"hashAlgorithm": "SHA3_256",
					"privateKey": "$ADVANCED_KEY"
				}
			}
		}
	}`)
	const key1 = "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
	const key2 = "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
	os.Setenv("EMULATOR_KEY", key1)
	os.Setenv("ADVANCED_KEY", key2)
	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)
	require.NoError(t, err)

	t.Run("Load and replace env variable", func(t *testing.T) {
		composer := config.NewLoader(afero.Afero{Fs: mockFS})
		composer.AddConfigParser(json.NewParser())
		conf, loadErr := composer.Load([]string{"test2-flow.json"})

		assert.NoError(t, loadErr)
		assert.Equal(t, 2, len(conf.Accounts))

		acc1, _ := conf.Accounts.ByName("advanced")
		assert.Equal(t, fmt.Sprintf("0x%s", key2), acc1.Key.PrivateKey.String())

		acc2, _ := conf.Accounts.ByName("emulator-account")
		assert.Equal(t, fmt.Sprintf("0x%s", key1), acc2.Key.PrivateKey.String())
	})

	t.Run("Save and remove replaced env variable", func(t *testing.T) {
		composer := config.NewLoader(afero.Afero{Fs: mockFS})
		composer.AddConfigParser(json.NewParser())
		conf, err := composer.Load([]string{"test2-flow.json"})
		require.NoError(t, err)

		testConf := "test-flow.json"
		err = composer.Save(conf, testConf)
		require.NoError(t, err)

		content, err := afero.ReadFile(mockFS, testConf)
		require.NoError(t, err)
		assert.NotContains(t, string(content), key1)
		assert.NotContains(t, string(content), key2)
		assert.Contains(t, string(content), "$EMULATOR_KEY")
		assert.Contains(t, string(content), "$ADVANCED_KEY")
	})

	t.Run("Fail not present env variable", func(t *testing.T) {
		b := []byte(`{
			"accounts": {
				"emulator-account": {
					"address": "f8d6e0586b0a20c7",
					"key": "$NOT_EXISTS"
				}
			}
		}`)

		mockFS := afero.NewMemMapFs()
		_ = afero.WriteFile(mockFS, "test.json", b, 0644)
		composer := config.NewLoader(afero.Afero{Fs: mockFS})
		composer.AddConfigParser(json.NewParser())

		_, err = composer.Load([]string{"test.json"})
		assert.EqualError(t, err, "required environment variable NOT_EXISTS not set")
	})

}

func Test_LoadAccountFileType(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": {
					"type": "file",
					"location": "./test.pkey"
				}
			}
		}
	}`)
	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, config.GlobalPath(), b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load(config.DefaultPaths())
	require.NoError(t, loadErr)

	acc, err := conf.Accounts.ByName("emulator-account")
	assert.NoError(t, err)

	assert.Len(t, conf.Accounts, 1)
	assert.Equal(t, filepath.FromSlash("./test.pkey"), acc.Key.Location)
}
