package config

import (
	"fmt"
)

type Networks []Network

// Network defines the configuration for a Flow network.
type Network struct {
	Name string
	Host string
}

// GetByName get network by name
func (n *Networks) GetByName(name string) *Network {
	for _, network := range *n {
		if network.Name == name {
			return &network
		}
	}

	return nil
}

// AddOrUpdate add new network or update if already present
func (n *Networks) AddOrUpdate(name string, network Network) {
	for i, existingNetwork := range *n {
		if existingNetwork.Name == name {
			(*n)[i] = network
			return
		}
	}

	*n = append(*n, network)
}

func (n *Networks) Remove(name string) error {
	network := n.GetByName(name)
	if network == nil {
		return fmt.Errorf("network named %s does not exist in configuration", name)
	}

	for i, network := range *n {
		if network.Name == name {
			*n = append((*n)[0:i], (*n)[i+1:]...) // remove item
		}
	}

	return nil
}

// DefaultEmulatorNetwork get default emulator network
func DefaultEmulatorNetwork() Network {
	return Network{
		Name: "emulator",
		Host: "127.0.0.1:3569",
	}
}

// DefaultTestnetNetwork get default testnet network
func DefaultTestnetNetwork() Network {
	return Network{
		Name: "testnet",
		Host: "access.devnet.nodes.onflow.org:9000",
	}
}

// DefaultMainnetNetwork get default mainnet network
func DefaultMainnetNetwork() Network {
	return Network{
		Name: "mainnet",
		Host: "access.mainnet.nodes.onflow.org:9000",
	}
}

// DefaultNetworks gets all default networks
func DefaultNetworks() Networks {
	return Networks{
		DefaultEmulatorNetwork(),
		DefaultTestnetNetwork(),
		DefaultMainnetNetwork(),
	}
}
