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

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

// jsonConfig implements JSON format for persisting and parsing configuration.
type jsonConfig struct {
	Emulators   jsonEmulators   `json:"emulators"`
	Contracts   jsonContracts   `json:"contracts"`
	Networks    jsonNetworks    `json:"networks"`
	Accounts    jsonAccounts    `json:"accounts"`
	Deployments jsonDeployments `json:"deployments"`
}

func (j *jsonConfig) transformToConfig() (*config.Config, error) {
	emulators, err := j.Emulators.transformToConfig()
	if err != nil {
		return nil, err
	}

	contracts, err := j.Contracts.transformToConfig()
	if err != nil {
		return nil, err
	}

	networks, err := j.Networks.transformToConfig()
	if err != nil {
		return nil, err
	}

	accounts, err := j.Accounts.transformToConfig()
	if err != nil {
		return nil, err
	}

	deployments, err := j.Deployments.transformToConfig()
	if err != nil {
		return nil, err
	}

	conf := &config.Config{
		Emulators:   emulators,
		Contracts:   contracts,
		Networks:    networks,
		Accounts:    accounts,
		Deployments: deployments,
	}

	return conf, nil
}

func transformConfigToJSON(config *config.Config) jsonConfig {
	return jsonConfig{
		Emulators:   transformEmulatorsToJSON(config.Emulators),
		Contracts:   transformContractsToJSON(config.Contracts),
		Networks:    transformNetworksToJSON(config.Networks),
		Accounts:    transformAccountsToJSON(config.Accounts),
		Deployments: transformDeploymentsToJSON(config.Deployments),
	}
}

type oldFormat struct {
	Host     interface{} `json:"host"`
	Accounts interface{} `json:"accounts"`
}

func oldConfigFormat(raw []byte) bool {
	var conf oldFormat

	err := json.Unmarshal(raw, &conf)
	if err != nil { // ignore errors in this case
		return false
	}

	if conf.Host != nil {
		return true
	}

	return false
}

// Parser for JSON configuration format.
type Parser struct{}

// NewParser returns a JSON parser.
func NewParser() *Parser {
	return &Parser{}
}

// Serialize configuration to raw.
func (p *Parser) Serialize(conf *config.Config) ([]byte, error) {
	jsonConf := transformConfigToJSON(conf)

	data, err := json.MarshalIndent(jsonConf, "", "\t")
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Deserialize configuration to config structure.
func (p *Parser) Deserialize(raw []byte) (*config.Config, error) {
	// check if old format of config and return an error
	if oldConfigFormat(raw) {
		return nil, config.ErrOutdatedFormat
	}

	var jsonConf jsonConfig
	err := json.Unmarshal(raw, &jsonConf)
	if err != nil {
		return nil, fmt.Errorf("configuration syntax error: %w", err)
	}

	return jsonConf.transformToConfig()
}

// SupportsFormat check if the file format is supported.
func (p *Parser) SupportsFormat(extension string) bool {
	return extension == ".json"
}
