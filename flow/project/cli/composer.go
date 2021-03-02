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

package cli

import (
	"errors"
	"fmt"
	"github.com/onflow/flow-cli/flow/project/cli/config"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// ErrDoesNotExist is error to be returned when config file does not exists
var ErrDoesNotExist = errors.New("project config file does not exist")

const DefaultConfigPath = "flow.json"

// Exists checks if file exists on the specified path
func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Parser is interface for any configuration format parser to implement
type Parser interface {
	Serialize(*config.Config) ([]byte, error)
	Deserialize([]byte) (*config.Config, error)
	SupportsFormat(string) bool
}

type ConfigParsers []Parser

func (c *ConfigParsers) FindForFormat(extension string) Parser {
	for _, parser := range *c {
		if parser.SupportsFormat(extension) {
			return parser
		}
	}

	return nil
}

type Composer struct {
	af            *afero.Afero
	configParsers ConfigParsers
}

func NewComposer(filesystem afero.Fs) *Composer {
	af := &afero.Afero{Fs: filesystem}
	return &Composer{af: af}
}

func (c *Composer) AddConfigParser(format Parser) {
	c.configParsers = append(c.configParsers, format)
}

// Load configuration
func (c *Composer) loadSingle(path string) (*config.Config, error) {
	raw, err := c.loadFile(path)

	if err != nil {
		return nil, err
	}

	preprocessor := config.NewPreprocessor(c.af.Fs)
	preProcessed := preprocessor.Run(raw)

	configParser := c.configParsers.FindForFormat(
		filepath.Ext(path),
	)

	if configParser == nil {
		return nil, errors.New(fmt.Sprintf("Parser not found for config: %s", path))
	}

	return configParser.Deserialize(preProcessed)
}

func (c *Composer) Save(conf *config.Config, path string) error {
	configFormat := c.configParsers.FindForFormat(
		filepath.Ext(path),
	)

	data, err := configFormat.Serialize(conf)
	if err != nil {
		return err
	}

	err = c.af.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Load and compose multiple configurations
func (c *Composer) Load(paths []string) (*config.Config, error) {
	var baseConf *config.Config

	for _, path := range paths {
		conf, err := c.loadSingle(path)
		if err != nil {
			return nil, err
		}

		// if first baseConf just save it
		if baseConf == nil {
			baseConf = conf
			continue
		}

		// if not first overwrite first with this one
		for _, account := range conf.Accounts {
			baseConf.Accounts.SetForName(account.Name, account)
		}
		for _, network := range conf.Networks {
			baseConf.Networks.SetForName(network.Name, network)
		}
		for _, contract := range conf.Contracts {
			baseConf.Contracts.SetForName(contract.Name, contract)
		}
		for _, deployment := range conf.Deployments {
			baseConf.Deployments.AddIfMissing(deployment)
		}
	}

	return baseConf, nil
}

func (c *Composer) loadFile(path string) ([]byte, error) {
	raw, err := c.af.ReadFile(path)

	// TODO: better handle
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDoesNotExist
		}

		return nil, err
	}

	return raw, nil
}
