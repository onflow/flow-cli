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

	"github.com/onflow/flow-cli/pkg/flowcli/config"
)

type jsonNetworks map[string]jsonNetwork

// transformToConfig transforms json structures to config structure
func (j jsonNetworks) transformToConfig() config.Networks {
	networks := make(config.Networks, 0)

	for networkName, n := range j {
		network := config.Network{
			Name: networkName,
			Host: n.Host,
		}

		networks = append(networks, network)
	}

	return networks
}

// transformToJSON transforms config structure to json structures for saving
func transformNetworksToJSON(networks config.Networks) jsonNetworks {
	jsonNetworks := jsonNetworks{}

	for _, n := range networks {
		jsonNetworks[n.Name] = jsonNetwork{
			Host: n.Host,
		}
	}

	return jsonNetworks
}

type jsonNetwork struct {
	Host string
}

func (j *jsonNetwork) UnmarshalJSON(b []byte) error {
	var host string
	err := json.Unmarshal(b, &host)
	if err != nil {
		return err
	}

	j.Host = host
	return nil
}

func (j jsonNetwork) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Host)
}
