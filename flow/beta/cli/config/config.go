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
	"github.com/thoas/go-funk"
)

type KeyType string

const (
	KeyTypeHex       KeyType = "hex"        // Hex private key with in memory signer
	KeyTypeGoogleKMS KeyType = "google-kms" // Google KMS signer
	KeyTypeShell     KeyType = "shell"      // Exec out to a shell script
)

// Config main configuration structure
type Config struct {
	Emulator  map[string]EmulatorConfigProfile `json:"emulator"`
	Networks  NetworkCollection                `json:"networks"`
	Aliases   map[string]map[string]string     `json:"aliases"`
	Contracts ContractCollection               `json:"contracts"`
	Accounts  AccountCollection                `json:"accounts"`
	Deploy    DeployCollection                 `json:"deploy"`
}

// EmulatorConfigProfile is emulator config
type EmulatorConfigProfile struct {
	Port       int                `json:"port"`
	ServiceKey EmulatorServiceKey `json:"serviceKey"`
}

// EmulatorServiceKey is the service key for emulator
type EmulatorServiceKey struct {
	PrivateKey string
	SigAlgo    crypto.SignatureAlgorithm
	HashAlgo   crypto.HashAlgorithm
}

// emulatorServiceKeyJSON internal structure for parsing
type emulatorServiceKeyJSON struct {
	PrivateKey string `json:"privateKey"`
	SigAlgo    string `json:"signatureAlgorithm"`
	HashAlgo   string `json:"hashAlgorithm"`
}

// Network collection of networks
type NetworkCollection struct {
	Networks []Network
}

// Network config sets host and chain id
type Network struct {
	Host    string       `json:"host"`
	ChainID flow.ChainID `json:"chain"`
}

// Deploy structure for contract
type Deploy struct {
	Network   string
	Account   string
	Contracts []string
}

type DeployCollection struct {
	Deploys []Deploy
}

// Contract is config for contract
type Contract struct {
	Name    string
	Source  string
	Network string
}

// ContractCollection contains contracts with names
type ContractCollection struct {
	Contracts []Contract
}

// AccountCollection contains accounts with names
type AccountCollection struct {
	Accounts map[string]Account
}

// Account is main config for each account
type Account struct {
	Name    string
	Address flow.Address `json:"address"`
	ChainID flow.ChainID `json:"chain"`
	Keys    []AccountKey `json:"keys"`
}

// AccountKey is config for account key
type AccountKey struct {
	Type     KeyType
	Index    int
	SigAlgo  crypto.SignatureAlgorithm
	HashAlgo crypto.HashAlgorithm
	Context  map[string]string
}

// accountKeyJSON is internal struct for parsing key json
type accountKeyJSON struct {
	Type     KeyType           `json:"type"`
	Index    int               `json:"index"`
	SigAlgo  string            `json:"signatureAlgorithm"`
	HashAlgo string            `json:"hashAlgorithm"`
	Context  map[string]string `json:"context"`
}

// UnmarshalJSON EmulatorServiceKey is parer for emulator service key
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

// MarshalJSON EmuatorServiceKey is encoding service key to json
func (k EmulatorServiceKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(emulatorServiceKeyJSON{
		PrivateKey: k.PrivateKey,
		SigAlgo:    k.SigAlgo.String(),
		HashAlgo:   k.HashAlgo.String(),
	})
}

//UnmarshalJSON collection to our structure
func (d *DeployCollection) UnmarshalJSON(b []byte) error {
	raw := make(map[string]map[string]json.RawMessage)
	d.Deploys = make([]Deploy, 0)

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// go over each network
	for network, v := range raw {
		// for each network go through all accounts
		for account, c := range v {
			deploy := new(Deploy)
			deploy.Network = network
			deploy.Account = account

			// try to parse contracts as array of strings
			contracts := []string{}
			err := json.Unmarshal(c, &contracts)
			if err == nil { // simple format
				deploy.Contracts = contracts
			} else { // advanced fromat
				//TODO: implement format with contract init values
			}

			d.Deploys = append(d.Deploys, *deploy)
		}
	}

	return nil
}

