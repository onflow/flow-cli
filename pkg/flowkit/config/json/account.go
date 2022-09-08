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

package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type jsonAccounts map[string]account

// transformAddress returns address based on address and chain id.
func transformAddress(address string) (flow.Address, error) {
	// only allow service for emulator
	if address == "service" {
		return flow.ServiceAddress(flow.Emulator), nil
	}

	if flow.HexToAddress(address) == flow.EmptyAddress {
		return flow.EmptyAddress, fmt.Errorf("could not parse address: %s", address)
	}

	return flow.HexToAddress(address), nil
}

// transformSimpleToConfig transforms simple internal account to config account.
func transformSimpleToConfig(accountName string, a simpleAccount) (*config.Account, error) {
	pkey, err := crypto.DecodePrivateKeyHex(
		crypto.ECDSA_P256,
		strings.TrimPrefix(a.Key, "0x"),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for account: %s", accountName)
	}

	address, err := transformAddress(a.Address)
	if err != nil {
		return nil, err
	}

	return &config.Account{
		Name:    accountName,
		Address: address,
		Key: config.AccountKey{
			Type:       config.KeyTypeHex,
			Index:      0,
			SigAlgo:    crypto.ECDSA_P256,
			HashAlgo:   crypto.SHA3_256,
			PrivateKey: pkey,
		},
	}, nil
}

// transformAdvancedToConfig transforms advanced internal account to config account.
func transformAdvancedToConfig(accountName string, a advancedAccount) (*config.Account, error) {
	sigAlgo := crypto.StringToSignatureAlgorithm(a.Key.SigAlgo)
	hashAlgo := crypto.StringToHashAlgorithm(a.Key.HashAlgo)

	if a.Key.Type != config.KeyTypeHex && a.Key.Type != config.KeyTypeGoogleKMS && a.Key.Type != config.KeyTypeBip44 {
		return nil, fmt.Errorf("invalid key type for account %s", accountName)
	}

	if a.Key.ResourceID != "" && a.Key.PrivateKey != "" {
		return nil, fmt.Errorf("only provide value for private key or resource ID on account %s", accountName)
	}

	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm for account %s", accountName)
	}

	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("invalid hash algorithm for account %s", accountName)
	}

	address, err := transformAddress(a.Address)
	if err != nil {
		return nil, err
	}

	key := config.AccountKey{
		Type:     a.Key.Type,
		Index:    a.Key.Index,
		SigAlgo:  sigAlgo,
		HashAlgo: hashAlgo,
	}

	switch a.Key.Type {
	case config.KeyTypeHex:
		if a.Key.PrivateKey == "" {
			return nil, fmt.Errorf("missing private key value for hex key type on account %s", accountName)
		}
		pKey, err := crypto.DecodePrivateKeyHex(
			sigAlgo,
			strings.TrimPrefix(a.Key.PrivateKey, "0x"),
		)
		if err != nil {
			return nil, err
		}

		key.PrivateKey = pKey
	case config.KeyTypeBip44:
		if a.Key.Mnemonic == "" {
			return nil, fmt.Errorf("missing mnemonic value for bip44 key type on account %s", accountName)
		}
		key.Mnemonic = a.Key.Mnemonic
		key.DerivationPath = a.Key.DerivationPath
		if key.DerivationPath == "" {
			key.DerivationPath = "m/44'/539'/0'/0/0"
		}

	case config.KeyTypeGoogleKMS:
		if a.Key.ResourceID == "" {
			return nil, fmt.Errorf("missing resource ID value for key on account %s", accountName)
		}
		key.ResourceID = a.Key.ResourceID
	}

	return &config.Account{
		Name:    accountName,
		Address: address,
		Key:     key,
	}, nil
}

// transformToConfig transforms json structures to config structure.
func (j jsonAccounts) transformToConfig() (config.Accounts, error) {
	accounts := make(config.Accounts, 0)

	for accountName, a := range j {
		var account *config.Account
		var err error
		if a.Simple.Address != "" {
			account, err = transformSimpleToConfig(accountName, a.Simple)
			if err != nil {
				return nil, err
			}
		} else { // advanced format
			account, err = transformAdvancedToConfig(accountName, a.Advanced)
			if err != nil {
				return nil, err
			}
		}

		accounts = append(accounts, *account)
	}

	return accounts, nil
}

