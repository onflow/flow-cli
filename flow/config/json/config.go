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

package json

import (
	"encoding/json"
	"github.com/onflow/flow-cli/flow/project/cli/config"
)

type jsonConfig struct {
	Emulators   jsonEmulators   `json:"emulators"`
	Contracts   jsonContracts   `json:"contracts"`
	Networks    jsonNetworks    `json:"networks"`
	Accounts    jsonAccounts    `json:"accounts"`
	Deployments jsonDeployments `json:"deployments"`
}

func (j *jsonConfig) transformToConfig() *config.Config {
	return &config.Config{
		Emulators:   j.Emulators.transformToConfig(),
		Contracts:   j.Contracts.transformToConfig(),
		Networks:    j.Networks.transformToConfig(),
		Accounts:    j.Accounts.transformToConfig(),
		Deployments: j.Deployments.transformToConfig(),
	}
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

func NewParser() *jsonConfig {
	return new(jsonConfig)
}

// Save saves the configuration to the specified path file in JSON format.
func (j *jsonConfig) Serialize(conf *config.Config) ([]byte, error) {
	jsonConf := transformConfigToJSON(conf)

	data, err := json.MarshalIndent(jsonConf, "", "\t")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (j *jsonConfig) Deserialize(raw []byte) (*config.Config, error) {
	var jsonConf jsonConfig
	err := json.Unmarshal(raw, &jsonConf)

// TODO: allow path to change
const DefaultConfigPath = "flow.json"

func (j *jsonConfig) SupportsFormat(extension string) bool {
	return extension == ".json"
}
