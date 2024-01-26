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
)

// Config contains all the configuration for CLI and implements getters and setters for properties.
// Config is agnostic to format from which it is built and it doesn't provide persistence functionality.
//
// Emulators contains all the emulator config
// Contracts contains all contracts definitions and their sources
// Networks defines all the Flow networks addresses
// Accounts defines Flow accounts and their addresses, private key and more properties
// Deployments describes which contracts should be deployed to which accounts
type Config struct {
	Emulators    Emulators
	Contracts    Contracts
	Dependencies Dependencies
	Networks     Networks
	Accounts     Accounts
	Deployments  Deployments
}

type KeyType string

const (
	KeyTypeHex       KeyType = "hex"
	KeyTypeGoogleKMS KeyType = "google-kms"
	KeyTypeBip44     KeyType = "bip44"
	KeyTypeFile      KeyType = "file"
)

// Validate the configuration values.
func (c *Config) Validate() error {
	for _, con := range c.Contracts {
		for _, alias := range con.Aliases {
			_, err := c.Networks.ByName(alias.Network)
			if alias.Network != "" && err != nil {
				return fmt.Errorf("contract %s alias contains nonexisting network %s", con.Name, alias.Network)
			}
		}
	}

	for _, em := range c.Emulators {
		if _, err := c.Accounts.ByName(em.ServiceAccount); err != nil {
			return fmt.Errorf("emulator %s contains nonexisting service account %s", em.Name, em.ServiceAccount)
		}
	}

	for _, d := range c.Deployments {
		if _, err := c.Networks.ByName(d.Network); err != nil {
			return fmt.Errorf("deployment contains nonexisting network %s", d.Network)
		}

		for _, con := range d.Contracts {
			if _, err := c.Contracts.ByName(con.Name); err != nil {
				return fmt.Errorf("deployment contains nonexisting contract %s", con.Name)
			}
		}

		if _, err := c.Accounts.ByName(d.Account); err != nil {
			return fmt.Errorf("deployment contains nonexisting account %s", d.Account)
		}
	}

	return nil
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Emulators: DefaultEmulators,
		Networks:  DefaultNetworks,
	}
}

var ErrOutdatedFormat = errors.New("you are using old configuration format")

const DefaultPath = "flow.json"

func IsDefaultPath(paths []string) bool {
	return len(paths) == 2 && paths[0] == GlobalPath() && paths[1] == DefaultPath
}

// GlobalPath gets global path based on home dir.
func GlobalPath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s/%s", dirname, DefaultPath)
}

// DefaultPaths determines default paths for configuration.
func DefaultPaths() []string {
	return []string{
		GlobalPath(),
		DefaultPath,
	}
}
