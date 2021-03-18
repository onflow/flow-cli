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
	"strings"

	"github.com/onflow/flow-cli/pkg/flow/util"

	"github.com/onflow/flow-cli/pkg/flow/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type jsonAccounts map[string]jsonAccount

// transformChainID return chain id based on address and chain id
func transformChainID(rawChainID string, rawAddress string) flow.ChainID {
	if rawAddress == "service" && rawChainID == "" {
		return flow.Emulator
	}

	if rawChainID == "" {
		address := flow.HexToAddress(strings.ReplaceAll(rawAddress, "0x", ""))
		chainID, _ := util.GetAddressNetwork(address)
		return chainID
	}

	return flow.ChainID(rawChainID)
}

// transformAddress returns address based on address and chain id
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

// transformToConfig transforms json structures to config structure
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
						config.PrivateKeyField: a.Simple.Keys,
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

func isDefaultKeyFormat(keys []config.AccountKey) bool {
	return len(keys) == 1 && keys[0].Index == 0 &&
		keys[0].Type == config.KeyTypeHex &&
		keys[0].SigAlgo == crypto.ECDSA_P256 &&
		keys[0].HashAlgo == crypto.SHA3_256
}

func transformSimpleAccountToJSON(a config.Account) jsonAccount {
	return jsonAccount{
		Simple: jsonAccountSimple{
			Address: a.Address.String(),
			Chain:   a.ChainID.String(),
			Keys:    a.Keys[0].Context[config.PrivateKeyField],
		},
	}
}

func transformAdvancedAccountToJSON(a config.Account) jsonAccount {
	var keys []jsonAccountKey

	for _, k := range a.Keys {
		keys = append(keys, jsonAccountKey{
			Type:     k.Type,
			Index:    k.Index,
			SigAlgo:  k.SigAlgo.String(),
			HashAlgo: k.HashAlgo.String(),
			Context:  k.Context,
		})
	}

	return jsonAccount{
		Advanced: jsonAccountAdvanced{
			Address: a.Address.String(),
			Chain:   a.ChainID.String(),
			Keys:    keys,
		},
	}
}

// transformToJSON transforms config structure to json structures for saving
func transformAccountsToJSON(accounts config.Accounts) jsonAccounts {
	jsonAccounts := jsonAccounts{}

	for _, a := range accounts {
		// if simple
		if isDefaultKeyFormat(a.Keys) {
			jsonAccounts[a.Name] = transformSimpleAccountToJSON(a)
		} else { // if advanced
			jsonAccounts[a.Name] = transformAdvancedAccountToJSON(a)
		}
	}

	return jsonAccounts
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

	// TODO: better error handling - here we just return error from advanced case
	return err
}

func (j jsonAccount) MarshalJSON() ([]byte, error) {
	if j.Simple != (jsonAccountSimple{}) {
		return json.Marshal(j.Simple)
	}

	return json.Marshal(j.Advanced)
}
