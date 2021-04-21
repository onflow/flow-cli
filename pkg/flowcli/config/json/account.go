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
)

type jsonAccounts map[string]jsonAccount

// transformAddress returns address based on address and chain id
func transformAddress(address string) flow.Address {
	// only allow service for emulator
	if address == "service" {
		return flow.ServiceAddress(flow.Emulator)
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
				Address: transformAddress(a.Simple.Address),
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
				Address: transformAddress(a.Advanced.Address),
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
			Key:     a.Key.Context[config.PrivateKeyField],
		},
	}
}

func transformAdvancedAccountToJSON(a config.Account) jsonAccount {
	return jsonAccount{
		Advanced: jsonAccountAdvanced{
			Address: a.Address.String(),
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
}

type jsonAccountAdvanced struct {
	Address string         `json:"address"`
	Key     jsonAccountKey `json:"key"`
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
