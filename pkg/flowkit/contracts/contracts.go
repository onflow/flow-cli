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
	"path"
	"strings"

	"github.com/onflow/cadence"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/flow-go-sdk"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

// Contract contains all the values a contract needs for deployment to the network.
//
// All the contract dependencies are defined here and later used when deploying on the network to
// define the order of deployments. We also define the account to which the contract needs to be deployed,
// and arguments used to deploy. Code contains replaced import statements with concrete addresses.
type Contract struct {
	index          int64
	name           string
	source         string
	accountAddress flow.Address
	accountName    string
	code           string
	args           []cadence.Value
	program        *ast.Program
	dependencies   map[string]*Contract
	aliases        map[string]flow.Address
}

func newContract(
	index int,
	contractName,
	contractSource,
	contractCode string,
	accountAddress flow.Address,
	accountName string,
	args []cadence.Value,
) (*Contract, error) {
	program, err := parser.ParseProgram([]byte(contractCode), nil)
	if err != nil {
		return nil, err
	}

	return &Contract{
		index:          int64(index),
		name:           contractName,
		source:         contractSource,
		accountAddress: accountAddress,
		accountName:    accountName,
		code:           contractCode,
		program:        program,
		args:           args,
		dependencies:   make(map[string]*Contract),
		aliases:        make(map[string]flow.Address),
	}, nil
}

func (c *Contract) ID() int64 {
	return c.index
}

func (c *Contract) Name() string {
	return c.name
}

func (c *Contract) Source() string {
	return c.source
}

func (c *Contract) Code() string {
	return c.code
}

func (c *Contract) Args() []cadence.Value {
	return c.args
}

func (c *Contract) TranspiledCode() string {
	code := c.code

	for source, dep := range c.dependencies {
		code = strings.Replace(
			code,
			fmt.Sprintf(`"%s"`, source),
			fmt.Sprintf("0x%s", dep.Target()),
			1,
		)
	}

	for location, target := range c.aliases {
		code = strings.Replace(
			code,
			fmt.Sprintf(`"%s"`, location),
			fmt.Sprintf("0x%s", target),
			1,
		)
	}

	return code
}
func (c *Contract) AccountName() string {
	return c.accountName
}
func (c *Contract) Target() flow.Address {
	return c.accountAddress
}

func (c *Contract) Dependencies() map[string]*Contract {
	return c.dependencies
}

func (c *Contract) imports() []string {
	imports := make([]string, 0)

	for _, imp := range c.program.ImportDeclarations() {
		location, ok := imp.Location.(common.StringLocation)
		if ok {
			imports = append(imports, location.String())
		}
	}

	return imports
}

func (c *Contract) addDependency(source string, dep *Contract) {
	c.dependencies[source] = dep
}

func (c *Contract) addAlias(location string, target flow.Address) {
	c.aliases[location] = target
}

func absolutePath(basePath, relativePath string) string {
	return path.Join(path.Dir(basePath), relativePath)
}

// Deployments is a collection of contracts to deploy.
//
// Containing functionality to build a dependency tree between contracts and sort them based on that.
type Deployments struct {
	contracts         []*Contract
	loader            Loader
	aliases           map[string]string
	contractsBySource map[string]*Contract
}

func New(loader Loader, aliases map[string]string) *Deployments {
	return &Deployments{
		loader:            loader,
		aliases:           aliases,
		contractsBySource: make(map[string]*Contract),
	}
}

func (c *Deployments) Contracts() []*Contract {
	return c.contracts
}

// Sort contracts by deployment order.
//
// Order of sorting is dependent on the possible imports contracts contains, since
// any imported contract must be deployed before deploying the contract with that import.
func (c *Deployments) Sort() error {
	err := c.resolveImports()
	if err != nil {
		return err
	}

	sorted, err := sortByDeploymentOrder(c.contracts)
	if err != nil {
		return err
	}

	c.contracts = sorted
	return nil
}

func (c *Deployments) Add(
	name,
	source string,
	accountAddress flow.Address,
	accountName string,
	args []cadence.Value,
) error {
	contractCode, err := c.loader.Load(source)
	if err != nil {
		return err
	}

	contract, err := newContract(
		len(c.contracts),
		name,
		source,
		string(contractCode),
		accountAddress,
		accountName,
		args,
	)
	if err != nil {
		return err
	}

	c.contracts = append(c.contracts, contract)
	c.contractsBySource[contract.source] = contract

	return nil
}

// resolveImports checks every contract import and builds a dependency tree.
func (c *Deployments) resolveImports() error {
	for _, contract := range c.contracts {
		for _, source := range contract.imports() {
			importPath := c.loader.Normalize(contract.source, source)

			importAlias, isAlias := c.aliases[importPath]
			importContract, isContract := c.contractsBySource[importPath]

			if isContract {
				contract.addDependency(source, importContract)
			} else if isAlias {
				contract.addAlias(source, flow.HexToAddress(importAlias))
			} else {
				return fmt.Errorf("import from %s could not be found: %s, make sure import path is correct.", contract.name, importPath)
			}
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
func sortByDeploymentOrder(contracts []*Contract) ([]*Contract, error) {
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

func nodeSetsToContractSets(nodes [][]graph.Node) [][]*Contract {
	contracts := make([][]*Contract, len(nodes))

	for i, s := range nodes {
		contracts[i] = nodesToContracts(s)
	}

	return contracts
}

func nodesToContracts(nodes []graph.Node) []*Contract {
	contracts := make([]*Contract, len(nodes))

	for i, s := range nodes {
		contracts[i] = s.(*Contract)
	}

	return contracts
}

// CyclicImportError is returned when contract contain cyclic imports one to the
// other which is not possible to be resolved and deployed.
type CyclicImportError struct {
	Cycles [][]*Contract
}

func (e *CyclicImportError) contractNames() [][]string {
	cycles := make([][]string, 0, len(e.Cycles))

	for _, cycle := range e.Cycles {
		contracts := make([]string, 0, len(cycle))
		for _, contract := range cycle {
			contracts = append(contracts, contract.Name())
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
