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
)

type Config struct {
	Emulators   Emulators
	Contracts   Contracts
	Networks    Networks
	Accounts    Accounts
	Deployments Deployments
}

type KeyType string

const (
	KeyTypeHex                        KeyType = "hex"
	KeyTypeGoogleKMS                  KeyType = "google-kms"
	DefaultEmulatorConfigName                 = "default"
	DefaultEmulatorServiceAccountName         = "emulator-account"
)

func (c *Config) Validate() error {
	for _, con := range c.Contracts {
		if c.Networks.GetByName(con.Network) == nil {
			return fmt.Errorf("contract %s contains nonexisting network %s", con.Name, con.Network)
		}
	}

	for _, em := range c.Emulators {
		if c.Accounts.GetByName(em.ServiceAccount) == nil {
			return fmt.Errorf("emulator %s contains nonexisting service account %s", em.Name, em.ServiceAccount)
		}
	}

	for _, d := range c.Deployments {
		if c.Networks.GetByName(d.Network) == nil {
			return fmt.Errorf("deployment contains nonexisting network %s", d.Network)
		}

		if c.Accounts.GetByName(d.Account) == nil {
			return fmt.Errorf("deployment contains nonexisting account %s", d.Account)
		}
	}

	return nil
}

// DefaultConfig gets default configuration
func DefaultConfig() *Config {
	return &Config{
		Emulators: DefaultEmulators(),
		Networks:  DefaultNetworks(),
	}
}

var ErrOutdatedFormat = errors.New("you are using old configuration format")

const DefaultPath = "flow.json"

// GlobalPath gets global path based on home dir
func GlobalPath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s/%s", dirname, DefaultPath)
}

// DefaultPaths determines default paths for configuration
func DefaultPaths() []string {
	return []string{
		GlobalPath(),
		DefaultPath,
	}
}
