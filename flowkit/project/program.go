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
	"regexp"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
)

type Program struct {
	code            []byte
	args            []cadence.Value
	location        string
	astProgram      *ast.Program
	developmentCode []byte
}

func NewProgram(code []byte, args []cadence.Value, location string) (*Program, error) {
	astProgram, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return nil, err
	}

	return &Program{
		code:            code,
		args:            args,
		location:        location,
		astProgram:      astProgram,
		developmentCode: code, // has converted import syntax e.g. 'import "Foo"'
	}, nil
}

// Imports builds an array of all the import locations
// It currently supports getting import locations as identifiers or as strings. Strings locations
// can represent a file or an account name, whereas identifiers represent contract names.
func (p *Program) Imports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range p.astProgram.ImportDeclarations() {
		// we parse all string locations, that are all imports that look like "import X from "Y"" or "import "X""
		_, isStringImport := importDeclaration.Location.(common.StringLocation)
		if isStringImport {
			imports = append(imports, importDeclaration.Location.String())
		}
	}

	return imports
}

func (p *Program) HasImports() bool {
	return len(p.Imports()) > 0
}

func (p *Program) replaceImport(from string, to string) *Program {
	code := string(p.Code())

	pathRegex := regexp.MustCompile(fmt.Sprintf(`import\s+(\w+)\s+from\s+"%s"`, from))
	identifierRegex := regexp.MustCompile(fmt.Sprintf(`import\s+"(%s)"`, from))

	replacement := fmt.Sprintf(`import $1 from 0x%s`, to)
	code = pathRegex.ReplaceAllString(code, replacement)
	code = identifierRegex.ReplaceAllString(code, replacement)

	p.code = []byte(code)
	p.reload()
	return p
}

func (p *Program) Location() string {
	return p.location
}

func (p *Program) Code() []byte {
	return p.code
}

func (p *Program) DevelopmentCode() []byte {
	return p.developmentCode
}

func (p *Program) Name() (string, error) {
	if len(p.astProgram.CompositeDeclarations()) > 1 || len(p.astProgram.InterfaceDeclarations()) > 1 ||
		len(p.astProgram.CompositeDeclarations())+len(p.astProgram.InterfaceDeclarations()) > 1 {
		return "", fmt.Errorf("the code must declare exactly one contract or contract interface")
	}

	for _, compositeDeclaration := range p.astProgram.CompositeDeclarations() {
		if compositeDeclaration.CompositeKind == common.CompositeKindContract {
			return compositeDeclaration.Identifier.Identifier, nil
		}
	}

	for _, interfaceDeclaration := range p.astProgram.InterfaceDeclarations() {
		if interfaceDeclaration.CompositeKind == common.CompositeKindContract {
			return interfaceDeclaration.Identifier.Identifier, nil
		}
	}

	return "", fmt.Errorf("unable to determine contract name")
}

func (p *Program) ConvertImports() {
	code := string(p.code)
	addressImportRegex := regexp.MustCompile(`import\s+(\w+)\s+from\s+0x[0-9a-fA-F]+`)
	modifiedCode := addressImportRegex.ReplaceAllString(code, `import "$1"`)

	p.developmentCode = []byte(modifiedCode)
}

func (p *Program) reload() {
	astProgram, err := parser.ParseProgram(nil, p.code, parser.Config{})
	if err != nil {
		return
	}

	p.astProgram = astProgram
	p.ConvertImports()
}
