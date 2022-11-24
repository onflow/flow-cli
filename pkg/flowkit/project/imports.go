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
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"path"
)

type ImportReplacer interface {
	Replace()
}

// FileImports implements file import replacements functionality for the project contracts with optionally included aliases.
type FileImports struct {
	contracts []*Contract
	aliases   flowkit.Aliases
}

func NewFileImports(contracts []*Contract, aliases flowkit.Aliases) *FileImports {
	return &FileImports{
		contracts: contracts,
		aliases:   aliases,
	}
}

func (f *FileImports) Replace(program *flowkit.Program) (*flowkit.Program, error) {
	imports := program.Imports()
	sourceTarget := f.getSourceTarget()

	for _, imp := range imports {
		target, found := sourceTarget[absolutePath(program.Location(), imp)]
		if !found {
			return nil, fmt.Errorf("import %s could not be resolved from the configuration", imp)
		}
		program.ReplaceImport(imp, target)
	}

	return program, nil
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
