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

package resolver

import (
	"fmt"
	"path"
	"strings"

	"github.com/onflow/cadence"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/flow-go-sdk"
)

// Program contains all the values the code needs for sending to the network.
//
// Program can represent a contract, script, or transaction code.
// All the code dependencies are defined here and later used when sending to network to
// define the order of execution. We also define the account to which the code needs to be executed on,
// and arguments used to send. Code contains replaced import statements with concrete addresses.
type Program struct {
	index          int64
	location       string
	name           string
	accountAddress flow.Address
	accountName    string
	code           string
	args           []cadence.Value
	program        *ast.Program
	dependencies   map[string]*Program
	aliases        map[string]flow.Address
}

func newProgram(
	index int,
	location,
	code string,
	accountAddress flow.Address,
	accountName string,
	args []cadence.Value,
) (*Program, error) {
	program, err := parser.ParseProgram([]byte(code), nil)
	if err != nil {
		return nil, err
	}

	return &Program{
		index:          int64(index),
		location:       location,
		name:           parseName(program),
		accountAddress: accountAddress,
		accountName:    accountName,
		code:           code,
		program:        program,
		args:           args,
		dependencies:   make(map[string]*Program),
		aliases:        make(map[string]flow.Address),
	}, nil
}

func (c *Program) ID() int64 {
	return c.index
}

func (c *Program) Name() string {
	return c.name
}

func (c *Program) Location() string {
	return c.location
}

// Code gets the original code without import replacements.
func (c *Program) Code() string {
	return c.code
}

func (c *Program) Args() []cadence.Value {
	return c.args
}

// ReplacedImports in the code with the corresponding network addresses.
func (c *Program) ReplacedImports() string {
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
func (c *Program) AccountName() string {
	return c.accountName
}
func (c *Program) Target() flow.Address {
	return c.accountAddress
}

func (c *Program) Dependencies() map[string]*Program {
	return c.dependencies
}

func (c *Program) HasImports() bool {
	return len(c.imports()) > 0
}

func (c *Program) isContract() bool {
	return len(c.program.CompositeDeclarations())+len(c.program.CompositeDeclarations()) == 1
}

func (c *Program) imports() []string {
	imports := make([]string, 0)

	for _, imp := range c.program.ImportDeclarations() {
		location, ok := imp.Location.(common.StringLocation)
		if ok {
			imports = append(imports, location.String())
		}
	}

	return imports
}

func (c *Program) addDependency(location string, dep *Program) {
	c.dependencies[location] = dep
}

func (c *Program) addAlias(location string, target flow.Address) {
	c.aliases[location] = target
}

func parseName(program *ast.Program) string {
	for _, compositeDeclaration := range program.CompositeDeclarations() {
		if compositeDeclaration.CompositeKind == common.CompositeKindContract {
			return compositeDeclaration.Identifier.Identifier
		}
	}

	for _, interfaceDeclaration := range program.InterfaceDeclarations() {
		if interfaceDeclaration.CompositeKind == common.CompositeKindContract {
			return interfaceDeclaration.Identifier.Identifier
		}
	}

	return ""
}

func absolutePath(basePath, relativePath string) string {
	return path.Join(path.Dir(basePath), relativePath)
}
