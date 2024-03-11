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

package super

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flowkit/v2"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/cobra"
)

type generateFlagsDef struct {
	Directory string `default:"" flag:"dir" info:"Directory to generate files in"`
}

var generateFlags = generateFlagsDef{}

var GenerateCommand = &cobra.Command{
	Use:     "generate",
	Short:   "Generate template files for common Cadence code",
	GroupID: "super",
	Aliases: []string{"g"},
}

var GenerateContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "contract <name>",
		Short:   "Generate Cadence smart contract template",
		Example: "flow generate contract HelloWorld",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  generateContract,
}

var GenerateTransactionCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "transaction <name>",
		Short:   "Generate a Cadence transaction template",
		Example: "flow generate transaction SomeTransaction",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  generateTransaction,
}

var GenerateScriptCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "script <name>",
		Short:   "Generate a Cadence script template",
		Example: "flow generate script SomeScript",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  generateScript,
}

func init() {
	GenerateContractCommand.AddToParent(GenerateCommand)
	GenerateTransactionCommand.AddToParent(GenerateCommand)
	GenerateScriptCommand.AddToParent(GenerateCommand)
}

func generateContract(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	return generateNew(args, "contract", logger, state)
}

func generateTransaction(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	return generateNew(args, "transaction", logger, state)
}

func generateScript(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	return generateNew(args, "script", logger, state)
}

func generateNew(
	args []string,
	templateType string,
	logger output.Logger,
	state *flowkit.State,
) (result command.Result, err error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}

	name := args[0]
	var filename string

	// Don't add .cdc extension if it's already there
	if strings.HasSuffix(name, ".cdc") {
		filename = name
	} else {
		filename = fmt.Sprintf("%s.cdc", name)
	}

	var fileToWrite string
	var basePath string

	if generateFlags.Directory != "" {
		basePath = generateFlags.Directory
	} else {
		switch templateType {
		case "contract":
			basePath = "cadence/contracts"
		case "script":
			basePath = "cadence/scripts"
		case "transaction":
			basePath = "cadence/transactions"
		default:
			return nil, fmt.Errorf("invalid template type: %s", templateType)
		}
	}

	switch templateType {
	case "contract":
		fileToWrite = fmt.Sprintf(`
access(all)
contract %s {
    init() {}
}`, name)
	case "script":
		fileToWrite = `access(all)
fun main() {
    // Script details here
}`
	case "transaction":
		fileToWrite = `transaction() {
    prepare(account: &Account) {}

    execute {}
}`
	default:
		return nil, fmt.Errorf("invalid template type: %s", templateType)
	}

	filenameWithBasePath := filepath.Join(basePath, filename)

	// Check file existence
	if _, err := state.ReaderWriter().ReadFile(filenameWithBasePath); err == nil {
		return nil, fmt.Errorf("file already exists: %s", filenameWithBasePath)
	}

	// Ensure the directory exists
	if err := state.ReaderWriter().MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("error creating directories: %w", err)
	}

	err = state.ReaderWriter().WriteFile(filenameWithBasePath, []byte(fileToWrite), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing file: %w", err)
	}

	logger.Info(fmt.Sprintf("Generated new %s: %s at %s", templateType, name, filenameWithBasePath))

	if templateType == "contract" {
		state.Contracts().AddOrUpdate(config.Contract{Name: name, Location: filenameWithBasePath})
		err = state.SaveDefault()
		if err != nil {
			return nil, fmt.Errorf("error saving to flow.json: %w", err)
		}
	}

	return nil, err
}
