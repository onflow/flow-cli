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
	"strconv"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk"
)

// StringToAccount converts string values to account.
func StringToAccount(
	name string,
	address string,
	index string,
	sigAlgo string,
	hashAlgo string,
	key string,
) (*Account, error) {
	parsedAddress, err := StringToAddress(address)
	if err != nil {
		return nil, err
	}

	parsedIndex, err := StringToKeyIndex(index)
	if err != nil {
		return nil, err
	}

	parsedKey, err := StringToHexKey(key, sigAlgo)
	if err != nil {
		return nil, err
	}

	accountKey := AccountKey{
		Type:       KeyTypeHex,
		Index:      parsedIndex,
		SigAlgo:    crypto.StringToSignatureAlgorithm(sigAlgo),
		HashAlgo:   crypto.StringToHashAlgorithm(hashAlgo),
		PrivateKey: parsedKey,
	}

	return &Account{
		Name:    name,
		Address: *parsedAddress,
		Key:     accountKey,
	}, nil
}

// StringToKeyIndex converts string key index to valid key index integer.
func StringToKeyIndex(value string) (int, error) {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid index, must be a number")
	}
	if v < 0 {
		return 0, fmt.Errorf("invalid index, must be positive")
	}

	return v, nil
}

// StringToAddress converts string to valid Flow address.
func StringToAddress(value string) (*flow.Address, error) {
	address := flow.HexToAddress(value)

	if !address.IsValid(flow.Mainnet) &&
		!address.IsValid(flow.Testnet) &&
		!address.IsValid(flow.Emulator) {
		return nil, fmt.Errorf("invalid address")
	}

	return &address, nil
}

// StringToHexKey converts string private key and signature algorithm to private key.
func StringToHexKey(key string, sigAlgo string) (crypto.PrivateKey, error) {
	privateKey, err := crypto.DecodePrivateKeyHex(
		crypto.StringToSignatureAlgorithm(sigAlgo),
		key,
	)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// StringToContracts converts strings to contracts.
func StringToContracts(
	name string,
	source string,
	emulatorAlias string,
	testnetAlias string,
	mainnetAlias string,
) []Contract {
	contracts := make([]Contract, 0)

	if emulatorAlias != "" {
		contracts = append(contracts, Contract{
			Name:    name,
			Source:  source,
			Network: DefaultEmulatorNetwork().Name,
			Alias:   emulatorAlias,
		})
	}

	if mainnetAlias != "" {
		contracts = append(contracts, Contract{
			Name:    name,
			Source:  source,
			Network: DefaultMainnetNetwork().Name,
			Alias:   mainnetAlias,
		})
	}

	if testnetAlias != "" {
		contracts = append(contracts, Contract{
			Name:    name,
			Source:  source,
			Network: DefaultTestnetNetwork().Name,
			Alias:   testnetAlias,
		})
	}

	if emulatorAlias == "" && testnetAlias == "" {
		contracts = append(contracts, Contract{
			Name:    name,
			Source:  source,
			Network: "",
			Alias:   "",
		})
	}

	return contracts
}

// StringToNetwork converts string to network.
func StringToNetwork(name, host, networkKey string) Network {
	return Network{
		Name: name,
		Host: host,
		Key:  networkKey,
	}
}

// StringToDeployment converts string to deployment.
func StringToDeployment(network string, account string, contracts []string) Deployment {
	parsedContracts := make([]ContractDeployment, 0)

	for _, c := range contracts {
		// prevent adding multiple contracts with same name
		exists := false
		for _, p := range parsedContracts {
			if p.Name == c {
				exists = true
			}
		}
		if exists {
			continue
		}

		parsedContracts = append(
			parsedContracts,
			ContractDeployment{
				Name: c,
				Args: nil,
			})
	}

	return Deployment{
		Network:   network,
		Account:   account,
		Contracts: parsedContracts,
	}
}
