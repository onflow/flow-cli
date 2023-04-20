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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworks_ByName(t *testing.T) {
	// Define a sample network configuration.
	networks := Networks{
		{Name: "flow-local", Host: "localhost:3569", Key: "flow-local-key"},
		{Name: "flow-testnet", Host: "localhost:3570", Key: "flow-testnet-key"},
	}

	// Test getting an existing network.
	network, err := networks.ByName("flow-local")
	assert.NoError(t, err)
	assert.Equal(t, "flow-local", network.Name)
	assert.Equal(t, "localhost:3569", network.Host)
	assert.Equal(t, "flow-local-key", network.Key)

	// Test getting a non-existent network.
	network, err = networks.ByName("flow-mainnet")
	assert.Error(t, err)
	assert.Nil(t, network)
	assert.EqualError(t, err, "network named flow-mainnet does not exist in configuration")
}

func TestNetworks_AddOrUpdate(t *testing.T) {
	// Define a sample network configuration.
	networks := Networks{
		{Name: "flow-local", Host: "localhost:3569", Key: "flow-local-key"},
	}

	// Test adding a new network.
	networks.AddOrUpdate(Network{Name: "flow-testnet", Host: "localhost:3570", Key: "flow-testnet-key"})
	assert.Equal(t, 2, len(networks))

	// Test updating an existing network.
	networks.AddOrUpdate(Network{Name: "flow-local", Host: "localhost:3580", Key: "flow-local-key-updated"})
	assert.Equal(t, "localhost:3580", networks[0].Host)
}

func TestNetworks_Remove(t *testing.T) {
	// Define a sample network configuration.
	networks := Networks{
		{Name: "flow-local", Host: "localhost:3569", Key: "flow-local-key"},
		{Name: "flow-testnet", Host: "localhost:3570", Key: "flow-testnet-key"},
	}

	// Test removing an existing network.
	err := networks.Remove("flow-local")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(networks))

	// Test removing a non-existent network.
	err = networks.Remove("flow-mainnet")
	assert.Error(t, err)
	assert.EqualError(t, err, "network named flow-mainnet does not exist in configuration")
}
