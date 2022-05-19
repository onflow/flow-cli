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
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type jsonNetworks map[string]jsonNetwork

// transformToConfig transforms json structures to config structure.
func (j jsonNetworks) transformToConfig() (config.Networks, error) {
	networks := make(config.Networks, 0)

	for networkName, n := range j {
		if n.Advanced.Key != "" && n.Advanced.Host != "" {
			err := util.ValidateECDSAP256Pub(n.Advanced.Key)
			if err != nil {
				return nil, fmt.Errorf("invalid key %s for network with name %s", n.Advanced.Key, networkName)
			}

			networks = append(networks, config.Network{
				Name: networkName,
				Host: n.Advanced.Host,
				Key:  n.Advanced.Key,
			})
		} else if n.Simple.Host != "" {
			networks = append(networks, config.Network{
				Name: networkName,
				Host: n.Simple.Host,
			})
		} else {
			return nil, fmt.Errorf("failed to transform networks configuration")
		}
	}

	return networks, nil
}

// transformNetworksToJSON transforms config structure to json structures for saving.
func transformNetworksToJSON(networks config.Networks) jsonNetworks {
	jsonNetworks := jsonNetworks{}

	for _, n := range networks {
		if n.Key != "" {
			jsonNetworks[n.Name] = transformAdvancedNetworkToJSON(n)
		} else {
			jsonNetworks[n.Name] = transformSimpleNetworkToJSON(n)
		}
	}

	return jsonNetworks
}

func transformSimpleNetworkToJSON(n config.Network) jsonNetwork {
	return jsonNetwork{
		Simple: simpleNetwork{
			Host: n.Host,
		},
	}
}

func transformAdvancedNetworkToJSON(n config.Network) jsonNetwork {
	return jsonNetwork{
		Advanced: advancedNetwork{
			Host: n.Host,
			Key:  n.Key,
		},
	}
}

type jsonNetwork struct {
	Simple   simpleNetwork
	Advanced advancedNetwork
}

type simpleNetwork struct {
	Host string `json:"host"`
}

type advancedNetwork struct {
	Host string `json:"host"`
	Key  string `json:"key"`
}

func (j *jsonNetwork) UnmarshalJSON(b []byte) error {
	var host string
	err := json.Unmarshal(b, &host)
	if err == nil {
		j.Simple.Host = host
		return nil
	}

	// ignore advanced schema from previous configuration format
	var advanced advancedNetwork
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		j.Advanced.Host = advanced.Host
		j.Advanced.Key = advanced.Key
	}

	return err
}

func (j jsonNetwork) MarshalJSON() ([]byte, error) {
	if j.Simple != (simpleNetwork{}) {
		return json.Marshal(j.Simple.Host)
	}

	return json.Marshal(j.Advanced)
}
