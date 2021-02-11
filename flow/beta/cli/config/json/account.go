package json

import (
	"encoding/json"
	"errors"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type jsonAccounts map[string]jsonAccount

func (j jsonAccounts) transformToConfig() config.Accounts {
	accounts := make(config.Accounts, 0)

	for accountName, a := range j {
		var account config.Account
		// simple format
		if a.Simple.Address != "" {
			account = config.Account{
				Name:    accountName,
				Address: flow.HexToAddress(a.Simple.Address), //TODO: improve (0x handle, validation)
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
				Address: flow.HexToAddress(a.Advanced.Address), //REF: merge with logic above - code dup
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
}

type jsonAccountAdvanced struct {
	Address string           `json:"address"`
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

func (j jsonAccount) UnmarshalJSON(b []byte) error {
	var val interface{}

	err := json.Unmarshal(b, &val)
	if err != nil {
		return err
	}

	switch typedVal := val.(type) {
	case jsonAccountSimple:
		j.Simple = typedVal
	case jsonAccountAdvanced:
		j.Advanced = typedVal
	default:
		return errors.New("invalid account definition")
	}

	return nil
}
