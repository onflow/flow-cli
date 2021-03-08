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

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// ErrDoesNotExist is error to be returned when config file does not exists
var ErrDoesNotExist = errors.New("project config file does not exist")

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
	Serialize(*Config) ([]byte, error)
	Deserialize([]byte) (*Config, error)
	SupportsFormat(string) bool
}

// ConfigParsers lsit of all configuration parsers
type ConfigParsers []Parser

// FindForFormat finds a parser that can parse a specific format based on extension
func (c *ConfigParsers) FindForFormat(extension string) Parser {
	for _, parser := range *c {
		if parser.SupportsFormat(extension) {
			return parser
		}
	}

	return nil
}

// Loader contains actions for composing and modification of configuration
type Loader struct {
	af               *afero.Afero
	configParsers    ConfigParsers
	composedFromFile map[string]string
}

// NewLoader creates a new loader
func NewLoader(filesystem afero.Fs) *Loader {
	af := &afero.Afero{Fs: filesystem}
	return &Loader{
		af:               af,
		composedFromFile: map[string]string{},
	}
}

// AddConfigParser adds new parssers for configuration
func (l *Loader) AddConfigParser(format Parser) {
	l.configParsers = append(l.configParsers, format)
}

// Save save configuration to a path with correct serializer
func (l *Loader) Save(conf *Config, path string) error {
	configFormat := l.configParsers.FindForFormat(
		filepath.Ext(path),
	)

	data, err := configFormat.Serialize(conf)
	if err != nil {
		return err
	}

	err = l.af.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Load and compose multiple configurations
func (l *Loader) Load(paths []string) (*Config, error) {
	var baseConf *Config

	for _, path := range paths {
		raw, err := l.loadFile(path)
		if err != nil {
			return nil, err
		}

		preProcessed := l.preprocess(raw)
		configParser := l.configParsers.FindForFormat(filepath.Ext(path))
		if configParser == nil {
			return nil, fmt.Errorf("Parser not found for config: %s", path)
		}

		conf, err := configParser.Deserialize(preProcessed)
		if err != nil {
			return nil, err
		}

		// if first conf just assign as baseConf
		if baseConf == nil {
			baseConf = conf
			continue
		}

		l.composeConfig(baseConf, conf)
	}

	baseConf, err := l.postprocess(baseConf)
	if err != nil {
		return nil, err
	}

	return baseConf, nil
}

// preprocess configuration - all manipulations to the raw configuration format happens here
func (l *Loader) preprocess(raw []byte) []byte {
	raw, accountsFromFile := ProcessorRun(raw)

	// add all imports from files preprocessor detected for later processing
	l.composedFromFile = accountsFromFile

	return raw
}

// postprocess configuration - do all stateful changes to configuration structures here after it is parsed
func (l *Loader) postprocess(baseConf *Config) (*Config, error) {
	for name, path := range l.composedFromFile {
		raw, err := l.loadFile(path)
		if err != nil {
			return nil, err
		}

		configParser := l.configParsers.FindForFormat(filepath.Ext(path))
		if configParser == nil {
			return nil, fmt.Errorf("Parser not found for config: %s", path)
		}

		conf, err := configParser.Deserialize(raw)
		if err != nil {
			return nil, err
		}

		// create an empty config with single account so we don't include all accounts in file
		accountConf := &Config{
			Accounts: []Account{*conf.Accounts.GetByName(name)},
		}

		l.composeConfig(baseConf, accountConf)
	}

	return baseConf, nil
}

// composeConfig - here we merge multiple configuration files from right to left
func (l *Loader) composeConfig(baseConf *Config, conf *Config) {
	// if not first overwrite first with this one
	for _, account := range conf.Accounts {
		baseConf.Accounts.AddOrUpdate(account.Name, account)
	}
	for _, network := range conf.Networks {
		baseConf.Networks.AddOrUpdate(network.Name, network)
	}
	for _, contract := range conf.Contracts {
		baseConf.Contracts.AddOrUpdate(contract.Name, contract)
	}
	for _, deployment := range conf.Deployments {
		baseConf.Deployments.AddOrUpdate(deployment)
	}
}

// simple file loader
func (l *Loader) loadFile(path string) ([]byte, error) {
	raw, err := l.af.ReadFile(path)

	// TODO: better handle
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
			//return nil, ErrDoesNotExist
		}

		return nil, err
	}

	return raw, nil
}
