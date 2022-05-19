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
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-go-sdk"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

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
	program, err := parser2.ParseProgram(contractCode)
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

func (c *Contract) Code() string {
	return c.code
}

func (c *Contract) Args() []cadence.Value {
	return c.args
}

func (c *Contract) TranspiledCode() string {
	code := c.code

	for location, dep := range c.dependencies {
		code = strings.Replace(
			code,
			fmt.Sprintf(`"%s"`, location),
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

func (c *Contract) addDependency(location string, dep *Contract) {
	c.dependencies[location] = dep
}

func (c *Contract) addAlias(location string, target flow.Address) {
	c.aliases[location] = target
}

func absolutePath(basePath, relativePath string) string {
	return path.Join(path.Dir(basePath), relativePath)
}

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
