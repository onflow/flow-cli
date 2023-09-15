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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/flowkit/config"
)

type jsonAccounts map[string]account

const (
	defaultHashAlgo = crypto.SHA3_256
	defaultSigAlgo  = crypto.ECDSA_P256
)

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
	key := config.AccountKey{
		Type:     config.KeyTypeHex,
		SigAlgo:  defaultSigAlgo,
		HashAlgo: defaultHashAlgo,
	}

	replaced, original, err := tryReplaceEnv(a.Key)
	if err != nil {
		return nil, err
	}
	if replaced != "" {
		key.Env = original
		a.Key = replaced
	}

	pkey, err := crypto.DecodePrivateKeyHex(
		config.DefaultSigAlgo,
		strings.TrimPrefix(a.Key, "0x"),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for account: %s", accountName)
	}
	key.PrivateKey = pkey

	replacedAddress, _, err := tryReplaceEnv(a.Address)
	if err != nil {
		return nil, err
	}
	if replacedAddress != "" {
		a.Address = replacedAddress
	}

	address, err := transformAddress(a.Address)
	if err != nil {
		return nil, err
	}

	return &config.Account{
		Name:    accountName,
		Address: address,
		Key:     key,
	}, nil
}

// tryReplaceEnv checks if value matches env regex, if it does it check whether the value was set in env,
// if not set then it errors, otherwise it replaces the value with set env variable, and also returns the original key.
func tryReplaceEnv(value string) (replaced string, original string, err error) {
	envRegex, err := regexp.Compile(`^\$\{(\w+)\}|\$(\w+)$`)
	if err != nil {
		return "", "", err
	}

	if !envRegex.MatchString(value) {
		return
	}

	found := envRegex.FindAllStringSubmatch(value, -1)
	if len(found) == 0 {
		return // should not happen
	}

	envVar := found[0][1]
	if found[0][2] != "" { // second regex
		envVar = found[0][2]
	}
	if os.Getenv(envVar) == "" {
		return "", "", fmt.Errorf("required environment variable %s not set", envVar)
	}

	original = value
	replaced = os.ExpandEnv(value)

	return
}

// transformAdvancedToConfig transforms advanced internal account to config account.
func transformAdvancedToConfig(accountName string, a advancedAccount) (*config.Account, error) {
	sigAlgo := config.DefaultSigAlgo // default to ecdsa as default
	if a.Key.SigAlgo != "" {
		sigAlgo = crypto.StringToSignatureAlgorithm(a.Key.SigAlgo)
	}

	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm for account %s", accountName)
	}

	hashAlgo := config.DefaultHashAlgo // default to sha3 as default
	if a.Key.HashAlgo != "" {
		hashAlgo = crypto.StringToHashAlgorithm(a.Key.HashAlgo)
	}

	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("invalid hash algorithm for account %s", accountName)
	}

	validTypes := []config.KeyType{config.KeyTypeHex, config.KeyTypeFile, config.KeyTypeBip44, config.KeyTypeGoogleKMS}
	if !slices.Contains(validTypes, a.Key.Type) {
		return nil, fmt.Errorf("invalid key type for account %s", accountName)
	}

	// check that only one is provided because the values are mutually exclusive
	set := false
	for _, v := range []string{a.Key.ResourceID, a.Key.PrivateKey, a.Key.Location} {
		if v == "" {
			continue
		}
		if set {
			return nil, fmt.Errorf("can only provide one property (resource ID, private key, location) on account %s", accountName)
		}
		set = true
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

		replaced, original, err := tryReplaceEnv(a.Key.PrivateKey)
		if err != nil {
			return nil, err
		}
		if replaced != "" {
			key.Env = original
			a.Key.PrivateKey = replaced
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

	case config.KeyTypeFile:
		if a.Key.Location == "" {
			return nil, fmt.Errorf("missing location to a file containing the private key value for the account %s", accountName)
		}
		key.Location = filepath.FromSlash(a.Key.Location)
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
		if a.Key.IsDefault() {
			jsonAccounts[a.Name] = transformSimpleAccountToJSON(a)
		} else {
			jsonAccounts[a.Name] = transformAdvancedAccountToJSON(a)
		}
	}

	return jsonAccounts
}

func transformSimpleAccountToJSON(a config.Account) account {
	key := strings.TrimPrefix(a.Key.PrivateKey.String(), "0x")
	if a.Key.Env != "" {
		key = a.Key.Env // if we used env vars then use it when saving
	}

	return account{
		Simple: simpleAccount{
			Address: a.Address.String(),
			Key:     key,
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
		Type: key.Type,
	}

	if key.Index != 0 { // only set if non-default
		advancedKey.Index = key.Index
	}

	if key.SigAlgo != config.DefaultSigAlgo { // only set if non-default
		advancedKey.SigAlgo = key.SigAlgo.String()
	}

	if key.HashAlgo != config.DefaultHashAlgo { // only set if non-default
		advancedKey.HashAlgo = key.HashAlgo.String()
	}

	switch key.Type {
	case config.KeyTypeHex:
		advancedKey.PrivateKey = strings.TrimPrefix(key.PrivateKey.String(), "0x")
		if key.Env != "" {
			advancedKey.PrivateKey = key.Env // if we used env vars then use it when saving
		}
	case config.KeyTypeBip44:
		advancedKey.Mnemonic = key.Mnemonic
		advancedKey.DerivationPath = key.DerivationPath
	case config.KeyTypeGoogleKMS:
		advancedKey.ResourceID = key.ResourceID
	case config.KeyTypeFile:
		advancedKey.Location = filepath.ToSlash(key.Location)
	}

	return advancedKey
}

type account struct {
	Simple   simpleAccount
	Advanced advancedAccount
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
	Index    int            `json:"index,omitempty"`
	SigAlgo  string         `json:"signatureAlgorithm,omitempty"`
	HashAlgo string         `json:"hashAlgorithm,omitempty"`
	// hex key type
	PrivateKey string `json:"privateKey,omitempty"`
	// bip44 key type
	Mnemonic       string `json:"mnemonic,omitempty"`
	DerivationPath string `json:"derivationPath,omitempty"`
	// kms key type
	ResourceID string `json:"resourceID,omitempty"`
	// key location
	Location string `json:"location,omitempty"`
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

type formatType int

const (
	simpleFormat         formatType = 0
	advancedFormat       formatType = 1
	simpleFormatPre022   formatType = 2 // pre v.022 format
	advancedFormatPre022 formatType = 3 // pre v.022 format
)

func decideFormat(b []byte) (formatType, error) {
	var raw map[string]any
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
	if j.Simple != (simpleAccount{}) {
		return json.Marshal(j.Simple)
	}

	return json.Marshal(j.Advanced)
}

func (a account) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Ref: "#/$defs/simpleAccount",
			},
			{
				Ref: "#/$defs/advancedAccount",
			},
			{
				Ref: "#/$defs/simpleAccountPre022",
			},
			{
				Ref: "#/$defs/advanceAccountPre022",
			},
		},
		Definitions: map[string]*jsonschema.Schema{
			"simpleAccount": jsonschema.Reflect(simpleAccount{}),
			"advancedAccount": jsonschema.Reflect(advancedAccount{}),
			"simpleAccountPre022": jsonschema.Reflect(simpleAccountPre022{}),
			"advanceAccountPre022": jsonschema.Reflect(advanceAccountPre022{}),
		},
	}
}
