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

package project

import (
	"fmt"
	"path"

	"github.com/onflow/flow-go-sdk"
)

type Account interface {
	Name() string
	Address() flow.Address
}

// ImportReplacer implements file import replacements functionality for the project contracts with optionally included aliases.
type ImportReplacer struct {
	contracts []*Contract
	aliases   Aliases
	accounts  []Account
}

func NewImportReplacer(contracts []*Contract, aliases Aliases, accounts []Account) *ImportReplacer {
	return &ImportReplacer{
		contracts: contracts,
		aliases:   aliases,
		accounts:  accounts,
	}
}

func (i *ImportReplacer) Replace(program *Program) (*Program, error) {
	imports := program.imports()
	contractsLocations := i.getContractsLocations()
	accountsLocations := i.getAccountsLocations()

	for _, imp := range imports {
		importLocation := path.Clean(absolutePath(program.Location(), imp))
		address, isPath := contractsLocations[importLocation]
		if isPath {
			program.replaceImport(imp, address)
			continue
		}
		// replace identifier imports
		address, isIdentifier := contractsLocations[imp]
		if isIdentifier {
			program.replaceImport(imp, address)
			continue
		}
		// replace account name imports
		address, isName := accountsLocations[imp]
		if isName {
			program.replaceImport(imp, address)
			continue
		}

		return nil, fmt.Errorf("import %s could not be resolved from provided contracts", imp)
	}

	return program, nil
}

// getContractsLocations return a map with contract locations as keys and addresses where they are deployed as values.
func (i *ImportReplacer) getContractsLocations() map[string]string {
	locationAddress := make(map[string]string)
	for _, contract := range i.contracts {
		locationAddress[path.Clean(contract.Location())] = contract.AccountAddress.String()
		// add also by name since we might use the new import schema
		locationAddress[contract.Name] = contract.AccountAddress.String()
	}

	for source, target := range i.aliases {
		locationAddress[path.Clean(source)] = flow.HexToAddress(target).String()
	}

	return locationAddress
}

func (i *ImportReplacer) getAccountsLocations() map[string]string {
	accountAddress := make(map[string]string)
	for _, account := range i.accounts {
		accountAddress[account.Name()] = account.Address().String()
	}
	return accountAddress
}

func absolutePath(basePath, relativePath string) string {
	return path.Join(path.Dir(basePath), relativePath)
}
