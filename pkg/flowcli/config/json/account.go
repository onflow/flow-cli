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

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

type jsonAccounts map[string]jsonAccount

// transformChainID return chain id based on address and chain id
func transformChainID(rawChainID string, rawAddress string) flow.ChainID {
	if rawAddress == "service" && rawChainID == "" {
		return flow.Emulator
	}

	if rawChainID == "" {
		address := flow.HexToAddress(rawAddress)
		chainID, _ := util.GetAddressNetwork(address)
		return chainID
	}

	return flow.ChainID(rawChainID)
}

// transformAddress returns address based on address and chain id
func transformAddress(address string, rawChainID string) flow.Address {
	chainID := transformChainID(rawChainID, address)

	if address == "service" {
		return flow.ServiceAddress(chainID)
	}

	return flow.HexToAddress(address)
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
				Key: config.AccountKey{
					Type:     config.KeyTypeHex,
					Index:    0,
					SigAlgo:  crypto.ECDSA_P256,
					HashAlgo: crypto.SHA3_256,
					Context: map[string]string{
						config.PrivateKeyField: a.Simple.Key,
					},
				},
			}
		} else { // advanced format
			account = config.Account{
				Name:    accountName,
				ChainID: transformChainID(a.Advanced.Chain, a.Advanced.Address),
				Address: transformAddress(a.Advanced.Address, a.Advanced.Chain),
				Key: config.AccountKey{
					Type:     a.Advanced.Key.Type,
					Index:    a.Advanced.Key.Index,
					SigAlgo:  crypto.StringToSignatureAlgorithm(a.Advanced.Key.SigAlgo),
					HashAlgo: crypto.StringToHashAlgorithm(a.Advanced.Key.HashAlgo),
					Context:  a.Advanced.Key.Context,
				},
			}
		}

		accounts = append(accounts, account)
	}

	return accounts
}

func isDefaultKeyFormat(key config.AccountKey) bool {
	return key.Index == 0 &&
		key.Type == config.KeyTypeHex &&
		key.SigAlgo == crypto.ECDSA_P256 &&
		key.HashAlgo == crypto.SHA3_256
}

func transformSimpleAccountToJSON(a config.Account) jsonAccount {
	return jsonAccount{
		Simple: jsonAccountSimple{
			Address: a.Address.String(),
			Chain:   a.ChainID.String(),
			Key:     a.Key.Context[config.PrivateKeyField],
		},
	}
}

func transformAdvancedAccountToJSON(a config.Account) jsonAccount {
	return jsonAccount{
		Advanced: jsonAccountAdvanced{
			Address: a.Address.String(),
			Chain:   a.ChainID.String(),
			Key: jsonAccountKey{
				Type:     a.Key.Type,
				Index:    a.Key.Index,
				SigAlgo:  a.Key.SigAlgo.String(),
				HashAlgo: a.Key.HashAlgo.String(),
				Context:  a.Key.Context,
			},
		},
	}
}

// transformToJSON transforms config structure to json structures for saving
func transformAccountsToJSON(accounts config.Accounts) jsonAccounts {
	jsonAccounts := jsonAccounts{}

	for _, a := range accounts {
		// if simple
		if isDefaultKeyFormat(a.Key) {
			jsonAccounts[a.Name] = transformSimpleAccountToJSON(a)
		} else { // if advanced
			jsonAccounts[a.Name] = transformAdvancedAccountToJSON(a)
		}
	}

	return jsonAccounts
}

type jsonAccountSimple struct {
	Address string `json:"address"`
	Key     string `json:"key"`
	Chain   string `json:"chain"`
}

type jsonAccountAdvanced struct {
	Address string         `json:"address"`
	Chain   string         `json:"chain"`
	Key     jsonAccountKey `json:"key"`
}

type jsonAccountKey struct {
	Type     config.KeyType    `json:"type"`
	Index    int               `json:"index"`
	SigAlgo  string            `json:"signatureAlgorithm"`
	HashAlgo string            `json:"hashAlgorithm"`
	Context  map[string]string `json:"context"`
}

type jsonAccountSimpleOld struct {
	Address string `json:"address"`
	Keys    string `json:"keys"`
}

type jsonAccountAdvancedOld struct {
	Address string           `json:"address"`
	Keys    []jsonAccountKey `json:"keys"`
}

type jsonAccount struct {
	Simple   jsonAccountSimple
	Advanced jsonAccountAdvanced
}

type FormatType int

const (
	simpleFormat      FormatType = 0
	advancedFormat    FormatType = 1
	simpleOldFormat   FormatType = 2
	advancedOldFormat FormatType = 3
)

func decideFormat(b []byte) (FormatType, error) {
	var raw map[string]interface{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return 0, err
	}

	if raw["keys"] != nil {
		switch raw["keys"].(type) {
		case string:
			return simpleOldFormat, nil
		default:
			return advancedOldFormat, nil
		}
	}

	switch raw["key"].(type) {
	case string:
		return simpleFormat, nil
	default:
		return advancedFormat, nil
	}
}

func (j *jsonAccount) UnmarshalJSON(b []byte) error {

	format, err := decideFormat(b)
	if err != nil {
		return err
	}

	switch format {
	case simpleFormat:
		var simple jsonAccountSimple
		err = json.Unmarshal(b, &simple)
		j.Simple = simple

	case advancedFormat:
		var advanced jsonAccountAdvanced
		err = json.Unmarshal(b, &advanced)
		j.Advanced = advanced

	case simpleOldFormat:
		var simpleOld jsonAccountSimpleOld
		err = json.Unmarshal(b, &simpleOld)
		j.Simple = jsonAccountSimple{
			Address: simpleOld.Address,
			Key:     simpleOld.Keys,
		}

	case advancedOldFormat:
		var advancedOld jsonAccountAdvancedOld
		err = json.Unmarshal(b, &advancedOld)
		j.Advanced = jsonAccountAdvanced{
			Address: advancedOld.Address,
			Key:     advancedOld.Keys[0],
		}
	}

	return err
}

func (j jsonAccount) MarshalJSON() ([]byte, error) {
	if j.Simple != (jsonAccountSimple{}) {
		return json.Marshal(j.Simple)
	}

	return json.Marshal(j.Advanced)
}
