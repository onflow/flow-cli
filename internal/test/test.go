/*
 * Flow CLI
 *
 * Copyright 2022 Dapper Labs, Inc.
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

package test

import (
	"bytes"
	"fmt"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"path"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

type flagsTests struct{}

var testFlags = flagsTests{}

var Command = &command.Command{
	Cmd: &cobra.Command{
		Use:     "test <filename>",
		Short:   "Run Cadence tests",
		Example: `flow test script.cdc`,
		Args:    cobra.MinimumNArgs(1),
		GroupID: "tools",
	},
	Flags: &testFlags,
	RunS:  run,
}

func run(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	filename := args[0]

	code, err := state.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading script file: %w", err)
	}

	logger.StartProgress("Running tests...")
	defer logger.StopProgress()

	result, err := testCode(code, filename, state)
	if err != nil {
		return nil, err
	}

	return &Result{
		Results: result,
	}, nil
}

func testCode(code []byte, scriptPath string, state *flowkit.State) (cdcTests.Results, error) {
	runner := cdcTests.NewTestRunner().
		WithImportResolver(importResolver(scriptPath, state)).
		WithFileResolver(fileResolver(scriptPath, state))

	return runner.RunTests(string(code))
}

func importResolver(scriptPath string, state *flowkit.State) cdcTests.ImportResolver {
	return func(location common.Location) (string, error) {
		stringLocation, isFileImport := location.(common.StringLocation)
		if !isFileImport {
			return "", fmt.Errorf("cannot import from %s", location)
		}

		importedContract, err := resolveContract(stringLocation, state)
		if err != nil {
			return "", err
		}

		importedContractFilePath := absolutePath(scriptPath, importedContract.Location)

		contractCode, err := state.ReadFile(importedContractFilePath)
		if err != nil {
			return "", err
		}

		return string(contractCode), nil
	}
}

func resolveContract(stringLocation common.StringLocation, state *flowkit.State) (config.Contract, error) {
	relativePath := stringLocation.String()
	for _, contract := range *state.Contracts() {
		if contract.Location == relativePath {
			return contract, nil
		}
	}

	return config.Contract{}, fmt.Errorf("cannot find contract with location '%s' in configuration", relativePath)
}

func fileResolver(scriptPath string, state *flowkit.State) cdcTests.FileResolver {
	return func(path string) (string, error) {
		importFilePath := absolutePath(scriptPath, path)

		content, err := state.ReadFile(importFilePath)
		if err != nil {
			return "", err
		}

		return string(content), nil
	}
}

func absolutePath(basePath, filePath string) string {
	if path.IsAbs(filePath) {
		return filePath
	}

	return path.Join(path.Dir(basePath), filePath)
}

type Result struct {
	cdcTests.Results
}

var _ command.Result = &Result{}

func (r *Result) JSON() any {
	results := make([]map[string]string, 0, len(r.Results))

	for _, result := range r.Results {
		results = append(results, map[string]string{
			"testName": result.TestName,
			"error":    result.Error.Error(),
		})
	}

	return results
}

func (r *Result) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, cdcTests.PrettyPrintResults(r.Results))

	_ = writer.Flush()

	return b.String()
}

func (r *Result) Oneliner() string {
	return cdcTests.PrettyPrintResults(r.Results)
}
