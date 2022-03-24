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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigNetworkSimple(t *testing.T) {
	b := []byte(`{
    "testnet": "access.testnet.nodes.onflow.org:9000"
	}`)

	var jsonNetworks jsonNetworks
	err := json.Unmarshal(b, &jsonNetworks)
	assert.NoError(t, err)

	networks, err := jsonNetworks.transformToConfig()
	assert.NoError(t, err)

	network, err := networks.ByName("testnet")
	assert.NoError(t, err)
	assert.Equal(t, network.Host, "access.testnet.nodes.onflow.org:9000")
	assert.Equal(t, network.Name, "testnet")
}

func Test_ConfigNetworkMultiple(t *testing.T) {
	b := []byte(`{
    		"emulator": "127.0.0.1:3569",
    		"testnet": {
				"host": "access.testnet.nodes.onflow.org:9000",
				"key": "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd"
			}
	}`)

	var jsonNetworks jsonNetworks
	err := json.Unmarshal(b, &jsonNetworks)
	assert.NoError(t, err)

	networks, err := jsonNetworks.transformToConfig()
	assert.NoError(t, err)

	network, err := networks.ByName("testnet")
	assert.NoError(t, err)
	assert.Equal(t, network.Host, "access.testnet.nodes.onflow.org:9000")
	assert.Equal(t, network.Name, "testnet")
	assert.Equal(t, network.Key, "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd")

	emulator, err := networks.ByName("emulator")
	assert.NoError(t, err)
	assert.Equal(t, emulator.Name, "emulator")
	assert.Equal(t, emulator.Host, "127.0.0.1:3569")
}

func Test_TransformNetworkToJSON(t *testing.T) {
	b := []byte(`{"emulator":"127.0.0.1:3569","testnet":{"host":"access.testnet.nodes.onflow.org:9000","key":"5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd"}}`)

	var jsonNetworks jsonNetworks
	err := json.Unmarshal(b, &jsonNetworks)
	assert.NoError(t, err)

	networks, err := jsonNetworks.transformToConfig()
	assert.NoError(t, err)

	j := transformNetworksToJSON(networks)
	x, _ := json.Marshal(j)

	assert.Equal(t, string(b), string(x))
}

func Test_IgnoreOldFormat(t *testing.T) {
	b := []byte(`{"emulator":"127.0.0.1:3569","testnet":{"host":"access.testnet.nodes.onflow.org:9000","key":"5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd"},"mainnet":{"host": "access.mainnet.nodes.onflow.org:9000","chain":"flow-mainnet","key":"5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd"}}`)

	var jsonNetworks jsonNetworks
	err := json.Unmarshal(b, &jsonNetworks)
	assert.NoError(t, err)

	conf, err := jsonNetworks.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, jsonNetworks, 3)

	testnet, err := conf.ByName("testnet")
	assert.NoError(t, err)

	mainnet, err := conf.ByName("mainnet")
	assert.NoError(t, err)

	assert.Equal(t, testnet.Host, "access.testnet.nodes.onflow.org:9000")
	assert.Equal(t, testnet.Key, "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd")
	assert.Equal(t, mainnet.Host, "access.mainnet.nodes.onflow.org:9000")
	assert.Equal(t, mainnet.Key, "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd")
}

func Test_TransformConfigAdvanced(t *testing.T) {
	t.Run("should returned advanced config with no errors", func(t *testing.T) {
		b := []byte(`{"testnet":{"host":"access.testnet.nodes.onflow.org:9000", "key": "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd"}}`)
		var jsonNetworks jsonNetworks
		err := json.Unmarshal(b, &jsonNetworks)
		assert.NoError(t, err)

		conf, err := jsonNetworks.transformToConfig()
		assert.NoError(t, err)

		testnet, err := conf.ByName("testnet")
		assert.NoError(t, err)
		assert.Equal(t, testnet.Host, "access.testnet.nodes.onflow.org:9000")
		assert.Equal(t, testnet.Key, "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd")
	})
	t.Run("should return error if advanced config does not have key", func(t *testing.T) {
		b := []byte(`{"testnet":{"host":"access.testnet.nodes.onflow.org:9000"}}`)
		var jsonNetworks jsonNetworks
		err := json.Unmarshal(b, &jsonNetworks)
		assert.NoError(t, err)

		_, err = jsonNetworks.transformToConfig()
		assert.Error(t, err)
	})
	t.Run("should return error if advanced config does not have host", func(t *testing.T) {
		b := []byte(`{"testnet":{"key": "5000676131ad3e22d853a3f75a5b5d0db4236d08dd6612e2baad771014b5266a242bccecc3522ff7207ac357dbe4f225c709d9b273ac484fed5d13976a39bdcd"}}`)
		var jsonNetworks jsonNetworks
		err := json.Unmarshal(b, &jsonNetworks)
		assert.NoError(t, err)

		_, err = jsonNetworks.transformToConfig()
		assert.Error(t, err)
	})
	t.Run("should return error if advanced config provides invalid network key", func(t *testing.T) {
		b := []byte(`{"testnet":{"host":"access.testnet.nodes.onflow.org:9000","key": "0xpublickey"}}`)
		var jsonNetworks jsonNetworks
		err := json.Unmarshal(b, &jsonNetworks)
		assert.NoError(t, err)

		_, err = jsonNetworks.transformToConfig()
		assert.Error(t, err)
	})
}
