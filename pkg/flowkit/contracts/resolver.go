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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-go-sdk"
)

// Resolver handles resolving imports in Cadence code.
type Resolver struct {
	code    []byte
	program *ast.Program
}

// NewResolver creates a new resolver.
func NewResolver(code []byte) (*Resolver, error) {
	program, err := parser2.ParseProgram(string(code))
	if err != nil {
		return nil, err
	}

	return &Resolver{
		code:    code,
		program: program,
	}, nil
}

// ResolveImports resolves imports in code to addresses.
//
// resolving is done based on code file path and is resolved to
// addresses defined in configuration for contracts or their aliases.
//
func (r *Resolver) ResolveImports(
	codePath string,
	contracts []flowkit.Contract,
	aliases flowkit.Aliases,
) ([]byte, error) {
	imports := r.getFileImports()
	sourceTarget := r.getSourceTarget(contracts, aliases)

	for _, imp := range imports {
		target := sourceTarget[absolutePath(codePath, imp)]
		if target != "" {
			r.code = r.replaceImport(imp, target)
		} else {
			return nil, fmt.Errorf("import %s could not be resolved from the configuration", imp)
		}
	}

	return r.code, nil
}

// replaceImport replaces import from path to address.
func (r *Resolver) replaceImport(from string, to string) []byte {
	return []byte(strings.Replace(
		string(r.code),
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	))
}

// getSourceTarget return a map with contract paths as keys and addresses as values.
func (r *Resolver) getSourceTarget(
	contracts []flowkit.Contract,
	aliases flowkit.Aliases,
) map[string]string {
	sourceTarget := make(map[string]string)
	for _, contract := range contracts {
		sourceTarget[path.Clean(contract.Source)] = contract.AccountAddress.String()
	}

	for source, target := range aliases {
		sourceTarget[path.Clean(source)] = flow.HexToAddress(target).String()
	}

	return sourceTarget
}

// HasFileImports checks if there is a file import statement present in Cadence code.
func (r *Resolver) HasFileImports() bool {
	return len(r.getFileImports()) > 0
}

// getFileImports returns all cadence file imports from Cadence code as an array.
func (r *Resolver) getFileImports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range r.program.ImportDeclarations() {
		_, isFileImport := importDeclaration.Location.(common.StringLocation)

		if isFileImport {
			imports = append(imports, importDeclaration.Location.String())
		}
	}

	return imports
}
