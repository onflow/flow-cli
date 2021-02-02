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
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type KeyType string

const (
	serviceAccountName = "service"

	KeyTypeHex       KeyType = "hex"        // Hex private key with in memory signer
	KeyTypeGoogleKMS KeyType = "google-kms" // Google KMS signer
	KeyTypeShell     KeyType = "shell"      // Exec out to a shell script

	defaultKeyType = KeyTypeHex
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
	Address flow.Address
	ChainID flow.ChainID
	Keys    []AccountKey
}

type accountJSON struct {
	Address string       `json:"address"`
	Chain   string       `json:"chain"`
	Keys    []AccountKey `json:"keys"`
}

func (a *Account) UnmarshalJSON(b []byte) error {
	var s accountJSON

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	chainID, err := stringToChainID(s.Chain)
	if err != nil {
		return err
	}

	if s.Address == serviceAccountName {
		a.Address = flow.ServiceAddress(chainID)
	} else {
		a.Address = flow.HexToAddress(s.Address)
	}

	a.ChainID = chainID
	a.Keys = s.Keys

	return nil
}

func (a Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(accountJSON{
		Address: a.Address.Hex(),
		Chain:   chainIDToString(a.ChainID),
		Keys:    a.Keys,
	})
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

func stringToChainID(s string) (flow.ChainID, error) {
	switch strings.TrimSpace(s) {
	case "emulator":
		return flow.Emulator, nil
	case "testnet":
		return flow.Testnet, nil
	case "mainnet":
		return flow.Mainnet, nil
	case "":
		return "", errors.New("chain cannot be empty")
	default:
		return "", fmt.Errorf(`invalid chain: "%s"`, s)
	}
}

func chainIDToString(chainID flow.ChainID) string {
	switch chainID {
	case flow.Emulator:
		return "emulator"
	case flow.Testnet:
		return "testnet"
	case flow.Mainnet:
		return "mainnet"
	default:
		return ""
	}
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
