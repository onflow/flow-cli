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
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/onflow/flow-cli/flow/project/cli/config"
)

type jsonConfig struct {
	Emulators jsonEmulators `json:"emulators"`
	Contracts jsonContracts `json:"contracts"`
	Networks  jsonNetworks  `json:"networks"`
	Accounts  jsonAccounts  `json:"accounts"`
	Deploys   jsonDeploys   `json:"deploys"`
}

func (j jsonConfig) transformToConfig() *config.Config {
	return &config.Config{
		Emulators: j.Emulators.transformToConfig(),
		Contracts: j.Contracts.transformToConfig(),
		Networks:  j.Networks.transformToConfig(),
		Accounts:  j.Accounts.transformToConfig(),
		Deploys:   j.Deploys.transformToConfig(),
	}
}

func transformConfigToJSON(config *config.Config) jsonConfig {
	return jsonConfig{
		Emulators: transformEmulatorsToJSON(config.Emulators),
		Contracts: transformContractsToJSON(config.Contracts),
		Networks:  transformNetworksToJSON(config.Networks),
		Accounts:  transformAccountsToJSON(config.Accounts),
		Deploys:   transformDeploysToJSON(config.Deploys),
	}
}

// Save saves the configuration to the specified path file in JSON format.
func Save(conf *config.Config, path string) error {
	jsonConf := transformConfigToJSON(conf)

	data, err := json.MarshalIndent(jsonConf, "", "\t")
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

func Load(path string) (*config.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDoesNotExist
		}

		return nil, err
	}

	d := json.NewDecoder(f)
	conf := new(jsonConfig)
	err = d.Decode(conf)

	if err != nil {
		fmt.Printf("%s contains invalid json: %s\n", path, err.Error())
		os.Exit(1)
	}

	return conf.transformToConfig(), nil
}

// Exists checks if file exists on the specified path
func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
