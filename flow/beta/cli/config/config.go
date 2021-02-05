/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type KeyType string

const (
	KeyTypeHex       KeyType = "hex"        // Hex private key with in memory signer
	KeyTypeGoogleKMS KeyType = "google-kms" // Google KMS signer
	KeyTypeShell     KeyType = "shell"      // Exec out to a shell script
)

type Config struct {
	Emulator        map[string]EmulatorConfigProfile `json:"emulator"`
	Networks        map[string]Network               `json:"networks"`
	Aliases         map[string]map[string]string     `json:"aliases"`
	ContractBundles map[string]ContractBundle        `json:"contracts"`
	Accounts        map[string]Account               `json:"accounts"`
}

type EmulatorConfigProfile struct {
	Port       int                `json:"port"`
	ServiceKey EmulatorServiceKey `json:"serviceKey"`
}

type EmulatorServiceKey struct {
	PrivateKey string
	SigAlgo    crypto.SignatureAlgorithm
	HashAlgo   crypto.HashAlgorithm
}

type emulatorServiceKeyJSON struct {
	PrivateKey string `json:"privateKey"`
	SigAlgo    string `json:"signatureAlgorithm"`
	HashAlgo   string `json:"hashAlgorithm"`
}

func (k *EmulatorServiceKey) UnmarshalJSON(b []byte) error {
	var s emulatorServiceKeyJSON

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	k.PrivateKey = s.PrivateKey
	k.SigAlgo = crypto.StringToSignatureAlgorithm(s.SigAlgo)
	k.HashAlgo = crypto.StringToHashAlgorithm(s.HashAlgo)

	return nil
}

func (k EmulatorServiceKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(emulatorServiceKeyJSON{
		PrivateKey: k.PrivateKey,
		SigAlgo:    k.SigAlgo.String(),
		HashAlgo:   k.HashAlgo.String(),
	})
}

type Network struct {
	Host    string       `json:"host"`
	ChainID flow.ChainID `json:"chain"`
}

type ContractBundle struct {
	Source map[string]string `json:"source"`
	Target map[string]string `json:"target"`
}

type Account struct {
	Address string       `json:"address"`
	ChainID flow.ChainID `json:"chain"`
	Keys    []AccountKey `json:"keys"`
}

type AccountKey struct {
	Type     KeyType
	Index    int
	SigAlgo  crypto.SignatureAlgorithm
	HashAlgo crypto.HashAlgorithm
	Context  map[string]string
}

type accountKeyJSON struct {
	Type     KeyType           `json:"type"`
	Index    int               `json:"index"`
	SigAlgo  string            `json:"signatureAlgorithm"`
	HashAlgo string            `json:"hashAlgorithm"`
	Context  map[string]string `json:"context"`
}

func (a *Account) UnmarshalJSON(b []byte) error {

	raw := make(map[string]json.RawMessage)
	json.Unmarshal(b, &raw)

	json.Unmarshal(raw["address"], &a.Address)
	json.Unmarshal(raw["chain"], &a.ChainID)
	err := json.Unmarshal(raw["keys"], &a.Keys)

	// if error trying unmarshal into key structure then we try unmarshal a string
	if err != nil {
		var keysString string
		json.Unmarshal(raw["keys"], &keysString)

		var keys []AccountKey
		json.Unmarshal([]byte(`[{
			"type": "hex",
			"index": 0,
			"signatureAlgorithm": "ECDSA_P256",
			"hashAlgorithm": "SHA3_256",
			"context": {
				"privateKey": "`+keysString+`"
			}
		}]`), &keys)
		a.Keys = keys
	}

	return nil
}

func (a *AccountKey) UnmarshalJSON(b []byte) error {
	var s accountKeyJSON

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	a.SigAlgo = crypto.StringToSignatureAlgorithm(s.SigAlgo)
	a.HashAlgo = crypto.StringToHashAlgorithm(s.HashAlgo)

	a.Type = s.Type
	a.Index = s.Index
	a.Context = s.Context

	return nil
}

func (a AccountKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(accountKeyJSON{
		SigAlgo:  a.SigAlgo.String(),
		HashAlgo: a.HashAlgo.String(),
		Type:     a.Type,
		Index:    a.Index,
		Context:  a.Context,
	})
}

func Save(conf *Config, path string) error {
	data, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0777)
	if err != nil {
		return err
	}

	return nil
}

var ErrDoesNotExist = errors.New("project config file does not exist")

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDoesNotExist
		}

		return nil, err
	}

	d := json.NewDecoder(f)

	conf := new(Config)

	if err := d.Decode(conf); err != nil {
		fmt.Printf("%s contains invalid json: %s\n", path, err.Error())
		os.Exit(1)
	}

	return conf, nil
}

func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
