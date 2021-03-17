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

	"github.com/onflow/flow-cli/pkg/flow/config"
	"github.com/onflow/flow-go-sdk"
)

type jsonNetworks map[string]jsonNetwork

// transformToConfig transforms json structures to config structure
func (j jsonNetworks) transformToConfig() config.Networks {
	networks := make(config.Networks, 0)

	for networkName, n := range j {
		var network config.Network

		if n.Host != "" {
			network = config.Network{
				Name: networkName,
				Host: n.Host,
			}
		} else {
			network = config.Network{
				Name:    networkName,
				Host:    n.Advanced.Host,
				ChainID: flow.ChainID(n.Advanced.ChainID),
			}
		}

		networks = append(networks, network)
	}

	return networks
}

// transformToJSON transforms config structure to json structures for saving
func transformNetworksToJSON(networks config.Networks) jsonNetworks {
	jsonNetworks := jsonNetworks{}

	for _, n := range networks {
		// if simple case
		if n.ChainID == "" {
			jsonNetworks[n.Name] = jsonNetwork{
				Host: n.Host,
			}
		} else { // if advanced case
			jsonNetworks[n.Name] = jsonNetwork{
				Advanced: advanced{
					Host:    n.Host,
					ChainID: n.ChainID.String(),
				},
			}
		}
	}

	return jsonNetworks
}

type advanced struct {
	Host    string `json:"host"`
	ChainID string `json:"chain"`
}

type jsonNetwork struct {
	Host     string
	Advanced advanced
}

func (j *jsonNetwork) UnmarshalJSON(b []byte) error {
	// simple
	var host string
	err := json.Unmarshal(b, &host)
	if err == nil {
		j.Host = host
		return nil
	}

	// advanced
	var advanced advanced
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		j.Advanced = advanced
	} else {
		return err
	}

	return nil
}

func (j jsonNetwork) MarshalJSON() ([]byte, error) {
	if j.Host != "" {
		return json.Marshal(j.Host)
	} else {
		return json.Marshal(j.Advanced)
	}
}
