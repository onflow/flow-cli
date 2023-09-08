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
	"path/filepath"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type deployContract struct {
	index int64
	*Contract
	program      *Program
	dependencies map[string]*deployContract
}

func (d *deployContract) ID() int64 {
	return d.index
}

func (d *deployContract) addDependency(location string, dep *deployContract) {
	d.dependencies[location] = dep
}

// Deployment contains logic to sort deployment order of contracts.
//
// Deployment makes sure the contract containing imports is deployed after all importing contracts are deployed.
// This way we can deploy all contracts without missing imports.
// Contracts are iterated and dependency graph is built which is then later sorted
type Deployment struct {
	contracts []*deployContract
	// map of contracts by their location specified in state
	contractsByLocation map[string]*deployContract
	contractsByName     map[string]*deployContract
	aliases             LocationAliases
}

// NewDeployment from the flowkit Contracts and loaded from the contract location using a loader.
func NewDeployment(contracts []*Contract, aliases LocationAliases) (*Deployment, error) {
	deployment := &Deployment{
		contractsByLocation: make(map[string]*deployContract),
		contractsByName:     make(map[string]*deployContract),
		aliases:             aliases,
	}

	for _, contract := range contracts {
		err := deployment.add(contract)
		if err != nil {
			return nil, err
		}
	}

	return deployment, nil
}

func (d *Deployment) add(contract *Contract) error {
	program, err := NewProgram(contract.code, contract.Args, contract.location)
	if err != nil {
		return err
	}

	c := &deployContract{
		index:        int64(len(d.contracts)),
		Contract:     contract,
		program:      program,
		dependencies: make(map[string]*deployContract),
	}

	d.contracts = append(d.contracts, c)
	d.contractsByLocation[filepath.Clean(c.Location())] = c
	d.contractsByName[c.Name] = c

	return nil
}

// Sort contracts by deployment order.
//
// Order of sorting is dependent on the possible imports contract contains, since
// any imported contract must be deployed before deploying the contract with that import.
// Only applicable to contracts.
func (d *Deployment) Sort() ([]*Contract, error) {
	if d.conflictExists() {
		return nil, fmt.Errorf("the same contract cannot be deployed to multiple accounts on the same network")
	}

	err := d.buildDependencies()
	if err != nil {
		return nil, err
	}

	sorted, err := sortByDeploymentOrder(d.contracts)
	if err != nil {
		return nil, err
	}

	contracts := make([]*Contract, len(d.contracts))
	for i, s := range sorted {
		contracts[i] = s.Contract
	}

	return contracts, nil
}

// conflictExists returns true if the same contract is configured to deploy to more than one account for the same network.
func (d *Deployment) conflictExists() bool {
	uniq := make(map[string]bool)
	for _, c := range d.contracts {
		if _, exists := uniq[c.Name]; exists {
			return true
		}
		uniq[c.Name] = true
	}

	return false
}

// buildDependencies iterates over all contracts and checks the imports which are added as its dependencies.
func (d *Deployment) buildDependencies() error {
	for _, contract := range d.contracts {
		for _, location := range contract.program.imports() {
			// find contract by the path import
			importPath := absolutePath(contract.location, location)
			importContract, isPath := d.contractsByLocation[importPath]
			if isPath {
				contract.addDependency(location, importContract)
				continue
			}
			// find contract by identifier import - new schema
			importContract, isIdentifier := d.contractsByName[location]
			if isIdentifier {
				contract.addDependency(location, importContract)
				continue
			}

			// if aliased then skip, not a dependency
			if _, exists := d.aliases[importPath]; exists {
				continue
			}
			if _, exists := d.aliases[location]; exists {
				continue
			}

			return fmt.Errorf(
				"import from %s could not be found: %s, make sure import path is correct, and the contract is added to deployments or has an alias",
				contract.Name,
				location,
			)
		}
	}

	return nil
}

// sortByDeploymentOrder sorts the given set of contracts in order of deployment.
//
// The resulting ordering ensures that each contract is deployed after all of its
// dependencies are deployed. This function returns an error if an import cycle exists.
//
// This function constructs a directed graph in which contracts are nodes and imports are edges.
// The ordering is computed by performing a topological sort on the constructed graph.
func sortByDeploymentOrder(contracts []*deployContract) ([]*deployContract, error) {
	g := simple.NewDirectedGraph()

	for _, c := range contracts {
		g.AddNode(c)
	}

	for _, c := range contracts {
		for _, dep := range c.dependencies {
			g.SetEdge(g.NewEdge(dep, c))
		}
	}

	sorted, err := topo.SortStabilized(g, nil)
	if err != nil {
		switch topoErr := err.(type) {
		case topo.Unorderable:
			return nil, &CyclicImportError{Cycles: nodeSetsToContractSets(topoErr)}
		default:
			return nil, err
		}
	}

	return nodesToContracts(sorted), nil
}

func nodeSetsToContractSets(nodes [][]graph.Node) [][]*deployContract {
	contracts := make([][]*deployContract, len(nodes))

	for i, s := range nodes {
		contracts[i] = nodesToContracts(s)
	}

	return contracts
}

func nodesToContracts(nodes []graph.Node) []*deployContract {
	contracts := make([]*deployContract, len(nodes))

	for i, s := range nodes {
		contracts[i] = s.(*deployContract)
	}

	return contracts
}

// CyclicImportError is returned when contract contain cyclic imports one to the
// other which is not possible to be resolved and deployed.
type CyclicImportError struct {
	Cycles [][]*deployContract
}

func (e *CyclicImportError) contractNames() [][]string {
	cycles := make([][]string, 0, len(e.Cycles))

	for _, cycle := range e.Cycles {
		contracts := make([]string, 0, len(cycle))
		for _, contract := range cycle {
			contracts = append(contracts, contract.Name)
		}

		cycles = append(cycles, contracts)
	}

	return cycles
}

func (e *CyclicImportError) Error() string {
	return fmt.Sprintf(
		"contracts: import cycle(s) detected: %v",
		e.contractNames(),
	)
}
