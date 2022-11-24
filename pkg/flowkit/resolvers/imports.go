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

package resolvers

import (
	"fmt"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"path"
	"strings"
)

type ImportReplacer interface {
	Replace()
}

// FileImports implements file import replacements functionality for the project contracts with optionally included aliases.
type FileImports struct {
	contracts []*flowkit.Contract
	aliases   flowkit.Aliases
}

func NewFileImports(contracts []*flowkit.Contract, aliases flowkit.Aliases) *FileImports {
	return &FileImports{
		contracts: contracts,
		aliases:   aliases,
	}
}

// getFileImports returns all cadence file imports from Cadence code as an array.
func (f *FileImports) getFileImports(program *ast.Program) []string {
	imports := make([]string, 0)

	for _, importDeclaration := range program.ImportDeclarations() {
		_, isFileImport := importDeclaration.Location.(common.StringLocation)

		if isFileImport {
			imports = append(imports, importDeclaration.Location.String())
		}
	}

	return imports
}

func (f *FileImports) Replace(program *flowkit.Program, codePath string) (*flowkit.Program, error) {
	imports := program.Imports()
	sourceTarget := f.getSourceTarget()

	for _, imp := range imports {
		target, found := sourceTarget[absolutePath(codePath, imp)]
		if !found {
			return nil, fmt.Errorf("import %s could not be resolved from the configuration", imp)
		}
		program.ReplaceImport(imp, target)
	}

	return program, nil
}

// replaceImport replaces import from path to address.
func (f *FileImports) replaceImport(code []byte, from string, to string) []byte {
	return []byte(strings.Replace(
		string(code),
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	))
}

// getSourceTarget return a map with contract paths as keys and addresses as values.
func (f *FileImports) getSourceTarget() map[string]string {
	sourceTarget := make(map[string]string)
	for _, contract := range f.contracts {
		sourceTarget[path.Clean(contract.Location)] = contract.AccountAddress.String()
	}

	for source, target := range f.aliases {
		sourceTarget[path.Clean(source)] = flow.HexToAddress(target).String()
	}

	return sourceTarget
}

func absolutePath(basePath, relativePath string) string {
	return path.Join(path.Dir(basePath), relativePath)
}
