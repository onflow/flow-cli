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

package config

import (
	"fmt"
)

var (
	EmptyNetwork    = Network{}
	EmulatorNetwork = Network{
		Name: "emulator",
		Host: "127.0.0.1:3569",
	}
	TestingNetwork = Network{
		Name: "testing",
		Host: "127.0.0.1:3569",
	}
	TestnetNetwork = Network{
		Name: "testnet",
		Host: "access.devnet.nodes.onflow.org:9000",
	}
	MainnetNetwork = Network{
		Name: "mainnet",
		Host: "access.mainnet.nodes.onflow.org:9000",
	}
	CrescendoNetwork = Network{
		Name: "crescendo",
		Host: "access.crescendo.nodes.onflow.org:9000",
	}
	DefaultNetworks = Networks{
		EmulatorNetwork,
		TestingNetwork,
		TestnetNetwork,
		MainnetNetwork,
		CrescendoNetwork,
	}
)

type Networks []Network

// Network defines the configuration for a Flow network.
type Network struct {
	Name string
	Host string
	Key  string
}

// ByName get network by name or return an error if not found.
func (n *Networks) ByName(name string) (*Network, error) {
	for _, network := range *n {
		if network.Name == name {
			return &network, nil
		}
	}

	return nil, fmt.Errorf("network named %s does not exist in configuration", name)
}

// AddOrUpdate add new network or update if already present.
func (n *Networks) AddOrUpdate(network Network) {
	for i, existingNetwork := range *n {
		if existingNetwork.Name == network.Name {
			(*n)[i] = network
			return
		}
	}

	*n = append(*n, network)
}

// Remove network by the name.
func (n *Networks) Remove(name string) error {
	_, err := n.ByName(name)
	if err != nil {
		return err
	}

	for i, network := range *n {
		if network.Name == name {
			*n = append((*n)[0:i], (*n)[i+1:]...) // remove item
		}
	}

	return nil
}
