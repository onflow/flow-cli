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
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
)

func ParseSourceString(s string) (network, address, contractName string, err error) {
	parts := strings.Split(s, "://")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid dependency source format: %s", s)
	}
	network = parts[0]

	subParts := strings.Split(parts[1], ".")
	if len(subParts) != 2 {
		return "", "", "", fmt.Errorf("invalid dependency source format: %s", s)
	}
	address = subParts[0]
	contractName = subParts[1]

	return network, address, contractName, nil
}

type Source struct {
	NetworkName  string
	Address      flow.Address
	ContractName string
}

type Dependency struct {
	Name    string
	Source  Source
	Version string
	Aliases Aliases
}

type Dependencies []Dependency

func (d *Dependencies) ByName(name string) *Dependency {
	for i, dep := range *d {
		if dep.Name == name {
			return &(*d)[i]
		}
	}

	return nil
}

func (d *Dependencies) AddOrUpdate(dep Dependency) {
	for i, dependency := range *d {
		if dependency.Name == dep.Name {
			(*d)[i] = dep
			return
		}
	}

	*d = append(*d, dep)
}
