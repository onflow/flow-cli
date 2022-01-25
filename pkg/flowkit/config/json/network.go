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

package json

import (
	"encoding/json"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type jsonNetworks map[string]jsonNetwork

// transformToConfig transforms json structures to config structure.
func (j jsonNetworks) transformToConfig() (config.Networks, error) {
	var networks config.Networks

	for networkName, n := range j {
		network := config.Network{
			Name:           networkName,
			Host:           n.Host,
			HostNetworkKey: n.NetworkKey,
		}

		networks = append(networks, network)
	}

	return networks, nil
}

// transformToJSON transforms config structure to json structures for saving.
func transformNetworksToJSON(networks config.Networks) jsonNetworks {
	jsonNetworks := jsonNetworks{}

	for _, n := range networks {
		jsonNetworks[n.Name] = jsonNetwork{
			Host:       n.Host,
			NetworkKey: n.HostNetworkKey,
		}
	}

	return jsonNetworks
}

type jsonNetwork struct {
	Host       string
	NetworkKey string
}

type advancedNetwork struct {
	Host  string
	Chain string
}

func (j *jsonNetwork) UnmarshalJSON(b []byte) error {
	var m map[string]string
	err := json.Unmarshal(b, &m)
	if err == nil {
		j.Host = m["host"]
		j.NetworkKey = m["network-key"]
		return nil
	}

	// ignore advanced schema from previous configuration format
	var advanced advancedNetwork
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		j.Host = advanced.Host
	}

	return err
}

func (j jsonNetwork) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"host": j.Host, "network-key": j.NetworkKey})
}
