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
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Contract struct {
	index        int64
	bundleName   string
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
	bundleName,
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
		bundleName:   bundleName,
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

type SourceResolver func(source string) (string, error)

func FilesystemResolver(source string) (string, error) {
	codeBytes, err := ioutil.ReadFile(source)
	if err != nil {
		return "", err
	}

	return string(codeBytes), nil
}

type Preprocessor struct {
	aliases   map[string]string
	contracts map[string]*Contract
	resolver  SourceResolver
}

func NewPreprocessor(resolver SourceResolver) *Preprocessor {
	return &Preprocessor{
		contracts: make(map[string]*Contract),
		resolver:  resolver,
	}
}

func (p *Preprocessor) AddContractSource(
	bundleName,
	contractName,
	contractSource string,
	target flow.Address,
) error {
	contractCode, err := p.resolver(contractSource)
	if err != nil {
		return err
	}

	c, err := newContract(
		len(p.contracts),
		bundleName,
		contractName,
		contractSource,
		contractCode,
		target,
	)
	if err != nil {
		return err
	}

	p.contracts[c.source] = c

	return nil
}

func (p *Preprocessor) PrepareForDeployment() ([]*Contract, error) {

	for _, c := range p.contracts {

		for _, location := range c.imports() {
			importPath := absolutePath(c.source, location)

			importContract, isContract := p.contracts[importPath]
			if isContract {
				c.addDependency(location, importContract)
			}
		}
	}

	sorted, err := sortByDeploymentOrder(p.contracts)
	if err != nil {
		return nil, err
	}

	return sorted, nil
}

func sortByDeploymentOrder(contracts map[string]*Contract) ([]*Contract, error) {
	g := simple.NewDirectedGraph()

	for _, c := range contracts {
		g.AddNode(c)
	}

	for _, c := range contracts {
		for _, dep := range c.dependencies {
			g.SetEdge(g.NewEdge(dep, c))
		}
	}

	sorted, err := topo.Sort(g)
	if err != nil {
		return nil, err
	}

	results := make([]*Contract, len(sorted))

	for i, s := range sorted {
		results[i] = s.(*Contract)
	}

	return results, nil
}

func absolutePath(source, location string) string {
	return path.Join(path.Dir(source), location)
}
