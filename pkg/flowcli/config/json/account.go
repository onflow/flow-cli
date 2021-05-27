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
	"fmt"
	"strconv"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
)

type jsonAccounts map[string]account

// transformAddress returns address based on address and chain id
func transformAddress(address string) flow.Address {
	// only allow service for emulator
	if address == "service" {
		return flow.ServiceAddress(flow.Emulator)
	}

	return flow.HexToAddress(address)
}

func transformKeyBaseToConfig(key keyAdvanced) config.AccountKeyBase {
	index, _ := strconv.Atoi(key["Index"])

	return config.AccountKeyBase{
		Type:     config.KeyType(key["Type"]),
		Index:    index,
		SigAlgo:  crypto.StringToSignatureAlgorithm(key["SigAlgo"]),
		HashAlgo: crypto.StringToHashAlgorithm(key["HashAlgo"]),
	}
}

func transformKeyToConfig(key keyAdvanced) interface{} {
	keyType := config.KeyType(key["Type"])
	base := transformKeyBaseToConfig(key)

	if keyType == config.KeyTypeHex {
		pKey, _ := crypto.DecodePrivateKey(base.SigAlgo, []byte(key["PrivateKey"]))
		return config.AccountKeyHex{
			AccountKeyBase: base,
			PrivateKey:     pKey,
		}
	} else {
		return config.AccountKeyKMS{
			AccountKeyBase: base,
			ResourceID:     key["ResourceID"],
		}
	}
}

// todo add error in return and do validation
// transformToConfig transforms json structures to config structure
func (j jsonAccounts) transformToConfig() config.Accounts {
	accounts := make(config.Accounts, 0)

	for accountName, a := range j {
		var account config.Account
		// simple format
		if a.Simple.Address != "" {
			pkey, _ := crypto.DecodePrivateKeyHex(
				crypto.ECDSA_P256,
				strings.ReplaceAll(a.Simple.Key, "0x", ""),
			)
			account = config.Account{
				Name:    accountName,
				Address: transformAddress(a.Simple.Address),
				Key: config.AccountKeyHex{
					AccountKeyBase: config.AccountKeyBase{
						Type:     config.KeyTypeHex,
						Index:    0,
						SigAlgo:  crypto.ECDSA_P256,
						HashAlgo: crypto.SHA3_256,
					},
					PrivateKey: pkey,
				},
			}
		} else { // advanced format
			account = config.Account{
				Name:    accountName,
				Address: transformAddress(a.Advanced.Address),
				Key:     transformKeyToConfig(a.Advanced.Key),
			}
		}

		accounts = append(accounts, account)
	}

	return accounts
}

func isDefaultKeyFormat(key config.AccountKeyBase) bool {
	return key.Index == 0 &&
		key.Type == config.KeyTypeHex &&
		key.SigAlgo == crypto.ECDSA_P256 &&
		key.HashAlgo == crypto.SHA3_256
}

func transformSimpleAccountToJSON(a config.Account) account {
	return account{
		Simple: accountSimple{
			Address: a.Address.String(),
			Key:     a.Key.(config.AccountKeyHex).PrivateKey.String(),
		},
	}
}

func transformAdvancedAccountToJSON(a config.Account) account {
	hexKey := a.Key.(config.AccountKeyHex)
	baseKey := a.Key.(config.AccountKeyBase)
	// todo kms

	return account{
		Advanced: accountAdvanced{
			Address: a.Address.String(),
			Key: map[string]string{
				"Type":       string(baseKey.Type),
				"Index":      fmt.Sprintf("%d", baseKey.Index),
				"SigAlgo":    baseKey.SigAlgo.String(),
				"HashAlgo":   baseKey.HashAlgo.String(),
				"PrivateKey": hexKey.PrivateKey.String(),
			},
		},
	}
}

// transformToJSON transforms config structure to json structures for saving
func transformAccountsToJSON(accounts config.Accounts) jsonAccounts {
	jsonAccounts := jsonAccounts{}

	for _, a := range accounts {
		if isDefaultKeyFormat(a.Key.(config.AccountKeyBase)) {
			jsonAccounts[a.Name] = transformSimpleAccountToJSON(a)
		} else { // if advanced
			jsonAccounts[a.Name] = transformAdvancedAccountToJSON(a)
		}
	}

	return jsonAccounts
}

type account struct {
	Simple   accountSimple
	Advanced accountAdvanced
}

type accountSimple struct {
	Address string `json:"address"`
	Key     string `json:"key"`
}

type accountAdvanced struct {
	Address string      `json:"address"`
	Key     keyAdvanced `json:"key"`
}

type keyAdvanced map[string]string

// support for pre v0.22 formats
type accountSimplePre022 struct {
	Address string `json:"address"`
	Keys    string `json:"keys"`
}

// support for pre v0.22 formats
type accountAdvancedPre022 struct {
	Address string        `json:"address"`
	Keys    []keyAdvanced `json:"keys"`
}

type FormatType int

const (
	simpleFormat         FormatType = 0
	advancedFormat       FormatType = 1
	simpleFormatPre022   FormatType = 2 // pre v.022 format
	advancedFormatPre022 FormatType = 3 // pre v.022 format
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
			return simpleFormatPre022, nil
		default:
			return advancedFormatPre022, nil
		}
	}

	switch raw["key"].(type) {
	case string:
		return simpleFormat, nil
	default:
		return advancedFormat, nil
	}
}

func (j *account) UnmarshalJSON(b []byte) error {

	format, err := decideFormat(b)
	if err != nil {
		return err
	}

	// todo refacto switch in array of parsers
	switch format {
	case simpleFormat:
		var simple accountSimple
		err = json.Unmarshal(b, &simple)
		j.Simple = simple

	case simpleFormatPre022:
		var simpleOld accountSimplePre022
		err = json.Unmarshal(b, &simpleOld)
		j.Simple = accountSimple{
			Address: simpleOld.Address,
			Key:     simpleOld.Keys,
		}

	case advancedFormatPre022:
		var advancedOld accountAdvancedPre022
		err = json.Unmarshal(b, &advancedOld)
		j.Advanced = accountAdvanced{
			Address: advancedOld.Address,
			Key:     advancedOld.Keys[0],
		}

	case advancedFormat:
		var advanced accountAdvanced
		err = json.Unmarshal(b, &advanced)
		j.Advanced = advanced
	}

	return err
}

func (j account) MarshalJSON() ([]byte, error) {
	if j.Simple != (accountSimple{}) {
		return json.Marshal(j.Simple)
	}

	return json.Marshal(j.Advanced)
}