//UnmarshalJSON contracts so we can convert it to our structure
func (c *ContractCollection) UnmarshalJSON(b []byte) error {
	raw := make(map[string]json.RawMessage)
	sourceNetwork := make(map[string]string)
	c.Contracts = make([]Contract, 0)

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	for name, value := range raw {
		err := json.Unmarshal(value, &sourceNetwork)
		// advanced schema
		if err == nil && len(sourceNetwork) > 0 {
			for network, source := range sourceNetwork {
				contract := new(Contract)
				contract.Name = name
				contract.Network = network
				contract.Source = source
				c.Contracts = append(c.Contracts, *contract)
			}
		} else { // basic schema
			contract := new(Contract)
			contract.Name = name
			json.Unmarshal(value, &contract.Source)
			c.Contracts = append(c.Contracts, *contract)
		}
	}

	return nil
}

// UnmarshalJSON account collection to get the key name for the account as well
func (c *AccountCollection) UnmarshalJSON(b []byte) error {
	c.Accounts = make(map[string]Account)
	raw := make(map[string]json.RawMessage)

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	for name, value := range raw {
		account := new(Account)
		err := json.Unmarshal(value, &account)
		if err != nil {
			return err
		}

		account.Name = name

		c.Accounts[name] = *account
	}

	return nil
}

// UnmarshalJSON Account decodes json config for account
// and has two options for keys - string and key object
func (a *Account) UnmarshalJSON(b []byte) error {
	raw := make(map[string]json.RawMessage)

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	json.Unmarshal(raw["chain"], &a.ChainID)

	var address string
	json.Unmarshal(raw["address"], &address)

	// TODO: address validation format
	if address == "service" {
		if a.ChainID == "" {
			a.ChainID = flow.Emulator //TODO: find better way for defaults in general
		}

		a.Address = flow.ServiceAddress(a.ChainID)
	} else {
		address = strings.ReplaceAll(address, "0x", "") // remove 0x if present
		a.Address = flow.HexToAddress(address)
	}

	// advanced key format
	err = json.Unmarshal(raw["keys"], &a.Keys)
	// basic key format
	if err != nil {
		var keysString string
		json.Unmarshal(raw["keys"], &keysString)

		var keys []AccountKey
		// default values REF: maybe refactor to variables
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
	} else {
		return err
	}

	return nil
}

// UnmarshalJSON AccountKey decodes json object
// to defined types for algo, hash, index etc
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

// MarshalJSON AccountKey convert to json format
func (a AccountKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(accountKeyJSON{
		SigAlgo:  a.SigAlgo.String(),
		HashAlgo: a.HashAlgo.String(),
		Type:     a.Type,
		Index:    a.Index,
		Context:  a.Context,
	})
}

// Save configuration to a path file in json format
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

// ErrDoesNotExist is error to be returned when config file does not exists
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

// Exists checks if file exists on the specified path
func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

/** ====================================
Config structure helpers
*/
//TODO: better handle error case out of index

// getForNetwork get accounts and contracts for network
func (c *ContractCollection) GetForNetwork(network string) []Contract {
	return funk.Filter(c.Contracts, func(c Contract) bool {
		return c.Network == network
	}).([]Contract)
}

// getByNameAndNetwork get contract array for account and network
func (c *ContractCollection) GetByNameAndNetwork(
	name string,
	network string,
) Contract {
	contracts := funk.Filter(c.Contracts, func(c Contract) bool {
		return c.Network == network && c.Name == name
	}).([]Contract)

	// if we don't find contract by name and network return default for name
	if len(contracts) == 0 {
		return c.GetByName(name)
	}

	return contracts[0]
}

// GetByName get contract from collection by name
func (c *ContractCollection) GetByName(name string) Contract {
	return funk.Filter(c.Contracts, func(c Contract) bool {
		return c.Name == name
	}).([]Contract)[0]
}

// GetByNetwork returns all contracts for specific network
func (c *ContractCollection) GetByNetwork(network string) []Contract {
	return funk.Filter(c.Contracts, func(c Contract) bool {
		return c.Network == network || c.Network == "" // if network is not defined return for all set value
	}).([]Contract)
}

// GetAccountByName get account from account collection by name
func (c *AccountCollection) GetByName(name string) Account {
	return c.Accounts[name]
}

// GetByNetwork get deploys needded for specific network
func (d *DeployCollection) GetByNetwork(network string) []Deploy {
	return funk.Filter(d.Deploys, func(d Deploy) bool {
		return d.Network == network
	}).([]Deploy)
}

func (d *DeployCollection) GetByAccountAndNetwork(
	account string,
	network string,
) []Deploy {
	return funk.Filter(d.Deploys, func(d Deploy) bool {
		return d.Account == account && d.Network == network
	}).([]Deploy)
}