// transformToJSON transforms config structure to json structures for saving.
func transformAccountsToJSON(accounts config.Accounts) jsonAccounts {
	jsonAccounts := jsonAccounts{}

	for _, a := range accounts {
		if a.Location != "" {
			jsonAccounts[a.Name] = transformFromFileAccountToJSON(a)
		} else if isDefaultKeyFormat(a.Key) && !a.UseAdvanceFormat {
			jsonAccounts[a.Name] = transformSimpleAccountToJSON(a)
		} else {
			jsonAccounts[a.Name] = transformAdvancedAccountToJSON(a)
		}
	}

	return jsonAccounts
}

func transformFromFileAccountToJSON(a config.Account) account {
	return account{
		FromFile: fromFileAccount{
			FromFile: a.Location,
		},
	}
}

func transformSimpleAccountToJSON(a config.Account) account {
	return account{
		Simple: simpleAccount{
			Address: a.Address.String(),
			Key:     strings.TrimPrefix(a.Key.PrivateKey.String(), "0x"),
		},
	}
}

func transformAdvancedAccountToJSON(a config.Account) account {
	return account{
		Advanced: advancedAccount{
			Address: a.Address.String(),
			Key:     transformAdvancedKeyToJSON(a.Key),
		},
	}
}

func transformAdvancedKeyToJSON(key config.AccountKey) advanceKey {
	advancedKey := advanceKey{
		Type:     key.Type,
		Index:    key.Index,
		SigAlgo:  key.SigAlgo.String(),
		HashAlgo: key.HashAlgo.String(),
	}

	switch key.Type {
	case config.KeyTypeHex:
		advancedKey.PrivateKey = strings.TrimPrefix(key.PrivateKey.String(), "0x")
	case config.KeyTypeBip44:
		advancedKey.Mnemonic = key.Mnemonic
		advancedKey.DerivationPath = key.DerivationPath
	case config.KeyTypeGoogleKMS:
		advancedKey.ResourceID = key.ResourceID
	}

	return advancedKey
}

func isDefaultKeyFormat(key config.AccountKey) bool {
	return key.Index == 0 &&
		key.Type == config.KeyTypeHex &&
		key.SigAlgo == crypto.ECDSA_P256 &&
		key.HashAlgo == crypto.SHA3_256
}

type account struct {
	FromFile fromFileAccount
	Simple   simpleAccount
	Advanced advancedAccount
}

type fromFileAccount struct {
	FromFile string `json:"fromFile"`
}

type simpleAccount struct {
	Address string `json:"address"`
	Key     string `json:"key"`
}

type advancedAccount struct {
	Address string     `json:"address"`
	Key     advanceKey `json:"key"`
}

type advanceKey struct {
	Type     config.KeyType `json:"type"`
	Index    int            `json:"index"`
	SigAlgo  string         `json:"signatureAlgorithm"`
	HashAlgo string         `json:"hashAlgorithm"`
	// hex key type
	PrivateKey string `json:"privateKey,omitempty"`
	// bip44 key type
	Mnemonic       string `json:"mnemonic,omitempty"`
	DerivationPath string `json:"derivationPath,omitempty"`
	// kms key type
	ResourceID string `json:"resourceID,omitempty"`
	// old key format
	Context map[string]string `json:"context,omitempty"`
}

// support for pre v0.22 formats
type simpleAccountPre022 struct {
	Address string `json:"address"`
	Keys    string `json:"keys"`
}

// support for pre v0.22 formats
type advanceAccountPre022 struct {
	Address string       `json:"address"`
	Keys    []advanceKey `json:"keys"`
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

	switch format {
	case simpleFormat:
		var simple simpleAccount
		err = json.Unmarshal(b, &simple)
		j.Simple = simple

	case simpleFormatPre022:
		var simpleOld simpleAccountPre022
		err = json.Unmarshal(b, &simpleOld)
		j.Simple = simpleAccount{
			Address: simpleOld.Address,
			Key:     simpleOld.Keys,
		}

	case advancedFormatPre022:
		var advancedOld advanceAccountPre022
		err = json.Unmarshal(b, &advancedOld)

		j.Advanced = advancedAccount{
			Address: advancedOld.Address,
			Key:     advancedOld.Keys[0],
		}
		j.Advanced.Key.PrivateKey = advancedOld.Keys[0].Context["privateKey"]

	case advancedFormat:
		var advanced advancedAccount
		err = json.Unmarshal(b, &advanced)
		j.Advanced = advanced
	}

	return err
}

func (j account) MarshalJSON() ([]byte, error) {
	if j.FromFile != (fromFileAccount{}) {
		return json.Marshal(j.FromFile)
	}

	if j.Simple != (simpleAccount{}) {
		return json.Marshal(j.Simple)
	}

	return json.Marshal(j.Advanced)
}
