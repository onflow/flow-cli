package json

import (
	"encoding/json"
	"strings"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type jsonAccounts map[string]jsonAccount

func transformChainID(rawChainID string, rawAddress string) flow.ChainID {
	if rawAddress == "service" && rawChainID == "" {
		return flow.Emulator
	}
	return flow.ChainID(rawChainID)
}

func transformAddress(rawAddress string, rawChainID string) flow.Address {
	var address flow.Address
	chainID := transformChainID(rawChainID, rawAddress)

	if rawAddress == "service" {
		address = flow.ServiceAddress(chainID)
	} else {
		rawAddress = strings.ReplaceAll(rawAddress, "0x", "") // remove 0x if present
		address = flow.HexToAddress(rawAddress)
	}

	return address
}

func (j jsonAccounts) transformToConfig() config.Accounts {
	accounts := make(config.Accounts, 0)

	for accountName, a := range j {
		var account config.Account
		// simple format
		if a.Simple.Address != "" {
			account = config.Account{
				Name:    accountName,
				ChainID: transformChainID(a.Simple.Chain, a.Simple.Address),
				Address: transformAddress(a.Simple.Address, a.Simple.Chain),
				Keys: []config.AccountKey{{
					Type:     config.KeyTypeHex,
					Index:    0,
					SigAlgo:  crypto.ECDSA_P256,
					HashAlgo: crypto.SHA3_256,
					Context: map[string]string{
						"privateKey": a.Simple.Keys,
					},
				}},
			}
		} else { // advanced format
			keys := make([]config.AccountKey, 0)
			for _, key := range a.Advanced.Keys {
				key := config.AccountKey{
					Type:     key.Type,
					Index:    key.Index,
					SigAlgo:  crypto.StringToSignatureAlgorithm(key.SigAlgo),
					HashAlgo: crypto.StringToHashAlgorithm(key.HashAlgo),
					Context:  key.Context,
				}
				keys = append(keys, key)
			}

			account = config.Account{
				Name:    accountName,
				ChainID: transformChainID(a.Advanced.Chain, a.Advanced.Address),
				Address: transformAddress(a.Advanced.Address, a.Advanced.Chain),
				Keys:    keys,
			}
		}

		accounts = append(accounts, account)
	}

	return accounts
}

type jsonAccountSimple struct {
	Address string `json:"address"`
	Keys    string `json:"keys"`
	Chain   string `json:"chain"`
}

type jsonAccountAdvanced struct {
	Address string           `json:"address"`
	Chain   string           `json:"chain"`
	Keys    []jsonAccountKey `json:"keys"`
	//TODO: define more properties
}

type jsonAccountKey struct {
	Type     config.KeyType    `json:"type"`
	Index    int               `json:"index"`
	SigAlgo  string            `json:"signatureAlgorithm"`
	HashAlgo string            `json:"hashAlgorithm"`
	Context  map[string]string `json:"context"`
}

type jsonAccount struct {
	Simple   jsonAccountSimple
	Advanced jsonAccountAdvanced
}

func (j *jsonAccount) UnmarshalJSON(b []byte) error {

	// try simple format
	var simple jsonAccountSimple
	err := json.Unmarshal(b, &simple)
	if err == nil {
		j.Simple = simple
		return nil
	}

	// try advanced format
	var advanced jsonAccountAdvanced
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		j.Advanced = advanced
		return nil
	}

	return err
}
