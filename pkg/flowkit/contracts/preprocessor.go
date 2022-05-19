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

package contracts

import (
	"fmt"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-go-sdk"
)

// Preprocessor loads the contract and replaces the imports with addresses.
type Preprocessor struct {
	loader            Loader
	aliases           map[string]string
	contracts         []*Contract
	contractsBySource map[string]*Contract
}

// NewPreprocessor creates a new preprocessor.
func NewPreprocessor(loader Loader, aliases map[string]string) *Preprocessor {
	return &Preprocessor{
		loader:            loader,
		aliases:           aliases,
		contracts:         make([]*Contract, 0),
		contractsBySource: make(map[string]*Contract),
	}
}

// AddContractSource adds a new contract and the target to resolve the imports to.
func (p *Preprocessor) AddContractSource(
	contractName,
	contractSource string,
	accountAddress flow.Address,
	accountName string,
	args []cadence.Value,
) error {
	contractCode, err := p.loader.Load(contractSource)
	if err != nil {
		return err
	}

	c, err := newContract(
		len(p.contracts),
		contractName,
		contractSource,
		string(contractCode),
		accountAddress,
		accountName,
		args,
	)
	if err != nil {
		return err
	}

	p.contracts = append(p.contracts, c)
	p.contractsBySource[c.source] = c

	return nil
}

// ResolveImports for the contracts checking the import path and getting an alias or location of contract.
func (p *Preprocessor) ResolveImports() error {
	for _, c := range p.contracts {
		for _, location := range c.imports() {
			importPath := p.loader.Normalize(c.source, location)
			importAlias, isAlias := p.aliases[importPath]
			importContract, isContract := p.contractsBySource[importPath]

			if isContract {
				c.addDependency(location, importContract)
			} else if isAlias {
				c.addAlias(location, flow.HexToAddress(importAlias))
			} else {
				return fmt.Errorf("import from %s could not be found: %s, make sure import path is correct.", c.name, importPath)
			}
		}
	}

	return nil
}

func (p *Preprocessor) ContractBySource(contractSource string) *Contract {
	return p.contractsBySource[contractSource]
}

func (p *Preprocessor) ContractDeploymentOrder() ([]*Contract, error) {
	sorted, err := sortByDeploymentOrder(p.contracts)
	if err != nil {
		return nil, err
	}

	return sorted, nil
}
