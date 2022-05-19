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

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrDoesNotExist is error to be returned when config file does not exists.
var ErrDoesNotExist = errors.New("missing configuration")

// Exists checks if file exists on the specified path.
func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Parser is interface for any configuration format parser to implement.
type Parser interface {
	Serialize(*Config) ([]byte, error)
	Deserialize([]byte) (*Config, error)
	SupportsFormat(string) bool
}

type ReaderWriter interface {
	ReadFile(source string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// Parsers is a list of all configuration parsers.
type Parsers []Parser

// FindForFormat finds a parser that can parse a specific format based on extension.
func (c *Parsers) FindForFormat(extension string) Parser {
	for _, parser := range *c {
		if parser.SupportsFormat(extension) {
			return parser
		}
	}

	return nil
}

// Loader contains actions for composing and modifying configuration.
type Loader struct {
	readerWriter     ReaderWriter
	configParsers    Parsers
	composedFromFile map[string]string
}

// NewLoader returns a new loader.
func NewLoader(readerWriter ReaderWriter) *Loader {
	return &Loader{
		readerWriter:     readerWriter,
		composedFromFile: map[string]string{},
	}
}

// AddConfigParser adds a new configuration parser.
func (l *Loader) AddConfigParser(format Parser) {
	l.configParsers = append(l.configParsers, format)
}

// Save saves a configuration to a path with correct serializer.
func (l *Loader) Save(conf *Config, path string) error {
	configFormat := l.configParsers.FindForFormat(
		filepath.Ext(path),
	)

	if configFormat == nil {
		return fmt.Errorf("parser not found for format")
	}

	data, err := configFormat.Serialize(conf)
	if err != nil {
		return err
	}

	err = l.readerWriter.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (l *Loader) loadConfig(confPath string) (*Config, error) {
	raw, err := l.loadFile(confPath)

	if err != nil {
		return nil, err
	}

	preProcessed := l.preprocess(raw)
	configParser := l.configParsers.FindForFormat(filepath.Ext(confPath))
	if configParser == nil {
		return nil, fmt.Errorf("parser not found for config: %s", confPath)
	}

	return configParser.Deserialize(preProcessed)
}

// Load loads configuration from one or more file paths.
//
// If more than one path is specified, their contents are merged
// together into on configuration object.
func (l *Loader) Load(paths []string) (*Config, error) {
	// special case for default configs
	// try to load local config and only if not found try to load global config
	if IsDefaultPath(paths) {
		conf, err := l.loadConfig(DefaultPath)
		if err == nil { // if we could load it then process it
			return l.postprocess(conf)
		}
		if !errors.Is(err, ErrDoesNotExist) {
			return nil, err
		}

		conf, err = l.loadConfig(GlobalPath())
		if err != nil {
			return nil, ErrDoesNotExist
		} else {
			return l.postprocess(conf)
		}
	}

	var baseConf *Config
	for _, confPath := range paths {
		conf, err := l.loadConfig(confPath)
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

	// if no config was loaded - neither local nor global return an error.
	if baseConf == nil {
		return nil, ErrDoesNotExist
	}

	return l.postprocess(baseConf)
}

// preprocess does all manipulations to the raw configuration format happens here.
func (l *Loader) preprocess(raw []byte) []byte {
	raw, accountsFromFile := ProcessorRun(raw)

	// add all imports from files preprocessor detected for later processing
	l.composedFromFile = accountsFromFile

	return raw
}

// postprocess does all stateful changes to configuration structures here after it is parsed.
func (l *Loader) postprocess(baseConf *Config) (*Config, error) {
	for name, path := range l.composedFromFile {
		raw, err := l.loadFile(path)
		if err != nil {
			return nil, err
		}

		configParser := l.configParsers.FindForFormat(filepath.Ext(path))
		if configParser == nil {
			return nil, fmt.Errorf("parser not found for config: %s", path)
		}

		conf, err := configParser.Deserialize(raw)
		if err != nil {
			return nil, err
		}

		account, err := conf.Accounts.ByName(name)
		if err != nil {
			return nil, err
		}

		// create an empty config with single account so we don't include all accounts in file
		accountConf := &Config{
			Accounts: []Account{*account},
		}

		l.composeConfig(baseConf, accountConf)
	}

	// validate as part of post processing
	err := baseConf.Validate()
	if err != nil {
		return nil, err
	}

	return baseConf, nil
}

// composeConfig merges multiple configuration files from right to left.
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

// loadFile simple file loader.
func (l *Loader) loadFile(path string) ([]byte, error) {
	raw, err := l.readerWriter.ReadFile(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDoesNotExist
		}

		return nil, err
	}

	return raw, nil
}
