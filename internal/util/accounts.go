/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package util

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
)

// GetAccountByContractName retrieves an account by contract name for a specific network
func GetAccountByContractName(state *flowkit.State, contractName string, network config.Network) (*accounts.Account, error) {
	deployments := state.Deployments().ByNetwork(network.Name)
	var accountName string
	for _, d := range deployments {
		for _, c := range d.Contracts {
			if c.Name == contractName {
				accountName = d.Account
				break
			}
		}
	}
	if accountName == "" {
		return nil, fmt.Errorf("contract not found in state")
	}

	accs := state.Accounts()
	if accs == nil {
		return nil, fmt.Errorf("no accounts found in state")
	}

	var account *accounts.Account
	for _, a := range *accs {
		if accountName == a.Name {
			account = &a
			break
		}
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found in state", accountName)
	}

	return account, nil
}

// GetAddressByContractName retrieves an address by contract name for a specific network
func GetAddressByContractName(state *flowkit.State, contractName string, network config.Network) (flow.Address, error) {
	account, err := GetAccountByContractName(state, contractName, network)
	if err != nil {
		return flow.Address{}, err
	}

	return flow.HexToAddress(account.Address.Hex()), nil
}

// GenerateTestPrivateKey generates a deterministic private key for testing
func GenerateTestPrivateKey() crypto.PrivateKey {
	seed := make([]byte, crypto.MinSeedLength)
	for i := range seed {
		seed[i] = byte(i)
	}
	privKey, _ := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	return privKey
}

// GetAccountsByNetworks returns all accounts that are valid for the specified networks
func GetAccountsByNetworks(state *flowkit.State, networks []string) []accounts.Account {
	var filteredAccounts []accounts.Account

	allAccounts := *state.Accounts()
	for _, account := range allAccounts {
		for _, network := range networks {
			if IsAddressValidForNetwork(account.Address, network) {
				filteredAccounts = append(filteredAccounts, account)
				break // Found a matching network, no need to check others
			}
		}
	}

	return filteredAccounts
}

// GetTestnetAccounts returns all accounts that have testnet-valid addresses
func GetTestnetAccounts(state *flowkit.State) []accounts.Account {
	return GetAccountsByNetworks(state, []string{"testnet"})
}

// GetEmulatorAccounts returns all accounts that have emulator-valid addresses
func GetEmulatorAccounts(state *flowkit.State) []accounts.Account {
	return GetAccountsByNetworks(state, []string{"emulator"})
}

// ResolveAddressOrAccountNameForNetworks resolves addresses for specified networks (supports multiple networks)
func ResolveAddressOrAccountNameForNetworks(input string, state *flowkit.State, supportedNetworks []string) (flow.Address, error) {
	address := flow.HexToAddress(input)

	// Check if the direct address is valid for any of the supported networks
	for _, network := range supportedNetworks {
		if IsAddressValidForNetwork(address, network) {
			return address, nil
		}
	}

	// If it's a valid address for unsupported networks, reject it
	if IsAddressValidForNetwork(address, "mainnet") || IsAddressValidForNetwork(address, "testnet") || IsAddressValidForNetwork(address, "emulator") {
		networksStr := strings.Join(supportedNetworks, " and ")
		return flow.EmptyAddress, fmt.Errorf("unsupported address %s, only supported for %s addresses", address.String(), networksStr)
	}

	// Try to resolve as account name
	account, err := state.Accounts().ByName(input)
	if err != nil {
		return flow.EmptyAddress, fmt.Errorf("could not find account with name %s", input)
	}

	// Check if the account's address is valid for any supported network
	for _, network := range supportedNetworks {
		if IsAddressValidForNetwork(account.Address, network) {
			return account.Address, nil
		}
	}

	networksStr := strings.Join(supportedNetworks, " and ")
	return flow.EmptyAddress, fmt.Errorf("account %s has address %s which is not valid for %s addresses", input, account.Address.String(), networksStr)
}
