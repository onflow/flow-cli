/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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
	"io/ioutil"
	"path"
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-go-sdk"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Contract struct {
	index        int64
	name         string
	source       string
	target       flow.Address
	code         string
	program      *ast.Program
	dependencies map[string]*Contract
	aliases      map[string]flow.Address
}

func newContract(
	index int,
	contractName,
	contractSource,
	contractCode string,
	target flow.Address,
) (*Contract, error) {
	program, err := parser2.ParseProgram(contractCode)
	if err != nil {
		return nil, err
	}

	return &Contract{
		index:        int64(index),
		name:         contractName,
		source:       contractSource,
		target:       target,
		code:         contractCode,
		program:      program,
		dependencies: make(map[string]*Contract),
		aliases:      make(map[string]flow.Address),
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

func (c *Contract) Target() flow.Address {
	return c.target
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

type Loader interface {
	Load(source string) (string, error)
	Normalize(base, relative string) string
}

type FilesystemLoader struct{}

func (f FilesystemLoader) Load(source string) (string, error) {
	codeBytes, err := ioutil.ReadFile(source)
	if err != nil {
		return "", err
	}

	return string(codeBytes), nil
}

func (f FilesystemLoader) Normalize(base, relative string) string {
	return absolutePath(base, relative)
}

func absolutePath(basePath, relativePath string) string {
	return path.Join(path.Dir(basePath), relativePath)
}

func (c *Contract) addAlias(location string, target flow.Address) {
	c.aliases[location] = target
}

type Preprocessor struct {
	loader            Loader
	aliases           map[string]string
	contracts         []*Contract
	contractsBySource map[string]*Contract
}

func NewPreprocessor(loader Loader, aliases map[string]string) *Preprocessor {
	return &Preprocessor{
		loader:            loader,
		aliases:           aliases,
		contracts:         make([]*Contract, 0),
		contractsBySource: make(map[string]*Contract),
	}
}

func (p *Preprocessor) AddContractSource(
	contractName,
	contractSource string,
	target flow.Address,
) error {
	contractCode, err := p.loader.Load(contractSource)
	if err != nil {
		return err
	}

	c, err := newContract(
		len(p.contracts),
		contractName,
		contractSource,
		contractCode,
		target,
	)
	if err != nil {
		return err
	}

	p.contracts = append(p.contracts, c)
	p.contractsBySource[c.source] = c

	return nil
}

func (p *Preprocessor) ResolveImports() error {
	for _, c := range p.contracts {
		for _, location := range c.imports() {
			importPath := p.loader.Normalize(c.source, location)
			importAlias, isAlias := p.aliases[importPath]
			importContract, isContract := p.contractsBySource[importPath]

			if isContract {
				c.addDependency(location, importContract)
			} else if isAlias {
				c.addAlias(location, flow.HexToAddress(
					strings.ReplaceAll(importAlias, "0x", ""), // REF: go-sdk should handle this
				))
			} else {
				return fmt.Errorf("Import from %s could not be found: %s, make sure import path is correct.", c.name, importPath)
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
		// TODO: add dedicated error types
		return nil, err
	}

	return sorted, nil
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
