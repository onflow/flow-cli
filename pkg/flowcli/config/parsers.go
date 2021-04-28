package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk"
)

// StringToAccount converts string values to account
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
		Type:     KeyTypeHex,
		Index:    parsedIndex,
		SigAlgo:  crypto.StringToSignatureAlgorithm(sigAlgo),
		HashAlgo: crypto.StringToHashAlgorithm(hashAlgo),
		Context: map[string]string{
			PrivateKeyField: strings.ReplaceAll(parsedKey.String(), "0x", ""),
		},
	}

	return &Account{
		Name:    name,
		Address: *parsedAddress,
		Key:     accountKey,
	}, nil
}

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

func StringToAddress(value string) (*flow.Address, error) {
	address := flow.HexToAddress(value)

	if !address.IsValid(flow.Mainnet) &&
		!address.IsValid(flow.Testnet) &&
		!address.IsValid(flow.Emulator) {
		return nil, fmt.Errorf("invalid address")
	}

	return &address, nil
}

func StringToHexKey(key string, sigAlgo string) (*crypto.PrivateKey, error) {
	privateKey, err := crypto.DecodePrivateKeyHex(
		crypto.StringToSignatureAlgorithm(sigAlgo),
		key,
	)
	if err != nil {
		return nil, err
	}

	return &privateKey, nil
}

func StringToContracts(
	name string,
	source string,
	emulatorAlias string,
	testnetAlias string,
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

func StringToNetwork(name string, host string) Network {
	return Network{
		Name: name,
		Host: host,
	}
}

func StringToDeployment(network string, account string, contracts []string) Deploy {
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

	return Deploy{
		Network:   network,
		Account:   account,
		Contracts: parsedContracts,
	}
}
